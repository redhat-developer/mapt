package fx

import "iter"

// Take returns an iterator that takes at most n values from its source.
func Take[T any](it iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range it {
			if n <= 0 || !yield(v) {
				return
			}
			n--
		}
	}
}

// Take2 returns an iterator that takes at most n values from the input slice.
func Take2[T, U any](it iter.Seq2[T, U], n int) iter.Seq2[T, U] {
	return func(yield func(T, U) bool) {
		for t, u := range it {
			if n <= 0 || !yield(t, u) {
				return
			}
			n--
		}
	}
}

// Skip returns an iterator that skips n values from its source.
func Skip[T any](it iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range it {
			if n <= 0 {
				if !yield(v) {
					return
				}
			} else {
				n--
			}
		}
	}
}

// Skip2 returns an iterator that skips n values from its source.
func Skip2[T, U any](it iter.Seq2[T, U], n int) iter.Seq2[T, U] {
	return func(yield func(T, U) bool) {
		for t, u := range it {
			if n <= 0 {
				if !yield(t, u) {
					return
				}
			} else {
				n--
			}
		}
	}
}

// First returns the first element of it, if any elements exist.
func First[T any](it iter.Seq[T]) (t T, ok bool) {
	for t := range it {
		return t, true
	}
	return t, false
}

// First2 returns the first element of it, if any elements exist.
func First2[T, U any](it iter.Seq2[T, U]) (t T, u U, ok bool) {
	for t, u := range it {
		return t, u, true
	}
	return t, u, false
}

// Last returns the last element of it, if any elements exist.
func Last[T any](it iter.Seq[T]) (last T, ok bool) {
	for t := range it {
		last, ok = t, true
	}
	return last, ok
}

// Last2 returns the last element of it, if any elements exist.
func Last2[T, U any](it iter.Seq2[T, U]) (lastT T, lastU U, ok bool) {
	for t, u := range it {
		lastT, lastU, ok = t, u, true
	}
	return lastT, lastU, ok
}
