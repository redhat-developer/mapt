package spot

import (
	"github.com/adrianriobo/qenvs/pkg/manager"
	qenvsContext "github.com/adrianriobo/qenvs/pkg/manager/context"
	"github.com/adrianriobo/qenvs/pkg/provider/aws"
	resourcesUtil "github.com/adrianriobo/qenvs/pkg/util/resources"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type SpotOptionRequest struct {
	Prefix             string
	ProductDescription string
	InstaceTypes       []string
	AMIName            string
	AMIArch            string
}

type SpotOptionResult struct {
	Region           string
	AvailabilityZone string
	AVGPrice         float64
	MaxPrice         float64
	Score            int64
}

// Create wil get the information for the best spot choice it is backed
// within a stack and state to allow idempotency, otherwise run 2nd time a create
// may bring other region with best option and then all dependant resources from other
// stacks would need to be updated

// So create will check if stack with state already exists, if that is the case it will
// pick info from its outputs
// If stack does not exists it will create it
func (r SpotOptionRequest) Create() (*SpotOptionResult, error) {
	stack, err := manager.CheckStack(manager.Stack{
		StackName:   qenvsContext.GetStackInstanceName("spotOption"),
		ProjectName: qenvsContext.GetInstanceName(),
		BackedURL:   qenvsContext.GetBackedURL()})
	if err != nil {
		return r.createStack()
	} else {
		return getOutputs(stack)
	}
}

// Check if spot option stack was created on the backed url
func Exist() bool {
	s, err := manager.CheckStack(manager.Stack{
		StackName:   qenvsContext.GetStackInstanceName("spotOption"),
		ProjectName: qenvsContext.GetInstanceName(),
		BackedURL:   qenvsContext.GetBackedURL()})
	return err == nil && s != nil
}

// Destroy the stack
func Destroy() (err error) {
	stack := manager.Stack{
		StackName:           qenvsContext.GetStackInstanceName("spotOption"),
		ProjectName:         qenvsContext.GetInstanceName(),
		BackedURL:           qenvsContext.GetBackedURL(),
		ProviderCredentials: aws.DefaultCredentials}
	return manager.DestroyStack(stack)
}

// function to create the stack
func (r SpotOptionRequest) createStack() (*SpotOptionResult, error) {
	stack := manager.Stack{
		StackName:           qenvsContext.GetStackInstanceName("spotOption"),
		ProjectName:         qenvsContext.GetInstanceName(),
		BackedURL:           qenvsContext.GetBackedURL(),
		ProviderCredentials: aws.DefaultCredentials,
		DeployFunc:          r.deployer,
	}
	stackResult, err := manager.UpStack(stack)
	if err != nil {
		return nil, err
	}
	return &SpotOptionResult{
		Region:           stackResult.Outputs["region"].Value.(string),
		AvailabilityZone: stackResult.Outputs["az"].Value.(string),
		MaxPrice:         stackResult.Outputs["max"].Value.(float64),
		AVGPrice:         stackResult.Outputs["avg"].Value.(float64),
		Score:            int64(stackResult.Outputs["score"].Value.(float64))}, nil
}

// deployer function to create the logic to get the best spot option
// and it will export the data from the best spot option to the stack state
func (r SpotOptionRequest) deployer(ctx *pulumi.Context) error {
	so, err := NewBestSpotOption(ctx,
		resourcesUtil.GetResourceName(r.Prefix, "bso", "bso"),
		r.ProductDescription, r.InstaceTypes, r.AMIName, r.AMIArch)
	if err != nil {
		return err
	}
	ctx.Export("region", pulumi.String(so.Region))
	ctx.Export("az", pulumi.String(so.AvailabilityZone))
	ctx.Export("max", pulumi.Float64(so.MaxPrice))
	ctx.Export("avg", pulumi.Float64(so.AVGPrice))
	ctx.Export("score", pulumi.Float64(so.Score))
	return nil
}

// function to get outputs from an existing stack
func getOutputs(stack *auto.Stack) (*SpotOptionResult, error) {
	outputs, err := manager.GetOutputs(stack)
	if err != nil {
		return nil, err
	}
	return &SpotOptionResult{
		Region:           outputs["region"].Value.(string),
		AvailabilityZone: outputs["az"].Value.(string),
		MaxPrice:         outputs["max"].Value.(float64),
		AVGPrice:         outputs["avg"].Value.(float64),
		Score:            int64(outputs["score"].Value.(float64))}, nil
}
