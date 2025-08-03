// Package powerkit provides a high-level API for querying and controlling
// macOS power management features.
package powerkit

import (
	"math"
)

// truncate rounds a float down to two decimal places. This is used
// consistently across the library for formatting final values.
func truncate(f float64) float64 {
	return math.Trunc(f*100) / 100
}

// findMinMax finds the minVal and maxVal values in a slice of integers.
// (This is also a helper, so it's good to move it here from calculations.go)
func findMinMax(a []int) (minVal int, maxVal int) {
	if len(a) == 0 {
		return 0, 0
	}
	minVal = a[0]
	maxVal = a[0]
	for _, value := range a {
		if value < minVal {
			minVal = value
		}
		if value > maxVal {
			maxVal = value
		}
	}
	return minVal, maxVal
}
