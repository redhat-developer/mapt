package data

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

var ErrECSClusterNotFound = fmt.Errorf("cluster not found")

func GetCluster(ctx context.Context, clusterName, region string) (*string, error) {
	cfg, err := getConfig(ctx, region)
	if err != nil {
		return nil, err
	}
	client := ecs.NewFromConfig(cfg)
	listClustersOutput, err := client.ListClusters(
		ctx,
		&ecs.ListClustersInput{})
	if err != nil {
		return nil, err
	}
	if listClustersOutput == nil || len(listClustersOutput.ClusterArns) == 0 {
		return nil, ErrECSClusterNotFound
	}
	cls, err := client.DescribeClusters(
		ctx,
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
