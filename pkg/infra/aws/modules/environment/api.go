package environment

import (
	"fmt"
	"os"
	"path"

	"github.com/adrianriobo/qenvs/pkg/infra/aws"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute/macm1"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute/rhel"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/network"
	spotprice "github.com/adrianriobo/qenvs/pkg/infra/aws/modules/spot-price"
	supportMatrix "github.com/adrianriobo/qenvs/pkg/infra/aws/support-matrix"
	utilInfra "github.com/adrianriobo/qenvs/pkg/infra/util"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

func Create(projectName, backedURL string, public bool, targetHostID string) (err error) {
	// Check which supported host
	host, err := supportMatrix.GetHost(targetHostID)
	if err != nil {
		return err
	}
	// Check with spot price environment requirements
	availabilityZones, spotPrice, plugin, err :=
		getEnvironmentInfo(projectName, backedURL, host)
	if err != nil {
		return err
	}
	// Based on spot price info the full environment will be created
	request := corporateEnvironmentRequest{
		name: projectName,
		network: &network.NetworkRequest{
			Name:               fmt.Sprintf("%s-%s", projectName, "network"),
			CIDR:               network.DefaultCIDRNetwork,
			AvailabilityZones:  availabilityZones,
			PublicSubnetsCIDRs: network.DefaultCIDRPublicSubnets[:1],
			SingleNatGateway:   false,
		},
	}
	// Add request values for requested host
	manageRequest(&request, host, public, projectName, spotPrice)
	// Create stack
	stack := utilInfra.Stack{
		StackName:   stackCreateEnvironmentName,
		ProjectName: projectName,
		BackedURL:   backedURL,
		Plugin:      *plugin,
		DeployFunc:  request.deployer,
	}
	// Exec stack
	stackResult, err := utilInfra.UpStack(stack)
	if err != nil {
		return err
	}
	// Write host access info to disk
	if err = manageResults(stackResult, host, public, "/tmp/qenvs"); err != nil {
		return err
	}
	logging.Debug("Environment has been created")
	return nil
}

func Destroy(projectName, backedURL string) (err error) {
	stack := utilInfra.Stack{
		StackName:   stackCreateEnvironmentName,
		ProjectName: projectName,
		BackedURL:   backedURL,
		Plugin:      aws.PluginAWSDefault}
	err = utilInfra.DestroyStack(stack)
	if err != nil {
		return
	}
	err = spotprice.Destroy(projectName, backedURL)
	return
}

func getEnvironmentInfo(projectName, backedURL string,
	host *supportMatrix.SupportedHost) ([]string, string, *utilInfra.PluginInfo, error) {
	var availabilityZones = network.DefaultAvailabilityZones[:1]
	var spotPrice string
	var plugin = aws.PluginAWSDefault
	if host.Spot {
		spg, err := spotprice.Create(projectName, backedURL, host.ID)
		if err != nil {
			return nil, "", nil, err
		}
		availabilityZones = []string{spg.AvailabilityZone}
		spotPrice = fmt.Sprintf("%f", spg.MaxPrice)
		// plugin will use the region from the best spot price
		plugin = aws.GetPluginAWS(
			map[string]string{
				aws.CONFIG_AWS_REGION: spg.Region})
	}
	return availabilityZones, spotPrice, &plugin, nil
}

func manageRequest(request *corporateEnvironmentRequest,
	host *supportMatrix.SupportedHost, public bool, projectName, spotPrice string) {
	switch host.Type {
	case supportMatrix.RHEL:
		request.rhel = &rhel.RHELRequest{
			VersionMajor: "8",
			Request: compute.Request{
				ProjecName: projectName,
				Public:     public,
				SpotPrice:  spotPrice,
				Specs:      host,
			}}

	case supportMatrix.MacM1:
		request.macm1 = &macm1.MacM1Request{
			Request: compute.Request{
				ProjecName: projectName,
				Public:     public,
				Specs:      host,
			},
		}
	}

}

func manageResults(stackResult auto.UpResult,
	host *supportMatrix.SupportedHost, public bool,
	destinationFolder string) error {
	if !public {
		if err := writeOutputs(stackResult, destinationFolder, map[string]string{
			fmt.Sprintf("%s-%s", compute.OutputPrivateKey, "bastion"): "bastion_id_rsa",
			fmt.Sprintf("%s-%s", compute.OutputHost, "bastion"):       "bastion_host",
			fmt.Sprintf("%s-%s", compute.OutputUsername, "bastion"):   "bastion_username",
		}); err != nil {
			return err
		}
	}
	switch host.Type {
	case supportMatrix.RHEL:
		if err := writeOutputs(stackResult, destinationFolder, map[string]string{
			fmt.Sprintf("%s-%s", compute.OutputPrivateKey, supportMatrix.OL_RHEL.ID): "rhel_id_rsa",
			fmt.Sprintf("%s-%s", compute.OutputHost, supportMatrix.OL_RHEL.ID):       "rhel_host",
			fmt.Sprintf("%s-%s", compute.OutputUsername, supportMatrix.OL_RHEL.ID):   "rhel_username",
		}); err != nil {
			return err
		}
	case supportMatrix.MacM1:
		if err := writeOutputs(stackResult, destinationFolder, map[string]string{
			fmt.Sprintf("%s-%s", compute.OutputPrivateKey, supportMatrix.G_MAC_M1.ID): "macm1_id_rsa",
			fmt.Sprintf("%s-%s", compute.OutputHost, supportMatrix.G_MAC_M1.ID):       "macm1_host",
			fmt.Sprintf("%s-%s", compute.OutputUsername, supportMatrix.G_MAC_M1.ID):   "macm1_username",
		}); err != nil {
			return err
		}
	}
	return nil
}

func writeOutputs(stackResult auto.UpResult, destinationFolder string, results map[string]string) (err error) {
	for k, v := range results {
		if err = writeOutput(stackResult, k, destinationFolder, v); err != nil {
			return err
		}
	}
	return
}

func writeOutput(stackResult auto.UpResult, outputkey, destinationFolder, destinationFilename string) error {
	value, ok := stackResult.Outputs[outputkey].Value.(string)
	if !ok {
		return fmt.Errorf("error getting %s", outputkey)
	}
	err := os.WriteFile(path.Join(destinationFolder, destinationFilename), []byte(value), 0600)
	if err != nil {
		return err
	}
	return nil
}
