package data

import (
	"fmt"
	"os"
	"slices"

	v "github.com/IBM-Cloud/power-go-client/clients/instance"
	ps "github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM-Cloud/power-go-client/power/models"

	"github.com/IBM/go-sdk-core/v5/core"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
)

const powerURLRegex = "%s.power-iaas.cloud.ibm.com"

func powerURL(region string) string { return fmt.Sprintf(powerURLRegex, region) }

type PiImageArgs struct {
	CloudInstanceId string
	Name            string
}

func GetImage(mCtx *mc.Context, args *PiImageArgs) (*string, error) {
	pc, err := piImagesClient(mCtx, args.CloudInstanceId)
	if err != nil {
		return nil, err
	}
	sis, err := pc.GetAllStockImages(false, false)
	if err != nil {
		return nil, err
	}
	idx := slices.IndexFunc(sis.Images,
		func(si *models.ImageReference) bool {
			return *si.Name == args.Name
		})
	if idx != -1 {
		return sis.Images[idx].ImageID, nil
	}
	return nil, fmt.Errorf("no stock image %s available", args.Name)

}

func piImagesClient(mCtx *mc.Context, cloudInstanceId string) (*v.IBMPIImageClient, error) {
	options := &ps.IBMPIOptions{
		Authenticator: &core.IamAuthenticator{
			ApiKey: os.Getenv("IBMCLOUD_API_KEY"),
		},
		UserAccount: os.Getenv("IBMCLOUD_ACCOUNT"),
		Zone:        os.Getenv("IC_ZONE"),
		URL:         powerURL(os.Getenv("IC_REGION")),
		Debug:       mCtx.Debug(),
	}
	session, err := ps.NewIBMPISession(options)
	if err != nil {
		return nil, err
	}
	return v.NewIBMPIImageClient(mCtx.Context(), session, cloudInstanceId), nil
}
