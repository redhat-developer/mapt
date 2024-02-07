package compute

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type ReplaceRootVolumeRequest struct {
	Region     string
	InstanceID string
	AMIID      string
}

// This function will replace the root volume for a running ec2 instance
// and will delete the replaced volume
func ReplaceRootVolume(r ReplaceRootVolumeRequest) (*ec2.CreateReplaceRootVolumeTaskOutput, error) {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(r.Region))
	if err != nil {
		return nil, err
	}
	client := ec2.NewFromConfig(cfg)
	deleteReplacedRootVolume := true
	return client.CreateReplaceRootVolumeTask(
		context.Background(),
		&ec2.CreateReplaceRootVolumeTaskInput{
			InstanceId:               &r.InstanceID,
			DeleteReplacedRootVolume: &deleteReplacedRootVolume,
			ImageId:                  &r.AMIID,
		})
}
