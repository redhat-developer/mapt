package openshiftsnc

import (
	_ "embed"
	"encoding/base64"

	"github.com/redhat-developer/mapt/pkg/util/file"
)

type CloudConfigDataValues struct {
	// user auth information
	Username string
	PubKey   string
	// IP
	PublicIP string
	// Secret information will be moved through ssm
	SSMPullSecretName        string
	SSMKubeAdminPasswordName string
	SSMDeveloperPasswordName string
	// Unprotected, used for azure
	PullSecret    string
	PassDeveloper string
	PassKubeadmin string
}

var AWSCloudConfigRequiredProfiles = []string{"arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"}

func GenCloudConfig(data CloudConfigDataValues, cloudConfig []byte) (*string, error) {
	templateConfig := string(cloudConfig[:])
	cc, err := file.Template(data, templateConfig)
	if err != nil {
		return nil, err
	}
	ccB64 := base64.StdEncoding.EncodeToString([]byte(cc))
	return &ccB64, nil
}
