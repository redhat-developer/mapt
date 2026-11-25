package aks

const (
	stackAKS         = "stackCreateAKS"
	azureAKSID       = "aaks"
	outputKubeconfig = "aksKubeconfig"

	// DefaultSystemPoolVMSize is the VM size used by the system pool when no
	// other size is set.
	DefaultSystemPoolVMSize = "Standard_A2_v2"
	// DefaultWorkersDesired is the number of nodes on the user pool when no
	// other value is set.
	DefaultWorkersDesired = 1
)
