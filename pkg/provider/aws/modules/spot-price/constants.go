package spotprice

const (
	StackName string = "Spot-Price"

	StackOutputRegion           string = "Region"
	StackOutputAvailabilityZone string = "AvailabilityZone"
	StackOutputAVGPrice         string = "AVGPrice"
	StackOutputMaxPrice         string = "MaxPrice"
	StackOutputScore            string = "Score"

	maxSpotPlacementScoreResults int64 = 10

	// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSpotPriceHistory.html
	spotQueryFilterProductDescription string = "product-description"

	spsMaxScore = 10

	pulumiType string = "rh:qe:aws:bsb"
)
