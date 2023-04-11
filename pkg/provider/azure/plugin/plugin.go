package plugin

import "github.com/adrianriobo/qenvs/pkg/manager/plugin"

const (
	pluginName    string = "azure-native"
	pluginVersion string = "v1.98.1"
)

func GetClouProviderPlugin(fixedCredentials map[string]string) plugin.PluginInfo {
	return plugin.PluginInfo{
		Name:              pluginName,
		Version:           pluginVersion,
		SetCredentialFunc: nil,
		FixedCredentials:  fixedCredentials}
}

var DefaultPlugin = GetClouProviderPlugin(nil)
