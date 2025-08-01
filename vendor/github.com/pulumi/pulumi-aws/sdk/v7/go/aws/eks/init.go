// Code generated by pulumi-language-go DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package eks

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type module struct {
	version semver.Version
}

func (m *module) Version() semver.Version {
	return m.version
}

func (m *module) Construct(ctx *pulumi.Context, name, typ, urn string) (r pulumi.Resource, err error) {
	switch typ {
	case "aws:eks/accessEntry:AccessEntry":
		r = &AccessEntry{}
	case "aws:eks/accessPolicyAssociation:AccessPolicyAssociation":
		r = &AccessPolicyAssociation{}
	case "aws:eks/addon:Addon":
		r = &Addon{}
	case "aws:eks/cluster:Cluster":
		r = &Cluster{}
	case "aws:eks/fargateProfile:FargateProfile":
		r = &FargateProfile{}
	case "aws:eks/identityProviderConfig:IdentityProviderConfig":
		r = &IdentityProviderConfig{}
	case "aws:eks/nodeGroup:NodeGroup":
		r = &NodeGroup{}
	case "aws:eks/podIdentityAssociation:PodIdentityAssociation":
		r = &PodIdentityAssociation{}
	default:
		return nil, fmt.Errorf("unknown resource type: %s", typ)
	}

	err = ctx.RegisterResource(typ, name, nil, r, pulumi.URN_(urn))
	return
}

func init() {
	version, err := internal.PkgVersion()
	if err != nil {
		version = semver.Version{Major: 1}
	}
	pulumi.RegisterResourceModule(
		"aws",
		"eks/accessEntry",
		&module{version},
	)
	pulumi.RegisterResourceModule(
		"aws",
		"eks/accessPolicyAssociation",
		&module{version},
	)
	pulumi.RegisterResourceModule(
		"aws",
		"eks/addon",
		&module{version},
	)
	pulumi.RegisterResourceModule(
		"aws",
		"eks/cluster",
		&module{version},
	)
	pulumi.RegisterResourceModule(
		"aws",
		"eks/fargateProfile",
		&module{version},
	)
	pulumi.RegisterResourceModule(
		"aws",
		"eks/identityProviderConfig",
		&module{version},
	)
	pulumi.RegisterResourceModule(
		"aws",
		"eks/nodeGroup",
		&module{version},
	)
	pulumi.RegisterResourceModule(
		"aws",
		"eks/podIdentityAssociation",
		&module{version},
	)
}
