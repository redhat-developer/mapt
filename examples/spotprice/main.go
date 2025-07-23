/*
An example tool to find the best provider to launch an
spot instance based on the specifications provided:
cpus, memory, and os

Credentials needs to setup as expected by the cloud provider
SDKs, underneath mapt calls the credential setup helpers
*/

package main

import (
	"fmt"
	"os"

	computerequest "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spot "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot/types"
)

func main() {
	// // Setup AWS credentials; can also be set by exporting the following
	// // variables in the shell
	// os.Setenv("AWS_ACCESS_KEY_ID", "replace_with_aws_access_key_id")
	// os.Setenv("AWS_SECRET_ACCESS_KEY", "replace_with_aws_secret_key")
	// os.Setenv("AWS_DEFAULT_REGION", "ap-south-1")

	// // Setup Azure credentials; can also be set by exporting the following
	// // variables in the shell
	// os.Setenv("ARM_TENANT_ID", "replace_arm_tenant_id")
	// os.Setenv("ARM_SUBSCRIPTION_ID", "replace_with_arm_subscription_id")
	// os.Setenv("ARM_CLIENT_ID", "replace_with_client_id")
	// os.Setenv("ARM_CLIENT_SECRET", "replace_with_client_secret")

	// Get the lowest price for the above spec across
	// all the supported cloud providers

	// By compute paras
	// spi, err := spot.GetLowestPrice(
	// 	&spotTypes.SpotRequestArgs{
	// 		ComputeRequest: &computerequest.ComputeRequestArgs{
	// 			GPUManufacturer: "NVIDIA",
	// 			CPUs:            4,
	// 			MemoryGib:       8,
	// 			Arch:            computerequest.Amd64,
	// 			NestedVirt:      false,
	// 		}}, spot.AWS)
	// if err != nil {
	// 	fmt.Printf("Error: %v", err)
	// 	os.Exit(1)
	// }

	p4ComputeTypes := []string{"p4d.24xlarge", "p4de.24xlarge"}
	p5ComputeTypes := []string{"p5.48xlarge", "p5e.48xlarge", "p5en.48xlarge"}
	g6ComputeTypes := []string{"g6.24xlarge", "g6.48xlarge", "g6e.24xlarge", "g6e.48xlarge"}

	runByComputeTypes(p4ComputeTypes, spot.AWS)
	runByComputeTypes(p5ComputeTypes, spot.AWS)
	runByComputeTypes(g6ComputeTypes, spot.AWS)

	// https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/azure/azure_instance_types.go
	nd96MI300XComputeTypes := []string{"Standard_ND96is_MI300X_v5", "Standard_ND96isr_MI300X_v5"}
	nd96H100ComputeTypes := []string{"Standard_ND96isr_H100_v5", "Standard_NCC40ads_H100_v5", "Standard_NC40ads_H100_v5", "Standard_NC80adis_H100_v5"}
	nc96A100ComputeTypes := []string{"Standard_NC96ads_A100_v4", "Standard_ND96amsr_A100_v4", "Standard_ND96asr_v4"}

	runByComputeTypes(nd96MI300XComputeTypes, spot.Azure)
	runByComputeTypes(nd96H100ComputeTypes, spot.Azure)
	runByComputeTypes(nc96A100ComputeTypes, spot.Azure)

}

func runByComputeTypes(ct []string, p spot.Provider) {
	spi, err := spot.GetLowestPrice(
		&spotTypes.SpotRequestArgs{
			ComputeRequest: &computerequest.ComputeRequestArgs{
				ComputeSizes: ct,
			},
		}, p)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	for _, sr := range spi {
		fmt.Printf("Price: %f, Instance Type: %s, Region: %s, Availability Zone: %s and ChanceLevel %d\n",
			sr.Price,
			sr.ComputeType,
			sr.Region,
			sr.AvailabilityZone,
			sr.ChanceLevel,
		)
	}
}
