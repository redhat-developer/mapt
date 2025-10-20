package keypair

import (
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/ec2"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

type KeyPairRequest struct {
	Name string
}

type KeyPairResources struct {
	AWSKeyPair *ec2.KeyPair
	PrivateKey *tls.PrivateKey
}

func (r KeyPairRequest) Create(ctx *pulumi.Context, mCtx *mc.Context) (*KeyPairResources, error) {
	kr, err := r.create(ctx, mCtx, r.Name, nil)
	if mCtx.Debug() {
		kr.PrivateKey.PrivateKeyPem.ApplyT(
			func(privateKey string) (*string, error) {
				logging.Debugf("%s", privateKey)
				return nil, nil
			})
	}
	return kr, err
}

// This will create the private on each update even when no changes are applied
func (r KeyPairRequest) CreateAlways(ctx *pulumi.Context, mCtx *mc.Context) (*KeyPairResources, error) {
	return r.create(ctx, mCtx, util.RandomID(r.Name), []pulumi.ResourceOption{
		pulumi.ReplaceOnChanges([]string{"name"}),
		pulumi.Aliases([]pulumi.Alias{{Name: pulumi.String(r.Name)}})})
}

func (r KeyPairRequest) create(ctx *pulumi.Context, mCtx *mc.Context, name string,
	options []pulumi.ResourceOption) (*KeyPairResources, error) {
	privateKey, err := tls.NewPrivateKey(
		ctx,
		name,
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
			KeyName:           pulumi.String(r.Name),
			PublicKeyMaterial: privateKey.PublicKeyOpenssh,
		})
	if err != nil {
		return nil, err
	}
	return &KeyPairResources{
			AWSKeyPair: k,
			PrivateKey: privateKey},
		nil
}
