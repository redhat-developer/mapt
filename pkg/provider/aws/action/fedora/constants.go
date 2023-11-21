package fedora

var (
	stackName            = "stackFedoraBaremetal"
	awsFedoraDedicatedID = "afd"

	diskSize int = 200

	amiRegex       = "Fedora-Cloud-Base-%s*"
	amiOwner       = "125523088429"
	amiUserDefault = "fedora"

	requiredInstanceTypes = []string{"c5.metal", "c5d.metal", "c5n.metal"}

	outputHost           = "afdHost"
	outputUsername       = "afdUsername"
	outputUserPrivateKey = "afdPrivatekey"
)
