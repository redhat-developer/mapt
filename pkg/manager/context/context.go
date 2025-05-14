package context

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	utilMaps "github.com/redhat-developer/mapt/pkg/util/maps"
	"golang.org/x/exp/maps"
)

var (
	// mapt image to make self use. OCI image value is passed during building time
	// this is intended for full build process, when building mapt binary we need to ensure
	// OCI image already exists to make use of it
	OCI = "quay.io/redhat-developer/mapt:v0.0.0-unset"
)

const (
	tagKeyOrigin      = "origin"
	origin            = "mapt"
	TagKeyProjectName = "projectName"
	TagKeyRunID       = "runid"
)

type ContextArgs struct {
	ProjectName   string
	BackedURL     string
	ResultsOutput string
	Debug         bool
	DebugLevel    uint
	Tags          map[string]string
	// serverless here is used to set the credentials based on
	// roles inherid by tasks as serverless
	// see SetAWSCredentials function
	// take into account that the name may change as the approach to get
	// credentials from role is more general approach
	Serverless bool
	// This forces destroy even when lock exists
	ForceDestroy bool
	// integrations
	GHRunnerArgs *github.GithubRunnerArgs
	CirrusPWArgs *cirrus.PersistentWorkerArgs
	// Spot Bid Safe Limit
	SpotPriceIncreaseRate int
}

type context struct {
	runID                 string
	projectName           string
	backedURL             string
	resultsOutput         string
	debug                 bool
	debugLevel            uint
	serverless            bool
	forceDestroy          bool
	spotPriceIncreaseRate int
	tags                  map[string]string
	tagsAsPulumiStringMap pulumi.StringMap
}

// mapt context
var mc *context

type Provider interface {
	Init(backedURL string) error
}

func Init(ca *ContextArgs, provider Provider) error {
	mc = &context{
		runID:                 CreateRunID(),
		projectName:           ca.ProjectName,
		backedURL:             ca.BackedURL,
		resultsOutput:         ca.ResultsOutput,
		debug:                 ca.Debug,
		debugLevel:            ca.DebugLevel,
		tags:                  ca.Tags,
		serverless:            ca.Serverless,
		forceDestroy:          ca.ForceDestroy,
		spotPriceIncreaseRate: ca.SpotPriceIncreaseRate,
	}
	addCommonTags()
	// Init provider
	if err := provider.Init(ca.BackedURL); err != nil {
		return err
	}
	// Manage integrations
	if err := manageIntegration(ca); err != nil {
		return err
	}
	logging.Debugf("context initialized for %s", mc.runID)
	return nil
}

func RunID() string { return mc.runID }

func ProjectName() string { return mc.projectName }

func SetProjectName(projectName string) { mc.projectName = projectName }

func BackedURL() string { return mc.backedURL }

func GetResultsOutputPath() string { return mc.resultsOutput }

func GetTags() map[string]string { return mc.tags }

func ResourceTags() pulumi.StringMap { return ResourceTagsWithCustom(nil) }

func Debug() bool { return mc.debug }

func DebugLevel() uint { return mc.debugLevel }

func IsServerless() bool { return mc.serverless }

func IsForceDestroy() bool { return mc.forceDestroy }

func SpotPriceIncreaseRate() int { return mc.spotPriceIncreaseRate }

// It will create a runID
// if context has been intialized it will set it as the runID for the context
// otherwise it will return the value (one time value)
func CreateRunID() string {
	runID := util.RandomID(origin)
	if mc != nil {
		mc.runID = runID
	}
	return runID
}

// Get tags ready to be added to any pulumi resource
// in addition we cas set specific custom tags
func ResourceTagsWithCustom(customTags map[string]string) pulumi.StringMap {
	lTags := make(map[string]string)
	maps.Copy(lTags, mc.tags)
	if customTags != nil {
		maps.Copy(lTags, customTags)
	}
	if mc.tagsAsPulumiStringMap == nil {
		mc.tagsAsPulumiStringMap = utilMaps.Convert(lTags,
			func(name string) string { return name },
			func(value string) pulumi.StringInput { return pulumi.String(value) })
	}
	return mc.tagsAsPulumiStringMap
}

func StackNameByProject(stackName string) string {
	return fmt.Sprintf("%s-%s", stackName, mc.projectName)
}

func addCommonTags() {
	if mc.tags == nil {
		mc.tags = make(map[string]string)
	}
	mc.tags[tagKeyOrigin] = origin
	mc.tags[TagKeyProjectName] = mc.projectName
}

func manageIntegration(ca *ContextArgs) error {
	if ca.GHRunnerArgs != nil {
		ca.GHRunnerArgs.Name = RunID()
		github.Init(ca.GHRunnerArgs)
	}
	if ca.CirrusPWArgs != nil {
		ca.CirrusPWArgs.Name = RunID()
		cirrus.Init(ca.CirrusPWArgs)
	}
	return nil
}
