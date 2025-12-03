package compute

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

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
func ReplaceRootVolume(ctx context.Context, r ReplaceRootVolumeRequest) (*string, error) {
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
