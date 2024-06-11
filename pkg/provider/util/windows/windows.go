package windows

import (
	_ "embed"
	"encoding/base64"
	"fmt"

	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/util/file"
)

//go:embed userdata.ps1.tpl
var userdataTemplate []byte

//go:embed setup.ps1
var SetupScript []byte

type userDataValues struct {
	Username      string
	Password      string
	AuthorizedKey string
	Hostname      string
}

func GetUserdata(ctx *pulumi.Context, resourceID string,
	username, hostname string, privateKey *tls.PrivateKey,
	password *random.RandomPassword) (pulumi.StringPtrInput, error) {
	udBase64 := pulumi.All(password.Result, privateKey.PublicKeyOpenssh).ApplyT(
		func(args []interface{}) (string, error) {
			password := args[0].(string)
			authorizedKey := args[1].(string)
			userdata, err := file.Template(
				userDataValues{
					username,
					password,
					authorizedKey,
					hostname},
				fmt.Sprintf("%s-%s", "windows-userdata", resourceID),
				string(userdataTemplate[:]))
			if err != nil {
				return "", err
			}
			return base64.StdEncoding.EncodeToString([]byte(userdata)), nil
		}).(pulumi.StringOutput)
	return udBase64, nil
}
