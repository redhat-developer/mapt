package maps

import (
	"cmp"
	"iter"
	"maps"
	"slices"

	"github.com/pgavlin/fx/v2"
)

// Sorted returns an iterator over key-value pairs from m. Pairs are ordered by keys.
func Sorted[M ~map[K]V, K cmp.Ordered, V any](m M) iter.Seq2[K, V] {
	return fx.UnpackAll(SortedPairs(m))
}

// SortedPairs returns an iterator over key-value pairs from m. Pairs are ordered by keys.
func SortedPairs[M ~map[K]V, K cmp.Ordered, V any](m M) iter.Seq[fx.Pair[K, V]] {
	pairs := slices.Collect(fx.PackAll(maps.All(m)))
	slices.SortFunc(pairs, func(a, b fx.Pair[K, V]) int { return cmp.Compare(a.Fst, b.Fst) })
	return slices.Values(pairs)
}
