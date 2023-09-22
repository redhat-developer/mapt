package amireplication

import (
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/manager"
	"github.com/adrianriobo/qenvs/pkg/provider/aws"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const stackName = "amiReplicate"

func (r ReplicatedRequest) runStackAsync(backedURL, region, operation string, errChan chan error) {
	errChan <- r.runStack(backedURL, region, operation)
}

func (r ReplicatedRequest) runStack(backedURL, region, operation string) error {
	stack := manager.Stack{
		StackName:   fmt.Sprintf("%s-%s", stackName, region),
		ProjectName: r.ProjectName,
		BackedURL:   backedURL,
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{aws.CONFIG_AWS_REGION: region}),
		DeployFunc: r.deployer,
	}

	var err error
	if operation == operationCreate {
		_, err = manager.UpStack(stack)
	} else {
		err = manager.DestroyStack(stack)
	}

	if err != nil {
		return err
	}
	return nil
}

func (r ReplicatedRequest) deployer(ctx *pulumi.Context) error {
	_, err := ec2.NewAmiCopy(ctx,
		r.AMITargetName,
		&ec2.AmiCopyArgs{
			Description: pulumi.String(
				fmt.Sprintf("Replica of %s from %s", r.AMISourceID, r.AMISourceRegion)),
			SourceAmiId:     pulumi.String(r.AMISourceID),
			SourceAmiRegion: pulumi.String(r.AMISourceRegion),
			Tags: pulumi.StringMap{
				"Name":    pulumi.String(r.AMITargetName),
				"Project": pulumi.String(r.ProjectName),
			},
		})
	if err != nil {
		return err
	}
	return nil
}

func (r ReplicatedRequest) Replicate(ctx *pulumi.Context) error {
	return r.deployer(ctx)
}
