// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ec2pricing

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/mitchellh/go-homedir"
	"github.com/patrickmn/go-cache"
	"go.uber.org/multierr"
)

const (
	SpotCacheFileName = "spot-pricing-cache.gob"
)

type SpotPricing struct {
	Region         string
	FullRefreshTTL time.Duration
	DirectoryPath  string
	cache          *cache.Cache
	ec2Client      ec2.DescribeSpotPriceHistoryAPIClient
	logger         *log.Logger
	sync.RWMutex
}

type spotPricingEntry struct {
	Timestamp time.Time
	SpotPrice float64
	Zone      string
}

func LoadSpotCacheOrNew(ctx context.Context, ec2Client ec2.DescribeSpotPriceHistoryAPIClient, region string, fullRefreshTTL time.Duration, directoryPath string, days int) (*SpotPricing, error) {
	expandedDirPath, err := homedir.Expand(directoryPath)
	if err != nil {
		return nil, fmt.Errorf("unable to load spot pricing cache directory %s: %w", expandedDirPath, err)
	}
	spotPricing := &SpotPricing{
		Region:         region,
		FullRefreshTTL: fullRefreshTTL,
		DirectoryPath:  expandedDirPath,
		ec2Client:      ec2Client,
		cache:          cache.New(fullRefreshTTL, fullRefreshTTL),
		logger:         log.New(io.Discard, "", 0),
	}
	if fullRefreshTTL <= 0 {
		if err := spotPricing.Clear(); err != nil {
			return nil, err
		}
		return spotPricing, nil
	}
	gob.Register([]*spotPricingEntry{})
	// Start the cache refresh job
	go spotPricing.spotCacheRefreshJob(ctx, days)
	spotCache, err := loadSpotCacheFrom(fullRefreshTTL, region, expandedDirPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("a spot pricing cache file could not be loaded: %w", err)
	}
	if err != nil {
		spotCache = cache.New(0, 0)
	}
	spotPricing.cache = spotCache
	return spotPricing, nil
}

func loadSpotCacheFrom(itemTTL time.Duration, region string, expandedDirPath string) (*cache.Cache, error) {
	file, err := os.Open(getSpotCacheFilePath(region, expandedDirPath))
	if err != nil {
		return nil, err
	}
	decoder := gob.NewDecoder(file)
	spotTimeSeries := &map[string]cache.Item{}
	if err := decoder.Decode(spotTimeSeries); err != nil {
		return nil, err
	}
	c := cache.NewFrom(itemTTL, itemTTL, *spotTimeSeries)
	c.DeleteExpired()
	return c, nil
}

func getSpotCacheFilePath(region string, directoryPath string) string {
	return filepath.Join(directoryPath, fmt.Sprintf("%s-%s", region, SpotCacheFileName))
}

func (c *SpotPricing) spotCacheRefreshJob(ctx context.Context, days int) {
	if c.FullRefreshTTL <= 0 {
		return
	}
	refreshTicker := time.NewTicker(c.FullRefreshTTL)
	for range refreshTicker.C {
		if err := c.Refresh(ctx, days); err != nil {
			c.logger.Printf("Periodic Spot Cache Refresh Error: %v", err)
		}
	}
}

func (c *SpotPricing) SetLogger(logger *log.Logger) {
	c.logger = logger
}

func (c *SpotPricing) Refresh(ctx context.Context, days int) error {
	c.Lock()
	defer c.Unlock()
	spotInstanceTypeCosts, err := c.fetchSpotPricingTimeSeries(ctx, "", days)
	if err != nil {
		return fmt.Errorf("there was a problem refreshing the spot instance type pricing cache: %v", err)
	}
	for instanceTypeAndZone, cost := range spotInstanceTypeCosts {
		c.cache.SetDefault(instanceTypeAndZone, cost)
	}
	if err := c.Save(); err != nil {
		return fmt.Errorf("unable to save the refreshed spot instance type pricing cache file: %v", err)
	}
	return nil
}

func (c *SpotPricing) Get(ctx context.Context, instanceType ec2types.InstanceType, zone string, days int) (float64, error) {
	entries, ok := c.cache.Get(string(instanceType))
	if zone != "" && ok {
		if !c.contains(zone, entries.([]*spotPricingEntry)) {
			ok = false
		}
	}
	if !ok {
		c.RLock()
		defer c.RUnlock()
		zonalSpotPricing, err := c.fetchSpotPricingTimeSeries(ctx, instanceType, days)
		if err != nil {
			return -1, fmt.Errorf("there was a problem fetching spot instance type pricing for %s: %v", instanceType, err)
		}
		for instanceType, costs := range zonalSpotPricing {
			c.cache.SetDefault(instanceType, costs)
		}
	}

	entries, ok = c.cache.Get(string(instanceType))
	if !ok {
		return -1, fmt.Errorf("unable to get spot pricing for %s in zone %s for %d days back", instanceType, zone, days)
	}
	return c.calculateSpotAggregate(c.filterOn(zone, entries.([]*spotPricingEntry))), nil
}

func (c *SpotPricing) contains(zone string, entries []*spotPricingEntry) bool {
	for _, entry := range entries {
		if entry.Zone == zone {
			return true
		}
	}
	return false
}

func (c *SpotPricing) calculateSpotAggregate(spotPriceEntries []*spotPricingEntry) float64 {
	if len(spotPriceEntries) == 0 {
		return 0.0
	}
	if len(spotPriceEntries) == 1 {
		return spotPriceEntries[0].SpotPrice
	}
	// Sort slice by timestamp in descending order from the end time (most likely, now)
	sort.Slice(spotPriceEntries, func(i, j int) bool {
		return spotPriceEntries[i].Timestamp.After(spotPriceEntries[j].Timestamp)
	})

	endTime := spotPriceEntries[0].Timestamp
	startTime := spotPriceEntries[len(spotPriceEntries)-1].Timestamp
	totalDuration := endTime.Sub(startTime).Minutes()

	priceSum := float64(0)
	for i, entry := range spotPriceEntries {
		duration := spotPriceEntries[int(math.Max(float64(i-1), 0))].Timestamp.Sub(entry.Timestamp).Minutes()
		priceSum += duration * entry.SpotPrice
	}
	return priceSum / totalDuration
}

func (c *SpotPricing) filterOn(zone string, pricingEntries []*spotPricingEntry) []*spotPricingEntry {
	filtered := []*spotPricingEntry{}
	for _, entry := range pricingEntries {
		// this takes the first zone, might be better to do all zones instead...
		if zone == "" {
			zone = entry.Zone
		}
		if entry.Zone == zone {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// Count of items in the cache.
func (c *SpotPricing) Count() int {
	return c.cache.ItemCount()
}

func (c *SpotPricing) Save() error {
	if c.FullRefreshTTL <= 0 || c.Count() == 0 {
		return nil
	}
	if err := os.Mkdir(c.DirectoryPath, 0o755); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	file, err := os.Create(getSpotCacheFilePath(c.Region, c.DirectoryPath))
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	return encoder.Encode(c.cache.Items())
}

func (c *SpotPricing) Clear() error {
	c.Lock()
	defer c.Unlock()
	c.cache.Flush()
	if err := os.Remove(getSpotCacheFilePath(c.Region, c.DirectoryPath)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// fetchSpotPricingTimeSeries makes a bulk request to the ec2 api to retrieve all spot instance type pricing for the past n days
// If instanceType is empty, it will fetch for all instance types.
func (c *SpotPricing) fetchSpotPricingTimeSeries(ctx context.Context, instanceType ec2types.InstanceType, days int) (map[string][]*spotPricingEntry, error) {
	start := time.Now()
	calls := 0
	defer func() {
		c.logger.Printf("Took %s and %d calls to collect Spot pricing", time.Since(start), calls)
	}()
	spotTimeSeries := map[string][]*spotPricingEntry{}
	endTime := time.Now().UTC()
	startTime := endTime.Add(time.Hour * time.Duration(24*-1*days))
	spotPriceHistInput := ec2.DescribeSpotPriceHistoryInput{
		ProductDescriptions: []string{productDescription},
		StartTime:           &startTime,
		EndTime:             &endTime,
	}
	if instanceType != "" {
		spotPriceHistInput.InstanceTypes = append(spotPriceHistInput.InstanceTypes, instanceType)
	}
	var processingErr error

	p := ec2.NewDescribeSpotPriceHistoryPaginator(c.ec2Client, &spotPriceHistInput)

	// Iterate through the Amazon S3 object pages.
	for p.HasMorePages() {
		calls++
		spotHistoryOutput, err := p.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get a spot pricing page, %w", err)
		}

		for _, history := range spotHistoryOutput.SpotPriceHistory {
			spotPrice, errFloat := strconv.ParseFloat(*history.SpotPrice, 64)
			if errFloat != nil {
				processingErr = multierr.Append(processingErr, errFloat)
				continue
			}
			spotTimeSeries[string(history.InstanceType)] = append(spotTimeSeries[string(history.InstanceType)], &spotPricingEntry{
				Timestamp: *history.Timestamp,
				SpotPrice: spotPrice,
				Zone:      *history.AvailabilityZone,
			})
		}
	}

	return spotTimeSeries, processingErr
}
