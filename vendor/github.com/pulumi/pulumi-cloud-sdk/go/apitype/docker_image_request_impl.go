// Copyright 2026, Pulumi Corporation.  All rights reserved.

package apitype

import (
	"encoding/json"

	jv "github.com/pulumi/pulumi-cloud-sdk/go/apitype/jsonvalue"
)

type dockerImageRequestJSON DockerImageRequest

func (d *DockerImageRequest) UnmarshalJSON(bytes []byte) error {
	var image dockerImageRequestJSON
	if err := json.Unmarshal(bytes, &image); err == nil {
		d.Reference, d.Credentials = image.Reference, image.Credentials
		return err
	}

	var reference jv.Value[string]
	if err := json.Unmarshal(bytes, &reference); err != nil {
		return err
	}
	d.Reference, d.Credentials = reference, jv.Null[DockerImageCredentialsRequest]()
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
