package maps

func Convert[X comparable, Y any, Z comparable, V any](source map[X]Y,
	convertX func(x X) Z, convertY func(y Y) V) map[Z]V {
	var result = make(map[Z]V)
	for k, v := range source {
		result[convertX(k)] = convertY(v)
	}
	return result
}

// Append 2 maps if a key value exists on both foo will take preference
func Append[X comparable, Y any](foo map[X]Y, bar map[X]Y) map[X]Y {
	var result = make(map[X]Y)
	for k, v := range bar {
		result[k] = v
	}
	for k, v := range foo {
		result[k] = v
	}
	return result
}
