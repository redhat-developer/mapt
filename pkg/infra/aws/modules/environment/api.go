package environment

import (
	"fmt"
	"os"

	"github.com/adrianriobo/qenvs/pkg/infra/aws"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/bastion"
	utilInfra "github.com/adrianriobo/qenvs/pkg/util/infra"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
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
	bastionPrivateKey, ok := stackResult.Outputs[bastion.OutputPrivateKey].Value.(string)
	if !ok {
		return fmt.Errorf("error getting private key for bastion")
	}
	err = os.WriteFile("/tmp/qenvs/id_rsa", []byte(bastionPrivateKey), 0600)
	if err != nil {
		return err
	}
	logging.Debug("Environment has been created")
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
