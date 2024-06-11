package resources

import "fmt"

// Returns the unique name to identify a resoruces within
// pulumi context
func GetResourceName(prefix, maptComponentID, resourceTypeAbbrev string) string {
	if len(prefix) > 0 {
		return fmt.Sprintf("%s-%s-%s", prefix, maptComponentID, resourceTypeAbbrev)
	}
	return fmt.Sprintf("%s-%s", maptComponentID, resourceTypeAbbrev)
}
