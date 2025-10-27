package fx

import "iter"

// A Pair is a pair of (possibly-differently) typed values.
type Pair[T, U any] struct {
	Fst T
	Snd U
}

// Unpack() extracts the contained values from the Pair.
func (p Pair[T, U]) Unpack() (T, U) {
	return p.Fst, p.Snd
}

// Pack creates a Pair from a pair of values.
func Pack[T, U any](fst T, snd U) Pair[T, U] {
	return Pair[T, U]{Fst: fst, Snd: snd}
}

// PackAll transforms a sequence of (K, V) pairs into a sequence of Pair[K, V] values.
func PackAll[K, V any](it iter.Seq2[K, V]) iter.Seq[Pair[K, V]] {
	return func(yield func(p Pair[K, V]) bool) {
		for k, v := range it {
			if !yield(Pack(k, v)) {
				return
			}
		}
	}
}

// UnpackAll transforms a sequence of Pair[K, V] values into a sequence of (K, V) pairs.
func UnpackAll[K, V any](it iter.Seq[Pair[K, V]]) iter.Seq2[K, V] {
	return func(yield func(k K, v V) bool) {
		for p := range it {
			if !yield(p.Unpack()) {
				return
			}
		}
	}
}
