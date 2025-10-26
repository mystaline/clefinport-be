package functions

// Empty returns an empty value.
func Empty[T any]() T {
	var zero T
	return zero
}

func MakePointer[T any](v T) *T {
	return &v
}

// src: https://github.com/samber/lo/blob/bbd44ff2a9c7f4bd4a28ecd9cfdfc41e3339453b/type_manipulation.go#L34

// FromPtr returns the pointer value or empty.
func FromPtr[T any](x *T) T {
	if x == nil {
		return Empty[T]()
	}

	return *x
}

// src: https://github.com/samber/lo/blob/bbd44ff2a9c7f4bd4a28ecd9cfdfc41e3339453b/type_manipulation.go#L43

// FromPtrOr returns the pointer value or the fallback value.
func FromPtrOr[T any](x *T, fallback T) T {
	if x == nil {
		return fallback
	}

	return *x
}
