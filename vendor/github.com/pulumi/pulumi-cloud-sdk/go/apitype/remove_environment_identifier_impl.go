// Copyright 2026, Pulumi Corporation.  All rights reserved.

package apitype

import (
	"encoding/json"
	"fmt"
)

// NOTE: It's possible to set a name, slug, and description fo the Pulumi team
// that is independent from those values on the GitHub-side. However, for the
// time being we don't. And just set those to mirror the values from GitHub.

func (r *RemoveEnvironmentIdentifier) UnmarshalJSON(data []byte) error {
	var raw any

	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	switch raw := raw.(type) {
	case string:
		*r = RemoveEnvironmentIdentifier{
			EnvName: raw,
		}
	case map[string]any:
		pn, ok := raw["projectName"]
		if !ok {
			return fmt.Errorf("expected removeEnvironment to have 'projectName' field")
		}

		en, ok := raw["envName"]
		if !ok {
			return fmt.Errorf("expected removeEnvironment to have 'envName' field")
		}

		*r = RemoveEnvironmentIdentifier{
			ProjectName: pn.(string),
			EnvName:     en.(string),
		}
	default:
		return fmt.Errorf("expected removeEnvironment to be a string or object")
	}

	return nil
}
