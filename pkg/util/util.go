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

func IfWithError[T any](cond bool, vtrue, vfalse func() (T, error)) (T, error) {
	if cond {
		return vtrue()
	}
	return vfalse()
}

// In case vtrue value depends on a variable checked on condition which could be nil
// as params are evaluated within the If function invokation it will produce a panic error
// in that case we will pass the vtrue as a function which will be evaluated only if condition is met
//
// i.e. If(foo != nil, foo.bar, "") In this case if foo is nill this will error with panic as the evaluation will try access foo which is nil
//
// so, in this case we will use IfNillable:
// bar = func() { return foo.bar}
// IfNillable(foo != nil, bar, "")
func IfNillable[T any](cond bool, vtrueNillable func() T, vfalse T) T {
	if cond {
		return vtrueNillable()
	}
	return vfalse
}

// Return a new array list with those items from source which
// return true from filter function
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
