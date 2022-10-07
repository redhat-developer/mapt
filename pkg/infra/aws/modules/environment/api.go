package environment

import (
	"github.com/adrianriobo/qenvs/pkg/infra/aws"
	utilInfra "github.com/adrianriobo/qenvs/pkg/util/infra"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
)

type corporateEnvironmentRequest struct {
	name string
}

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
	_, err := utilInfra.UpStack(stack)
	if err != nil {
		return err
	}
	logging.Debug("Environment has been created")
	return nil
}

func DestroyNetwork(projectName, backedURL string) error {
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
