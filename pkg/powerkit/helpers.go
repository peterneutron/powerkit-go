// pkg/powerkit/helpers.go
package powerkit

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Foundation -lobjc
#import <Foundation/Foundation.h>

// Returns the major version as an int (e.g. 26)
int get_os_major_version() {
    NSOperatingSystemVersion v = [[NSProcessInfo processInfo] operatingSystemVersion];
    return (int)v.majorVersion;
}
*/
import "C"

import (
	"math"
	"sync"
)

var (
	osMajorVersion     int
	osVersionCheckOnce sync.Once
)

// getOSMajorVersion retrieves the major version number of macOS (e.g., 15 for "Tahoe").
// It uses sync.Once to ensure the Cgo call is only made once.
func getOSMajorVersion() int {
	osVersionCheckOnce.Do(func() {
		// Directly call the C function to get the integer version.
		// This is much cleaner and more reliable than parsing strings.
		osMajorVersion = int(C.get_os_major_version())
	})
	return osMajorVersion
}

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
