package azure

import (
	"github.com/redhat-developer/mapt/pkg/manager"
	"github.com/redhat-developer/mapt/pkg/manager/credentials"
)

func GetClouProviderCredentials(fixedCredentials map[string]string) credentials.ProviderCredentials {
	return credentials.ProviderCredentials{
		SetCredentialFunc: nil,
		FixedCredentials:  fixedCredentials}
}

var DefaultCredentials = GetClouProviderCredentials(nil)

func Destroy(projectName, backedURL, stackName string) error {
	stack := manager.Stack{
		StackName:           stackName,
		ProjectName:         projectName,
		BackedURL:           backedURL,
		ProviderCredentials: DefaultCredentials}
	return manager.DestroyStack(stack)
}
