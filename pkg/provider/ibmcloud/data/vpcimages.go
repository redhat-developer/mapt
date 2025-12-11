package data

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/IBM/vpc-go-sdk/vpcv1"

	"github.com/IBM/go-sdk-core/v5/core"
)

const (
	VPC_ARCH_X86_64 vpcArch = "amd64"
	VPC_ARCH_IBMZ   vpcArch = "s390x"
)

type vpcArch string

type VPCImageArgs struct {
	Name string
	Arch vpcArch
}

func GetVPCImage(args *VPCImageArgs) (*string, error) {
	return getVPCImage(nil, args)
}

func getVPCImage(next *string, args *VPCImageArgs) (*string, error) {
	vpcService, err := vpcService()
	if err != nil {
		return nil, err
	}
	images, _, err := vpcService.ListImages(
		&vpcv1.ListImagesOptions{
			Start: next,
		})
	if err != nil {
		return nil, err
	}
	idx := slices.IndexFunc(images.Images,
		func(i vpcv1.Image) bool {
			return *i.OperatingSystem.Architecture == string(args.Arch) &&
				strings.Contains(*i.Name, args.Name)
		})
	if idx != -1 {
		return images.Images[idx].ID, nil
	}
	if images.Next != nil {
		next, err := images.GetNextStart()
		if err != nil {
			return nil, err
		}
		return getVPCImage(next, args)
	}
	return nil, fmt.Errorf("no image %s available", args.Name)
}

func vpcService() (*vpcv1.VpcV1, error) {
	serviceURL, err := vpcv1.GetServiceURLForRegion(os.Getenv("IC_REGION"))
	if err != nil {
		return nil, err
	}
	return vpcv1.NewVpcV1(&vpcv1.VpcV1Options{
		Authenticator: &core.IamAuthenticator{
			ApiKey: os.Getenv("IBMCLOUD_API_KEY"),
		},
		URL: serviceURL,
	})
}
