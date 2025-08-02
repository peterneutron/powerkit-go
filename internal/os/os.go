//go:build darwin
// +build darwin

package os

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

import "sync"

var (
	majorVersion     int
	versionCheckOnce sync.Once
)

// GetMajorVersion retrieves and caches the major version number of macOS.
// It uses sync.Once to ensure the Cgo call is only performed once.
func GetMajorVersion() int {
	versionCheckOnce.Do(func() {
		v := int(C.get_os_major_version())

		// Defensive check: If Cgo returns 0, it indicates an error or a very old OS.
		// We fall back to a sensible default to ensure logic in other packages
		// doesn't fail.
		if v == 0 {
			majorVersion = 15 // Fallback to a known modern version.
		} else {
			majorVersion = v
		}
	})
	return majorVersion
}
