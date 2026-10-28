// Copyright 2026, Pulumi Corporation.  All rights reserved.

package apitype

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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

// secretExplicitValue is the object form the pulumi CLI decodes under API v9. Accepted on
// read, never emitted.
type secretExplicitValue struct {
	IsSecret   bool   `json:"isSecret"`
	Value      string `json:"value,omitempty"`
	Ciphertext []byte `json:"ciphertext,omitempty"`
}

type secretYamlExplicitValue struct {
	IsSecret   bool   `yaml:"isSecret"`
	Value      string `yaml:"value,omitempty"`
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

// secretValueKeys records which object keys carry a non-null value; null counts as absent
// because some clients serialize unset fields as null.
type secretValueKeys struct {
	isSecret   bool
	value      bool
	secret     bool
	ciphertext bool
	plaintext  bool
}

// foldCase matches encoding/json's case-insensitive keys; yaml.v3 matches exactly, so a
// folded YAML key would pass the check and then go undecoded.
func (k *secretValueKeys) add(name string, isNull bool, foldCase bool) error {
	match := func(key string) bool {
		if foldCase {
			return strings.EqualFold(name, key)
		}
		return name == key
	}
	var field *bool
	switch {
	case match("isSecret"):
		field = &k.isSecret
	case match("value"):
		field = &k.value
	case match("secret"):
		field = &k.secret
	case match("ciphertext"):
		field = &k.ciphertext
	case match("plaintext"):
		field = &k.plaintext
	default:
		return fmt.Errorf("invalid SecretValue: unknown key %q (expected one of "+
			"'isSecret', 'value', 'secret', 'ciphertext', 'plaintext')", name)
	}
	if !isNull {
		*field = true
	}
	return nil
}

func (k secretValueKeys) validate() error {
	switch {
	case k.isSecret && (k.secret || k.plaintext):
		return errors.New("invalid SecretValue: 'isSecret' cannot be combined with 'secret' or 'plaintext'")
	case k.value && !k.isSecret:
		return errors.New("invalid SecretValue: 'value' requires 'isSecret'")
	}
	return nil
}

// An object with no content is an empty secret, which is what older MarshalJSON emitted for
// one, unless `plaintext` was given explicitly.
func (v *SecretValue) setLegacy(plaintext string, secret string, ciphertext []byte, plaintextPresent bool) error {
	if len(ciphertext) == 0 {
		ciphertext = nil
	}
	hasSecretContent := secret != "" || ciphertext != nil
	switch {
	case plaintext != "" && hasSecretContent:
		return errors.New("invalid SecretValue: 'plaintext' cannot be combined with 'secret' or 'ciphertext'")
	case plaintext != "" || (plaintextPresent && !hasSecretContent):
		v.Value = plaintext
		v.Ciphertext = nil
		v.Secret = false
	default:
		v.Value = secret
		v.Ciphertext = ciphertext
		v.Secret = true
	}
	return nil
}

func (v *SecretValue) setExplicit(isSecret bool, value string, ciphertext []byte) error {
	if len(ciphertext) == 0 {
		ciphertext = nil
	}
	if !isSecret && ciphertext != nil {
		return errors.New("invalid SecretValue: 'ciphertext' requires 'isSecret: true'")
	}
	v.Value = value
	v.Ciphertext = ciphertext
	v.Secret = isSecret
	return nil
}

// Keys are checked before decoding because the struct decoders ignore unknown keys, which
// turned an explicit object or a mistyped key into an empty secret.
func (v *SecretValue) UnmarshalJSON(bytes []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(bytes, &raw); err != nil {
		var plaintext string
		if err := json.Unmarshal(bytes, &plaintext); err != nil {
			return err
		}
		v.Value = plaintext
		v.Ciphertext = nil
		v.Secret = false
		return nil
	}

	var keys secretValueKeys
	for name, value := range raw {
		if err := keys.add(name, string(value) == "null", true); err != nil {
			return err
		}
	}
	if err := keys.validate(); err != nil {
		return err
	}

	if keys.isSecret {
		var explicit secretExplicitValue
		if err := json.Unmarshal(bytes, &explicit); err != nil {
			return err
		}
		return v.setExplicit(explicit.IsSecret, explicit.Value, explicit.Ciphertext)
	}

	var legacy secretWorkflowValue
	if err := json.Unmarshal(bytes, &legacy); err != nil {
		return err
	}
	return v.setLegacy(legacy.Plaintext, legacy.Secret, legacy.Ciphertext, keys.plaintext)
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
	if node.Kind != yaml.MappingNode {
		var plaintext string
		if err := node.Decode(&plaintext); err != nil {
			return err
		}
		v.Value = plaintext
		v.Ciphertext = nil
		v.Secret = false
		return nil
	}

	// Decoding into a map resolves aliases and `<<` merge keys before the key check.
	var fields map[string]yaml.Node
	if err := node.Decode(&fields); err != nil {
		return err
	}
	var keys secretValueKeys
	for name, value := range fields {
		if err := keys.add(name, value.ShortTag() == "!!null", false); err != nil {
			return err
		}
	}
	if err := keys.validate(); err != nil {
		return err
	}

	if keys.isSecret {
		var explicit secretYamlExplicitValue
		if err := node.Decode(&explicit); err != nil {
			return err
		}
		ciphertext, err := decodeYAMLCiphertext(explicit.Ciphertext)
		if err != nil {
			return err
		}
		return v.setExplicit(explicit.IsSecret, explicit.Value, ciphertext)
	}

	var legacy secretYamlWorkflowValue
	if err := node.Decode(&legacy); err != nil {
		return err
	}
	ciphertext, err := decodeYAMLCiphertext(legacy.Ciphertext)
	if err != nil {
		return err
	}
	return v.setLegacy(legacy.Plaintext, legacy.Secret, ciphertext, keys.plaintext)
}

func decodeYAMLCiphertext(encoded string) ([]byte, error) {
	if encoded == "" {
		return nil, nil
	}
	return base64.StdEncoding.DecodeString(encoded)
}
