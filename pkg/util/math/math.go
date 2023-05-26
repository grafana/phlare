package math

import (
	"golang.org/x/exp/constraints"
)

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Slightly improved functions from github.com/samber/lo:
// now the input is an ellipsis.

// MaxMany searches the maximum value of a collection.
// Returns zero value when collection is empty.
func MaxMany[T constraints.Ordered](collection ...T) T {
	var max T

	if len(collection) == 0 {
		return max
	}

	max = collection[0]

	for i := 1; i < len(collection); i++ {
		item := collection[i]

		if item > max {
			max = item
		}
	}

	return max
}

// MinMany search the minimum value of a collection.
// Returns zero value when collection is empty.
func MinMany[T constraints.Ordered](collection ...T) T {
	var min T

	if len(collection) == 0 {
		return min
	}

	min = collection[0]

	for i := 1; i < len(collection); i++ {
		item := collection[i]

		if item < min {
			min = item
		}
	}

	return min
}
