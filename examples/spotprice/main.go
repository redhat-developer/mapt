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

	"github.com/redhat-developer/mapt/pkg/spot"
)

func main() {
	// Setup AWS credentials; can also be set by exporting the following
	// variables in the shell
	os.Setenv("AWS_ACCESS_KEY_ID", "replace_with_aws_access_key_id")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "replace_with_aws_secret_key")
	os.Setenv("AWS_DEFAULT_REGION", "ap-south-1")

	// Setup Azure credentials; can also be set by exporting the following
	// variables in the shell
	os.Setenv("ARM_TENANT_ID", "replace_arm_tenant_id")
	os.Setenv("ARM_SUBSCRIPTION_ID", "replace_with_arm_subscription_id")
	os.Setenv("ARM_CLIENT_ID", "replace_with_client_id")
	os.Setenv("ARM_CLIENT_SECRET", "replace_with_client_secret")

	// The SpotRequest struct holds the desired
	// specification in terms of hardware specs
	// and OS requirements
	spotReq := spot.SpotRequest{
		CPUs:       4,
		MemoryGib:  8,
		Arch:       "amd64",
		NestedVirt: false,
		Os:         "fedora",
		OSVersion:  "39",
	}

	// Get the lowest price for only aws
	//spi, err := spotReq.GetAwsLowestPrice()

	// Get the lowest price for only azure
	//spi, err := spotReq.GetAzureLowestPrice()

	// Get the lowest price for the above spec across
	// all the supported cloud providers
	spi, err := spotReq.GetLowestPrice()
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	for provider, price := range spi {
		fmt.Printf("Provider: %s | Price: %f, Instance Type: %s, Region: %s, Availability Zone: %s\n",
			provider,
			price.Price,
			price.InstanceType,
			price.Region,
			price.AvailabilityZone,
		)
	}
}
