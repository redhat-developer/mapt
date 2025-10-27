package fx

// Some returns a pointer to its argument.
func Some[T any](v T) *T {
	return &v
}
