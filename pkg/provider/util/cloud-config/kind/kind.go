package kind

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"

	"github.com/redhat-developer/mapt/pkg/util/file"
)

type kindArch string

var (
	X86_64 kindArch = "amd64"
	Arm64  kindArch = "arm64"
)

type PortMapping struct {
	ContainerPort int    `json:"containerPort"`
	HostPort      int    `json:"hostPort"`
	Protocol      string `json:"protocol"`
}

type DataValues struct {
	Arch              kindArch
	KindVersion       string
	KindImage         string
	PublicIP          string
	Username          string
	ExtraPortMappings []PortMapping
}

//go:embed cloud-config
var CloudConfigTemplate []byte

func CloudConfig(data *DataValues) (*string, error) {
	templateConfig := string(CloudConfigTemplate[:])
	userdata, err := file.Template(data, templateConfig)
	ccB64 := base64.StdEncoding.EncodeToString([]byte(userdata))
	return &ccB64, err
}

func ParseExtraPortMappings(extraPortMappingsJSON string) ([]PortMapping, error) {
	if extraPortMappingsJSON == "" {
		return []PortMapping{}, nil
	}

	var portMappings []PortMapping
	if err := json.Unmarshal([]byte(extraPortMappingsJSON), &portMappings); err != nil {
		return nil, err
	}

	return portMappings, nil
}
