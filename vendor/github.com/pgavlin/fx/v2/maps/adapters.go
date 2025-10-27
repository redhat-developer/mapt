package maps

import (
	"iter"
	"maps"

	"github.com/pgavlin/fx/v2"
)

// All returns true if pred returns true for every element of the input slice.
func All[M ~map[T]U, T comparable, U any](m M, pred func(t T, u U) bool) bool {
	return fx.All2(maps.All(m), pred)
}

// Any returns true if pred returns true for any element of the input slice.
func Any[M ~map[T]U, T comparable, U any](m M, pred func(t T, u U) bool) bool {
	return fx.Any2(maps.All(m), pred)
}

// FMap returns a sequence of values computed by invoking fn on each element of the input slice and returning only mapped
// values for with fn returns true.
func FMap[M ~map[T]U, T comparable, U any, V any, W any](m M, fn func(t T, u U) (V, W, bool)) iter.Seq2[V, W] {
	return fx.FMap2(maps.All(m), fn)
}

// FMapPack returns a sequence of values computed by invoking fn on each element of the input slice and returning only mapped
// values for with fn returns true.
func FMapPack[M ~map[T]U, T comparable, U any, V any](m M, fn func(t T, u U) (V, bool)) iter.Seq[V] {
	return fx.FMap2Pack(maps.All(m), fn)
}

// Filter returns a sequence of values computed by invoking fn on each element
// of the input slice and returning only those elements for with fn returns
// true.
func Filter[M ~map[T]U, T comparable, U any](m M, fn func(t T, u U) bool) iter.Seq2[T, U] {
	return fx.Filter2(maps.All(m), fn)
}

// First returns the first element of it, if any elements exist.
func First[M ~map[T]U, T comparable, U any](m M) (T, U, bool) {
	return fx.First2(maps.All(m))
}

// Last returns the last element of it, if any elements exist.
func Last[M ~map[T]U, T comparable, U any](m M) (T, U, bool) {
	return fx.Last2(maps.All(m))
}

// Map invokes fn on each value in the input slice and returns the results.
func Map[M ~map[T]U, T comparable, U any, V any, W any](m M, fn func(t T, u U) (V, W)) iter.Seq2[V, W] {
	return fx.Map2(maps.All(m), fn)
}

// MapPack invokes fn on each value in the input slice and returns the results.
func MapPack[M ~map[T]U, T comparable, U any, V any](m M, fn func(t T, u U) V) iter.Seq[V] {
	return fx.Map2Pack(maps.All(m), fn)
}

// OfType returns a sequence composed of all elements in the input sequence where the second value is of type U.
func OfType[U any, M ~map[K]T, K comparable, T any](m M) iter.Seq2[K, U] {
	return fx.OfType2[U](maps.All(m))
}

// Reduce calls fn on each element of the input slice, passing in the
// current value of the accumulator with each invocation and updating the
// accumulator to the result of fn after each invocation.
func Reduce[M ~map[T]U, T comparable, U any, V any](m M, init V, fn func(acc V, t T, u U) V) V {
	return fx.Reduce2(maps.All(m), init, fn)
}

// Skip returns an iterator that skips n values from its source.
func Skip[M ~map[T]U, T comparable, U any](m M, n int) iter.Seq2[T, U] {
	return fx.Skip2(maps.All(m), n)
}

// Take returns an iterator that takes at most n values from the input slice.
func Take[M ~map[T]U, T comparable, U any](m M, n int) iter.Seq2[T, U] {
	return fx.Take2(maps.All(m), n)
}
