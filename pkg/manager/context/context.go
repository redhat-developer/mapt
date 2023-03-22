package context

import (
	"github.com/adrianriobo/qenvs/pkg/util/maps"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Struct holding the information for
// the pulumi execution
type context struct {
	// TODO change to instanceName?
	projectName string
	tags        pulumi.StringMap
}

var c context

func Init(projectName string, tags map[string]string) {
	c = context{
		projectName: projectName,
		tags: maps.Convert(tags,
			func(name string) string { return name },
			func(value string) pulumi.StringInput { return pulumi.String(value) }),
	}
}

func GetTags() pulumi.StringMap {
	return c.tags
}

func GetName() string {
	return c.projectName
}
