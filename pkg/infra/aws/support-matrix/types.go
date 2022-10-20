package supportmatrix

import "github.com/aws/aws-sdk-go/service/ec2"

type SupportedHost struct {
	ID          string
	Description string
	// Filter for machine
	ProductDescription string
	// Manually check the map on https://ap-south-1.console.aws.amazon.com/ec2/home?region=ap-south-1#SpotPlacementScore:
	//between instances types and Requirements
	InstaceTypes []string
	Requirements *ec2.InstanceRequirementsWithMetadataRequest
	// Requirements Requirements
	// true if spot instances are supported
	Spot bool
}

// type Requirements struct {
// 	RAM          int64
// 	Architecture string
// 	Baremetal    string
// }
