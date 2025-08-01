package spot

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/manager"
	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	spot "github.com/redhat-developer/mapt/pkg/provider/api/spot"
	"github.com/redhat-developer/mapt/pkg/provider/aws"
	"github.com/redhat-developer/mapt/pkg/provider/aws/data"
	resourcesUtil "github.com/redhat-developer/mapt/pkg/util/resources"
)

type SpotOptionRequest struct {
	MCtx               *mc.Context `validate:"required"`
	Prefix             string
	ProductDescription string
	InstaceTypes       []string
	AMIName            string
	AMIArch            string
}

func (r *SpotOptionRequest) validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(r)
}

type SpotOptionResult struct {
	Region           string
	AvailabilityZone string
	Price            float64
	Score            int64
}

type bestSpotOption struct {
	pulumi.ResourceState
	Option *spot.SpotResults
}

func NewBestSpotOption(ctx *pulumi.Context, mCtx *mc.Context, name string,
	productDescription string, instaceTypes []string,
	amiName, amiArch string, opts ...pulumi.ResourceOption) (*spot.SpotResults, error) {
	spotOption, err := data.SpotInfo(mCtx,
		&data.SpotInfoArgs{
			ProductDescription: &productDescription,
			InstaceTypes:       instaceTypes,
			AMIName:            &amiName,
			AMIArch:            &amiArch,
		})
	if err != nil {
		return nil, err
	}
	err = ctx.RegisterComponentResource("rh:qe:aws:bso",
		name,
		&bestSpotOption{
			Option: spotOption,
		},
		opts...)
	if err != nil {
		return nil, err
	}
	return spotOption, nil
}

// Create wil get the information for the best spot choice it is backed
// within a stack and state to allow idempotency, otherwise run 2nd time a create
// may bring other region with best option and then all dependant resources from other
// stacks would need to be updated

// So create will check if stack with state already exists, if that is the case it will
// pick info from its outputs
// If stack does not exists it will create it
func (r SpotOptionRequest) Create() (*SpotOptionResult, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}
	stack, err := manager.CheckStack(manager.Stack{
		StackName:   r.MCtx.StackNameByProject("spotOption"),
		ProjectName: r.MCtx.ProjectName(),
		BackedURL:   r.MCtx.BackedURL()})
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
func (r SpotOptionRequest) createStack() (*SpotOptionResult, error) {
	stack := manager.Stack{
		StackName:           r.MCtx.StackNameByProject("spotOption"),
		ProjectName:         r.MCtx.ProjectName(),
		BackedURL:           r.MCtx.BackedURL(),
		ProviderCredentials: aws.DefaultCredentials,
		DeployFunc:          r.deployer,
	}
	stackResult, err := manager.UpStack(r.MCtx, stack)
	if err != nil {
		return nil, err
	}
	return &SpotOptionResult{
		Region:           stackResult.Outputs["region"].Value.(string),
		AvailabilityZone: stackResult.Outputs["az"].Value.(string),
		Price:            stackResult.Outputs["max"].Value.(float64),
		Score:            int64(stackResult.Outputs["score"].Value.(float64))}, nil
}

// deployer function to create the logic to get the best spot option
// and it will export the data from the best spot option to the stack state
func (r SpotOptionRequest) deployer(ctx *pulumi.Context) error {
	so, err := NewBestSpotOption(ctx, r.MCtx,
		resourcesUtil.GetResourceName(r.Prefix, "bso", "bso"),
		r.ProductDescription, r.InstaceTypes, r.AMIName, r.AMIArch)
	if err != nil {
		return err
	}
	ctx.Export("region", pulumi.String(so.HostingPlace))
	ctx.Export("az", pulumi.String(so.AvailabilityZone))
	ctx.Export("max", pulumi.Float64(so.Price))
	ctx.Export("score", pulumi.Float64(so.ChanceLevel))
	return nil
}

// function to get outputs from an existing stack
func getOutputs(stack *auto.Stack) (*SpotOptionResult, error) {
	outputs, err := manager.GetOutputs(stack)
	if err != nil {
		return nil, err
	}
	if len(outputs) == 0 {
		return nil, errors.New("stack outputs are empty please destroy and re-create")
	}
	return &SpotOptionResult{
		Region:           outputs["region"].Value.(string),
		AvailabilityZone: outputs["az"].Value.(string),
		Price:            outputs["max"].Value.(float64),
		Score:            int64(outputs["score"].Value.(float64))}, nil
}
