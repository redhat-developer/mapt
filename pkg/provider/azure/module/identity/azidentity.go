package identity

import (
	"os"
	"strings"
)

var azIdentityEnvs = []string{
	"AZURE_TENANT_ID",
	"AZURE_SUBSCRIPTION_ID",
	"AZURE_CLIENT_ID",
	"AZURE_CLIENT_SECRET",
}

// Envs required for auth with go sdk
// https://learn.microsoft.com/es-es/azure/developer/go/azure-sdk-authentication?tabs=bash#service-principal-with-a-secret
// do not match standard envs for pulumi envs for auth with native sdk
// https://www.pulumi.com/registry/packages/azure-native/installation-configuration/#set-configuration-using-environment-variables
func SetAZIdentityEnvs() {
	for _, e := range azIdentityEnvs {
		os.Setenv(e,
			os.Getenv(strings.ReplaceAll(e, "AZURE", "ARM")))
	}
}
