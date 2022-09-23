package aws

// TODO check if map fits better
// var region = map[string][]string{
// 	"us-east": {"us-east-1", "us-east-2"},
// 	"us-west": {"us-west-1", "us-west-2"}}

var AvailabilityZones = [...]string{
	"us-east-1",
	"us-east-2",
	"us-west-1",
	"us-west-2",
	"ap-south-1",
	"ap-northeast-1",
	"ap-northeast-2",
	"ap-northeast-3",
	"ap-southeast-1",
	"ap-southeast-2",
	"ca-central-1",
	"eu-central-1",
	"eu-west-1",
	"eu-west-2",
	"eu-west-3",
	"eu-north-1",
	"sa-east-1"}

const (
	// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSpotPriceHistory.html
	spotQueryFilterProductDescription string = "product-description"

	StackSpotOutputSpotPrice        string = "spotPrice"
	StackSpotOutputAvailabilityZone string = "availabilityZone"
)
