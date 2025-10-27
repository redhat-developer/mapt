package fx

import "iter"

// OfType returns a sequence composed of all elements in the input sequence that are of type U.
func OfType[U, T any](it iter.Seq[T]) iter.Seq[U] {
	return FMap(it, func(v T) (U, bool) {
		u, ok := any(v).(U)
		return u, ok
	})
}

// OfType2 returns a sequence composed of all elements in the input sequence where the second value is of type U.
func OfType2[U, K, T any](it iter.Seq2[K, T]) iter.Seq2[K, U] {
	return FMap2(it, func(k K, v T) (K, U, bool) {
		u, ok := any(v).(U)
		return k, u, ok
	})
}

// FMap returns a sequence of values computed by invoking fn on each element of the input sequence and returning only
// mapped values for with fn returns true.
func FMap[T, U any](it iter.Seq[T], fn func(v T) (U, bool)) iter.Seq[U] {
	return func(yield func(v U) bool) {
		for v := range it {
			if u, ok := fn(v); ok {
				if !yield(u) {
					return
				}
			}
		}
	}
}

// FMapUnpack returns a sequence of values computed by invoking fn on each element of the input sequence and returning only
// mapped values for with fn returns true.
func FMapUnpack[T, U, V any](it iter.Seq[T], fn func(v T) (U, V, bool)) iter.Seq2[U, V] {
	return func(yield func(U, V) bool) {
		for t := range it {
			if u, v, ok := fn(t); ok {
				if !yield(u, v) {
					return
				}
			}
		}
	}
}

// FMap2 returns a sequence of values computed by invoking fn on each element of the input slice and returning only mapped
// values for with fn returns true.
func FMap2[T, U, V, W any](it iter.Seq2[T, U], fn func(t T, u U) (V, W, bool)) iter.Seq2[V, W] {
	return func(yield func(V, W) bool) {
		for t, u := range it {
			if l, w, ok := fn(t, u); ok {
				if !yield(l, w) {
					return
				}
			}
		}
	}
}

// FMap2Pack returns a sequence of values computed by invoking fn on each element of the input slice and returning only mapped
// values for with fn returns true.
func FMap2Pack[T, U, V any](it iter.Seq2[T, U], fn func(t T, u U) (V, bool)) iter.Seq[V] {
	return func(yield func(V) bool) {
		for t, u := range it {
			if v, ok := fn(t, u); ok {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// Map invokes fn on each value in the input sequence and returns the results.
func Map[T, U any](it iter.Seq[T], fn func(v T) U) iter.Seq[U] {
	return func(yield func(v U) bool) {
		for v := range it {
			if !yield(fn(v)) {
				return
			}
		}
	}
}

// MapUnpack invokes fn on each value in the input sequence and returns the results.
func MapUnpack[T, U, V any](it iter.Seq[T], fn func(v T) (U, V)) iter.Seq2[U, V] {
	return func(yield func(U, V) bool) {
		for v := range it {
			if !yield(fn(v)) {
				return
			}
		}
	}
}

// Map2 invokes fn on each value in the input slice and returns the results.
func Map2[T, U, V, W any](it iter.Seq2[T, U], fn func(t T, u U) (V, W)) iter.Seq2[V, W] {
	return func(yield func(V, W) bool) {
		for t, u := range it {
			if !yield(fn(t, u)) {
				return
			}
		}
	}
}

// Map2Pack invokes fn on each value in the input slice and returns the results.
func Map2Pack[T, U, V any](it iter.Seq2[T, U], fn func(t T, u U) V) iter.Seq[V] {
	return func(yield func(V) bool) {
		for t, u := range it {
			if !yield(fn(t, u)) {
				return
			}
		}
	}
}
