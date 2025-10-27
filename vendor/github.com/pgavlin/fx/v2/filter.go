package fx

import "iter"

// Filter returns a sequence of values computed by invoking fn on each element of the input sequence and returning only
// those elements for with fn returns true.
func Filter[T any](it iter.Seq[T], fn func(v T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range it {
			if fn(v) {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// Filter2 returns a sequence of values computed by invoking fn on each element
// of the input slice and returning only those elements for with fn returns
// true.
func Filter2[T, U any](it iter.Seq2[T, U], fn func(t T, u U) bool) iter.Seq2[T, U] {
	return func(yield func(T, U) bool) {
		for t, u := range it {
			if fn(t, u) {
				if !yield(t, u) {
					return
				}
			}
		}
	}
}
