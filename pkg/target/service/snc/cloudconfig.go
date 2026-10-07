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
	//   IBM Cloud: Secrets Manager secret name
	SecretStorePullSecretRef        string
	SecretStoreKubeAdminPasswordRef string
	SecretStoreDeveloperPasswordRef string
	// SecretStoreInstanceName is the name of the secret store instance.
	// Used by IBM Cloud cloud-init to discover the SM endpoint via Resource Controller.
	// Empty for AWS (endpoint inferred from instance metadata/role).
	SecretStoreInstanceName string
	// SecretStoreRegion is the cloud region of the secret store instance.
	// Used by IBM Cloud cloud-init to construct the SM endpoint URL after discovery.
	// Empty for AWS.
	SecretStoreRegion string
}

// cloudConfigRenderData extends DataValues with the provider-specific
// secret-fetch commands injected at render time.
type cloudConfigRenderData struct {
	DataValues
	SecretFetchCommands string
}

//go:embed cloud-config
var cloudConfigFile []byte

// ssmFetchSnippet fetches secrets from AWS SSM using the instance's IAM role
// (via podman aws-cli). Each fetch is a separate runcmd item.
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

// smFetchSnippet fetches secrets from IBM Cloud Secrets Manager.
// It runs as a single shell block that:
//  1. Gets an IAM token via VPC Instance Identity (no static credentials needed).
//  2. Polls the Resource Controller until the SM instance is active and returns its endpoint.
//  3. Polls SM for each secret by name (with retry) until the provisioner has stored them.
//
// The instance's trusted profile must grant:
//   - Viewer on the SM resource instance (for Resource Controller discovery)
//   - Reader on Secrets Manager (for secret retrieval)
const smFetchSnippet = `  - |
    set -euo pipefail
    II_TOKEN=$(curl -sf -X PUT "http://169.254.169.254/instance_identity/v1/token?version=2022-03-01" \
      -H "Metadata-Flavor: ibm" -d '{"expires_in":3600}' | jq -r '.access_token')
    IAM_TOKEN=$(curl -sf -X POST "https://iam.cloud.ibm.com/identity/token" \
      -H "Content-Type: application/x-www-form-urlencoded" \
      -d "grant_type=urn:ibm:params:oauth:grant-type:iam-authz&iam_token=Bearer ${II_TOKEN}" \
      | jq -r '.access_token')
    SM_ENDPOINT=""
    until [ -n "${SM_ENDPOINT}" ]; do
      SM_GUID=$(curl -sf -H "Authorization: Bearer ${IAM_TOKEN}" \
        "https://resource-controller.cloud.ibm.com/v2/resource_instances?name={{ .SecretStoreInstanceName }}" \
        | jq -r '.resources[]|select(.state=="active" and (.crn|contains(":secrets-manager:")))|.guid // empty' \
        | head -1)
      if [ -n "${SM_GUID}" ]; then
        SM_ENDPOINT="https://${SM_GUID}.{{ .SecretStoreRegion }}.secrets-manager.appdomain.cloud"
      else
        sleep 30
      fi
    done
    fetch_secret() {
      local name="$1" dest="$2" secret_id
      until secret_id=$(curl -sf -H "Authorization: Bearer ${IAM_TOKEN}" \
          "${SM_ENDPOINT}/api/v2/secrets?search=name:equals:${name}&secret_types=arbitrary" \
          | jq -r '.secrets[0].id // empty') && [ -n "${secret_id}" ]; do sleep 15; done
      curl -sf -H "Authorization: Bearer ${IAM_TOKEN}" \
        "${SM_ENDPOINT}/api/v2/secrets/${secret_id}/versions/current" \
        | jq -r '.secret_data.payload' > "${dest}"
      chmod 0644 "${dest}"
    }
    fetch_secret "{{ .SecretStorePullSecretRef }}" /opt/crc/pull-secret
    fetch_secret "{{ .SecretStoreKubeAdminPasswordRef }}" /opt/crc/pass_kubeadmin
    fetch_secret "{{ .SecretStoreDeveloperPasswordRef }}" /opt/crc/pass_developer
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
// Manager, fetched at boot using Instance Identity) and returns it
// MIME-wrapped for use as VPC instance user data.
func CloudConfigSM(data DataValues) (string, error) {
	cc, err := renderCloudConfig(data, smFetchSnippet)
	if err != nil {
		return "", err
	}
	return mimeWrapCloudConfig(cc), nil
}

// renderCloudConfig renders the provider-specific fetch snippet then injects
// it into the shared cloud-config template.
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
