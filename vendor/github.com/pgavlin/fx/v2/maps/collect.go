package maps

import (
	"iter"

	"github.com/pgavlin/fx/v2"
)

// Pairs returns a sequence of (key, value) entries in m represented as fx.Pairs.
//
// Equivalent to fx.PackAll(maps.All(m)).
func Pairs[M ~map[K]V, K comparable, V any](m M) iter.Seq[fx.Pair[K, V]] {
	return func(yield func(kvp fx.Pair[K, V]) bool) {
		for k, v := range m {
			if !yield(fx.Pack(k, v)) {
				return
			}
		}
	}
}

// TryCollect attempts to collect the key-value pairs in the input sequence into a map. If any pair in the input contains a
// non-nil error, TryCollect halts and returns the collected map up to that point (excluding the value returned with the error)
// and the error.
func TryCollect[K comparable, V any](it iter.Seq2[fx.Pair[K, V], error]) (map[K]V, error) {
	var m map[K]V
	for kvp, err := range it {
		if err != nil {
			return m, err
		}
		if m == nil {
			m = make(map[K]V)
		}
		m[kvp.Fst] = kvp.Snd
	}
	return m, nil
}
