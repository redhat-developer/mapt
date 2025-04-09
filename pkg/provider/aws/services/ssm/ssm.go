package ssm

import (
	"fmt"

	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/ssm"
	// "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ssm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/util"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

func AddSSM(ctx *pulumi.Context, prefix, id, value *string) (*string, *ssm.Parameter, error) {
	name := fmt.Sprintf("/%s/%s", *id, util.RandomID(*id))
	param, err := ssm.NewParameter(ctx,
		resourcesUtil.GetResourceName(*prefix, *id, "ssm"),
		&ssm.ParameterArgs{
			Name:  pulumi.String(name),
			Type:  ssm.ParameterTypeString,
			Value: pulumi.String(*value),
		})
	if err != nil {
		return nil, nil, err
	}
	return &name, param, nil
}

func AddSSMFromResource(ctx *pulumi.Context, prefix, id *string, value pulumi.StringInput) (*string, *ssm.Parameter, error) {
	name := fmt.Sprintf("/%s/%s", *id, util.RandomID(*id))
	param, err := ssm.NewParameter(ctx,
		resourcesUtil.GetResourceName(*prefix, *id, "ssm"),
		&ssm.ParameterArgs{
			Name:  pulumi.String(name),
			Type:  ssm.ParameterTypeString,
			Value: value,
		})
	if err != nil {
		return nil, nil, err
	}
	return &name, param, nil
}
