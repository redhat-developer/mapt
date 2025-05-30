package compute

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

var ErrNoCapacity error = fmt.Errorf("no capacity: dedicated host had not been allocated")

type ReplaceRootVolumeRequest struct {
	Region     string
	InstanceID string
	AMIID      string
	Wait       bool
}

// This function will replace the root volume for a running ec2 instance
// and will delete the replaced volume
// If wait flag is true on request the funcion will wait until the replace task succeed
// otherwise it will trigger the task and return the id to reference it
func ReplaceRootVolume(r ReplaceRootVolumeRequest) (*string, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(r.Region))
	if err != nil {
		return nil, err
	}
	client := ec2.NewFromConfig(cfg)
	deleteReplacedRootVolume := true
	// rrvt, err :=
	rrvt, err := client.CreateReplaceRootVolumeTask(
		ctx,
		&ec2.CreateReplaceRootVolumeTaskInput{
			InstanceId:               &r.InstanceID,
			DeleteReplacedRootVolume: &deleteReplacedRootVolume,
			ImageId:                  &r.AMIID,
		})
	if err != nil {
		return rrvt.ReplaceRootVolumeTask.ReplaceRootVolumeTaskId, err
	}
	taskState := rrvt.ReplaceRootVolumeTask.TaskState
	for r.Wait && taskState != types.ReplaceRootVolumeTaskStateSucceeded {
		drvt, err := client.DescribeReplaceRootVolumeTasks(ctx, &ec2.DescribeReplaceRootVolumeTasksInput{
			ReplaceRootVolumeTaskIds: []string{*rrvt.ReplaceRootVolumeTask.ReplaceRootVolumeTaskId},
		})
		if err != nil {
			return rrvt.ReplaceRootVolumeTask.ReplaceRootVolumeTaskId, nil
		}
		if len(drvt.ReplaceRootVolumeTasks) == 0 {
			return rrvt.ReplaceRootVolumeTask.ReplaceRootVolumeTaskId,
				fmt.Errorf("something wrong happened checkding the replace root volume task with id %s",
					*rrvt.ReplaceRootVolumeTask.ReplaceRootVolumeTaskId)
		}
		taskState = drvt.ReplaceRootVolumeTasks[0].TaskState
	}
	return rrvt.ReplaceRootVolumeTask.ReplaceRootVolumeTaskId, nil
}

// Create a dedicated host
func DedicatedHost(region, azId, instanceType *string, tags map[string]string) (*string, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(*region))
	if err != nil {
		return nil, err
	}
	client := ec2.NewFromConfig(cfg)
	logging.Debugf("Trying to get a dedicated host %s in %s", *instanceType, *azId)
	aResp, err := client.AllocateHosts(
		context.Background(),
		&ec2.AllocateHostsInput{
			AvailabilityZone: aws.String(*azId),
			InstanceType:     aws.String(*instanceType),
			Quantity:         aws.Int32(1),
		})
	if err != nil {
		return nil, err
	}
	if len(aResp.HostIds) == 0 {
		return nil, ErrNoCapacity
	}
	hostID := aResp.HostIds[0]
	if tags != nil {
		var t []types.Tag
		for k, v := range tags {
			t = append(t,
				types.Tag{
					Key:   aws.String(k),
					Value: aws.String(v)})
		}
		_, err = client.CreateTags(ctx,
			&ec2.CreateTagsInput{
				Resources: []string{hostID},
				Tags:      t,
			})
	}
	return &hostID, err
}

func DedicatedHostRelease(region, hostId *string) error {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(*region))
	if err != nil {
		return err
	}
	client := ec2.NewFromConfig(cfg)
	logging.Debugf("destroying dedicated host: %s", *hostId)
	output, err := client.ReleaseHosts(ctx,
		&ec2.ReleaseHostsInput{
			HostIds: []string{*hostId},
		})
	if err != nil {
		return err
	}
	if len(output.Unsuccessful) == 1 {
		return fmt.Errorf("unsuccessful releases: %v", output.Unsuccessful)
	}
	logging.Debugf("dedicated host: %s has been destroyed successfully", *hostId)
	return nil
}
