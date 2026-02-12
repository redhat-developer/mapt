package kind

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

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

type CloudConfigArgs struct {
	Arch              kindArch
	KindVersion       string
	KindImage         string
	PublicIP          string
	Username          string
	ExtraPortMappings []PortMapping
}

//go:embed cloud-config
var CloudConfigTemplate []byte

func (args *CloudConfigArgs) CloudConfig() (*string, error) {
	templateConfig := string(CloudConfigTemplate[:])
	userdata, err := file.Template(args, templateConfig)
	if err != nil {
		return nil, err
	}
	ccB64 := base64.StdEncoding.EncodeToString([]byte(userdata))
	return &ccB64, nil
}

func ParseExtraPortMappings(extraPortMappingsJSON string) ([]PortMapping, error) {
	if extraPortMappingsJSON == "" {
		return []PortMapping{}, nil
	}

	var portMappings []PortMapping
	if err := json.Unmarshal([]byte(extraPortMappingsJSON), &portMappings); err != nil {
		return nil, err
	}

	// Validate and normalize each port mapping
	for i := range portMappings {
		if err := validatePortMapping(&portMappings[i], i); err != nil {
			return nil, err
		}
	}

	return portMappings, nil
}

func validatePortMapping(pm *PortMapping, index int) error {
	if pm.ContainerPort <= 0 {
		return fmt.Errorf("port mapping %d: containerPort must be greater than 0, got %d", index, pm.ContainerPort)
	}

	if pm.HostPort <= 0 {
		return fmt.Errorf("port mapping %d: hostPort must be greater than 0, got %d", index, pm.HostPort)
	}

	if pm.Protocol == "" {
		return fmt.Errorf("port mapping %d: protocol is required", index)
	}

	// Validate protocol (case-insensitive)
	protocol := strings.ToUpper(pm.Protocol)
	if protocol != "TCP" && protocol != "UDP" {
		return fmt.Errorf("port mapping %d: protocol must be 'TCP' or 'UDP', got '%s'", index, pm.Protocol)
	}

	// Normalize protocol to uppercase
	pm.Protocol = protocol

	return nil
}
