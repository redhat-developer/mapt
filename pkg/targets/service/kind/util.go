package kind

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/provider/util/output"
)

const (
	// Outputs
	OKHost       = "kndHost"
	OKUsername   = "kndUsername"
	OKPrivateKey = "kndPrivatekey"
	OKKubeconfig = "kndKubeconfig"
	OKSpotPrice  = "kndSpotPrice"
)

func Results(mCtx *mc.Context, stackResult auto.UpResult, prefix *string) (*KindResults, error) {
	username, err := get[string](OKUsername, stackResult, prefix)
	if err != nil {
		return nil, err
	}
	privateKey, err := get[string](OKPrivateKey, stackResult, prefix)
	if err != nil {
		return nil, err
	}
	host, err := get[string](OKHost, stackResult, prefix)
	if err != nil {
		return nil, err
	}
	kubeconfig, err := get[string](OKKubeconfig, stackResult, prefix)
	if err != nil {
		return nil, err
	}
	spotPrice, err := get[float64](OKSpotPrice, stackResult, prefix)
	if err != nil {
		return nil, err
	}
	if mCtx.GetResultsOutputPath() != "" {
		if err := output.Write(stackResult, mCtx.GetResultsOutputPath(), map[string]string{
			fmt.Sprintf("%s-%s", *prefix, OKUsername):   "username",
			fmt.Sprintf("%s-%s", *prefix, OKPrivateKey): "id_rsa",
			fmt.Sprintf("%s-%s", *prefix, OKHost):       "host",
			fmt.Sprintf("%s-%s", *prefix, OKKubeconfig): "kubeconfig",
		}); err != nil {
			return nil, fmt.Errorf("failed to write results: %w", err)
		}
	}
	return &KindResults{
		Username:   username,
		PrivateKey: privateKey,
		Host:       host,
		Kubeconfig: kubeconfig,
		SpotPrice:  spotPrice,
	}, nil
}

func get[T any](name string, sr auto.UpResult, prefix *string) (*T, error) {
	key := fmt.Sprintf("%s-%s", *prefix, name)
	output, ok := sr.Outputs[key]
	if !ok {
		return nil, fmt.Errorf("output not found: %s", key)
	}
	value, ok := output.Value.(T)
	if !ok {
		return nil, fmt.Errorf("output for %s is not the right type", key)
	}
	return &value, nil
}
