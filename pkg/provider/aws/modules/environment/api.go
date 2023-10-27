package environment

import (
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/manager"
	"github.com/adrianriobo/qenvs/pkg/manager/credentials"
	"github.com/adrianriobo/qenvs/pkg/provider/aws"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/compute"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/compute/host/fedora"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/compute/host/rhel"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/compute/host/windows"
	"github.com/adrianriobo/qenvs/pkg/provider/aws/modules/compute/services/snc"
	network "github.com/adrianriobo/qenvs/pkg/provider/aws/modules/network/standard"
	spotprice "github.com/adrianriobo/qenvs/pkg/provider/aws/modules/spot-price"
	supportMatrix "github.com/adrianriobo/qenvs/pkg/provider/aws/support-matrix"
	"github.com/adrianriobo/qenvs/pkg/provider/util/output"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

func Create(projectName, backedURL, connectionDetailsOutput string,
	public bool, targetHostID string,
	rhMajorVersion, rhSubscriptionUsername, rhSubscriptionPassword string,
	fedoraMajorVersion, macosMajorVersion string) (err error) {
	// Check which supported host
	host, err := supportMatrix.GetHost(targetHostID)
	if err != nil {
		return err
	}
	// Check with spot price environment requirements
	availabilityZones, region, spotPrice, plugin, err :=
		getHostParameters(projectName, backedURL, host)
	if err != nil {
		return err
	}
	// Based on spot price info the full environment will be created
	request := singleHostRequest{
		name: projectName,
		network: &network.NetworkRequest{
			Name:               fmt.Sprintf("%s-%s", projectName, "network"),
			CIDR:               network.DefaultCIDRNetwork,
			AvailabilityZones:  availabilityZones,
			Region:             region,
			PublicSubnetsCIDRs: network.DefaultCIDRPublicSubnets[:1],
			SingleNatGateway:   false,
		},
	}
	// Add request values for requested host
	manageRequest(&request, host, public, projectName, spotPrice,
		rhMajorVersion, rhSubscriptionUsername, rhSubscriptionPassword,
		fedoraMajorVersion, macosMajorVersion)
	// Create stack
	stack := manager.Stack{
		StackName:           stackCreateEnvironmentName,
		ProjectName:         projectName,
		BackedURL:           backedURL,
		ProviderCredentials: *plugin,
		DeployFunc:          request.deployer,
	}
	// Exec stack
	stackResult, err := manager.UpStack(stack)
	if err != nil {
		// Even in case of failure we try to get credentials for manual interaction
		if err = manageResults(stackResult, &request, public, connectionDetailsOutput); err != nil {
			logging.Error(err)
		}
		return err
	}
	// Write host access info to disk
	if err = manageResults(stackResult, &request, public, connectionDetailsOutput); err != nil {
		return err
	}
	logging.Debug("Environment has been created")
	return nil
}

func Destroy(projectName, backedURL string) (err error) {
	stack := manager.Stack{
		StackName:           stackCreateEnvironmentName,
		ProjectName:         projectName,
		BackedURL:           backedURL,
		ProviderCredentials: aws.DefaultCredentials}
	err = manager.DestroyStack(stack)
	if err != nil {
		return
	}
	err = spotprice.Destroy(projectName, backedURL)
	return
}

// Function get host parameters for Az, Region, and price if spot, Plugin setup accordingly and error
func getHostParameters(projectName, backedURL string,
	host *supportMatrix.SupportedHost) ([]string, string, string, *credentials.ProviderCredentials, error) {
	var availabilityZones = network.DefaultAvailabilityZones[:1]
	var region string = network.DefaultRegion
	var spotPrice string
	var credentials = aws.DefaultCredentials
	if host.Spot {
		spg, err := spotprice.Create(projectName, backedURL, host.ID)
		if err != nil {
			return nil, "", "", nil, err
		}
		availabilityZones = []string{spg.AvailabilityZone}
		region = spg.Region
		spotPrice = fmt.Sprintf("%f", spg.MaxPrice)
		// plugin will use the region from the best spot price
		credentials = aws.GetClouProviderCredentials(
			map[string]string{
				aws.CONFIG_AWS_REGION: spg.Region})
	}
	return availabilityZones, region, spotPrice, &credentials, nil
}

func manageRequest(request *singleHostRequest,
	host *supportMatrix.SupportedHost, public bool,
	projectName, spotPrice string,
	rhMajorVersion, rhSubscriptionUsername, rhSubscriptionPassword string,
	fedoraMajorVersion, macosMajorVersion string) {
	switch host.ID {
	case supportMatrix.OL_RHEL.ID:
		request.hostRequested = &rhel.RHELRequest{
			VersionMajor:         rhMajorVersion,
			SubscriptionUsername: rhSubscriptionUsername,
			SubscriptionPassword: rhSubscriptionPassword,
			Request: compute.Request{
				ProjecName: projectName,
				Public:     public,
				SpotPrice:  spotPrice,
				Specs:      host,
			}}
	case supportMatrix.OL_Fedora.ID:
		request.hostRequested = &fedora.Request{
			VersionMajor: fedoraMajorVersion,
			Request: compute.Request{
				ProjecName: projectName,
				Public:     public,
				SpotPrice:  spotPrice,
				Specs:      host,
			}}
	case supportMatrix.OL_Windows.ID:
		request.hostRequested = &windows.WindowsRequest{
			Request: compute.Request{
				ProjecName: projectName,
				Public:     public,
				SpotPrice:  spotPrice,
				Specs:      host,
			}}
	case supportMatrix.S_SNC.ID:
		request.hostRequested = &snc.SNCRequest{
			RHELRequest: rhel.RHELRequest{VersionMajor: rhMajorVersion,
				SubscriptionUsername: rhSubscriptionUsername,
				SubscriptionPassword: rhSubscriptionPassword,
				Request: compute.Request{
					ProjecName: projectName,
					Public:     public,
					SpotPrice:  spotPrice,
					Specs:      host,
				}}}
	case supportMatrix.S_SNC.ID:
		request.hostRequested = &snc.SNCRequest{
			RHELRequest: rhel.RHELRequest{VersionMajor: rhMajorVersion,
				SubscriptionUsername: rhSubscriptionUsername,
				SubscriptionPassword: rhSubscriptionPassword,
				Request: compute.Request{
					ProjecName: projectName,
					Public:     public,
					SpotPrice:  spotPrice,
					Specs:      host,
				}}}
	}
}

func manageResults(stackResult auto.UpResult, request *singleHostRequest,
	public bool, destinationFolder string) error {
	// Currently support only one host on host operation
	// this should be change when create a environment with multiple hosts
	remoteHost := request.hostRequested
	if !public {
		remoteHost = request.bastion
	}
	if err := output.Write(stackResult, destinationFolder, map[string]string{
		remoteHost.GetRequest().OutputPrivateKey(): "id_rsa",
		remoteHost.GetRequest().OutputHost():       "host",
		remoteHost.GetRequest().OutputUsername():   "username",
		remoteHost.GetRequest().OutputPassword():   "password",
	}); err != nil {
		return err
	}
	return nil
}
