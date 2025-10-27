package fx

import "iter"

// Only returns a sequence that contains the single value v.
func Only[T any](v T) iter.Seq[T] {
	return func(yield func(T) bool) {
		yield(v)
	}
}

// Only2 returns a sequence that contains the single value v.
func Only2[T, U any](t T, u U) iter.Seq2[T, U] {
	return func(yield func(T, U) bool) {
		yield(t, u)
	}
}

// Empty returns an empty sequence.
func Empty[T any]() iter.Seq[T] {
	return func(_ func(T) bool) {}
}

// Empty2 returns an empty sequence.
func Empty2[T, U any]() iter.Seq2[T, U] {
	return func(_ func(T, U) bool) {}
}

// Enumerate returns a sequence of (index, value) pairs from the given length and element accessor.
func Enumerate[T any](len func() int, item func(int) T) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i := range len() {
			if !yield(i, item(i)) {
				return
			}
		}
	}
}

// Values returns a sequence of values from the given length and element accessor.
func Values[T any](len func() int, item func(int) T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := range len() {
			if !yield(item(i)) {
				return
			}
		}
	}
}
