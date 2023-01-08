package keys

import (
	_ "embed"
	"fmt"
	"io/ioutil"
)

//go:embed ami-0569ce8a44f2351be
var ami0569ce8a44f2351be []byte

func GetKey(amiSourceID string) ([]byte, error) {
	switch amiSourceID {
	case "ami-0569ce8a44f2351be":
		err := ioutil.WriteFile("ami-0569ce8a44f2351be", ami0569ce8a44f2351be, 0644)
		return ami0569ce8a44f2351be, err
	}
	return nil, fmt.Errorf("openspot-ng instance ami not supported")
}
