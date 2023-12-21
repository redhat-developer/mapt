package context

import (
	"crypto/rand"
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/util/logging"
	utilMaps "github.com/adrianriobo/qenvs/pkg/util/maps"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"golang.org/x/exp/maps"
)

const (
	originTagName  = "origin"
	originTagValue = "qenvs"
	instaceTagName = "instanceID"
)

// store details for the current execution
type context struct {
	id                    string
	instanceID            string
	backedURL             string
	resultsOutput         string
	tags                  map[string]string
	tagsAsPulumiStringMap pulumi.StringMap
}

var c context

func Init(instanceID, backedURL, resultsOutput string, tags map[string]string) {
	c = context{
		instanceID:    instanceID,
		id:            randomID(originTagValue),
		backedURL:     backedURL,
		resultsOutput: resultsOutput,
		tags:          tags,
	}
	addCommonTags()
	logging.Debugf("Context initialized for %s", c.id)
}

func InitBase(instanceID, backedURL string) {
	c = context{
		instanceID: instanceID,
		backedURL:  backedURL,
	}
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

func GetID() string {
	return c.id
}

func GetInstanceName() string {
	return c.instanceID
}

func GetBackedURL() string {
	return c.backedURL
}

func GetResultsOutputPath() string {
	return c.resultsOutput
}

func GetStackInstanceName(stackName string) string {
	return fmt.Sprintf("%s-%s", stackName, c.instanceID)
}

func addCommonTags() {
	c.tags[originTagName] = originTagValue
	c.tags[instaceTagName] = c.instanceID
}

func randomID(name string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s%x", name, b)
}
