// Copyright 2025-2026, Pulumi Corporation.  All rights reserved.

package apitype

import (
	"encoding/json"
)

// OpenAPIName returns the OpenAPI name of certain generated types.
func OpenAPIName(object interface{ openapiName() string }) string {
	return object.openapiName()
}

type teamNameSlice []string

func (teams *teamNameSlice) MarshalCSV() (string, error) {
	if teams == nil || len(*teams) == 0 {
		return "", nil
	}

	json, err := json.Marshal(*teams)
	if err != nil {
		return "", err
	}

	return string(json), nil
}

// RawProperty wraps a json.RawProperty for JSON or CSV export.
type RawProperty struct {
	json.RawMessage
}

// String provides compatibility with CSV export.
func (r RawProperty) String() string {
	return string(r.RawMessage)
}

type ResponseWithHeaders[R any, H any] struct {
	Response R
	Headers  H
}
