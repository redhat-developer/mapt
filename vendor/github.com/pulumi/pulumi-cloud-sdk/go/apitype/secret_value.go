// Copyright 2026, Pulumi Corporation.  All rights reserved.

package apitype

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

// A SecretValue describes a possibly-secret value.
//
// A value with Secret set to true may be represented as plaintext or ciphertext. A value with Secret set to false
// may only contain plaintext.
type SecretValue struct {
	Value      string // plaintext
	Ciphertext []byte // ciphertext
	Secret     bool
}

type secretWorkflowValue struct {
	Plaintext  string `json:"plaintext,omitempty"`
	Secret     string `json:"secret,omitempty"`
	Ciphertext []byte `json:"ciphertext,omitempty"`
}

type secretYamlWorkflowValue struct {
	Plaintext  string `yaml:"plaintext,omitempty"`
	Secret     string `yaml:"secret,omitempty"`
	Ciphertext string `yaml:"ciphertext,omitempty"`
}

func (v SecretValue) IsEmpty() bool {
	return len(strings.TrimSpace(v.Value)) == 0 && len(v.Ciphertext) == 0
}

func (v SecretValue) MarshalJSON() ([]byte, error) {
	switch {
	case len(v.Ciphertext) != 0 && v.Secret:
		return json.Marshal(secretWorkflowValue{Secret: v.Value, Ciphertext: v.Ciphertext})
	case len(v.Ciphertext) != 0:
		return json.Marshal(secretWorkflowValue{Ciphertext: v.Ciphertext})
	case v.Secret:
		return json.Marshal(secretWorkflowValue{Secret: v.Value})
	default:
		return json.Marshal(v.Value)
	}
}

func (v *SecretValue) UnmarshalJSON(bytes []byte) error {
	var secret secretWorkflowValue
	if err := json.Unmarshal(bytes, &secret); err == nil {
		if secret.Plaintext != "" {
			v.Value = secret.Plaintext
			v.Secret = false
			return nil
		}

		v.Value = secret.Secret
		v.Ciphertext = secret.Ciphertext
		v.Secret = true
		return nil
	}

	var plaintext string
	if err := json.Unmarshal(bytes, &plaintext); err != nil {
		return err
	}

	v.Value = plaintext
	v.Secret = false
	return nil
}

func (v SecretValue) MarshalYAML() (any, error) {
	switch {
	case len(v.Ciphertext) != 0:
		ciphertext := base64.StdEncoding.EncodeToString(v.Ciphertext)
		return secretYamlWorkflowValue{Ciphertext: ciphertext}, nil
	case v.Secret:
		return secretYamlWorkflowValue{Secret: v.Value}, nil
	default:
		return v.Value, nil
	}
}

func (v *SecretValue) UnmarshalYAML(node *yaml.Node) error {
	var secret secretYamlWorkflowValue
	if err := node.Decode(&secret); err == nil {
		ciphertext, err := base64.StdEncoding.DecodeString(secret.Ciphertext)
		if err != nil {
			return err
		}

		if secret.Plaintext != "" {
			v.Value = secret.Plaintext
			v.Secret = false
			return nil
		}

		v.Value = secret.Secret
		v.Ciphertext = ciphertext
		v.Secret = true
		return nil
	}

	var plaintext string
	if err := node.Decode(&plaintext); err != nil {
		return err
	}

	v.Value = plaintext
	v.Secret = false
	return nil
}
