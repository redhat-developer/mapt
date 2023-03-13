package spotprice

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

type BestSpotBid struct {
	pulumi.ResourceState
	Price *SpotPriceGroup
}

func BestSpotBidOrder(ctx *pulumi.Context, name string, targetHostID string, opts ...pulumi.ResourceOption) (*BestSpotBid, error) {
	price, err := BestSpotPriceInfo(targetHostID)
	if err != nil {
		return nil, err
	}
	bsb := &BestSpotBid{
		Price: price,
	}
	err = ctx.RegisterComponentResource(pulumiType, name, bsb, opts...)
	if err != nil {
		return nil, err
	}
	return bsb, nil
}
