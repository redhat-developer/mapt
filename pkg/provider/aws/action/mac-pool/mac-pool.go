package macpool

import (
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/aws/modules/iam"
	macPool "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/pool"
	macUtil "github.com/redhat-developer/mapt/pkg/provider/aws/modules/mac/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

const (
	stackName    = "stackMacPool"
	awsMacPoolID = "amp"
)

// func Create(ctx *maptContext.ContextArgs, r *PoolRequestArgs) error {
// 	logging.Debug("Run mac pool create")
// 	// Create mapt Context
// 	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
// 		return err
// 	}
// 	re := os.Getenv("AWS_DEFAULT_REGION")
// 	it := "c5d.xlarge"
// 	dhArgs := dedicatedhost.DedicatedHostArgs{
// 		Region:       &re,
// 		InstanceType: &it,
// 		Tags:         map[string]string{"test": "test"},
// 	}
// 	burl := maptContext.BackedURL()
// 	_, err := dedicatedhost.Create(&burl, false, &dhArgs)
// 	return err
// }

// func Destroy(ctx *maptContext.ContextArgs) (err error) {
// 	logging.Debug("Run mac pool destroy")
// 	// Create mapt Context
// 	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
// 		return err
// 	}
// 	return dedicatedhost.Destroy(nil)
// }

// Create works as an orchestrator for create n machines based on offered capacity
// if pool already exists just change the params for the HouseKeeper
// also the HouseKeep will take care of regulate the capacity

// Even if we want to destroy the pool we will set params to max size 0
func Create(ctx *maptContext.ContextArgs, r *PoolRequestArgs) error {
	logging.Debug("Run mac pool create")
	// Create mapt Context
	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
		return err
	}
	cs := manager.Stack{
		StackName:           maptContext.StackNameByProject(stackName),
		ProjectName:         maptContext.ProjectName(),
		BackedURL:           maptContext.BackedURL(),
		ProviderCredentials: aws.DefaultCredentials,
		DeployFunc:          r.deploy,
	}
	sr, _ := manager.UpStack(cs)
	return r.results(sr)
}

func Destroy(ctx *maptContext.ContextArgs) (err error) {
	logging.Debug("Run mac pool destroy")
	// Create mapt Context
	if err := maptContext.Init(ctx, aws.Provider()); err != nil {
		return err
	}
	return aws.DestroyStack(
		aws.DestroyStackRequest{
			Stackname: stackName,
		})
}

// House keeper is the function executed serverless to check if is there any
// machine non locked which had been running more than 24h.
// It should check if capacity allows to remove the machine
func HouseKeeper(ctx *maptContext.ContextArgs, r *HouseKeepRequestArgs) error {
	return houseKeeper(ctx, r)
}

func Request(ctx *maptContext.ContextArgs, r *RequestMachineArgs) error {
	// If remote run through serverless
	if ctx.Remote {
		// Generate ticket
		ticket, err := ticket()
		if err != nil {
			return err
		}
		if err = macPool.RequestRemote(ctx, &r.PoolName, &r.Architecture, &r.OSVersion, ticket); err != nil {
			return err
		}
		return writeTicket(ticket)
	}
	return request(ctx, r)
}

func Release(ctx *maptContext.ContextArgs, m *MachineRequestArgs, ticket string) error {
	// If remote run through serverless
	if ctx.Remote {
		return macPool.ReleaseRemote(ctx, ticket)
	}
	return macUtil.Release(ctx, ticket)
}

func (r *PoolRequestArgs) deploy(ctx *pulumi.Context) error {
	_, err := macPool.NewPool(ctx,
		resourcesUtil.GetResourceName(r.Prefix, awsMacPoolID, "mac-pool"),
		&macPool.PoolArgs{
			Region:          os.Getenv("AWS_DEFAULT_REGION"),
			Name:            r.Name,
			Arch:            r.Architecture,
			OSVersion:       r.OSVersion,
			OfferedCapacity: r.OfferedCapacity,
			MaxSize:         r.MaxSize,
		})
	return err
}

func (r *PoolRequestArgs) results(stackResult auto.UpResult) error {
	return iam.Results(stackResult, r.Name)
}
