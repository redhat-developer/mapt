package slices

import (
	"iter"
	"slices"

	"github.com/pgavlin/fx/v2"
)

// All returns true if pred returns true for every element of the input sequence.
func All[S ~[]T, T any](s S, pred func(v T) bool) bool {
	return fx.All(slices.Values(s), pred)
}

// Any returns true if pred returns true for any element of the input sequence.
func Any[S ~[]T, T any](s S, pred func(v T) bool) bool {
	return fx.Any(slices.Values(s), pred)
}

// Contains returns true if the input sequence contains t.
func Contains[S ~[]T, T comparable](s S, t T) bool {
	return fx.Contains(slices.Values(s), t)
}

// FMap returns a sequence of values computed by invoking fn on each element of the input sequence and returning only
// mapped values for with fn returns true.
func FMap[S ~[]T, T any, U any](s S, fn func(v T) (U, bool)) iter.Seq[U] {
	return fx.FMap(slices.Values(s), fn)
}

// FMapUnpack returns a sequence of values computed by invoking fn on each element of the input sequence and returning only
// mapped values for with fn returns true.
func FMapUnpack[S ~[]T, T any, U any, V any](s S, fn func(v T) (U, V, bool)) iter.Seq2[U, V] {
	return fx.FMapUnpack(slices.Values(s), fn)
}

// Filter returns a sequence of values computed by invoking fn on each element of the input sequence and returning only
// those elements for with fn returns true.
func Filter[S ~[]T, T any](s S, fn func(v T) bool) iter.Seq[T] {
	return fx.Filter(slices.Values(s), fn)
}

// First returns the first element of it, if any elements exist.
func First[S ~[]T, T any](s S) (T, bool) {
	return fx.First(slices.Values(s))
}

// Last returns the last element of it, if any elements exist.
func Last[S ~[]T, T any](s S) (T, bool) {
	return fx.Last(slices.Values(s))
}

// Map invokes fn on each value in the input sequence and returns the results.
func Map[S ~[]T, T any, U any](s S, fn func(v T) U) iter.Seq[U] {
	return fx.Map(slices.Values(s), fn)
}

// MapUnpack invokes fn on each value in the input sequence and returns the results.
func MapUnpack[S ~[]T, T any, U any, V any](s S, fn func(v T) (U, V)) iter.Seq2[U, V] {
	return fx.MapUnpack(slices.Values(s), fn)
}

// OfType returns a sequence composed of all elements in the input sequence that are of type U.
func OfType[U any, S ~[]T, T any](s S) iter.Seq[U] {
	return fx.OfType[U](slices.Values(s))
}

// Reduce calls fn on each element of the input sequence, passing in the current value of the accumulator with
// each invocation and updating the accumulator to the result of fn after each invocation.
func Reduce[S ~[]T, T any, U any](s S, init U, fn func(acc U, v T) U) U {
	return fx.Reduce(slices.Values(s), init, fn)
}

// Skip returns an iterator that skips n values from its source.
func Skip[S ~[]T, T any](s S, n int) iter.Seq[T] {
	return fx.Skip(slices.Values(s), n)
}

// Take returns an iterator that takes at most n values from its source.
func Take[S ~[]T, T any](s S, n int) iter.Seq[T] {
	return fx.Take(slices.Values(s), n)
}
