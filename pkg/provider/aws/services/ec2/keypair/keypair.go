package keypair

import (
	qenvsContext "github.com/adrianriobo/qenvs/pkg/manager/context"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
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
		},
		pulumi.ReplaceOnChanges([]string{"name"}))
	if err != nil {
		return nil, err
	}
	k, err := ec2.NewKeyPair(ctx,
		r.Name,
		&ec2.KeyPairArgs{
			PublicKey: privateKey.PublicKeyOpenssh,
			Tags:      qenvsContext.ResourceTags()})
	if err != nil {
		return nil, err
	}
	return &KeyPairResources{
			AWSKeyPair: k,
			PrivateKey: privateKey},
		nil
}
