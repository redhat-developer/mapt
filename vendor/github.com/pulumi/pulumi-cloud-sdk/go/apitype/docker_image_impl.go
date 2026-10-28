// Copyright 2026, Pulumi Corporation.  All rights reserved.

package apitype

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

type dockerImageJSON struct {
	Reference   string                  `json:"reference" yaml:"reference"`
	IsDefault   bool                    `json:"isDefault,omitempty" yaml:"isDefault,omitempty"`
	Credentials *DockerImageCredentials `json:"credentials,omitempty" yaml:"credentials,omitempty"`
}

func (d *DockerImage) MarshalJSON() ([]byte, error) {
	// If only Reference set, we can just marshal it as a string.
	if d.Credentials != nil || d.IsDefault {
		return json.Marshal(dockerImageJSON{
			Reference:   d.Reference,
			IsDefault:   d.IsDefault,
			Credentials: d.Credentials,
		})
	}
	return json.Marshal(d.Reference)
}

func (d *DockerImage) UnmarshalJSON(bytes []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(bytes, &raw); err == nil {
		if err := rejectUnknownDockerImageKeys("DockerImage", raw, "reference", "isDefault", "credentials"); err != nil {
			return err
		}
		var image dockerImageJSON
		if err := json.Unmarshal(bytes, &image); err != nil {
			return err
		}
		d.Reference = image.Reference
		d.IsDefault = image.IsDefault
		d.Credentials = image.Credentials
		return nil
	}

	var reference string
	if err := json.Unmarshal(bytes, &reference); err != nil {
		return err
	}
	d.Reference = reference
	d.IsDefault = false
	d.Credentials = nil
	return nil
}

func rejectUnknownDockerImageKeys(typeName string, raw map[string]json.RawMessage, known ...string) error {
	for name := range raw {
		if !slices.ContainsFunc(known, func(key string) bool { return strings.EqualFold(name, key) }) {
			return fmt.Errorf("invalid %s: unknown key %q (expected one of '%s')",
				typeName, name, strings.Join(known, "', '"))
		}
	}
	return nil
}
