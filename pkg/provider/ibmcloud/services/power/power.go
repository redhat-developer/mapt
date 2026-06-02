package power

import (
	"fmt"
	"os"
	"time"

	v "github.com/IBM-Cloud/power-go-client/clients/instance"
	ps "github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM/go-sdk-core/v5/core"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	icConstants "github.com/redhat-developer/mapt/pkg/provider/ibmcloud/constants"
	"github.com/redhat-developer/mapt/pkg/util/logging"
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
	if len(*createRespOk) == 0 {
		return nil, fmt.Errorf("create response is empty")
	}
	pInstanceId := *(*createRespOk)[0].PvmInstanceID
	if err := waitForInstance(mCtx, pc, pInstanceId); err != nil {
		return nil, err
	}
	return &pInstanceId, nil
}

func waitForInstance(mCtx *mc.Context, pc *v.IBMPIInstanceClient, instanceId string) error {
	for i := 0; i < 30; i++ { // retry up to ~5 minutes
		inst, err := pc.Get(instanceId)
		if err == nil && inst.Health.Status == "WARNING" {
			logging.Infof("instance %s is ready", instanceId)
			return nil
		}
		logging.Infof("instance %s not ready, retrying in 10s...", instanceId)
		select {
		case <-mCtx.Context().Done():
			return mCtx.Context().Err()
		case <-time.After(10 * time.Second):
		}
	}
	return fmt.Errorf("timed out waiting for instance %s to become ready", instanceId)
}

func client(mCtx *mc.Context, cloudInstanceId string) (*v.IBMPIInstanceClient, error) {
	options := &ps.IBMPIOptions{
		Authenticator: &core.IamAuthenticator{
			ApiKey: os.Getenv(icConstants.EnvIBMCloudAPIKey),
		},
		UserAccount: os.Getenv(icConstants.EnvIBMCloudAccount),
		Zone:        os.Getenv("IC_ZONE"),
		URL:         powerURL(os.Getenv("IC_REGION")),
		Debug:       mCtx.Debug(),
	}
	session, err := ps.NewIBMPISession(options)
	if err != nil {
		return nil, err
	}
	return v.NewIBMPIInstanceClient(mCtx.Context(), session, cloudInstanceId), nil
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
