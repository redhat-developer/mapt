package bastion

import (
	// "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/elb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type BastionRequest struct {
}

type BastionResources struct {
}

func (r BastionRequest) Create(ctx *pulumi.Context) (*BastionResources, error) {
	// _, err := elb.NewLoadBalancer(ctx, "bar", &elb.LoadBalancerArgs{})
	return nil, nil
}
