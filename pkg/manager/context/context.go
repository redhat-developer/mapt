package context

import (
	"fmt"

	"maps"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/redhat-developer/mapt/pkg/integrations/cirrus"
	"github.com/redhat-developer/mapt/pkg/integrations/github"
	"github.com/redhat-developer/mapt/pkg/util"
	"github.com/redhat-developer/mapt/pkg/util/logging"
	utilMaps "github.com/redhat-developer/mapt/pkg/util/maps"
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
	ProjectName string
	BackedURL   string
	//Optional
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
}

type Context struct {
	runID         string
	projectName   string
	backedURL     string
	resultsOutput string
	debug         bool
	debugLevel    uint
	serverless    bool
	forceDestroy  bool
	// spotPriceIncreaseRate int
	tags                  map[string]string
	tagsAsPulumiStringMap pulumi.StringMap
}

type Provider interface {
	Init(backedURL string) error
}

func InitNoState() *Context { return &Context{} }

func Init(ca *ContextArgs, provider Provider) (*Context, error) {
	c := &Context{
		runID:         util.RandomID(origin),
		projectName:   ca.ProjectName,
		backedURL:     ca.BackedURL,
		resultsOutput: ca.ResultsOutput,
		debug:         ca.Debug,
		debugLevel:    ca.DebugLevel,
		tags:          ca.Tags,
		serverless:    ca.Serverless,
		forceDestroy:  ca.ForceDestroy,
	}
	addCommonTags(c)
	// Init provider
	if err := provider.Init(ca.BackedURL); err != nil {
		return nil, err
	}
	// Manage integrations
	if err := manageIntegration(c, ca); err != nil {
		return nil, err
	}
	logging.Debugf("context initialized for %s", c.runID)
	return c, nil
}

func (c *Context) RunID() string { return c.runID }

func (c *Context) ProjectName() string { return c.projectName }

func (c *Context) SetProjectName(projectName string) { c.projectName = projectName }

func (c *Context) BackedURL() string { return c.backedURL }

func (c *Context) GetResultsOutputPath() string { return c.resultsOutput }

func (c *Context) GetTags() map[string]string { return c.tags }

func (c *Context) ResourceTags() pulumi.StringMap { return c.ResourceTagsWithCustom(nil) }

func (c *Context) Debug() bool { return c.debug }

func (c *Context) DebugLevel() uint { return c.debugLevel }

func (c *Context) IsServerless() bool { return c.serverless }

func (c *Context) IsForceDestroy() bool { return c.forceDestroy }

// Get tags ready to be added to any pulumi resource
// in addition we cas set specific custom tags
func (c *Context) ResourceTagsWithCustom(customTags map[string]string) pulumi.StringMap {
	lTags := make(map[string]string)
	maps.Copy(lTags, c.tags)
	if customTags != nil {
		maps.Copy(lTags, customTags)
	}
	if c.tagsAsPulumiStringMap == nil {
		c.tagsAsPulumiStringMap = utilMaps.Convert(lTags,
			func(name string) string { return name },
			func(value string) pulumi.StringInput { return pulumi.String(value) })
	}
	return c.tagsAsPulumiStringMap
}

func (c *Context) StackNameByProject(stackName string) string {
	return fmt.Sprintf("%s-%s", stackName, c.projectName)
}

func addCommonTags(c *Context) {
	if c.tags == nil {
		c.tags = make(map[string]string)
	}
	c.tags[tagKeyOrigin] = origin
	c.tags[TagKeyProjectName] = c.projectName
}

func manageIntegration(c *Context, ca *ContextArgs) error {
	if ca.GHRunnerArgs != nil {
		ca.GHRunnerArgs.Name = c.RunID()
		github.Init(ca.GHRunnerArgs)
	}
	if ca.CirrusPWArgs != nil {
		ca.CirrusPWArgs.Name = c.RunID()
		cirrus.Init(ca.CirrusPWArgs)
	}
	return nil
}
