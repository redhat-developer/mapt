package util

import (
	"strings"
)

func If[T any](cond bool, vtrue, vfalse T) T {
	if cond {
		return vtrue
	}
	return vfalse
}

func ArrayCast[T any](source []interface{}) []T {
	var result []T
	for _, item := range source {
		result = append(result, item.(T))
	}
	return result
}

func ArrayConvert[T any, X any](source []*X,
	convert func(item *X) T) []T {
	var result []T
	for _, item := range source {
		result = append(result, convert(item))
	}
	return result
}

func SplitString(source, delimiter string) []string {
	splitted := strings.Split(source, delimiter)
	return If(len(splitted[0]) > 0, splitted, []string{})
}
