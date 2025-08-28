package spot

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	"github.com/redhat-developer/mapt/pkg/util"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type SpotStackArgs struct {
	Prefix             string
	ProductDescription string
	InstaceTypes       []string
	AMIName            string
	AMIArch            string
	Spot               *spotTypes.SpotArgs
}

type SpotStackResult struct {
	Region           string
	AvailabilityZone string
	InstanceType     []string
	Price            float64
	Score            int64
}

type spotStackRequest struct {
	mCtx               *mc.Context
	prefix             string
	productDescription string
	instaceTypes       []string
	amiName            string
	amiArch            string
	spot               *spotTypes.SpotArgs
}

func (r *SpotStackArgs) validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(r)
}

func (r *SpotStackArgs) toRequest(mCtx *mc.Context) *spotStackRequest {
	return &spotStackRequest{
		mCtx:               mCtx,
		prefix:             r.Prefix,
		productDescription: r.ProductDescription,
		instaceTypes:       r.InstaceTypes,
		amiName:            r.AMIName,
		amiArch:            r.AMIArch,
		spot:               r.Spot,
	}
}

// Create wil get the information for the best spot choice it is backed
// within a stack and state to allow idempotency, otherwise run 2nd time a create
// may bring other region with best option and then all dependant resources from other
// stacks would need to be updated

// So create will check if stack with state already exists, if that is the case it will
// pick info from its outputs
// If stack does not exists it will create it
func Create(mCtx *mc.Context, args *SpotStackArgs) (*SpotStackResult, error) {
	if err := args.validate(); err != nil {
		return nil, err
	}
	stack, err := manager.CheckStack(manager.Stack{
		StackName:   mCtx.StackNameByProject("spotOption"),
		ProjectName: mCtx.ProjectName(),
		BackedURL:   mCtx.BackedURL()})
	r := args.toRequest(mCtx)
	if err != nil {
		return r.createStack()
	} else {
		return getOutputs(stack)
	}
}

// Check if spot option stack was created on the backed url
func Exist(mCtx *mc.Context) bool {
	s, err := manager.CheckStack(manager.Stack{
		StackName:   mCtx.StackNameByProject("spotOption"),
		ProjectName: mCtx.ProjectName(),
		BackedURL:   mCtx.BackedURL()})
	return err == nil && s != nil
}

// Destroy the stack
func Destroy(mCtx *mc.Context) (err error) {
	stack := manager.Stack{
		StackName:           mCtx.StackNameByProject("spotOption"),
		ProjectName:         mCtx.ProjectName(),
		BackedURL:           mCtx.BackedURL(),
		ProviderCredentials: aws.DefaultCredentials}
	return manager.DestroyStack(mCtx, stack)
}

// function to create the stack
func (r *spotStackRequest) createStack() (*SpotStackResult, error) {
	stack := manager.Stack{
		StackName:           r.mCtx.StackNameByProject("spotOption"),
		ProjectName:         r.mCtx.ProjectName(),
		BackedURL:           r.mCtx.BackedURL(),
		ProviderCredentials: aws.DefaultCredentials,
		DeployFunc:          r.deployer,
	}
	stackResult, err := manager.UpStack(r.mCtx, stack)
	if err != nil {
		return nil, err
	}
	return &SpotStackResult{
		Region:           stackResult.Outputs["region"].Value.(string),
		AvailabilityZone: stackResult.Outputs["az"].Value.(string),
		InstanceType: util.ArrayCast[string](
			stackResult.Outputs["instancetypes"].Value.([]interface{})),
		Price: stackResult.Outputs["max"].Value.(float64),
		Score: int64(stackResult.Outputs["score"].Value.(float64))}, nil
}

// deployer function to create the logic to get the best spot option
// and it will export the data from the best spot option to the stack state
func (r *spotStackRequest) deployer(ctx *pulumi.Context) error {
	sia := &data.SpotInfoArgs{
		ProductDescription: &r.productDescription,
		InstaceTypes:       r.instaceTypes,
		AMIName:            &r.amiName,
		AMIArch:            &r.amiArch,
	}
	if r.spot != nil {
		sia.ExcludedRegions = r.spot.ExcludedHostingPlaces
		sia.SpotPriceIncreaseRate = &r.spot.IncreaseRate
		sia.SpotTolerance = &r.spot.Tolerance
	}
	so, err := NewBestSpotOption(ctx, r.mCtx,
		resourcesUtil.GetResourceName(r.prefix, "bso", "bso"),
		sia)
	if err != nil {
		return err
	}
	ctx.Export("region", pulumi.String(so.HostingPlace))
	ctx.Export("az", pulumi.String(so.AvailabilityZone))
	ctx.Export("instancetypes", pulumi.ToStringArray(so.ComputeType))
	ctx.Export("max", pulumi.Float64(so.Price))
	ctx.Export("score", pulumi.Float64(so.ChanceLevel))
	return nil
}

// function to get outputs from an existing stack
func getOutputs(stack *auto.Stack) (*SpotStackResult, error) {
	outputs, err := manager.GetOutputs(stack)
	if err != nil {
		return nil, err
	}
	if len(outputs) == 0 {
		return nil, errors.New("stack outputs are empty please destroy and re-create")
	}
	return &SpotStackResult{
		Region:           outputs["region"].Value.(string),
		AvailabilityZone: outputs["az"].Value.(string),
		Price:            outputs["max"].Value.(float64),
		InstanceType: util.ArrayCast[string](
			outputs["instancetypes"].Value.([]interface{})),
		Score: int64(outputs["score"].Value.(float64))}, nil

}
