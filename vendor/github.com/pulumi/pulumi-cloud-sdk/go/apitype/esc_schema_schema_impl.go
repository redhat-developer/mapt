// Copyright 2026, Pulumi Corporation.  All rights reserved.

package apitype

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math/big"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

func Never() *EscSchemaSchema {
	return &EscSchemaSchema{Never: true}
}

func Always() *EscSchemaSchema {
	return &EscSchemaSchema{Always: true}
}

func Ref(ref string) *EscSchemaSchema {
	return &EscSchemaSchema{Ref: ref}
}

func (s *EscSchemaSchema) UnmarshalJSON(data []byte) error {
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		if b {
			s.Always = true
			return nil
		}
		s.Never = true
		return nil
	}

	type rawSchema EscSchemaSchema
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	return dec.Decode((*rawSchema)(s))
}

func (s *EscSchemaSchema) MarshalJSON() ([]byte, error) {
	switch {
	case s.Never:
		return []byte("false"), nil
	case s.Always:
		return []byte("true"), nil
	default:
		type rawSchema EscSchemaSchema
		return json.Marshal((*rawSchema)(s))
	}
}

func (s *EscSchemaSchema) Schema() *EscSchemaSchema {
	return s
}

func (s *EscSchemaSchema) arrayItem(index int) *EscSchemaSchema {
	if s.Type != "array" {
		return Never()
	}
	if index < len(s.PrefixItems) {
		return s.PrefixItems[index]
	}
	return s.Items
}

func (s *EscSchemaSchema) Item(index int) *EscSchemaSchema {
	oneOf := make([]*EscSchemaSchema, 0, len(s.AnyOf)+len(s.OneOf)+1)
	for _, x := range s.AnyOf {
		oneOf = append(oneOf, orNever(x).Item(index))
	}
	for _, x := range s.OneOf {
		oneOf = append(oneOf, orNever(x).Item(index))
	}
	oneOf = append(oneOf, s.arrayItem(index))
	return union(oneOf)
}

func (s *EscSchemaSchema) objectProperty(name string) *EscSchemaSchema {
	if s.Type != "object" {
		return Never()
	}
	if p, ok := s.Properties[name]; ok {
		return p
	}
	return s.AdditionalProperties
}

func (s *EscSchemaSchema) Property(name string) *EscSchemaSchema {
	oneOf := make([]*EscSchemaSchema, 0, len(s.AnyOf)+len(s.OneOf)+1)
	for _, x := range s.AnyOf {
		oneOf = append(oneOf, orNever(x).Property(name))
	}
	for _, x := range s.OneOf {
		oneOf = append(oneOf, orNever(x).Property(name))
	}
	oneOf = append(oneOf, s.objectProperty(name))
	return union(oneOf)
}

func orNever(s *EscSchemaSchema) *EscSchemaSchema {
	if s == nil {
		return Never()
	}
	return s
}

func (s *EscSchemaSchema) IsRotateOnly() bool {
	return s.rotateOnly
}

func (s *EscSchemaSchema) GetRef() *EscSchemaSchema        { return s.ref }
func (s *EscSchemaSchema) GetMultipleOf() *big.Float       { return s.multipleOf }
func (s *EscSchemaSchema) GetMaximum() *big.Float          { return s.maximum }
func (s *EscSchemaSchema) GetExclusiveMaximum() *big.Float { return s.exclusiveMaximum }
func (s *EscSchemaSchema) GetMinimum() *big.Float          { return s.minimum }
func (s *EscSchemaSchema) GetExclusiveMinimum() *big.Float { return s.exclusiveMinimum }
func (s *EscSchemaSchema) GetMaxLength() *uint             { return s.maxLength }
func (s *EscSchemaSchema) GetMinLength() *uint             { return s.minLength }
func (s *EscSchemaSchema) GetPattern() *regexp.Regexp      { return s.pattern }
func (s *EscSchemaSchema) GetMaxItems() *uint              { return s.maxItems }
func (s *EscSchemaSchema) GetMinItems() *uint              { return s.minItems }
func (s *EscSchemaSchema) GetMaxProperties() *uint         { return s.maxProperties }
func (s *EscSchemaSchema) GetMinProperties() *uint         { return s.minProperties }

func (s *EscSchemaSchema) Compile() error {
	if s == nil || s.compiled {
		return nil
	}

	return s.compile(s)
}

// setRotateOnly transitively sets the rotateOnly flag on the input schema, making a
// copy of every node it touches. memo maps each original node to its rotateOnly copy
// so a cyclic $ref graph (e.g. a $defs entry that references itself) terminates
// instead of recursing forever: a re-entry guard on the original node alone isn't
// enough, since rotateOnly is only ever set on the copies, never on the originals.
func setRotateOnly(s *EscSchemaSchema) *EscSchemaSchema {
	return setRotateOnlyMemo(s, make(map[*EscSchemaSchema]*EscSchemaSchema))
}

func setRotateOnlyMemo(s *EscSchemaSchema, memo map[*EscSchemaSchema]*EscSchemaSchema) *EscSchemaSchema {
	if s == nil || s.rotateOnly {
		return s
	}
	if result, ok := memo[s]; ok {
		return result
	}

	result := new(EscSchemaSchema)
	*result = *s
	result.rotateOnly = true
	memo[s] = result

	rotateSlice := func(v []*EscSchemaSchema) []*EscSchemaSchema {
		var res []*EscSchemaSchema
		if v != nil {
			res = make([]*EscSchemaSchema, len(v))
			for i, schema := range v {
				res[i] = setRotateOnlyMemo(schema, memo)
			}
		}
		return res
	}

	rotateMap := func(v map[string]*EscSchemaSchema) map[string]*EscSchemaSchema {
		var res map[string]*EscSchemaSchema
		if v != nil {
			res = make(map[string]*EscSchemaSchema)
			for k, schema := range v {
				res[k] = setRotateOnlyMemo(schema, memo)
			}
		}
		return res
	}

	result.ref = setRotateOnlyMemo(s.ref, memo)
	result.AnyOf = rotateSlice(s.AnyOf)
	result.OneOf = rotateSlice(s.OneOf)
	result.PrefixItems = rotateSlice(s.PrefixItems)
	result.Items = setRotateOnlyMemo(s.Items, memo)
	result.AdditionalProperties = setRotateOnlyMemo(s.AdditionalProperties, memo)
	result.Properties = rotateMap(s.Properties)
	result.RotateOnly = slices.Collect(maps.Keys(s.Properties))

	return result
}

func (s *EscSchemaSchema) compile(root *EscSchemaSchema) error {
	if s == nil || s.compiled {
		return nil
	}
	s.compiled = true

	var err error
	if s.Ref != "" {
		if s.ref, err = parseRef(root, s.Ref); err != nil {
			return err
		}
		if err = s.ref.compile(root); err != nil {
			return err
		}
	}

	for _, s := range s.AnyOf {
		if err := s.compile(root); err != nil {
			return err
		}
	}
	for _, s := range s.OneOf {
		if err := s.compile(root); err != nil {
			return err
		}
	}

	for _, s := range s.PrefixItems {
		if err := s.compile(root); err != nil {
			return err
		}
	}
	if err := s.Items.compile(root); err != nil {
		return err
	}
	if err := s.AdditionalProperties.compile(root); err != nil {
		return err
	}
	for _, v := range s.Properties {
		if err := v.compile(root); err != nil {
			return err
		}
	}
	for _, name := range s.RotateOnly {
		// need to push the rotateOnly flag down onto the actual properties, so it is available to the evaluator while
		// evaluating object properties
		if p, ok := s.Properties[name]; ok {
			s.Properties[name] = setRotateOnly(p)
		}
	}

	if s.multipleOf, err = parseNumber(s.MultipleOf); err != nil {
		return err
	}
	if s.maximum, err = parseNumber(s.Maximum); err != nil {
		return err
	}
	if s.exclusiveMaximum, err = parseNumber(s.ExclusiveMaximum); err != nil {
		return err
	}
	if s.minimum, err = parseNumber(s.Minimum); err != nil {
		return err
	}
	if s.exclusiveMinimum, err = parseNumber(s.ExclusiveMinimum); err != nil {
		return err
	}
	if s.maxLength, err = parseUint(s.MaxLength); err != nil {
		return err
	}
	if s.minLength, err = parseUint(s.MinLength); err != nil {
		return err
	}
	if s.pattern, err = parseRegexp(s.Pattern); err != nil {
		return err
	}
	if s.maxItems, err = parseUint(s.MaxItems); err != nil {
		return err
	}
	if s.minItems, err = parseUint(s.MinItems); err != nil {
		return err
	}
	if s.maxProperties, err = parseUint(s.MaxProperties); err != nil {
		return err
	}
	if s.minProperties, err = parseUint(s.MinProperties); err != nil {
		return err
	}

	return nil
}

func parseRef(root *EscSchemaSchema, ref string) (*EscSchemaSchema, error) {
	refName, ok := strings.CutPrefix(ref, "#/$defs/")
	if !ok || strings.Contains(refName, "/") {
		return nil, errors.New("only fragment references of the form #/$defs/ref are supported")
	}

	refName, err := url.PathUnescape(refName)
	if err != nil {
		return nil, err
	}

	s, ok := root.Defs[refName]
	if !ok {
		return nil, fmt.Errorf("unknown subschema %v", ref)
	}
	return s, nil
}

func parseNumber(n json.Number) (*big.Float, error) {
	if n == "" {
		return nil, nil
	}
	f, _, err := big.ParseFloat(string(n), 10, 0, big.ToNearestEven)
	return f, err
}

func parseUint(n json.Number) (*uint, error) {
	if n == "" {
		return nil, nil
	}
	v64, err := strconv.ParseUint(string(n), 10, 0)
	if err != nil {
		return nil, err
	}
	v := uint(v64)
	return &v, nil
}

func parseRegexp(pattern string) (*regexp.Regexp, error) {
	if pattern == "" {
		return nil, nil
	}
	return regexp.Compile(pattern)
}

func union(oneOf []*EscSchemaSchema) *EscSchemaSchema {
	// Filter out Never schemas.
	n := 0
	rotateOnly := true
	for _, s := range oneOf {
		if s != nil && !s.Never {
			oneOf[n] = s
			rotateOnly = rotateOnly && s.rotateOnly
			n++
		}
	}
	oneOf = oneOf[:n]

	switch len(oneOf) {
	case 0:
		// If there are no schemas left, return Never.
		return Never()
	case 1:
		// If there is one schema left, return is.
		return oneOf[0]
	default:
		// Otherwise, return a OneOf.
		return &EscSchemaSchema{OneOf: oneOf, rotateOnly: rotateOnly}
	}
}
