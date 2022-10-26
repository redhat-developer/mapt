package ami

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const awsOwnerID string = "137112412989"
const redhatOwnerID string = "309956199498"

// Looks for the AMI ID on the current Region based on name
// it only allows images from AWS and self
func GetAMIByName(ctx *pulumi.Context,
	imageName, owner string, filters map[string]string) (*ec2.LookupAmiResult, error) {
	mostRecent := true
	lookupfilters := []ec2.GetAmiFilter{
		{
			Name:   "name",
			Values: []string{imageName},
		},
	}
	for k, v := range filters {
		lookupfilters = append(lookupfilters, ec2.GetAmiFilter{
			Name:   k,
			Values: []string{v},
		})
	}
	owners := []string{awsOwnerID, redhatOwnerID}
	if len(owner) > 0 {
		owners = append(owners, owner)
	}
	return ec2.LookupAmi(ctx, &ec2.LookupAmiArgs{
		Filters:    lookupfilters,
		Owners:     owners,
		MostRecent: &mostRecent,
	})
}
