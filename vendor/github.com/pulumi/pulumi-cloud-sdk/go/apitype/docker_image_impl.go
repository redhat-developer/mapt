// Copyright 2026, Pulumi Corporation.  All rights reserved.

package apitype

import "encoding/json"

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
	var image dockerImageJSON
	if err := json.Unmarshal(bytes, &image); err == nil {
		d.Reference, d.IsDefault, d.Credentials = image.Reference, image.IsDefault, image.Credentials
		return nil
	}

	var reference string
	if err := json.Unmarshal(bytes, &reference); err != nil {
		return err
	}
	d.Reference, d.IsDefault, d.Credentials = reference, false, nil
	return nil
}
