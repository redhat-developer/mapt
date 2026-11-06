// Copyright 2026, Pulumi Corporation.  All rights reserved.

package apitype

import (
	"encoding/json"

	jv "github.com/pulumi/pulumi-cloud-sdk/go/apitype/jsonvalue"
)

type dockerImageRequestJSON DockerImageRequest

func (d *DockerImageRequest) UnmarshalJSON(bytes []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(bytes, &raw); err == nil {
		err := rejectUnknownDockerImageKeys(
			"DockerImageRequest",
			raw,
			"reference",
			"isDefault",
			"credentials",
		)
		if err != nil {
			return err
		}
		var image dockerImageRequestJSON
		if err := json.Unmarshal(bytes, &image); err != nil {
			return err
		}
		d.Reference = image.Reference
		d.Credentials = image.Credentials
		return nil
	}

	var reference jv.Value[string]
	if err := json.Unmarshal(bytes, &reference); err != nil {
		return err
	}
	d.Reference = reference
	d.Credentials = jv.Null[DockerImageCredentialsRequest]()
	return nil
}

func (d *DockerImageRequest) MarshalJSON() ([]byte, error) {
	if !d.Credentials.Undefined() {
		return json.Marshal(dockerImageRequestJSON{
			Reference:   d.Reference,
			Credentials: d.Credentials,
		})
	}
	return json.Marshal(d.Reference)
}
