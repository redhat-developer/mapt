package keypair

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type KeyPairRequest struct {
	Name string
}

type KeyPairResources struct {
	AWSKeyPair *ec2.KeyPair
	PrivateKey *tls.PrivateKey
}

func (r KeyPairRequest) Create(ctx *pulumi.Context) (*KeyPairResources, error) {
	privateKey, err := tls.NewPrivateKey(
		ctx,
		r.Name,
		&tls.PrivateKeyArgs{
			Algorithm: pulumi.String("RSA"),
			RsaBits:   pulumi.Int(4096),
		})
	if err != nil {
		return nil, err
	}
	k, err := ec2.NewKeyPair(ctx,
		r.Name,
		&ec2.KeyPairArgs{
			PublicKey: privateKey.PublicKeyOpenssh})
	if err != nil {
		return nil, err
	}
	return &KeyPairResources{
			AWSKeyPair: k,
			PrivateKey: privateKey},
		nil
}
