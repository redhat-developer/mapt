package credentials

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/redhat-developer/mapt/pkg/util/logging"
)

const (
	metadataBaseURL              = "http://169.254.170.2"
	ecsCredentialsRelativeURIENV = "AWS_CONTAINER_CREDENTIALS_RELATIVE_URI"
	defaultAWSRegion             = "us-east-1"
)

// https://docs.aws.amazon.com/sdkref/latest/guide/feature-container-credentials.html
func SetCredentialsFromContainerRole() error {
	relativeURI := os.Getenv(ecsCredentialsRelativeURIENV)
	if relativeURI == "" {
		return fmt.Errorf("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI not set. Are you running in an ECS container?")
	}
	resp, err := http.Get(fmt.Sprintf("%s/%s", metadataBaseURL, relativeURI))
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
	if err := os.Setenv("AWS_DEFAULT_REGION", defaultAWSRegion); err != nil {
		return err
	}
	if err := os.Setenv("AWS_REGION", defaultAWSRegion); err != nil {
		return err
	}
	return nil
}
