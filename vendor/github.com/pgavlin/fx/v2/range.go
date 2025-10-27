package fx

import "iter"

// Range returns a sequence of each integer in the range [min, max).
func Range(min, max int) iter.Seq[int] {
	if min == 0 {
		return MaxRange(max)
	}
	return func(yield func(v int) bool) {
		for ; min < max; min++ {
			if !yield(min) {
				return
			}
		}
	}
}

// MaxRange returns a sequence of each integer in the range [0, max)
func MaxRange(max int) iter.Seq[int] {
	return func(yield func(v int) bool) {
		for n := range max {
			if !yield(n) {
				return
			}
		}
	}
}

// MinRange returns a sequence of each integer in the range [min, ∞)
func MinRange(min int) iter.Seq[int] {
	return func(yield func(v int) bool) {
		for ; ; min++ {
			if !yield(min) {
				return
			}
		}
	}
}
