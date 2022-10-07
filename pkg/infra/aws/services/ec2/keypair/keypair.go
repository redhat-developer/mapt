package keypair

import (
	"fmt"

	utilCrypto "github.com/adrianriobo/qenvs/pkg/util/crypto"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type KeyPairRequest struct {
	ProjectName string
	Name        string
}

type KeyPairResources struct {
	KeyPair *ec2.KeyPair
	KeyPEM  []byte
	PubPEM  []byte
}

func (r KeyPairRequest) Create(ctx *pulumi.Context) (*KeyPairResources, error) {
	key, pub := utilCrypto.CreateDefaultKey()
	kName := fmt.Sprintf("%s-%s", r.ProjectName, r.Name)
	k, err := ec2.NewKeyPair(ctx,
		kName,
		&ec2.KeyPairArgs{
			PublicKey: pulumi.String(pub[:])})
	if err != nil {
		return nil, err
	}
	return &KeyPairResources{
			KeyPair: k,
			KeyPEM:  key,
			PubPEM:  pub},
		nil
}
