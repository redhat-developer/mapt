package spotprice

const (
	StackGetSpotPriceName                   string = "Get-SpotPrice"
	StackGetSpotPriceOutputSpotPrice        string = "spotPrice"
	StackGetSpotPriceOutputAvailabilityZone string = "availabilityZone"
)

// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSpotPriceHistory.html
var spotQueryFilterProductDescription string = "product-description"
