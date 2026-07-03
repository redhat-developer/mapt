package openshift

import (
	"context"

	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/manager/credentials"
)

type OpenShift struct{}

func (o *OpenShift) Init(_ context.Context, _ string) (string, error) {
	return "", nil
}

func (o *OpenShift) DefaultHostingPlace() (*string, error) {
	return nil, nil
}

func Provider() *OpenShift {
	return &OpenShift{}
}

var NoCredentials = credentials.ProviderCredentials{}

func DestroyStack(mCtx *mc.Context, stackName string) error {
	return manager.DestroyStack(
		mCtx,
		manager.Stack{
			StackName:           mCtx.StackNameByProject(stackName),
			ProjectName:         mCtx.ProjectName(),
			BackedURL:           mCtx.BackedURL(),
			ProviderCredentials: NoCredentials,
		})
}
