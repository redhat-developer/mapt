package environment

import (
	"fmt"
	"os"
	"path"

	"github.com/adrianriobo/qenvs/pkg/infra/aws"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/compute/rhel"
	"github.com/adrianriobo/qenvs/pkg/infra/aws/modules/network"
	spotprice "github.com/adrianriobo/qenvs/pkg/infra/aws/modules/spot-price"
	utilInfra "github.com/adrianriobo/qenvs/pkg/infra/util"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

func Create(projectName, backedURL string, spot, public bool) (err error) {

	// TODO define Parse configuration for env

	// If option for best price in...
	var spotPriceInfo *spotprice.SpotPriceData
	var availabilityZones = network.DefaultAvailabilityZones[:1]
	var spotPrice string
	var plugin = aws.PluginAWSDefault
	if spot {
		spotPriceInfo, err = spotprice.BestSpotPriceInfo(
			[]string{},
			[]string{"c5n.metal"},
			"Red Hat Enterprise Linux")
		if err != nil {
			return err
		}
		availabilityZones = []string{spotPriceInfo.AvailabilityZone}
		spotPrice = spotPriceInfo.Price
		plugin = aws.GetPluginAWS(
			map[string]string{
				aws.CONFIG_AWS_REGION: spotPriceInfo.Region})
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
		rhel: &rhel.RHELRequest{
			Name:         fmt.Sprintf("%s-%s", projectName, "rhel"),
			VersionMajor: rhel.VERSION_8,
			Public:       public,
			SpotPrice:    spotPrice,
		},
	}

	// plugin will use the region from the best spot price
	stack := utilInfra.Stack{
		StackName:   stackCreateEnvironmentName,
		ProjectName: projectName,
		BackedURL:   backedURL,
		Plugin:      plugin,
		DeployFunc:  request.deployer,
	}
	// Exec stack
	stackResult, err := utilInfra.UpStack(stack)
	if err != nil {
		return err
	}

	// if option bastion
	// if err = bastionOutputs(stackResult); err != nil {
	// 	return err
	// }

	if err = writeOutput(stackResult, rhel.OutputPrivateKey, "/tmp/qenvs", "rhel_id_rsa"); err != nil {
		return err
	}
	if err = writeOutput(stackResult, rhel.OutputPrivateIP, "/tmp/qenvs", "rhel_host"); err != nil {
		return err
	}
	if err = writeOutput(stackResult, rhel.OutputUsername, "/tmp/qenvs", "rhel_username"); err != nil {
		return err
	}
	logging.Debug("Environment has been created")
	return nil
}

// func bastionOutputs(stackResult auto.UpResult) (err error) {
// 	if err = writeOutput(stackResult, bastion.OutputPrivateKey, "/tmp/qenvs", "bastion_id_rsa"); err != nil {
// 		return err
// 	}
// 	if err = writeOutput(stackResult, bastion.OutputPublicIP, "/tmp/qenvs", "bastion_host"); err != nil {
// 		return err
// 	}
// 	if err = writeOutput(stackResult, bastion.OutputUsername, "/tmp/qenvs", "bastion_username"); err != nil {
// 		return err
// 	}
// 	return
// }

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

func Destroy(projectName, backedURL string) error {
	stack := utilInfra.Stack{
		StackName:   stackCreateEnvironmentName,
		ProjectName: projectName,
		BackedURL:   backedURL,
		Plugin:      aws.PluginAWSDefault}
	_, err := utilInfra.DestroyStack(stack)
	if err != nil {
		return err
	}
	logging.Debugf("Environment has been destroyed")
	return nil
}
