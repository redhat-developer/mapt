package data

import (
	"fmt"
	"os"
	"slices"
	"strings"

	v "github.com/IBM-Cloud/power-go-client/clients/instance"
	ps "github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM/go-sdk-core/v5/core"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	icConstants "github.com/redhat-developer/mapt/pkg/provider/ibmcloud/constants"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

var DefaultSystemTypePriority = []string{"e1080", "e1050", "s1122", "s1022", "e980", "s922"}

type SystemTypeRequirements struct {
	CloudInstanceId string
	Zone            string
	ProcType        string
	PreferredType   string
}

type SystemTypeResult struct {
	Types []string
}

func GetAvailableSystemTypes(mCtx *mc.Context, args *SystemTypeRequirements) (*SystemTypeResult, error) {
	client, err := piDatacentersClient(mCtx, args.CloudInstanceId)
	if err != nil {
		return nil, fmt.Errorf("failed to create datacenters client: %w", err)
	}
	dc, err := client.Get(args.Zone)
	if err != nil {
		return nil, fmt.Errorf("failed to query datacenter %s: %w", args.Zone, err)
	}

	var general, dedicated []string
	if dc.CapabilitiesDetails != nil && dc.CapabilitiesDetails.SupportedSystems != nil {
		general = dc.CapabilitiesDetails.SupportedSystems.General
		dedicated = dc.CapabilitiesDetails.SupportedSystems.Dedicated
	}

	types, err := filterAndPrioritize(args.PreferredType, args.ProcType, general, dedicated)
	if err != nil {
		return nil, err
	}

	logging.Infof("zone %s supported system types (general=%v, dedicated=%v)", args.Zone, general, dedicated)
	logging.Infof("system types to attempt (priority order): %v", types)

	return &SystemTypeResult{Types: types}, nil
}

func filterAndPrioritize(preferred, procType string, general, dedicated []string) ([]string, error) {
	var supported []string
	if strings.EqualFold(procType, "dedicated") {
		supported = dedicated
	} else {
		supported = general
	}

	priority := buildPriorityList(preferred)

	var filtered []string
	for _, t := range priority {
		if slices.Contains(supported, t) {
			filtered = append(filtered, t)
		}
	}

	for _, t := range supported {
		if !slices.Contains(filtered, t) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf(
			"no system types from priority list %v are supported in zone (supported: %v)",
			priority, supported)
	}

	return filtered, nil
}

func buildPriorityList(preferred string) []string {
	if preferred == "" {
		return DefaultSystemTypePriority
	}

	result := []string{preferred}
	for _, t := range DefaultSystemTypePriority {
		if t != preferred {
			result = append(result, t)
		}
	}
	return result
}

func piDatacentersClient(mCtx *mc.Context, cloudInstanceId string) (*v.IBMPIDatacentersClient, error) {
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
	return v.NewIBMPIDatacenterClient(mCtx.Context(), session, cloudInstanceId), nil
}
