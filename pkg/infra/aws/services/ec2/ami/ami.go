package ami

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const awsOwnerID string = "137112412989"

// Looks for the AMI ID on the current Region based on name
// it only allows images from AWS and self
func GetAMIByName(ctx *pulumi.Context, imageName string) (*ec2.LookupAmiResult, error) {
	return ec2.LookupAmi(ctx, &ec2.LookupAmiArgs{
		ExecutableUsers: []string{
			"self",
		},
		Filters: []ec2.GetAmiFilter{
			{
				Name:   "name",
				Values: []string{imageName},
			},
		},
		MostRecent: pulumi.BoolRef(true),
		Owners:     []string{"self", awsOwnerID},
	}, nil)

}
