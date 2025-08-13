package spot

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	spot "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
)

type bestSpotOption struct {
	pulumi.ResourceState
	Option *spot.SpotResults
}

func NewBestSpotOption(ctx *pulumi.Context, mCtx *mc.Context, name string,
	args *data.SpotInfoArgs, opts ...pulumi.ResourceOption) (*spot.SpotResults, error) {
	spotOption, err := data.SpotInfo(mCtx, args)
	if err != nil {
		return nil, err
	}
	err = ctx.RegisterComponentResource("rh:qe:aws:bso",
		name,
		&bestSpotOption{
			Option: spotOption,
		},
		opts...)
	if err != nil {
		return nil, err
	}
	return spotOption, nil
}
