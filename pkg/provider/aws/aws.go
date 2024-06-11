package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsEC2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/manager/credentials"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	"github.com/redhat-developer/mapt/pkg/util/maps"
)

const (
	CONFIG_AWS_REGION     string = "aws:region"
	CONFIG_AWS_ACCESS_KEY string = "aws:accessKey"
	CONFIG_AWS_SECRET_KEY string = "aws:secretKey"
)

// pulumi config key : aws env credential
var envCredentials = map[string]string{
	CONFIG_AWS_REGION:     "AWS_DEFAULT_REGION",
	CONFIG_AWS_ACCESS_KEY: "AWS_ACCESS_KEY_ID",
	CONFIG_AWS_SECRET_KEY: "AWS_SECRET_ACCESS_KEY"}

var DefaultCredentials = GetClouProviderCredentials(nil)

func GetClouProviderCredentials(customCredentials map[string]string) credentials.ProviderCredentials {
	return credentials.ProviderCredentials{
		SetCredentialFunc: SetAWSCredentials,
		FixedCredentials:  customCredentials}
}

func SetAWSCredentials(ctx context.Context, stack auto.Stack, customCredentials map[string]string) error {
	for configKey, envKey := range envCredentials {
		if value, ok := customCredentials[configKey]; ok {
			if err := stack.SetConfig(ctx, configKey,
				auto.ConfigValue{Value: value}); err != nil {
				logging.Errorf("Failed setting credential: %v", err)
				return err
			}
		} else {
			if err := stack.SetConfig(ctx, configKey,
				auto.ConfigValue{Value: os.Getenv(envKey)}); err != nil {
				logging.Errorf("Failed setting credential: %v", err)
				return err
			}
		}
	}
	return nil
}

type DestroyStackRequest struct {
	Region    string
	BackedURL string
	Stackname string
}

func DestroyStack(s DestroyStackRequest) error {
	if len(s.Stackname) == 0 {
		return fmt.Errorf("stackname is required")
	}
	return manager.DestroyStack(manager.Stack{
		StackName:   maptContext.StackNameByProject(s.Stackname),
		ProjectName: maptContext.ProjectName(),
		BackedURL: util.If(len(s.BackedURL) > 0,
			s.BackedURL,
			maptContext.BackedURL()),
		ProviderCredentials: GetClouProviderCredentials(
			map[string]string{
				CONFIG_AWS_REGION: util.If(len(s.Region) > 0,
					s.Region,
					os.Getenv("AWS_DEFAULT_REGION"))})})
}

// Create a list of filters for tags based on the tags added by mapt
func GetTagsAsFilters() (filters []*awsEC2Types.Filter) {
	filterMap := maps.Convert(maptContext.GetTags(),
		func(name string) *string { return aws.String("tag:" + name) },
		func(value string) []string { return []string{value} })
	for k, v := range filterMap {
		filter := awsEC2Types.Filter{
			Name:   k,
			Values: v,
		}
		filters = append(filters, &filter)
	}
	return
}
