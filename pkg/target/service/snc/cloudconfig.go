package snc

import (
	_ "embed"
	"encoding/base64"

	"github.com/redhat-developer/mapt/pkg/util/file"
)

type DataValues struct {
	// user auth information
	Username string
	PubKey   string
	// IP
	PublicIP string
	// Secret information will be moved through ssm
	SSMPullSecretName        string
	SSMKubeAdminPasswordName string
	SSMDeveloperPasswordName string
}

//go:embed cloud-config
var CloudConfigFile []byte

func CloudConfig(data DataValues) (*string, error) {
	templateConfig := string(CloudConfigFile[:])
	cc, err := file.Template(data, templateConfig)
	if err != nil {
		return nil, err
	}
	ccB64 := base64.StdEncoding.EncodeToString([]byte(cc))
	return &ccB64, nil
}
