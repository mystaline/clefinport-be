package functions

import "reflect"

// IsZeroValue checks if a given struct is equal to its zero value.
// The zero value of a struct is the state where all its fields are uninitialized:
// - Numeric fields are 0.
// - String fields are "".
// - Boolean fields are false.
// - Pointers, slices, maps, and interfaces are nil.
// This function is particularly useful for structs, where field-by-field comparison
// would be tedious and error-prone.
//
// Parameters:
// - model: An interface{} representing the struct to check.
//
// Returns:
// - bool: true if the struct is equal to its zero value, false otherwise.
//
// Usage: This function simplifies checking if a struct is "empty" or in its default state,
// which can be difficult to do manually. It's especially helpful when dealing with nested structs
// or dynamic types where the struct's fields may not be known at compile time.
func IsZeroValue(v interface{}) bool {
	if v == nil {
		return true
	}

	return reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface())
}

// ComparePointers compares two pointers of a comparable type.
// It handles cases where one or both pointers may be nil, ensuring safe comparison
// without causing runtime panics.
//
// Parameters:
//   - v1: A pointer to a value of type T. Can be nil.
//   - v2: A pointer to a value of type T. Can be nil.
//
// Returns:
//   - bool: true if both pointers are nil or if the values they point to are equal.
//     false if one pointer is nil and the other is not, or if the values they
//     point to are not equal.
//
// Usage: This function is useful for safely comparing pointers of any comparable type,
// especially when nil values are a possibility. It avoids the need for manual nil checks
// and dereferencing, simplifying pointer comparison logic.
//
// Example:
//
//	x := 10
//	y := 10
//	var p1 *int = &x
//	var p2 *int = &y
//	var p3 *int = nil
//
//	ComparePointers(p1, p2) // true, because *p1 == *p2
//	ComparePointers(p1, p3) // false, because p3 is nil
//	ComparePointers(p3, p3) // true, because both are nil
func ComparePointers[T comparable](v1, v2 *T) bool {
	if v1 == nil && v2 == nil {
		return true
	}

	if v1 == nil || v2 == nil {
		return false
	}

	return *v1 == *v2
}
