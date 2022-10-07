package spotprice

const (
	// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSpotPriceHistory.html
	spotQueryFilterProductDescription string = "product-description"

	StackGetSpotPriceName                   string = "Get-SpotPrice"
	StackGetSpotPriceOutputSpotPrice        string = "spotPrice"
	StackGetSpotPriceOutputAvailabilityZone string = "availabilityZone"
)
