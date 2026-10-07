package secretsmanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"
	icConstants "github.com/redhat-developer/mapt/pkg/provider/ibmcloud/constants"
)

const (
	// EndpointURLFormat builds a Secrets Manager instance URL from its GUID and region.
	EndpointURLFormat = "https://%s.%s.secrets-manager.appdomain.cloud"
	// RegionEnv is the environment variable that specifies the IBM Cloud region.
	RegionEnv = "IC_REGION"

	// smCRNServiceName is the service segment present in every Secrets Manager CRN:
	// crn:v1:bluemix:public:secrets-manager:<region>:a/<account>:<guid>::
	smCRNServiceName = ":secrets-manager:"
)

// Client is a thin wrapper around the IBM Cloud Secrets Manager REST API v2.
// Authentication uses IBMCLOUD_API_KEY via IamAuthenticator — the same
// credential already required by the IBM Cloud Pulumi provider.
type Client struct {
	endpointURL string
	auth        core.Authenticator
	http        *http.Client
}

// NewClient creates a Secrets Manager client by auto-discovering the SM instance
// endpoint URL from the Resource Controller using IC_REGION. No extra env vars needed.
func NewClient() (*Client, error) {
	endpoint, err := getInstanceEndpointURL()
	if err != nil {
		return nil, err
	}
	apiKey := os.Getenv(icConstants.EnvIBMCloudAPIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("env var %s is not set", icConstants.EnvIBMCloudAPIKey)
	}
	return &Client{
		endpointURL: endpoint,
		auth:        &core.IamAuthenticator{ApiKey: apiKey},
		http:        &http.Client{},
	}, nil
}

// NewClientWithEndpoint creates a Secrets Manager client with a pre-known endpoint URL.
// Use this when the SM instance GUID is already known (e.g., from a Pulumi output)
// to avoid the Resource Controller discovery call.
func NewClientWithEndpoint(endpointURL string) (*Client, error) {
	apiKey := os.Getenv(icConstants.EnvIBMCloudAPIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("env var %s is not set", icConstants.EnvIBMCloudAPIKey)
	}
	return &Client{
		endpointURL: endpointURL,
		auth:        &core.IamAuthenticator{ApiKey: apiKey},
		http:        &http.Client{},
	}, nil
}

// getInstanceEndpointURL discovers the Secrets Manager instance endpoint for
// the current IC_REGION using the Resource Controller API.
// Errors if no instance or more than one instance is found in the region.
func getInstanceEndpointURL() (string, error) {
	region := os.Getenv(RegionEnv)
	if region == "" {
		return "", fmt.Errorf("env var %s is not set", RegionEnv)
	}
	apiKey := os.Getenv(icConstants.EnvIBMCloudAPIKey)
	if apiKey == "" {
		return "", fmt.Errorf("env var %s is not set", icConstants.EnvIBMCloudAPIKey)
	}

	rc, err := resourcecontrollerv2.NewResourceControllerV2(&resourcecontrollerv2.ResourceControllerV2Options{
		Authenticator: &core.IamAuthenticator{ApiKey: apiKey},
	})
	if err != nil {
		return "", fmt.Errorf("creating resource controller client: %w", err)
	}

	pager, err := rc.NewResourceInstancesPager(&resourcecontrollerv2.ListResourceInstancesOptions{})
	if err != nil {
		return "", fmt.Errorf("creating resource instances pager: %w", err)
	}
	all, err := pager.GetAll()
	if err != nil {
		return "", fmt.Errorf("listing resource instances: %w", err)
	}

	// Filter by CRN service segment and region. The catalog offering ID for
	// Secrets Manager is a UUID that varies by environment, so we match on the
	// CRN which always contains ":secrets-manager:" for SM instances.
	var active []resourcecontrollerv2.ResourceInstance
	for _, inst := range all {
		if inst.State != nil && *inst.State == "active" &&
			inst.RegionID != nil && *inst.RegionID == region &&
			inst.CRN != nil && strings.Contains(*inst.CRN, smCRNServiceName) {
			active = append(active, inst)
		}
	}

	switch len(active) {
	case 0:
		return "", fmt.Errorf("no active Secrets Manager instance found in region %s", region)
	case 1:
		inst := active[0]
		if inst.GUID == nil || inst.RegionID == nil {
			return "", fmt.Errorf("secrets manager instance is missing GUID or RegionID")
		}
		return fmt.Sprintf(EndpointURLFormat, *inst.GUID, *inst.RegionID), nil
	default:
		return "", fmt.Errorf("found %d active Secrets Manager instances in region %s — expected exactly one", len(active), region)
	}
}

type createSecretRequest struct {
	Name       string `json:"name"`
	SecretType string `json:"secret_type"`
	Payload    string `json:"payload"`
}

type secretSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type listSecretsResponse struct {
	TotalCount int             `json:"total_count"`
	Secrets    []secretSummary `json:"secrets"`
}

// EndpointURL returns the Secrets Manager instance endpoint URL.
func (c *Client) EndpointURL() string { return c.endpointURL }

// CreateArbitrarySecret creates an arbitrary secret with the given name and
// plaintext value. Returns the secret ID assigned by Secrets Manager.
func (c *Client) CreateArbitrarySecret(name, value string) (string, error) {
	body, err := json.Marshal(createSecretRequest{
		Name:       name,
		SecretType: "arbitrary",
		Payload:    value,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost,
		c.endpointURL+"/api/v2/secrets", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if err := c.auth.Authenticate(req); err != nil {
		return "", fmt.Errorf("authenticating SM request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create secret %q: HTTP %d: %s", name, resp.StatusCode, b)
	}
	var s secretSummary
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return "", err
	}
	return s.ID, nil
}

// DeleteSecretByName finds the first secret with the given exact name and
// deletes it. It is a no-op when the secret does not exist (idempotent).
func (c *Client) DeleteSecretByName(name string) error {
	id, err := c.findIDByName(name)
	if err != nil {
		return err
	}
	if id == "" {
		return nil
	}
	return c.deleteByID(id)
}

func (c *Client) findIDByName(name string) (string, error) {
	u := fmt.Sprintf("%s/api/v2/secrets?search=%s&secret_types=arbitrary",
		c.endpointURL, url.QueryEscape("name:equals:"+name))
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	if err := c.auth.Authenticate(req); err != nil {
		return "", fmt.Errorf("authenticating SM request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("list secrets: HTTP %d: %s", resp.StatusCode, b)
	}
	var list listSecretsResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return "", err
	}
	for _, s := range list.Secrets {
		if s.Name == name {
			return s.ID, nil
		}
	}
	return "", nil
}

func (c *Client) deleteByID(id string) error {
	req, err := http.NewRequest(http.MethodDelete,
		c.endpointURL+"/api/v2/secrets/"+id, nil)
	if err != nil {
		return err
	}
	if err := c.auth.Authenticate(req); err != nil {
		return fmt.Errorf("authenticating SM request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 && resp.StatusCode != 404 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete secret %s: HTTP %d: %s", id, resp.StatusCode, b)
	}
	return nil
}
