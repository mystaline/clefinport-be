package functions

// Filter is a generic utility function that processes a slice of any type (T)
// and returns a new slice containing only the elements that satisfy the provided
// comparison function (compareFn).
//
// Parameters:
// - src: A slice of elements of type T to be filtered.
// - compareFn: A function that takes two arguments:
//  1. An element of type T.
//  2. The index of the element within the source slice.
//     The function should return true if the element should be included in the
//     filtered slice, and false otherwise.
//
// Returns:
// - A new slice containing the elements from src for which compareFn returned true.
//
// Example usage:
//
// numbers := []int{1, 2, 3, 4, 5}
//
//	evenNumbers := Filter(numbers, func(n int, _ int) bool {
//	    return n%2 == 0
//	})
//
// fmt.Println(evenNumbers) // Output: [2, 4]
func Filter[T any](src []T, compareFn func(T, int) bool) []T {
	filtered := []T{}
	for i, v := range src {
		if compareFn(v, i) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// Map is a generic utility function that processes a slice of any type (T)
// and transforms each element into another type (R) using the provided mapping
// function (mapFn). It returns a new slice of type R.
//
// Parameters:
// - src: A slice of elements of type T to be transformed.
// - mapFn: A function that takes two arguments:
//  1. An element of type T.
//  2. The index of the element within the source slice.
//     The function should return the transformed element of type R.
//
// Returns:
// - A new slice containing the transformed elements of type R.
//
// Example usage:
//
// numbers := []int{1, 2, 3}
//
//	squaredNumbers := Map(numbers, func(n int, _ int) int {
//	    return n * n
//	})
//
// fmt.Println(squaredNumbers) // Output: [1, 4, 9]
func Map[T any, R any](src []T, mapFn func(T, int) R) []R {
	mapped := []R{}
	for i, v := range src {
		mapped = append(mapped, mapFn(v, i))
	}
	return mapped
}

// ForEach is a utility function that iterates over a slice of any type (T)
// and performs an operation on each element using the provided callback function (forEachFn).
// This function does not return any value, as it is meant for executing side effects.
//
// Parameters:
// - src: A slice of elements of type T to iterate over.
// - forEachFn: A function that takes two arguments:
//  1. An element of type T.
//  2. The index of the element within the source slice.
//     The function is called for each element in the slice.
//
// Returns:
// - Nothing. This function is used for executing side effects.
//
// Example usage:
//
// numbers := []int{1, 2, 3}
//
//	ForEach(numbers, func(n int, i int) {
//	    fmt.Printf("Index: %d, Value: %d\n", i, n)
//	})
//
// Output:
// Index: 0, Value: 1
// Index: 1, Value: 2
// Index: 2, Value: 3
func ForEach[T any](src []T, forEachFn func(T, int)) {
	for i, v := range src {
		forEachFn(v, i)
	}
}

// All is a utility function that checks if all elements in a slice satisfy a given condition defined by the findFn function.
// It returns true if all elements meet the condition, and false as soon as any element does not satisfy the condition.
//
// Parameters:
// - src: A slice of elements of type T to check.
// - findFn: A function that takes two arguments:
//  1. The element of type T.
//  2. The index of the element within the slice (int).
//     The function should return true if the element satisfies the condition, and false otherwise.
//
// Returns:
// - A boolean value indicating whether all elements satisfy the condition (true) or not (false).
//
// Example usage:
//
// numbers := []int{1, 2, 3, 4, 5}
//
//	result := All(numbers, func(n int, _ int) bool {
//	    return n > 0 // Check if all numbers are greater than 0
//	})
//
// fmt.Println(result) // Output: true
func All[T any](src []T, findFn func(T, int) bool) bool {
	for i, v := range src {
		if !findFn(v, i) {
			return false
		}
	}

	return true
}

// Any is a utility function that checks if any element in a slice satisfies a given condition defined by the findFn function.
// It returns true if at least one element meets the condition, and false if no element satisfies the condition.
//
// Parameters:
// - src: A slice of elements of type T to check.
// - findFn: A function that takes two arguments:
//  1. The element of type T.
//  2. The index of the element within the slice (int).
//     The function should return true if the element satisfies the condition, and false otherwise.
//
// Returns:
// - A boolean value indicating whether any element satisfies the condition (true) or not (false).
//
// Example usage:
//
// numbers := []int{1, 2, 3, 4, 5}
//
//	result := Any(numbers, func(n int, _ int) bool {
//	    return n == 3 // Check if there is any number equal to 3
//	})
//
// fmt.Println(result) // Output: true
func Any[T any](src []T, findFn func(T, int) bool) bool {
	for i, v := range src {
		if findFn(v, i) {
			return true
		}
	}

	return false
}

// Find is a generic utility function that searches through a slice of any type (T)
// and returns a pointer to the first element that satisfies the provided find function (findFn).
// If no element satisfies the condition, it returns nil.
//
// Parameters:
// - src: A slice of elements of type T to search through.
// - findFn: A function that takes two arguments:
//  1. An element of type T.
//  2. The index of the element within the source slice.
//     The function should return true if the element is the one you are looking for, and false otherwise.
//
// Returns:
// - A pointer to the first element that satisfies the findFn condition, or nil if no such element is found.
//
// Example usage:
//
// numbers := []int{1, 2, 3, 4, 5}
//
//	result := Find(numbers, func(n int, _ int) bool {
//	    return n == 3
//	})
//
//	if result != nil {
//	    fmt.Println(*result) // Output: 3
//	} else {
//
//	    fmt.Println("Not found")
//	}
func Find[T any](src []T, findFn func(T, int) bool) *T {
	for i, v := range src {
		if findFn(v, i) {
			return MakePointer(v)
		}
	}

	return nil
}

// IndexOf is a utility function that searches for the first occurrence of an element in a slice (src)
// that satisfies the provided condition (findFn). It returns the index of the first matching element,
// or -1 if no element satisfies the condition.
//
// Parameters:
//   - src: A slice of elements of type T to search through.
//   - findFn: A function that takes an element of type T and returns a boolean.
//     It should return true for the element you are looking for, and false otherwise.
//
// Returns:
// - The index of the first element in the slice that satisfies the findFn condition, or -1 if no such element is found.
//
// Example usage:
//
// numbers := []int{1, 2, 3, 4, 5}
//
//	index := IndexOf(numbers, func(n int) bool {
//	    return n == 3
//	})
//
// fmt.Println(index) // Output: 2
func IndexOf[T any](src []T, findFn func(T) bool) int {
	for i, v := range src {
		if findFn(v) {
			return i
		}
	}

	return -1
}

// Distinct is a utility function that removes duplicate elements from a slice of any comparable type (T).
// It returns a new slice containing only unique elements from the input slice.
//
// Parameters:
// - src: A slice of elements of type T from which duplicates will be removed.
//
// Returns:
// - A new slice of type T containing only unique elements from the source slice.
//
// Example usage:
//
// numbers := []int{1, 2, 2, 3, 4, 4, 5}
// uniqueNumbers := Distinct(numbers)
// fmt.Println(uniqueNumbers) // Output: [1, 2, 3, 4, 5]
func Distinct[T comparable](src []T) []T {
	distincted := []T{}
	for _, v := range src {
		found := Find(distincted, func(d T, _ int) bool {
			return d == v
		})

		if found == nil {
			distincted = append(distincted, v)
		}
	}

	return distincted
}

// DistinctBy is a utility function that removes duplicate elements from a slice of any type (T)
// based on a custom comparison function (compareFn). It returns a new slice containing only unique elements
// according to the logic defined in the comparison function.
//
// Parameters:
//   - src: A slice of elements of type T from which duplicates will be removed.
//   - compareFn: A function that takes two arguments of type T and returns a boolean.
//     It is used to compare two elements, and if it returns true, those elements are considered equal.
//
// Returns:
// - A new slice containing only the unique elements from the source slice, according to the compareFn logic.
//
// Example usage:
//
//	users := []User{
//	    {ID: 1, Name: "Alice"},
//	    {ID: 2, Name: "Bob"},
//	    {ID: 1, Name: "Alice"},
//	}
//
//	distinctUsers := DistinctBy(users, func(u1, u2 User) bool {
//	    return u1.ID == u2.ID // Compare by ID
//	})
//
// fmt.Println(distinctUsers) // Output: [{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}]
func DistinctBy[T any](src []T, compareFn func(src T, dest T) bool) []T {
	distincted := []T{}
	for _, v := range src {
		found := Find(distincted, func(d T, _ int) bool {
			return compareFn(v, d)
		})

		if found == nil {
			distincted = append(distincted, v)
		}
	}

	return distincted
}

// Reduce is a utility function that processes a slice of any type (T) and accumulates a result of type R
// based on the provided initial value and reduction function (reduceFn).
// It iterates through the slice, applying the reduction function to each element and returning a final accumulated value.
//
// Parameters:
// - src: A slice of elements of type T to be processed.
// - initialValue: The initial value of type R that the reduction process starts with.
// - reduceFn: A function that takes three arguments:
//  1. The accumulator of type R, which stores the accumulated result.
//  2. The current element from the slice of type T.
//  3. The current index of the element within the slice.
//     The function returns a new accumulator value of type R, which will be passed to the next iteration.
//
// Returns:
// - The final accumulated value of type R after processing all elements in the slice.
//
// Example usage:
//
// numbers := []int{1, 2, 3, 4}
//
//	sum := Reduce(numbers, 0, func(acc int, value int, _ int) int {
//	    return acc + value
//	})
//
// fmt.Println(sum) // Output: 10
func Reduce[T any, R any](
	src []T,
	initialValue R,
	reduceFn func(accumulator R, currentValue T, currentIndex int) R,
) R {
	value := initialValue
	for i, v := range src {
		value = reduceFn(value, v, i)
	}

	return value
}

// Contains checks whether a given value exists in a slice of comparable elements.
// It iterates through the slice and returns true if the value is found, otherwise false.
//
// Parameters:
// - src: A slice of elements of type T to be searched. The type T must be comparable.
// - value: The value of type T to search for within the slice.
//
// Returns:
// - bool: true if the value exists in the slice, false otherwise.
//
// Example usage:
//
//	numbers := []int{1, 2, 3, 4, 5}
//	fmt.Println(Contains(numbers, 3)) // Output: true
//	fmt.Println(Contains(numbers, 6)) // Output: false
//
//	fruits := []string{"apple", "banana", "cherry"}
//	fmt.Println(Contains(fruits, "banana")) // Output: true
//	fmt.Println(Contains(fruits, "grape"))  // Output: false
func Contains[T comparable](src []T, value T) bool {
	contain := false
	for _, v := range src {
		contain = v == value
		if contain {
			break
		}
	}

	return contain
}
