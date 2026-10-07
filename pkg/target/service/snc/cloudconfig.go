package snc

import (
	_ "embed"
	"encoding/base64"
	"strings"

	"github.com/redhat-developer/mapt/pkg/util/file"
)

type DataValues struct {
	Username string
	PubKey   string
	PublicIP string
	// Secret store references — meaning is provider-specific:
	//   AWS: SSM parameter name
	//   IBM Cloud: Secrets Manager secret ID (UUID)
	SecretStorePullSecretRef        string
	SecretStoreKubeAdminPasswordRef string
	SecretStoreDeveloperPasswordRef string
	// SecretStoreEndpointURL is the base URL of the secret store instance.
	// Empty for AWS (endpoint is inferred from the instance IAM role).
	// Required for IBM Cloud Secrets Manager.
	SecretStoreEndpointURL string
}

// cloudConfigRenderData extends DataValues with the provider-specific
// secret-fetch commands injected at render time.
type cloudConfigRenderData struct {
	DataValues
	SecretFetchCommands string
}

//go:embed cloud-config
var cloudConfigFile []byte

// ssmFetchSnippet is the runcmd snippet that fetches secrets from AWS SSM
// using the instance's IAM role (via podman aws-cli).
const ssmFetchSnippet = `  - export PS=$(podman run --rm docker.io/amazon/aws-cli ssm get-parameter --name "{{ .SecretStorePullSecretRef }}" --with-decryption --query "Parameter.Value" --output text)
  - echo ${PS} > /opt/crc/pull-secret
  - chmod 0644 /opt/crc/pull-secret
  - export KP=$(podman run --rm docker.io/amazon/aws-cli ssm get-parameter --name "{{ .SecretStoreKubeAdminPasswordRef }}" --with-decryption --query "Parameter.Value" --output text)
  - echo ${KP} > /opt/crc/pass_kubeadmin
  - chmod 0644 /opt/crc/pass_kubeadmin
  - export DV=$(podman run --rm docker.io/amazon/aws-cli ssm get-parameter --name "{{ .SecretStoreDeveloperPasswordRef }}" --with-decryption --query "Parameter.Value" --output text)
  - echo ${DV} > /opt/crc/pass_developer
  - chmod 0644 /opt/crc/pass_developer
`

// smFetchSnippet is the runcmd snippet that fetches secrets from IBM Cloud
// Secrets Manager using the VPC Instance Identity token (no embedded credentials).
const smFetchSnippet = `  # Obtain an IAM token via the VPC Instance Identity service — no static credentials needed.
  # The instance's trusted profile must have Secrets Manager Reader access.
  - export II_TOKEN=$(curl -sf -X PUT "http://169.254.169.254/instance_identity/v1/token?version=2022-03-01" -H "Metadata-Flavor: ibm" -d '{"expires_in":3600}' | jq -r '.access_token')
  - export IAM_TOKEN=$(curl -sf -X POST "https://iam.cloud.ibm.com/identity/token" -H "Content-Type: application/x-www-form-urlencoded" -d "grant_type=urn:ibm:params:oauth:grant-type:iam-authz&iam_token=Bearer ${II_TOKEN}" | jq -r '.access_token')
  - export PS=$(curl -sf -H "Authorization: Bearer ${IAM_TOKEN}" "{{ .SecretStoreEndpointURL }}/api/v2/secrets/{{ .SecretStorePullSecretRef }}/versions/current" | jq -r '.secret_data.payload')
  - echo "${PS}" > /opt/crc/pull-secret
  - chmod 0644 /opt/crc/pull-secret
  - export KP=$(curl -sf -H "Authorization: Bearer ${IAM_TOKEN}" "{{ .SecretStoreEndpointURL }}/api/v2/secrets/{{ .SecretStoreKubeAdminPasswordRef }}/versions/current" | jq -r '.secret_data.payload')
  - echo "${KP}" > /opt/crc/pass_kubeadmin
  - chmod 0644 /opt/crc/pass_kubeadmin
  - export DV=$(curl -sf -H "Authorization: Bearer ${IAM_TOKEN}" "{{ .SecretStoreEndpointURL }}/api/v2/secrets/{{ .SecretStoreDeveloperPasswordRef }}/versions/current" | jq -r '.secret_data.payload')
  - echo "${DV}" > /opt/crc/pass_developer
  - chmod 0644 /opt/crc/pass_developer
`

// CloudConfig renders the cloud-config for AWS (secrets via SSM) and returns
// it base64-encoded for use as EC2 user data.
func CloudConfig(data DataValues) (*string, error) {
	cc, err := renderCloudConfig(data, ssmFetchSnippet)
	if err != nil {
		return nil, err
	}
	ccB64 := base64.StdEncoding.EncodeToString([]byte(cc))
	return &ccB64, nil
}

// CloudConfigSM renders the cloud-config for IBM Cloud (secrets via Secrets
// Manager using Instance Identity) and returns it MIME-wrapped for VPC user data.
func CloudConfigSM(data DataValues) (string, error) {
	cc, err := renderCloudConfig(data, smFetchSnippet)
	if err != nil {
		return "", err
	}
	return mimeWrapCloudConfig(cc), nil
}

// renderCloudConfig renders the provider-specific fetch snippet then injects it
// into the shared cloud-config template.
func renderCloudConfig(data DataValues, snippetTmpl string) (string, error) {
	snippet, err := file.Template(data, snippetTmpl)
	if err != nil {
		return "", err
	}
	return file.Template(cloudConfigRenderData{DataValues: data, SecretFetchCommands: snippet},
		string(cloudConfigFile))
}

func mimeWrapCloudConfig(rawCC string) string {
	const boundary = "MAPT-CLOUD-CONFIG"
	encoded := base64.StdEncoding.EncodeToString([]byte(rawCC))
	return strings.Join([]string{
		"MIME-Version: 1.0",
		`Content-Type: multipart/mixed; boundary="` + boundary + `"`,
		"",
		"--" + boundary,
		`Content-Type: text/cloud-config; charset="us-ascii"`,
		"Content-Transfer-Encoding: base64",
		"",
		encoded,
		"--" + boundary + "--",
		"",
	}, "\n")
}
