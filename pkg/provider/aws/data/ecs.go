package data

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/redhat-developer/mapt/pkg/util"
	"golang.org/x/exp/slices"
)

var ErrECSClusterNotFound = fmt.Errorf("cluster not found")

func GetCluster(clusterName, region string) (*string, error) {
	cfg, err := getConfig(region)
	if err != nil {
		return nil, err
	}
	client := ecs.NewFromConfig(cfg)
	listClustersOutput, err := client.ListClusters(
		context.TODO(),
		&ecs.ListClustersInput{})
	if err != nil {
		return nil, err
	}
	if listClustersOutput == nil || len(listClustersOutput.ClusterArns) == 0 {
		return nil, ErrECSClusterNotFound
	}
	cls, err := client.DescribeClusters(
		context.TODO(),
		&ecs.DescribeClustersInput{
			Clusters: listClustersOutput.ClusterArns,
		})
	if err != nil {
		return nil, err
	}
	for _, c := range cls.Clusters {
		if *c.ClusterName == clusterName {
			return c.ClusterArn, nil
		}
	}
	return nil, ErrECSClusterNotFound
}

func ActiveTasks(taskArns []string) (*string, error) {
	cfg, err := getGlobalConfig()
	if err != nil {
		return nil, err
	}
	client := ecs.NewFromConfig(cfg)
	tDefs, err := client.ListTaskDefinitions(
		context.TODO(),
		&ecs.ListTaskDefinitionsInput{
			Status: types.TaskDefinitionStatusActive,
		})
	if err != nil {
		return nil, err
	}
	activeTasks := util.ArrayFilter(
		tDefs.TaskDefinitionArns,
		func(arn string) bool {
			return slices.Contains(taskArns, arn)
		})
	if len(activeTasks) != 1 {
		return nil, fmt.Errorf("there should be exactly one active task")
	}
	return &activeTasks[0], nil
}
