package maps

func Convert[X comparable, Y any, Z comparable, V any](source map[X]Y,
	convertX func(x X) Z, convertY func(y Y) V) map[Z]V {
	var result = make(map[Z]V)
	for k, v := range source {
		result[convertX(k)] = convertY(v)
	}
	return result
}

func Keys[X comparable, Y any](m map[X]Y) []X {
	keys := make([]X, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func KeysFiltered[X comparable, Y any](m map[X]Y, matchFilter func(y Y) bool) []X {
	keys := make([]X, 0, len(m))
	for k, v := range m {
		if matchFilter(v) {
			keys = append(keys, k)
		}
	}
	return keys
}

func Values[X comparable, Y any](m map[X]Y) []Y {
	values := make([]Y, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}
