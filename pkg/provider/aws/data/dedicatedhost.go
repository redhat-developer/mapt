package data

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsEC2 "github.com/aws/aws-sdk-go/service/ec2"
)

// This function check on all regions for the dedicated host
// and return its state if found or error if no host is found within the hostID
func GetDedicatedHostState(hostID string) (*string, error) {
	regions, err := GetRegions()
	if err != nil {
		return nil, err
	}
	s := make(chan *string, len(regions))
	e := make(chan string, 1)
	defer close(s)
	defer close(e)
	var wg sync.WaitGroup
	for _, region := range regions {
		wg.Add(1)
		lRegion := region
		go func(c chan *string) {
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
		return oState, nil
	case <-e:
		return nil, fmt.Errorf("not host with hostID %s on any region", hostID)
	}
}

// This funcion check on a specific region if a hosts with the id or the filters exists
// and returns its state
func getDedicatedHostStateByRegion(regionName, hostID string) (*string, error) {
	config := aws.Config{}
	if len(regionName) > 0 {
		config.Region = aws.String(regionName)
	}
	sess, err := session.NewSession(&config)
	if err != nil {
		return nil, err
	}
	svc := awsEC2.New(sess)
	h, err := svc.DescribeHosts(&awsEC2.DescribeHostsInput{
		HostIds: aws.StringSlice([]string{hostID}),
	})
	if err != nil {
		return nil, err
	}
	if len(h.Hosts) == 0 {
		return nil, fmt.Errorf("dedicated host was not found on current region")
	}
	return h.Hosts[0].State, nil
}
