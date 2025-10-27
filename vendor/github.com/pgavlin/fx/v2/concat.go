package fx

import (
	"iter"
	"slices"
)

// Concat returns an iterator that returns values from each iterator in sequence.
func Concat[T any](iters ...iter.Seq[T]) iter.Seq[T] {
	return ConcatMany(slices.Values(iters))
}

// ConcatMany returns an iterator that returns values from each iterator in sequence.
func ConcatMany[T any](iters iter.Seq[iter.Seq[T]]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for it := range iters {
			for v := range it {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// Concat2 returns an iterator that returns values from each iterator in sequence.
func Concat2[T, U any](iters ...iter.Seq2[T, U]) iter.Seq2[T, U] {
	return ConcatMany2(slices.Values(iters))
}

// ConcatMany2 returns an iterator that returns values from each iterator in sequence.
func ConcatMany2[T, U any](iters iter.Seq[iter.Seq2[T, U]]) iter.Seq2[T, U] {
	return func(yield func(T, U) bool) {
		for it := range iters {
			for t, u := range it {
				if !yield(t, u) {
					return
				}
			}
		}
	}
}
