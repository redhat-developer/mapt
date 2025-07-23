package ami

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	awsConstants "github.com/redhat-developer/mapt/pkg/provider/aws/constants"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

type replicateArgs struct {
	AMITargetName   string
	AMISourceID     string
	AMISourceRegion string
}

type replicateRequest struct {
	mCtx            *mc.Context `validate:"required"`
	amiTargetName   string
	amiSourceID     string
	amiSourceRegion string
}

func (r *replicateRequest) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(r)
}

func CreateReplica(mCtxArgs *mc.ContextArgs, args *replicateArgs) (err error) {
	mCtx, err := mc.Init(mCtxArgs, aws.Provider())
	if err != nil {
		return err
	}
	return manageReplica(
		&replicateRequest{
			mCtx:            mCtx,
			amiTargetName:   args.AMITargetName,
			amiSourceID:     args.AMISourceID,
			amiSourceRegion: args.AMISourceRegion,
		},
		"create")
}

func DestroyReplica(mCtxArgs *mc.ContextArgs) (err error) {
	mCtx, err := mc.Init(mCtxArgs, aws.Provider())
	if err != nil {
		return err
	}
	return manageReplica(&replicateRequest{
		mCtx: mCtx,
	}, "destroy")
}

func manageReplica(r *replicateRequest, operation string) (err error) {
	regions, err := data.GetRegions()
	if err != nil {
		return err
	}
	errChan := make(chan error)
	for _, region := range regions {
		// Do not replicate on source region
		if region != r.amiSourceRegion {
			go r.runStackAsync(region, operation, errChan)
		}
	}
	hasErrors := false
	for _, region := range regions {
		if region != r.amiSourceRegion {
			if err := <-errChan; err != nil {
				logging.Errorf("%v", err)
				hasErrors = true
			}
		}
	}
	if hasErrors {
		return fmt.Errorf("there are errors on some replications. Check the logs to get information")
	}
	return nil
}

func (r replicateRequest) runStackAsync(region, operation string, errChan chan error) {
	errChan <- r.runStack(region, operation)
}

func (r replicateRequest) runStack(region, operation string) error {
	stack := manager.Stack{
		StackName:   fmt.Sprintf("%s-%s", "amiReplicate", region),
		ProjectName: r.mCtx.ProjectName(),
		BackedURL:   r.mCtx.BackedURL(),
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{awsConstants.CONFIG_AWS_REGION: region}),
		DeployFunc: r.deployer,
	}

	var err error
	if operation == "create" {
		_, err = manager.UpStack(r.mCtx, stack,
			manager.ManagerOptions{Baground: true})
	} else {
		err = manager.DestroyStack(r.mCtx, stack,
			manager.ManagerOptions{Baground: true})
	}

	if err != nil {
		return err
	}
	return nil
}

func (r replicateRequest) deployer(ctx *pulumi.Context) error {
	_, err := ec2.NewAmiCopy(ctx,
		r.amiTargetName,
		&ec2.AmiCopyArgs{
			Description: pulumi.String(
				fmt.Sprintf("Replica of %s from %s", r.amiSourceID, r.amiSourceRegion)),
			SourceAmiId:     pulumi.String(r.amiSourceID),
			SourceAmiRegion: pulumi.String(r.amiSourceRegion),
			Tags: pulumi.StringMap{
				"Name":    pulumi.String(r.amiTargetName),
				"Project": pulumi.String(r.mCtx.ProjectName()),
			},
		})
	if err != nil {
		return err
	}
	return nil
}

func (r replicateRequest) Replicate(ctx *pulumi.Context) error {
	return r.deployer(ctx)
}
