package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsEC2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/redhat-developer/mapt/pkg/manager"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/manager/credentials"
	awsConstants "github.com/redhat-developer/mapt/pkg/provider/aws/constants"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	"github.com/redhat-developer/mapt/pkg/util/maps"
)

type AWS struct{}

func (a *AWS) Init(backedURL string) error {
	// Manage remote state requirements, if backedURL
	// is on a different region we need to change to that region
	// in order to interact with the state
	return manageRemoteState(backedURL)
}

func Provider() *AWS {
	return &AWS{}
}

// Under some circumstances it is possible we need to update Location initial configuration
// due to usage of remote backed url. i.e. https://github.com/redhat-developer/mapt/issues/392

// This function will check if backed url is remote and if so change initial values to be able to
// use it.
func manageRemoteState(backedURL string) error {
	if data.ValidateS3Path(backedURL) {
		awsRegion, err := data.GetBucketLocationFromS3Path(backedURL)
		if err != nil {
			return err
		}
		if err := os.Setenv("AWS_DEFAULT_REGION", *awsRegion); err != nil {
			return err
		}
		if err := os.Setenv("AWS_REGION", *awsRegion); err != nil {
			return err
		}
		return nil
	}
	return nil
}

// pulumi config key : aws env credential
var envCredentials = map[string]string{
	awsConstants.CONFIG_AWS_REGION:        "AWS_DEFAULT_REGION",
	awsConstants.CONFIG_AWS_NATIVE_REGION: "AWS_DEFAULT_REGION",
	awsConstants.CONFIG_AWS_ACCESS_KEY:    "AWS_ACCESS_KEY_ID",
	awsConstants.CONFIG_AWS_SECRET_KEY:    "AWS_SECRET_ACCESS_KEY"}

var DefaultCredentials = GetClouProviderCredentials(nil)

func GetClouProviderCredentials(customCredentials map[string]string) credentials.ProviderCredentials {
	return credentials.ProviderCredentials{
		SetCredentialFunc: SetAWSCredentials,
		FixedCredentials:  customCredentials}
}

func SetAWSCredentials(ctx context.Context, stack auto.Stack, customCredentials map[string]string) error {
	if maptContext.IsServerless() {
		if err := setCredentialsForServerless(); err != nil {
			return err
		}
	}
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
	logging.Debug("Running destroy operation")
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
				awsConstants.CONFIG_AWS_REGION: util.If(len(s.Region) > 0,
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

// https://docs.aws.amazon.com/sdkref/latest/guide/feature-container-credentials.html
// When running as serverless Credendials should be retrieved based on the role
// for the serverless task being executed as so we need to get them and set as Envs
// to continue with default behavior
func setCredentialsForServerless() error {
	relativeURI := os.Getenv(awsConstants.ECSCredentialsRelativeURIENV)
	if relativeURI == "" {
		return fmt.Errorf("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI not set. Are you running in an ECS container?")
	}
	resp, err := http.Get(fmt.Sprintf("%s/%s", awsConstants.MetadataBaseURL, relativeURI))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch credentials, status code: %d", resp.StatusCode)
	}
	// Parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to fetch credentials, status code: %d", resp.StatusCode)

	}
	var credentials struct {
		AccessKeyID     string `json:"AccessKeyId"`
		SecretAccessKey string `json:"SecretAccessKey"`
		SessionToken    string `json:"Token"`
		Expiration      string `json:"Expiration"`
	}
	err = json.Unmarshal(body, &credentials)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		os.Exit(1)
	}
	logging.Debug("We are runnging on serverless mode so we will set the ephemeral Envs for it")
	if err := os.Setenv("AWS_ACCESS_KEY_ID", credentials.AccessKeyID); err != nil {
		return err
	}
	if err := os.Setenv("AWS_SECRET_ACCESS_KEY", credentials.SecretAccessKey); err != nil {
		return err
	}
	if err := os.Setenv("AWS_SESSION_TOKEN", credentials.SessionToken); err != nil {
		return err
	}
	if err := os.Setenv("AWS_DEFAULT_REGION", awsConstants.DefaultAWSRegion); err != nil {
		return err
	}
	if err := os.Setenv("AWS_REGION", awsConstants.DefaultAWSRegion); err != nil {
		return err
	}
	return nil
}
