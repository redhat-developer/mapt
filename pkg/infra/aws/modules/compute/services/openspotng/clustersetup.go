package openspotng

import (
	_ "embed"
	"fmt"
	"io/ioutil"

	"github.com/adrianriobo/qenvs/pkg/util/logging"
)

//go:embed clustersetup.tpl
var clusterSetupTemplate []byte

type clusterSetupValues struct {
	InternalIP        string
	ExternalIP        string
	PullScret         string
	DeveloperPassword string
	KubeadminPassword string
	RedHatPassword    string
}

func init() {
	err := ioutil.WriteFile("clustersetup.tpl", clusterSetupTemplate, 0644)
	if err != nil {
		logging.Errorf("error loading openspot-ng clustersetup: %v", err)
	}
}

func getClusterSetupTemplate() ([]byte, error) {
	if clusterSetupTemplate != nil {
		return clusterSetupTemplate, nil
	}
	return nil, fmt.Errorf("error loading clustersetup.tpl")
}
