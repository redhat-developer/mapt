package environment

import (
	"fmt"
	"os"
	"path"

	"github.com/adrianriobo/qenvs/pkg/infra/aws"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/bastion"
	utilInfra "github.com/adrianriobo/qenvs/pkg/infra/util"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

func Create(projectName, backedURL string) error {
	request := corporateEnvironmentRequest{
		name: projectName,
	}
	// Here will run the cost analyzer
	// based on amount of VMs the min spot price will be calculated
	// and region / az will be used accordingly
	stack := utilInfra.Stack{
		StackName:   stackCreateEnvironmentName,
		ProjectName: projectName,
		BackedURL:   backedURL,
		Plugin:      aws.PluginAWSDefault,
		DeployFunc:  request.deployer,
	}
	// Exec stack
	stackResult, err := utilInfra.UpStack(stack)
	if err != nil {
		return err
	}
	if err = writeOutput(stackResult, bastion.OutputPrivateKey, "/tmp/qenvs", "id_rsa"); err != nil {
		return err
	}
	if err = writeOutput(stackResult, bastion.OutputPublicIP, "/tmp/qenvs", "host"); err != nil {
		return err
	}
	if err = writeOutput(stackResult, bastion.OutputUsername, "/tmp/qenvs", "username"); err != nil {
		return err
	}
	logging.Debug("Environment has been created")
	return nil
}

func writeOutput(stackResult auto.UpResult, outputkey, destinationFolder, destinationFilename string) error {
	value, ok := stackResult.Outputs[outputkey].Value.(string)
	if !ok {
		return fmt.Errorf("error getting private key for bastion")
	}
	err := os.WriteFile(path.Join(destinationFolder, destinationFilename), []byte(value), 0600)
	if err != nil {
		return err
	}
	return nil
}

func Destroy(projectName, backedURL string) error {
	stack := utilInfra.Stack{
		StackName:   stackCreateEnvironmentName,
		ProjectName: projectName,
		BackedURL:   backedURL,
		Plugin:      aws.PluginAWSDefault}
	_, err := utilInfra.DestroyStack(stack)
	if err != nil {
		return err
	}
	logging.Debugf("Environment has been destroyed")
	return nil
}
