package snc

import (
	"fmt"
	"strings"
)

const (
	ibmCloudSNCID = "icsnc"

	defaultUser     = "core"
	defaultProfile  = "bx2-16x64"
	defaultDiskSize = 200

	imageNamePattern = "openshift-local-%s-%s"

	smPullSecretSuffix    = "pull-secret"
	smKubeAdminPassSuffix = "kubeadminpassword"
	smDeveloperPassSuffix = "devpassword"

	// smInstancePlan is the Secrets Manager plan used for ephemeral instances.
	// Use "standard" if a trial instance already exists in the region.
	smInstancePlan = "trial"
)

func imageName(version, arch string) string {
	v := strings.ReplaceAll(version, ".", "-")
	a := strings.ReplaceAll(arch, "_", "-")
	return fmt.Sprintf(imageNamePattern, v, a)
}
