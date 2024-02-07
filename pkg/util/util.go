package util

import (
	cRand "crypto/rand"
	"fmt"
	"math/rand"
	"strings"
)

func If[T any](cond bool, vtrue, vfalse T) T {
	if cond {
		return vtrue
	}
	return vfalse
}

func ArrayFilter[T any](source []T, filter func(item T) bool) []T {
	var result []T
	for _, item := range source {
		if filter(item) {
			result = append(result, item)
		}
	}
	return result
}

func ArrayCast[T any](source []interface{}) []T {
	var result []T
	for _, item := range source {
		result = append(result, item.(T))
	}
	return result
}

func ArrayConvert[T any, Y any](source []Y,
	convert func(item Y) T) []T {
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

// This function split a list based on a evaluation function
// the result is a map of lists
func SplitSlice[T any, Y comparable](source []T,
	identification func(item T) Y) (m map[Y][]T) {
	m = make(map[Y][]T)
	for _, item := range source {
		k := identification(item)
		if l, found := m[k]; found {
			m[k] = append(l, item)
		} else {
			m[k] = []T{item}
		}
	}
	return
}

func Average(source []float64) float64 {
	total := 0.0
	for _, v := range source {
		total += v
	}
	return total / float64(len(source))
}

func Max(source []float64) float64 {
	total := 0.0
	for _, v := range source {
		total += v
	}
	return total / float64(len(source))
}

func Random(max, min int) int {
	return rand.Intn(max-min+1) + min
}

func RandomItemFromArray[X any](source []X) X {
	return source[Random(len(source)-1, 0)]
}

func RandomID(name string) string {
	b := make([]byte, 4)
	_, _ = cRand.Read(b)
	return fmt.Sprintf("%s%x", name, b)
}
