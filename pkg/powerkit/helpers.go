// pkg/powerkit/helpers.go
package powerkit

import (
	"math"
)

// truncate rounds a float down to two decimal places. This is used
// consistently across the library for formatting final values.
func truncate(f float64) float64 {
	return math.Trunc(f*100) / 100
}

// findMinMax finds the min and max values in a slice of integers.
// (This is also a helper, so it's good to move it here from calculations.go)
func findMinMax(a []int) (min int, max int) {
	if len(a) == 0 {
		return 0, 0
	}
	min = a[0]
	max = a[0]
	for _, value := range a {
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}
	return min, max
}
