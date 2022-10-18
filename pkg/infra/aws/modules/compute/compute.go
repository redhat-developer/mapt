package compute

import (
	"github.com/adrianriobo/qenvs/pkg/infra/aws/services/ec2/keypair"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Compute interface {
	waitForInit(ctx *pulumi.Context) error
}

func ManageKeypair(ctx *pulumi.Context, keyPair *ec2.KeyPair,
	name, stackOutputKey string) (*ec2.KeyPair, *tls.PrivateKey, error) {
	if keyPair == nil {
		// create key
		keyResources, err := keypair.KeyPairRequest{
			Name: name}.Create(ctx)
		if err != nil {
			return nil, nil, err
		}
		ctx.Export(stackOutputKey, keyResources.PrivateKey.PrivateKeyPem)
		return keyResources.AWSKeyPair, keyResources.PrivateKey, nil
	}
	return keyPair, nil, nil
}
