package securitygroup

import (
	"fmt"

	"github.com/pulumi/pulumi-azure-native-sdk/network/v3"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
)

type IngressRules struct {
	Description string
	FromPort    int
	ToPort      int
	Protocol    string
	CidrBlocks  string
}

type SecurityGroupArgs struct {
	Name         string
	RG           *resources.ResourceGroup
	Location     *string
	IngressRules []IngressRules
}

type SecurityGroup = *network.NetworkSecurityGroup

func Create(ctx *pulumi.Context, mCtx *mc.Context, args *SecurityGroupArgs) (SecurityGroup, error) {
	nsg, err := network.NewNetworkSecurityGroup(ctx,
		args.Name,
		&network.NetworkSecurityGroupArgs{
			NetworkSecurityGroupName: pulumi.String(args.Name),
			ResourceGroupName:        args.RG.Name,
			Location:                 pulumi.String(*args.Location),
			SecurityRules:            securityRules(args.IngressRules),
			Tags:                     mCtx.ResourceTags(),
		})
	if err != nil {
		return nil, err
	}
	return nsg, nil
}

func securityRules(rules []IngressRules) (sra network.SecurityRuleTypeArray) {
	priority := 1000
	for _, r := range rules {
		priority++
		sr := network.SecurityRuleTypeArgs{
			Name:                     pulumi.String(r.Description),
			Access:                   pulumi.String("Allow"),
			Description:              pulumi.String(r.Description),
			Priority:                 pulumi.Int(priority),
			Direction:                pulumi.String("Inbound"),
			SourcePortRange:          pulumi.String("*"),
			DestinationPortRange:     pulumi.String(fmt.Sprint(r.ToPort)),
			Protocol:                 pulumi.String(r.Protocol),
			SourceAddressPrefix:      pulumi.String("*"),
			DestinationAddressPrefix: pulumi.String("*"),
		}
		sra = append(sra, sr)
	}
	return sra
}
