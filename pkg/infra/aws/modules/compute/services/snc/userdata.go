package snc

import (
	_ "embed"
	"fmt"
	"io/ioutil"

	"github.com/adrianriobo/qenvs/pkg/util/logging"
)

//go:embed userdata.tpl
var userdataTemplate []byte

type userDataValues struct {
	SubscriptionUsername string
	SubscriptionPassword string
}

func init() {
	err := ioutil.WriteFile("userdata.tpl", userdataTemplate, 0644)
	if err != nil {
		logging.Errorf("error loading snc userdata.tpl: %v", err)
	}
}

func getUserdataTemplate() ([]byte, error) {
	if userdataTemplate != nil {
		return userdataTemplate, nil
	}
	return nil, fmt.Errorf("error loading snc userdata.tpl")
}
