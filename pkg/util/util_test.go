package util

import (
	"reflect"
	"testing"
)

type identifier struct {
	a string
	b string
}

type data struct {
	a string
	b string
	c string
}

var items []data = []data{
	{a: "a", b: "b", c: "c"},
	{a: "a", b: "b", c: "d"},
	{a: "a", b: "foo", c: "c"},
	{a: "foo", b: "b", c: "c"}}

var expectedResult map[identifier][]data = map[identifier][]data{
	{a: "a", b: "b"}:   {{a: "a", b: "b", c: "c"}, {a: "a", b: "b", c: "d"}},
	{a: "a", b: "foo"}: {{a: "a", b: "foo", c: "c"}},
	{a: "foo", b: "b"}: {{a: "foo", b: "b", c: "c"}},
}

func TestSplitSlice(t *testing.T) {
	result := SplitSlice(items,
		func(item data) identifier {
			return identifier{
				a: item.a,
				b: item.b}
		})
	if !reflect.DeepEqual(result, expectedResult) {
		t.Fatalf("Split slice failed %v should match %v", result, expectedResult)
	}
}
