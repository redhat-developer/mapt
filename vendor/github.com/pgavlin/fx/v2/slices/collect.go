package slices

import "iter"

// TryCollect attempts to collect the values in the input sequence into a slice. If any pair in the input contains a
// non-nil error, TryCollect halts and returns the values collected up to that point (excluding the value returned
// with the error) and the error.
func TryCollect[T any](it iter.Seq2[T, error]) ([]T, error) {
	var s []T
	for t, err := range it {
		if err != nil {
			return s, err
		}
		s = append(s, t)
	}
	return s, nil
}
