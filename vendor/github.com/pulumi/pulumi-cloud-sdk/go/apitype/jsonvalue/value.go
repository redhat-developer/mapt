// Copyright 2024-2026, Pulumi Corporation.  All rights reserved.

package jsonvalue

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

type value[T any] struct {
	t *T
}

// Value represents a JSON value of type T.
type Value[T any] []value[T]

// Null returns the JSON null value of type T.
func Null[T any]() Value[T] {
	return Value[T]{value[T]{}}
}

// ValueOf returns an explicitly-defined JSON value of type T. If t is nil, the result is the JSON null value.
func ValueOf[T any](t *T) Value[T] {
	return Value[T]{value[T]{t: t}}
}

// NotNull returns an explicitly-defined JSON value of type T.
func NotNull[T any](t T) Value[T] {
	return Value[T]{value[T]{t: &t}}
}

// Undefined returns true if the value is undefined.
func (v Value[T]) Undefined() bool {
	return len(v) == 0
}

// Null returns true if the value is the JSON null value.
func (v Value[T]) Null() bool {
	return !v.Undefined() && v[0].t == nil
}

// Value returns the underlying value (if any) and a boolean that indicates whether or not the value is defined.
func (v Value[T]) Value() (*T, bool) {
	if v.Undefined() {
		return nil, false
	}
	return v[0].t, true
}

func (v Value[T]) ValueOrDefault() T {
	var t T
	if vv, ok := v.Value(); ok && vv != nil {
		t = *vv
	}
	return t
}

// Reflect returns a tuple of (value, null, undefined).
func (v Value[T]) Reflect() (any, bool, bool) {
	if v.Undefined() {
		return nil, false, true
	}
	return v[0].t, v.Null(), v.Undefined()
}

func (v Value[T]) MarshalJSON() ([]byte, error) {
	if v.Undefined() {
		return []byte("null"), nil
	}
	return json.Marshal(v[0].t)
}

func (v *Value[T]) UnmarshalJSON(bytes []byte) error {
	if len(*v) == 0 {
		*v = append(*v, value[T]{})
	}
	if err := json.Unmarshal(bytes, &(*v)[0].t); err != nil {
		return err
	}
	return nil
}

// Yaml decoder implementation differs from the JSON one by shortcircuiting on null values
// here https://github.com/go-yaml/yaml/blob/f6f7691b1fdeb513f56608cd2c32c51f8194bf51/decode.go#L407-L409
// causing it to never invoke the UnmarshalYAML interface making it behave as an undefined value
// and this test to fails:
//
// expected: serialization.Value[bool]{serialization.value[bool]{t:(*bool)(nil)}}
// actual  : serialization.Value[bool](nil)
func (v Value[T]) MarshalYAML() (any, error) {
	if v.Undefined() {
		return nil, nil
	}

	return v[0].t, nil
}

func (v *Value[T]) UnmarshalYAML(n *yaml.Node) error {
	if len(*v) == 0 {
		*v = append(*v, value[T]{})
	}

	if err := n.Decode(&(*v)[0].t); err != nil {
		return err
	}
	return nil
}
