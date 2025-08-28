package slices

import "slices"

// For float we need explictly check the values
func SortbyFloat[X any](s []X, float func(X) float64) {
	slices.SortFunc(s,
		func(a, b X) int {
			if float(a) < float(b) {
				return -1
			}
			if float(a) > float(b) {
				return 1
			}
			return 0
		})
}

// This function split a list based on a evaluation function
// the result is a map of lists
func Split[T any, Y comparable](source []T,
	identification func(item T) Y) (m map[Y][]T) {
	m = make(map[Y][]T)
	for _, item := range source {
		k := identification(item)
		if l, found := m[k]; found {
			m[k] = append(l, item)
		} else {
			m[k] = []T{item}
		}
	}
	return
}
