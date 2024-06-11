package ami

import (
	"fmt"
	"os"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

type ReplicatedRequest struct {
	ProjectName     string
	AMITargetName   string
	AMISourceID     string
	AMISourceRegion string
}

func CreateReplica(projectName, backedURL,
	amiID, amiName, amiSourceRegion string) (err error) {
	return manageReplica(projectName, backedURL, amiID, amiName, amiSourceRegion, "create")
}

func DestroyReplica(projectName, backedURL string) (err error) {
	return manageReplica(projectName, backedURL, "", "", "", "destroy")
}

func manageReplica(projectName, backedURL,
	amiID, amiName, amiSourceRegion, operation string) (err error) {

	request := ReplicatedRequest{
		ProjectName:     projectName,
		AMITargetName:   amiName,
		AMISourceID:     amiID,
		AMISourceRegion: amiSourceRegion}

	regions, err := data.GetRegions()
	if err != nil {
		logging.Errorf("failed to get regions")
		os.Exit(1)
	}
	errChan := make(chan error)
	for _, region := range regions {
		// Do not replicate on source region
		if region != amiSourceRegion {
			go request.runStackAsync(backedURL, region, operation, errChan)
		}
	}
	hasErrors := false
	for _, region := range regions {
		if region != amiSourceRegion {
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

func (r ReplicatedRequest) runStackAsync(backedURL, region, operation string, errChan chan error) {
	errChan <- r.runStack(backedURL, region, operation)
}

func (r ReplicatedRequest) runStack(backedURL, region, operation string) error {
	stack := manager.Stack{
		StackName:   fmt.Sprintf("%s-%s", "amiReplicate", region),
		ProjectName: r.ProjectName,
		BackedURL:   backedURL,
		ProviderCredentials: aws.GetClouProviderCredentials(
			map[string]string{aws.CONFIG_AWS_REGION: region}),
		DeployFunc: r.deployer,
	}

	var err error
	if operation == "create" {
		_, err = manager.UpStack(stack,
			manager.ManagerOptions{Baground: true})
	} else {
		err = manager.DestroyStack(stack,
			manager.ManagerOptions{Baground: true})
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
