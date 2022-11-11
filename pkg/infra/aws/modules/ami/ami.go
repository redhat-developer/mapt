package ami

import (
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/infra/aws"
	utilInfra "github.com/adrianriobo/qenvs/pkg/infra/util"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const stackName = "amiReplicate"

func (r replicatedRequest) runStackAsync(backedURL, region, operation string, errChan chan error) {
	errChan <- r.runStack(backedURL, region, operation)
}

func (r replicatedRequest) runStack(backedURL, region, operation string) error {
	stack := utilInfra.Stack{
		StackName:   fmt.Sprintf("%s-%s", stackName, region),
		ProjectName: r.projectName,
		BackedURL:   backedURL,
		Plugin:      aws.GetPluginAWS(map[string]string{aws.CONFIG_AWS_REGION: region}),
		DeployFunc:  r.deployer,
	}

	var err error
	if operation == operationCreate {
		_, err = utilInfra.UpStack(stack)
	} else {
		err = utilInfra.DestroyStack(stack)
	}

	if err != nil {
		return err
	}
	return nil
}

func (r replicatedRequest) deployer(ctx *pulumi.Context) error {
	_, err := ec2.NewAmiCopy(ctx,
		r.amiName,
		&ec2.AmiCopyArgs{
			Description: pulumi.String(
				fmt.Sprintf("Replica of %s from %s", r.amiID, r.amiSourceRegion)),
			SourceAmiId:     pulumi.String(r.amiID),
			SourceAmiRegion: pulumi.String(r.amiSourceRegion),
			Tags: pulumi.StringMap{
				"Name":    pulumi.String(r.amiName),
				"Project": pulumi.String(r.projectName),
			},
		})
	if err != nil {
		return err
	}
	return nil
}
