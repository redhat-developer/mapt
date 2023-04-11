package context

import (
	"crypto/rand"
	"fmt"

	"github.com/adrianriobo/qenvs/pkg/util/maps"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	originTagName  = "origin"
	originTagValue = "qenvs"
	instaceTagName = "instanceID"
)

// store details for the current execution
type context struct {
	id            string
	instanceID    string
	backedURL     string
	resultsOutput string
	tags          pulumi.StringMap
}

var c context

func Init(instanceID, backedURL, resultsOutput string, tags map[string]string) {
	c = context{
		instanceID:    instanceID,
		id:            randomID(originTagValue),
		backedURL:     backedURL,
		resultsOutput: resultsOutput,
		tags: maps.Convert(tags,
			func(name string) string { return name },
			func(value string) pulumi.StringInput { return pulumi.String(value) }),
	}
	addCommonTags()
}

func GetTags() pulumi.StringMap {
	return c.tags
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

func GetResultsOutput() string {
	return c.resultsOutput
}

func GetStackInstanceName(stackName string) string {
	return fmt.Sprintf("%s-%s", stackName, c.instanceID)
}

func addCommonTags() {
	c.tags[originTagName] = pulumi.String(originTagValue)
	c.tags[instaceTagName] = pulumi.String(c.instanceID)
}

func randomID(name string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s%x", name, b)
}
