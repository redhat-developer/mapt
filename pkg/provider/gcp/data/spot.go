package data

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	spot "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	computeBeta "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/option"
)

var (
	obtainabilityThresholds = map[spot.Tolerance]float64{
		spot.Lowest:  0.7,
		spot.Low:     0.5,
		spot.Medium:  0.4,
		spot.High:    0.2,
		spot.Highest: 0.0,
	}

	preemptionThresholds = map[spot.Tolerance]float64{
		spot.Lowest:  0.05,
		spot.Low:     0.10,
		spot.Medium:  0.15,
		spot.High:    0.20,
		spot.Highest: 1.00,
	}

	tier1Regions = map[string]bool{
		"us-central1":   true,
		"us-east4":      true,
		"europe-west1":  true,
		"europe-west4":  true,
	}

	// Older families tend to have higher spot availability
	preferredFamilies = map[string]int{
		"n1": 3, "n2": 3, "e2": 3,
		"n2d": 2, "t2d": 2, "c2": 2,
		"c3": 1, "c4": 1, "m3": 1, "n4": 1,
	}

	maxMachineTypesPerCapacityQuery = 5
)

type exhaustionEntry struct {
	expiry time.Time
}

type exhaustionCache struct {
	mu      sync.Mutex
	entries map[string]exhaustionEntry
}

func newExhaustionCache() *exhaustionCache {
	return &exhaustionCache{
		entries: make(map[string]exhaustionEntry),
	}
}

func (c *exhaustionCache) add(zone, machineType string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := zone + "/" + machineType
	c.entries[key] = exhaustionEntry{expiry: time.Now().Add(15 * time.Minute)}
}

func (c *exhaustionCache) isExhausted(zone, machineType string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := zone + "/" + machineType
	entry, ok := c.entries[key]
	if !ok {
		return false
	}
	if time.Now().After(entry.expiry) {
		delete(c.entries, key)
		return false
	}
	return true
}

type SpotSelector struct {
	cache *exhaustionCache
}

func NewSpotSelector() *SpotSelector {
	return &SpotSelector{
		cache: newExhaustionCache(),
	}
}

func (s *SpotSelector) Select(mCtx *mc.Context, args *spot.SpotRequestArgs) (*spot.SpotResults, error) {
	return s.selectSpot(mCtx, args, nil)
}

func (s *SpotSelector) selectSpot(mCtx *mc.Context, args *spot.SpotRequestArgs, excludedZones []string) (*spot.SpotResults, error) {
	ctx := mCtx.Context()
	tolerance := spot.DefaultTolerance
	if args.SpotParams != nil {
		tolerance = args.SpotParams.Tolerance
	}

	machineTypes := args.ComputeRequest.ComputeSizes
	if len(machineTypes) == 0 {
		region := mCtx.TargetHostingPlace()
		zones, err := GetAvailabilityZones(ctx, region)
		if err != nil {
			return nil, err
		}
		for _, zone := range zones {
			mts, err := NewComputeSelector().Select(ctx, args.ComputeRequest, zone)
			if err != nil {
				continue
			}
			machineTypes = mts
			break
		}
		if len(machineTypes) == 0 {
			return nil, fmt.Errorf("no matching machine types found for spot selection")
		}
	}

	regions, err := s.getRegions(ctx, args.SpotParams)
	if err != nil {
		return nil, err
	}

	candidates, err := s.queryCapacity(ctx, regions, machineTypes, tolerance)
	if err != nil {
		logging.Debugf("advice.capacity API failed, falling back to zone-based selection: %v", err)
		return s.fallbackSelection(mCtx, args, machineTypes, regions, tolerance)
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no zones meet the obtainability threshold for tolerance %d", tolerance)
	}

	candidates = s.filterExhausted(candidates, excludedZones)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("all candidate zones are exhausted or excluded")
	}

	candidates, err = s.filterByPreemption(ctx, candidates, tolerance)
	if err != nil {
		logging.Debugf("advice.capacityHistory API failed, skipping preemption filtering: %v", err)
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no zones meet the preemption rate threshold for tolerance %d", tolerance)
	}

	candidates, err = s.enrichWithPricing(ctx, candidates)
	if err != nil {
		logging.Debugf("price enrichment failed, using candidates without pricing: %v", err)
	}

	s.rankCandidates(candidates)

	best := candidates[0]
	price := best.price
	if price <= 0 {
		price = 0.10
	}

	safePrice := spot.SafePrice(price, nil)
	if args.SpotParams != nil && args.SpotParams.IncreaseRate > 0 {
		safePrice = spot.SafePrice(price, &args.SpotParams.IncreaseRate)
	}

	logging.Debugf("spot selection: zone=%s, machineType=%s, obtainability=%.2f, price=%.4f, safePrice=%.4f",
		best.zone, best.machineType, best.obtainability, best.price, safePrice)

	return &spot.SpotResults{
		ComputeType:      []string{best.machineType},
		Price:            safePrice,
		HostingPlace:     best.region,
		AvailabilityZone: best.zone,
	}, nil
}

func (s *SpotSelector) MarkExhausted(zone, machineType string) {
	s.cache.add(zone, machineType)
}

type spotCandidate struct {
	region         string
	zone           string
	machineType    string
	obtainability  float64
	preemptionRate float64
	price          float64
	score          float64
}

func (s *SpotSelector) getRegions(ctx context.Context, spotParams *spot.SpotArgs) ([]string, error) {
	project, err := getProject(ctx)
	if err != nil {
		return nil, err
	}

	svc, err := computeBeta.NewService(ctx, option.WithScopes(computeBeta.ComputeReadonlyScope))
	if err != nil {
		return nil, fmt.Errorf("error creating GCP compute beta service: %v", err)
	}

	regionList, err := svc.Regions.List(project).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("error listing regions: %v", err)
	}

	var regions []string
	for _, r := range regionList.Items {
		if r.Status != "UP" {
			continue
		}
		if spotParams != nil && len(spotParams.ExcludedHostingPlaces) > 0 {
			excluded := false
			for _, ex := range spotParams.ExcludedHostingPlaces {
				if r.Name == ex {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}
		regions = append(regions, r.Name)
	}

	if len(regions) == 0 {
		return nil, fmt.Errorf("no available GCP regions found")
	}

	logging.Debugf("spot: found %d candidate regions", len(regions))
	return regions, nil
}

func (s *SpotSelector) queryCapacity(ctx context.Context, regions []string, machineTypes []string, tolerance spot.Tolerance) ([]spotCandidate, error) {
	project, err := getProject(ctx)
	if err != nil {
		return nil, err
	}

	svc, err := computeBeta.NewService(ctx, option.WithScopes(computeBeta.ComputeScope))
	if err != nil {
		return nil, fmt.Errorf("error creating GCP compute beta service: %v", err)
	}

	minScore := obtainabilityThresholds[tolerance]

	// API supports max 5 machine types per query — batch them
	mtBatches := batchMachineTypes(machineTypes, maxMachineTypesPerCapacityQuery)

	var mu sync.Mutex
	var candidates []spotCandidate
	var wg sync.WaitGroup
	var firstErr error

	for _, region := range regions {
		for _, batch := range mtBatches {
			wg.Add(1)
			go func(region string, mts []string) {
				defer wg.Done()

				selections := make(map[string]computeBeta.CapacityAdviceRequestInstanceFlexibilityPolicyInstanceSelection)
				selections["sel0"] = computeBeta.CapacityAdviceRequestInstanceFlexibilityPolicyInstanceSelection{
					MachineTypes: mts,
				}

				req := &computeBeta.CapacityAdviceRequest{
					DistributionPolicy: &computeBeta.CapacityAdviceRequestDistributionPolicy{
						TargetShape: "ANY_SINGLE_ZONE",
					},
					InstanceFlexibilityPolicy: &computeBeta.CapacityAdviceRequestInstanceFlexibilityPolicy{
						InstanceSelections: selections,
					},
					InstanceProperties: &computeBeta.CapacityAdviceRequestInstanceProperties{
						Scheduling: &computeBeta.CapacityAdviceRequestInstancePropertiesScheduling{
							ProvisioningModel: "SPOT",
						},
					},
					Size: 1,
				}

				resp, err := svc.Advice.Capacity(project, region, req).Context(ctx).Do()
				if err != nil {
					logging.Debugf("advice.capacity failed for region %s: %v", region, err)
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
					return
				}

				mu.Lock()
				defer mu.Unlock()
				for _, rec := range resp.Recommendations {
					if rec.Scores == nil || rec.Scores.Obtainability < minScore {
						continue
					}
					for _, shard := range rec.Shards {
						if shard.ProvisioningModel != "SPOT" {
							continue
						}
						candidates = append(candidates, spotCandidate{
							region:        region,
							zone:          shard.Zone,
							machineType:   shard.MachineType,
							obtainability: rec.Scores.Obtainability,
						})
					}
				}
			}(region, batch)
		}
	}

	wg.Wait()

	if len(candidates) == 0 && firstErr != nil {
		return nil, firstErr
	}

	logging.Debugf("spot: found %d candidates above obtainability threshold %.2f", len(candidates), minScore)
	return candidates, nil
}

func (s *SpotSelector) filterExhausted(candidates []spotCandidate, excludedZones []string) []spotCandidate {
	var filtered []spotCandidate
	for _, c := range candidates {
		if s.cache.isExhausted(c.zone, c.machineType) {
			logging.Debugf("spot: skipping exhausted zone+type %s/%s", c.zone, c.machineType)
			continue
		}
		excluded := false
		for _, ez := range excludedZones {
			if c.zone == ez {
				excluded = true
				break
			}
		}
		if !excluded {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func (s *SpotSelector) filterByPreemption(ctx context.Context, candidates []spotCandidate, tolerance spot.Tolerance) ([]spotCandidate, error) {
	project, err := getProject(ctx)
	if err != nil {
		return candidates, err
	}

	svc, err := computeBeta.NewService(ctx, option.WithScopes(computeBeta.ComputeReadonlyScope))
	if err != nil {
		return candidates, err
	}

	maxRate := preemptionThresholds[tolerance]

	type historyKey struct {
		region      string
		machineType string
	}
	var mu sync.Mutex
	preemptionRates := make(map[historyKey]float64)
	prices := make(map[historyKey]float64)
	var wg sync.WaitGroup

	seen := make(map[historyKey]bool)
	for _, c := range candidates {
		key := historyKey{c.region, c.machineType}
		if seen[key] {
			continue
		}
		seen[key] = true
		wg.Add(1)
		go func(region, machineType string) {
			defer wg.Done()
			req := &computeBeta.CapacityHistoryRequest{
				InstanceProperties: &computeBeta.CapacityHistoryRequestInstanceProperties{
					MachineType: machineType,
					Scheduling: &computeBeta.CapacityHistoryRequestInstancePropertiesScheduling{
						ProvisioningModel: "SPOT",
					},
				},
				Types: []string{"PREEMPTION", "PRICE"},
			}

			resp, err := svc.Advice.CapacityHistory(project, region, req).Context(ctx).Do()
			if err != nil {
				logging.Debugf("advice.capacityHistory failed for %s/%s: %v", region, machineType, err)
				return
			}

			mu.Lock()
			defer mu.Unlock()

			if len(resp.PreemptionHistory) > 0 {
				var total float64
				for _, p := range resp.PreemptionHistory {
					total += p.PreemptionRate
				}
				avg := total / float64(len(resp.PreemptionHistory))
				preemptionRates[historyKey{region, machineType}] = avg
			}

			if len(resp.PriceHistory) > 0 {
				latest := resp.PriceHistory[len(resp.PriceHistory)-1]
				if latest.ListPrice != nil {
					prices[historyKey{region, machineType}] = moneyToFloat(latest.ListPrice)
				}
			}
		}(c.region, c.machineType)
	}

	wg.Wait()

	for i := range candidates {
		if p, ok := prices[historyKey{candidates[i].region, candidates[i].machineType}]; ok {
			candidates[i].price = p
		}
	}

	var filtered []spotCandidate
	for _, c := range candidates {
		key := historyKey{c.region, c.machineType}
		rate, hasRate := preemptionRates[key]
		if hasRate {
			c.preemptionRate = rate
			if rate > maxRate {
				logging.Debugf("spot: filtering out %s/%s: preemption rate %.3f exceeds threshold %.3f",
					c.zone, c.machineType, rate, maxRate)
				continue
			}
		}
		filtered = append(filtered, c)
	}

	return filtered, nil
}

func (s *SpotSelector) enrichWithPricing(ctx context.Context, candidates []spotCandidate) ([]spotCandidate, error) {
	hasPrices := false
	for _, c := range candidates {
		if c.price > 0 {
			hasPrices = true
			break
		}
	}
	if hasPrices {
		return candidates, nil
	}

	project, err := getProject(ctx)
	if err != nil {
		return candidates, err
	}

	svc, err := computeBeta.NewService(ctx, option.WithScopes(computeBeta.ComputeReadonlyScope))
	if err != nil {
		return candidates, err
	}

	type priceKey struct {
		region      string
		machineType string
	}
	var mu sync.Mutex
	prices := make(map[priceKey]float64)
	var wg sync.WaitGroup

	seen := make(map[priceKey]bool)
	for _, c := range candidates {
		key := priceKey{c.region, c.machineType}
		if seen[key] {
			continue
		}
		seen[key] = true
		wg.Add(1)
		go func(region, machineType string) {
			defer wg.Done()
			req := &computeBeta.CapacityHistoryRequest{
				InstanceProperties: &computeBeta.CapacityHistoryRequestInstanceProperties{
					MachineType: machineType,
					Scheduling: &computeBeta.CapacityHistoryRequestInstancePropertiesScheduling{
						ProvisioningModel: "SPOT",
					},
				},
				Types: []string{"PRICE"},
			}
			resp, err := svc.Advice.CapacityHistory(project, region, req).Context(ctx).Do()
			if err != nil {
				return
			}
			if len(resp.PriceHistory) > 0 {
				latest := resp.PriceHistory[len(resp.PriceHistory)-1]
				if latest.ListPrice != nil {
					mu.Lock()
					prices[priceKey{region, machineType}] = moneyToFloat(latest.ListPrice)
					mu.Unlock()
				}
			}
		}(c.region, c.machineType)
	}
	wg.Wait()

	for i := range candidates {
		key := priceKey{candidates[i].region, candidates[i].machineType}
		if p, ok := prices[key]; ok {
			candidates[i].price = p
		}
	}
	return candidates, nil
}

func (s *SpotSelector) rankCandidates(candidates []spotCandidate) {
	for i := range candidates {
		c := &candidates[i]
		price := c.price
		if price <= 0 {
			price = 0.10
		}
		priceScore := 1.0 / price

		hubWeight := 1.0
		region := extractRegion(c.zone)
		if tier1Regions[region] {
			hubWeight = 1.2
		}

		familyWeight := 1.0
		family := extractFamily(c.machineType)
		if w, ok := preferredFamilies[family]; ok {
			familyWeight = float64(w)
		}

		c.score = c.obtainability * priceScore * hubWeight * familyWeight
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})
}

func (s *SpotSelector) fallbackSelection(mCtx *mc.Context, args *spot.SpotRequestArgs, machineTypes []string, regions []string, tolerance spot.Tolerance) (*spot.SpotResults, error) {
	for _, region := range regions {
		zones, err := GetAvailabilityZones(mCtx.Context(), region)
		if err != nil {
			continue
		}
		for _, zone := range zones {
			if s.cache.isExhausted(zone, machineTypes[0]) {
				continue
			}
			zoneMTs, err := NewComputeSelector().Select(mCtx.Context(), args.ComputeRequest, zone)
			if err != nil || len(zoneMTs) == 0 {
				continue
			}

			logging.Debugf("spot fallback: selected zone %s with machine type %s", zone, zoneMTs[0])
			return &spot.SpotResults{
				ComputeType:      zoneMTs,
				Price:            0.10,
				HostingPlace:     region,
				AvailabilityZone: zone,
			}, nil
		}
	}
	return nil, fmt.Errorf("no suitable zones found for spot allocation (fallback)")
}

func batchMachineTypes(machineTypes []string, batchSize int) [][]string {
	var batches [][]string
	for i := 0; i < len(machineTypes); i += batchSize {
		end := i + batchSize
		if end > len(machineTypes) {
			end = len(machineTypes)
		}
		batches = append(batches, machineTypes[i:end])
	}
	return batches
}

func moneyToFloat(m *computeBeta.Money) float64 {
	return float64(m.Units) + float64(m.Nanos)/1e9
}

func extractRegion(zone string) string {
	parts := strings.Split(zone, "-")
	if len(parts) >= 2 {
		return strings.Join(parts[:len(parts)-1], "-")
	}
	return zone
}

func extractFamily(machineType string) string {
	parts := strings.Split(machineType, "-")
	if len(parts) > 0 {
		return parts[0]
	}
	return machineType
}
