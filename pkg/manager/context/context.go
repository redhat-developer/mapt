package context

import (
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/util"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
	utilMaps "github.com/adrianriobo/qenvs/pkg/util/maps"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"golang.org/x/exp/maps"
)

const (
	tagKeyOrigin      = "origin"
	origin            = "qenvs"
	TagKeyProjectName = "projectName"
	TagKeyRunID       = "runid"
)

// store details for the current execution
type context struct {
	runID                 string
	projectName           string
	backedURL             string
	resultsOutput         string
	tags                  map[string]string
	tagsAsPulumiStringMap pulumi.StringMap
}

var c *context

func Init(projectName, backedURL, resultsOutput string, tags map[string]string) {
	c = &context{
		runID:         CreateRunID(),
		projectName:   projectName,
		backedURL:     backedURL,
		resultsOutput: resultsOutput,
		tags:          tags,
	}
	addCommonTags()
	logging.Debugf("context initialized for %s", c.runID)
}

func InitBase(projectName, backedURL string) {
	c = &context{
		projectName: projectName,
		backedURL:   backedURL,
	}
}

// It will create a runID
// if context has been intialized it will set it as the runID for the context
// otherwise it will return the value (one time value)
func CreateRunID() string {
	runID := util.RandomID(origin)
	if c != nil {
		c.runID = runID
	}
	return runID
}

func GetTags() map[string]string {
	return c.tags
}

// Get tags ready to be added to any pulumi resource
func ResourceTags() pulumi.StringMap {
	return ResourceTagsWithCustom(nil)
}

// Get tags ready to be added to any pulumi resource
// in addition we cas set specific custom tags
func ResourceTagsWithCustom(customTags map[string]string) pulumi.StringMap {
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

func RunID() string {
	return c.runID
}

func ProjectName() string {
	return c.projectName
}

func BackedURL() string {
	return c.backedURL
}

func GetResultsOutputPath() string {
	return c.resultsOutput
}

func StackNameByProject(stackName string) string {
	return fmt.Sprintf("%s-%s", stackName, c.projectName)
}

func addCommonTags() {
	c.tags[tagKeyOrigin] = origin
	c.tags[TagKeyProjectName] = c.projectName
}
