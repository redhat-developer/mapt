package data

import (
	"fmt"
	"strings"

	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot/types"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

type SpotSelector struct{}

func NewSpotSelector() *SpotSelector { return &SpotSelector{} }

func (c *SpotSelector) Select(
	args *spotTypes.SpotRequestArgs) (*spotTypes.SpotResults, error) {
	return lowestPrice(args)
}

func lowestPrice(args *spotTypes.SpotRequestArgs) (*spotTypes.SpotResults, error) {
	var err error
	vms := args.ComputeRequest.ComputeSizes
	if len(vms) == 0 {
		vms, err =
			NewComputeSelector().Select(args.ComputeRequest)
		if err != nil {
			return nil, err
		}
	}
	logging.Debugf("vms: %s", strings.Join(vms, ","))
	return nil, fmt.Errorf("not implemented yet")
}
