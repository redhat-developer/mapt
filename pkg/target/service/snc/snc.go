package snc

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

func Results(stackResult auto.UpResult, prefix *string,
	resultOutputPath string, spotPrice *float64,
	disableClusterReadiness bool) (*SNCResults, error) {
	username, err := getResultOutput(OutputUsername, stackResult, prefix)
	if err != nil {
		return nil, err
	}
	privateKey, err := getResultOutput(OutputUserPrivateKey, stackResult, prefix)
	if err != nil {
		return nil, err
	}
	host, err := getResultOutput(OutputHost, stackResult, prefix)
	if err != nil {
		return nil, err
	}
	kubeAdminPass, err := getResultOutput(OutputKubeAdminPass, stackResult, prefix)
	if err != nil {
		return nil, err
	}
	kubeconfig := ""
	if !disableClusterReadiness {
		kubeconfig, err = getResultOutput(OutputKubeconfig, stackResult, prefix)
		if err != nil {
			return nil, err
		}
	}

	hostIPKey := fmt.Sprintf("%s-%s", *prefix, OutputHost)
	results := map[string]string{
		fmt.Sprintf("%s-%s", *prefix, OutputUsername):       "username",
		fmt.Sprintf("%s-%s", *prefix, OutputUserPrivateKey): "id_rsa",
		hostIPKey: "host",
		fmt.Sprintf("%s-%s", *prefix, OutputKubeconfig):    "kubeconfig",
		fmt.Sprintf("%s-%s", *prefix, OutputKubeAdminPass): "kubeadmin_pass",
		fmt.Sprintf("%s-%s", *prefix, OutputDeveloperPass): "developer_pass",
	}

	if len(resultOutputPath) == 0 {
		logging.Warn("conn-details-output flag not set; skipping writing output files.")
	} else {
		if err := output.Write(stackResult, resultOutputPath, results); err != nil {
			return nil, fmt.Errorf("failed to write results: %w", err)
		}
	}

	consoleURL := fmt.Sprintf(consoleURLRegex, host)
	if eip, ok := stackResult.Outputs[hostIPKey].Value.(string); ok {
		fmt.Printf("Cluster has been started you can access console at: %s.\n", fmt.Sprintf(consoleURLRegex, eip))
	}

	return &SNCResults{
		Username:      username,
		PrivateKey:    privateKey,
		Host:          host,
		Kubeconfig:    kubeconfig,
		KubeadminPass: kubeAdminPass,
		SpotPrice:     spotPrice,
		ConsoleUrl:    consoleURL,
	}, nil
}

func getResultOutput(name string, sr auto.UpResult, prefix *string) (string, error) {
	key := fmt.Sprintf("%s-%s", *prefix, name)
	output, ok := sr.Outputs[key]
	if !ok {
		return "", fmt.Errorf("output not found: %s", key)
	}
	value, ok := output.Value.(string)
	if !ok {
		return "", fmt.Errorf("output for %s is not a string", key)
	}
	return value, nil
}
