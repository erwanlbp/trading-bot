package util

func WrapPtr[T any](t T) *T {
	return &t
}
