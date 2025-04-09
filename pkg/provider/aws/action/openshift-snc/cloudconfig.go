package openshiftsnc

import (
	_ "embed"
	"encoding/base64"

	"github.com/redhat-developer/mapt/pkg/util/file"
)

type dataValues struct {
	// user auth information
	Username string
	PubKey   string
	// IP
	PublicIP string
	// Secret information will be moved through ssm
	SSMPullSecretName        string
	SSMCaCertName            string
	SSMKubeAdminPasswordName string
	SSMDeveloperPasswordName string
}

//go:embed cloud-config
var CloudConfig []byte

var cloudConfigRequiredProfiles = []string{"arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"}

func cloudConfig(data dataValues) (*string, error) {
	templateConfig := string(CloudConfig[:])
	cc, err := file.Template(data, templateConfig)
	if err != nil {
		return nil, err
	}
	ccB64 := base64.StdEncoding.EncodeToString([]byte(cc))
	return &ccB64, nil
}
