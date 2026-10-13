package snc

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/redhat-developer/mapt/pkg/util/file"
)

func TestCloudConfigQuotesPasswordValues(t *testing.T) {
	data := dataValues{
		Username:          "core",
		PubKey:            "ssh-rsa AAAATEST",
		PublicIP:          "192.0.2.1",
		PullSecretB64:     "cHJ1bGxlcnNlY3JldA==",
		KubeAdminPassword: "!admin-password",
		DeveloperPassword: "!developer-password",
	}

	rendered, err := file.Template(data, string(cloudConfigTemplate))
	if err != nil {
		t.Fatalf("render cloud config: %v", err)
	}

	var config struct {
		WriteFiles []struct {
			Path    string `yaml:"path"`
			Content string `yaml:"content"`
		} `yaml:"write_files"`
	}
	if err := yaml.Unmarshal([]byte(rendered), &config); err != nil {
		t.Fatalf("parse rendered cloud config: %v", err)
	}

	contents := make(map[string]string)
	for _, writeFile := range config.WriteFiles {
		contents[writeFile.Path] = writeFile.Content
	}
	if contents["/opt/crc/pass_kubeadmin"] != data.KubeAdminPassword {
		t.Errorf("kubeadmin password was not preserved: %q", contents["/opt/crc/pass_kubeadmin"])
	}
	if contents["/opt/crc/pass_developer"] != data.DeveloperPassword {
		t.Errorf("developer password was not preserved: %q", contents["/opt/crc/pass_developer"])
	}
}
