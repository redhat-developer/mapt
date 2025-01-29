package keypair

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	maptContext "github.com/redhat-developer/mapt/pkg/manager/context"
)

type KeyPairRequest struct {
	Name string
}

type KeyPairResources struct {
	AWSKeyPair *ec2.KeyPair
	PrivateKey *tls.PrivateKey
}

func (r KeyPairRequest) Create(ctx *pulumi.Context) (*KeyPairResources, error) {
	return r.create(ctx, nil)
}

// This will create the private on each update even when no changes are applied
func (r KeyPairRequest) CreateAlways(ctx *pulumi.Context) (*KeyPairResources, error) {
	return r.create(ctx, []pulumi.ResourceOption{pulumi.ReplaceOnChanges([]string{"*"})})
}

func (r KeyPairRequest) create(ctx *pulumi.Context,
	options []pulumi.ResourceOption) (*KeyPairResources, error) {
	privateKey, err := tls.NewPrivateKey(
		ctx,
		r.Name,
		&tls.PrivateKeyArgs{
			Algorithm: pulumi.String("RSA"),
			RsaBits:   pulumi.Int(4096),
		},
		options...)
	if err != nil {
		return nil, err
	}
	k, err := ec2.NewKeyPair(ctx,
		r.Name,
		&ec2.KeyPairArgs{
			PublicKey: privateKey.PublicKeyOpenssh,
			Tags:      maptContext.ResourceTags()})
	if err != nil {
		return nil, err
	}
	return &KeyPairResources{
			AWSKeyPair: k,
			PrivateKey: privateKey},
		nil
}
