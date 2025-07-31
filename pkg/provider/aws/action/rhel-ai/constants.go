package rhelai

import "fmt"

var (
	stackName          = "stackRHELAIBaremetal"
	awsRHELDedicatedID = "araid"

	diskSize int = 2000

	amiProduct     = "Red Hat Enterprise Linux"
	amiRegex       = "rhel-ai-nvidia-aws-%s*"
	amiOwner       = "391597328979"
	amiOwnerSelf   = "self"
	amiArch        = "x86_64"
	amiUserDefault = "cloud-user"

	// p4ComputeTypes = []string{"p4d.24xlarge", "p4de.24xlarge"}
	// p5ComputeTypes = []string{"p5.48xlarge", "p5e.48xlarge", "p5en.48xlarge"}
	// g6ComputeTypes = []string{"g6.24xlarge", "g6.48xlarge", "g6e.24xlarge", "g6e.48xlarge"}

	// instanceSpecs = instancetypes.AwsInstanceRequest{
	// 	CPUsRange: &selector.Int32RangeFilter{
	// 		LowerBound: 32,
	// 		UpperBound: 192,
	// 	},
	// 	MemoryRange: &selector.ByteQuantityRangeFilter{
	// 		LowerBound: bytequantity.FromGiB(uint64(192)),
	// 		UpperBound: bytequantity.FromTiB(uint64(4)),
	// 	},
	// 	Arch:            instancetypes.Amd64,
	// 	GPUs:            8,
	// 	GPUManufacturer: "NVIDIA",
	// 	// GPUModel:        "A100",
	// 	// GPUModel: "H100",
	// 	GPUModel: "L40S",
	// }

	outputHost           = "ardHost"
	outputUsername       = "ardUsername"
	outputUserPrivateKey = "ardPrivatekey"
)

func amiName(version *string) string { return fmt.Sprintf(amiRegex, *version) }

// NVIDIA A100 X2
// [2] NVIDIA A100 X4
// [3] NVIDIA A100 X8
// [4] NVIDIA H100 X2
// [5] NVIDIA H100 X4
// [6] NVIDIA H100 X8
// [7] NVIDIA L4 X8
// [8] NVIDIA L40S X4
// [9] NVIDIA L40S X8
