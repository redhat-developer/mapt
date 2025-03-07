package rhelai

import (
	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/bytequantity"
	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/selector"
	"github.com/redhat-developer/mapt/pkg/provider/util/instancetypes"
)

var (
	stackName          = "stackRHELAIBaremetal"
	awsRHELDedicatedID = "araid"

	diskSize int = 2000

	amiProductDescription = "Red Hat Enterprise Linux"
	amiName               = "rhel-ai-nvidia-aws-1.3.2"
	amiOwnerSelf          = "self"
	amiArch               = "x86_64"
	amiUserDefault        = "ec2-user"

	instanceSpecs = instancetypes.AwsInstanceRequest{
		CPUsRange: &selector.Int32RangeFilter{
			LowerBound: 32,
			UpperBound: 192,
		},
		MemoryRange: &selector.ByteQuantityRangeFilter{
			LowerBound: bytequantity.FromGiB(uint64(192)),
			UpperBound: bytequantity.FromTiB(uint64(4)),
		},
		Arch:            instancetypes.Amd64,
		GPUs:            8,
		GPUManufacturer: "NVIDIA",
		// GPUModel:        "A100",
		// GPUModel: "H100",
		GPUModel: "L40S",
	}

	outputHost           = "ardHost"
	outputUsername       = "ardUsername"
	outputUserPrivateKey = "ardPrivatekey"
)

// NVIDIA A100 X2
// [2] NVIDIA A100 X4
// [3] NVIDIA A100 X8
// [4] NVIDIA H100 X2
// [5] NVIDIA H100 X4
// [6] NVIDIA H100 X8
// [7] NVIDIA L4 X8
// [8] NVIDIA L40S X4
// [9] NVIDIA L40S X8
