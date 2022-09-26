package util

func If[T any](cond bool, vtrue, vfalse T) T {
	if cond {
		return vtrue
	}
	return vfalse
}

func ArrayConvert[T any](source []interface{}) []T {
	var result []T
	for _, item := range source {
		result = append(result, item.(T))
	}
	return result
}
