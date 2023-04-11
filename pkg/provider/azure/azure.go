package azure

import (
	"github.com/adrianriobo/qenvs/pkg/manager"
	"github.com/adrianriobo/qenvs/pkg/provider/azure/plugin"
)

func Destroy(projectName, backedURL, stackName string) error {
	stack := manager.Stack{
		StackName:           stackName,
		ProjectName:         projectName,
		BackedURL:           backedURL,
		CloudProviderPlugin: plugin.DefaultPlugin}
	return manager.DestroyStack(stack)
}
