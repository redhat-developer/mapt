package aws

import (
	"context"
	"os"

	"github.com/adrianriobo/qenvs/pkg/manager"
	qenvsContext "github.com/adrianriobo/qenvs/pkg/manager/context"
	"github.com/adrianriobo/qenvs/pkg/manager/credentials"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/adrianriobo/qenvs/pkg/util/maps"
	"github.com/aws/aws-sdk-go/aws"
	awsEC2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
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

func DestroyStackByRegion(region, stackname string) error {
	stack := manager.Stack{
		StackName:   qenvsContext.GetStackInstanceName(stackname),
		ProjectName: qenvsContext.GetInstanceName(),
		BackedURL:   qenvsContext.GetBackedURL(),
		ProviderCredentials: GetClouProviderCredentials(
			map[string]string{
				CONFIG_AWS_REGION: region})}
	return manager.DestroyStack(stack)
}

func DestroyStack(stackname string) error {
	return DestroyStackByRegion(os.Getenv("AWS_DEFAULT_REGION"), stackname)
}

// Create a list of filters for tags based on the tags added by qenvs
func GetTagsAsFilters() (filters []*awsEC2.Filter) {
	filterMap := maps.Convert(qenvsContext.GetTags(),
		func(name string) *string { return aws.String("tag:" + name) },
		func(value string) []*string { return []*string{aws.String(value)} })
	for k, v := range filterMap {
		filter := awsEC2.Filter{
			Name:   k,
			Values: v,
		}
		filters = append(filters, &filter)
	}
	return
}
