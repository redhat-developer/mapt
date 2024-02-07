package data

import (
	"context"
	"fmt"
	"sync"

	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/aws/aws-sdk-go-v2/config"
	"golang.org/x/exp/maps"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type DedicatedHostResquest struct {
	HostID string
	Tags   map[string]string
}

func GetDedicatedHost(hostID string) (*ec2Types.Host, error) {
	hosts, err := GetDedicatedHosts(DedicatedHostResquest{
		HostID: hostID,
	})
	if err != nil {
		return nil, err
	}
	if len(hosts) != 1 {
		return nil, fmt.Errorf("error getting the dedicated host %s", hostID)
	}
	return &hosts[0], nil
}

// This function check on all regions for the dedicated host
// and return the list of hosts matching the request params
func GetDedicatedHosts(r DedicatedHostResquest) ([]ec2Types.Host, error) {
	regions, err := GetRegions()
	if err != nil {
		return nil, err
	}
	h := make(chan []ec2Types.Host, len(regions))
	e := make(chan string, 1)
	defer close(h)
	defer close(e)
	var wg sync.WaitGroup
	for _, region := range regions {
		wg.Add(1)
		lRegion := region
		go func(h chan []ec2Types.Host) {
			defer wg.Done()
			if hosts, err := getDedicatedHostByRegion(
				r, lRegion); err == nil {
				h <- hosts
			}
		}(h)
	}
	go func(c chan string) {
		wg.Wait()
		c <- "done"
	}(e)
	var hosts []ec2Types.Host
	for {
		exit := false
		select {
		case oHosts := <-h:
			hosts = append(hosts, oHosts...)
		case <-e:
			exit = true
		}
		if exit {
			break
		}
	}
	return hosts, nil
}

func getDedicatedHostByRegion(r DedicatedHostResquest, regionName string) ([]ec2Types.Host, error) {
	var cfgOpts config.LoadOptionsFunc
	if len(regionName) > 0 {
		cfgOpts = config.WithRegion(regionName)
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(), cfgOpts)
	if err != nil {
		return nil, err
	}
	client := ec2.NewFromConfig(cfg)
	// Describe params
	stateKey := "state"
	dhi := &ec2.DescribeHostsInput{
		// state
		Filter: []ec2Types.Filter{
			{
				Name: &stateKey,
				Values: []string{
					string(ec2Types.AllocationStateAvailable),
					string(ec2Types.AllocationStatePending)},
			},
		},
	}
	if len(r.HostID) > 0 {
		dhi.HostIds = []string{r.HostID}
	}
	if len(r.Tags) > 0 {
		tagKey := "tag-key"
		dhi.Filter = append(dhi.Filter,
			ec2Types.Filter{
				Name:   &tagKey,
				Values: maps.Keys(r.Tags)})
	}
	h, err := client.DescribeHosts(context.Background(), dhi)
	if err != nil {
		return nil, err
	}
	if len(h.Hosts) == 0 {
		return nil, fmt.Errorf("dedicated host was not found on current region")
	}
	if len(r.Tags) > 0 {
		return util.ArrayFilter(h.Hosts,
			func(h ec2Types.Host) bool {
				return allTagsMatches(r.Tags, h.Tags)
			}), nil
	}
	return h.Hosts, nil
}
