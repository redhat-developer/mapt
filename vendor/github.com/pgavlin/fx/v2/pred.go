package fx

import "iter"

// Any returns true if pred returns true for any element of the input sequence.
func Any[T any](it iter.Seq[T], pred func(v T) bool) bool {
	for v := range it {
		if pred(v) {
			return true
		}
	}
	return false
}

// Any2 returns true if pred returns true for any element of the input slice.
func Any2[T, U any](it iter.Seq2[T, U], pred func(t T, u U) bool) bool {
	for t, u := range it {
		if pred(t, u) {
			return true
		}
	}
	return false
}

// Contains returns true if the input sequence contains t.
func Contains[T comparable](it iter.Seq[T], t T) bool {
	for v := range it {
		if v == t {
			return true
		}
	}
	return false
}

// All returns true if pred returns true for every element of the input sequence.
func All[T any](it iter.Seq[T], pred func(v T) bool) bool {
	for v := range it {
		if !pred(v) {
			return false
		}
	}
	return true
}

// All2 returns true if pred returns true for every element of the input slice.
func All2[T, U any](it iter.Seq2[T, U], pred func(t T, u U) bool) bool {
	for t, u := range it {
		if !pred(t, u) {
			return false
		}
	}
	return true
}

// And combines a list of predicates into a predicate that returns true if every predicate in the list returns true.
func And[T any](preds ...func(v T) bool) func(T) bool {
	return func(v T) bool {
		for _, p := range preds {
			if !p(v) {
				return false
			}
		}
		return true
	}
}

// And2 combines a list of predicates into a predicate that returns true if every predicate in the list returns true.
func And2[T, U any](preds ...func(t T, u U) bool) func(T, U) bool {
	return func(t T, u U) bool {
		for _, p := range preds {
			if !p(t, u) {
				return false
			}
		}
		return true
	}
}

// Or combines a list of predicates into a predicate that returns true if any predicate in the list returns true.
func Or[T any](preds ...func(v T) bool) func(T) bool {
	return func(v T) bool {
		for _, p := range preds {
			if p(v) {
				return true
			}
		}
		return false
	}
}

// Or2 combines a list of predicates into a predicate that returns true if any predicate in the list returns true.
func Or2[T, U any](preds ...func(t T, u U) bool) func(T, U) bool {
	return func(t T, u U) bool {
		for _, p := range preds {
			if p(t, u) {
				return true
			}
		}
		return false
	}
}

// Not inverts the result of a predicate.
func Not[T any](pred func(v T) bool) func(T) bool {
	return func(v T) bool { return !pred(v) }
}

// Not2 inverts the result of a predicate.
func Not2[T, U any](pred func(t T, u U) bool) func(T, U) bool {
	return func(t T, u U) bool { return !pred(t, u) }
}
