package functions

// src:
// - https://github.com/samber/lo/blob/dc8a175ce1dbd833bdd14404d95c41b3c72e4ae9/condition.go#L5
// - https://stackoverflow.com/a/59375088/13359128

// Ternary is a 1 line if/else statement.
//
// !! USE WITH DISCRETION FOR VERBOSITY ISSUES !!
//
// EXAMPLE:
//
// result := functions.Ternary(true, "Yes", "No")
// // "Yes"
//
// result := functions.Ternary(false, "Yes", "No")
// // "No"
func Ternary[T any](condition bool, ifOutput T, elseOutput T) T {
	if condition {
		return ifOutput
	}

	return elseOutput
}
