package fx

import "iter"

// Reduce calls fn on each element of the input sequence, passing in the current value of the accumulator with
// each invocation and updating the accumulator to the result of fn after each invocation.
func Reduce[T, U any](it iter.Seq[T], init U, fn func(acc U, v T) U) U {
	for v := range it {
		init = fn(init, v)
	}
	return init
}

// Reduce2 calls fn on each element of the input slice, passing in the
// current value of the accumulator with each invocation and updating the
// accumulator to the result of fn after each invocation.
func Reduce2[T, U, V any](it iter.Seq2[T, U], init V, fn func(acc V, t T, u U) V) V {
	for t, u := range it {
		init = fn(init, t, u)
	}
	return init
}
