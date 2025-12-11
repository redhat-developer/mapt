package ibmcloud

import (
	"context"
	"fmt"
	"os"

	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/manager/credentials"
)

const (
	LOCATION_ENV = "IC_REGION"
)

type IBMCloud struct{}

func (i *IBMCloud) Init(ctx context.Context, backedURL string) error {
	return nil
}

func (a *IBMCloud) DefaultHostingPlace() (*string, error) {
	hp := os.Getenv("IC_REGION")
	if len(hp) > 0 {
		return &hp, nil
	}
	return nil, fmt.Errorf("missing default value for IBM Cloud Region: IC_REGION")
}

func (a *IBMCloud) Zone() (*string, error) {
	hp := os.Getenv("IC_ZONE")
	if len(hp) > 0 {
		return &hp, nil
	}
	return nil, fmt.Errorf("missing default value for IBM Cloud Region: IC_ZONE")
}

func Provider() *IBMCloud {
	return &IBMCloud{}
}

func GetClouProviderCredentials(fixedCredentials map[string]string) credentials.ProviderCredentials {
	return credentials.ProviderCredentials{
		SetCredentialFunc: nil,
		FixedCredentials:  fixedCredentials}
}

var (
	DefaultCredentials = GetClouProviderCredentials(nil)
)

func Destroy(mCtx *mc.Context, stackName string) error {
	stack := manager.Stack{
		StackName:           mCtx.StackNameByProject(stackName),
		ProjectName:         mCtx.ProjectName(),
		BackedURL:           mCtx.BackedURL(),
		ProviderCredentials: DefaultCredentials}
	return manager.DestroyStack(mCtx, stack)
}

type gen2Location struct {
	region, zone string
}

type classicLocation string

var LocationMapping = map[classicLocation]gen2Location{
	"dal10": {region: "us-south", zone: "us-south-1"},
	"dal12": {region: "us-south", zone: "us-south-2"},
	"dal13": {region: "us-south", zone: "us-south-3"},
	"wdc06": {region: "us-east", zone: "us-east-1"},
	"tor01": {region: "ca-tor", zone: "ca-tor-1"},
	"mon01": {region: "ca-mon", zone: "us-south-2"},
	"lon04": {region: "eu-gb", zone: "eu-gb-1"},
	"fra04": {region: "eu-de", zone: "eu-de-1"},
	"fra05": {region: "eu-de", zone: "eu-de-2"},
	"syd04": {region: "au-syd", zone: "au-syd-1"},
	"tok04": {region: "jp-tok", zone: "jp-tok-1"}}

func ClassicLocation() *classicLocation {
	for k, v := range LocationMapping {
		if v.region == os.Getenv("IC_REGION") && v.zone == os.Getenv("IC_ZONE") {
			return &k
		}
	}
	return nil
}
