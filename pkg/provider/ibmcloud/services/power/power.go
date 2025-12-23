package power

import (
	"context"
	"fmt"
	"os"
	"time"

	v "github.com/IBM-Cloud/power-go-client/clients/instance"
	ps "github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM/go-sdk-core/v5/core"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
)

const powerURLRegex = "%s.power-iaas.cloud.ibm.com"

func powerURL(region string) string { return fmt.Sprintf(powerURLRegex, region) }

type PowerArgs struct {
	InstanceArgs    models.PVMInstanceCreate
	CloudInstanceId string
}

func New(mCtx *mc.Context, args *PowerArgs) (*string, error) {
	pc, err := client(mCtx, args.CloudInstanceId)
	if err != nil {
		return nil, err
	}
	allInstances, err := pc.GetAll()
	if err != nil {
		return nil, err
	}
	for _, in := range allInstances.PvmInstances {
		if *in.ServerName == *args.InstanceArgs.ServerName {
			return in.PvmInstanceID, nil
		}
	}
	createRespOk, err := pc.Create(convertToPVMInstanceCreate(args))
	if err != nil {
		return nil, err
	}
	for _, in := range *createRespOk {
		pInstanceId := *in.PvmInstanceID
		waitForInstance(pc, pInstanceId)
		return &pInstanceId, nil
	}
	return nil, fmt.Errorf("create response is empty")
}

func waitForInstance(pc *v.IBMPIInstanceClient, instanceId string) {
	for i := 0; i < 30; i++ { // retry up to ~5 minutes
		i, err := pc.Get(instanceId)
		if err == nil && i.Health.Status == "warning" {
			fmt.Println("Instance ready")
			break
		}
		fmt.Println("Instance not ready, retrying in 10s...")
		time.Sleep(10 * time.Second)
	}

}

func client(mCtx *mc.Context, cloudInstanceId string) (*v.IBMPIInstanceClient, error) {
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
	return v.NewIBMPIInstanceClient(context.Background(), session, cloudInstanceId), nil
}

func convertToPVMInstanceCreate(s *PowerArgs) *models.PVMInstanceCreate {
	return &models.PVMInstanceCreate{
		ServerName:  s.InstanceArgs.ServerName,
		Memory:      s.InstanceArgs.Memory,
		Processors:  s.InstanceArgs.Processors,
		ProcType:    s.InstanceArgs.ProcType,
		SysType:     s.InstanceArgs.SysType,
		ImageID:     s.InstanceArgs.ImageID,
		KeyPairName: s.InstanceArgs.KeyPairName,
		NetworkIDs:  s.InstanceArgs.NetworkIDs,
	}
}
