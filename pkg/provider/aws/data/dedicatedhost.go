package data

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awsEC2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awsEC2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// This function check on all regions for the dedicated host
// and return its state if found or error if no host is found within the hostID
func GetDedicatedHostState(hostID string) (*string, error) {
	regions, err := GetRegions()
	if err != nil {
		return nil, err
	}
	s := make(chan *awsEC2Types.AllocationState, len(regions))
	e := make(chan string, 1)
	defer close(s)
	defer close(e)
	var wg sync.WaitGroup
	for _, region := range regions {
		wg.Add(1)
		lRegion := region
		go func(c chan *awsEC2Types.AllocationState) {
			defer wg.Done()
			if state, err := getDedicatedHostStateByRegion(
				lRegion,
				hostID); err == nil {
				c <- state
			}
		}(s)
	}
	go func(c chan string) {
		wg.Wait()
		c <- "done"
	}(e)
	select {
	case oState := <-s:
		oStateAsString := fmt.Sprintf("%v", oState)
		return &oStateAsString, nil
	case <-e:
		return nil, fmt.Errorf("not host with hostID %s on any region", hostID)
	}
}

// This funcion check on a specific region if a hosts with the id or the filters exists
// and returns its state
func getDedicatedHostStateByRegion(regionName, hostID string) (*awsEC2Types.AllocationState, error) {
	var cfgOpts config.LoadOptionsFunc
	if len(regionName) > 0 {
		cfgOpts = config.WithRegion(regionName)
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(), cfgOpts)
	if err != nil {
		return nil, err
	}
	client := ec2.NewFromConfig(cfg)
	// config := aws.Config{}
	// if len(regionName) > 0 {
	// 	config.Region = aws.String(regionName)
	// }
	// sess, err := session.NewSession(&config)
	// if err != nil {
	// 	return nil, err
	// }
	// svc := awsEC2.New(sess)
	h, err := client.DescribeHosts(context.Background(),
		&awsEC2.DescribeHostsInput{
			HostIds: []string{hostID},
		})
	if err != nil {
		return nil, err
	}
	if len(h.Hosts) == 0 {
		return nil, fmt.Errorf("dedicated host was not found on current region")
	}
	return &h.Hosts[0].State, nil
}
