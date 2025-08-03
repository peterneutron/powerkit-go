//go:build darwin
// +build darwin

// Package os provides an internal wrapper for OS-specific queries, like getting
// the macOS version number.
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

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

var (
	majorVersion      int
	versionCheckOnce  sync.Once
	firmwareVersion   int
	firmwareCheckOnce sync.Once
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

// GetFirmwareMajorVersion executes `system_profiler` to get the system firmware
// version. It runs the command only once and caches the result for the
// lifetime of the application.
func GetFirmwareMajorVersion() int {
	firmwareCheckOnce.Do(func() {
		cmd := exec.Command("system_profiler", "SPHardwareDataType")
		out, err := cmd.Output()
		if err != nil {
			// If the command fails, we must fall back to a legacy value.
			// 0 will ensure the legacy SMC keys are used.
			firmwareVersion = 0
			return
		}

		version, err := parseFirmwareVersion(string(out))
		if err != nil {
			// If parsing fails, also fall back to legacy.
			firmwareVersion = 0
			return
		}
		firmwareVersion = version
	})
	return firmwareVersion
}

// parseFirmwareVersion scans the text output of `system_profiler` for the
// firmware version line and extracts the major version number.
func parseFirmwareVersion(output string) (int, error) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "System Firmware Version:") {
			// Line is like: "      System Firmware Version: 13822.0.233.0.0 ..."
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue // Malformed line
			}

			// valStr is " 13822.0.233.0.0 ..."
			valStr := strings.TrimSpace(parts[1])

			// versionPart is "13822.0.233.0.0"
			versionPart := strings.Fields(valStr)[0]

			// majorPart is "13822"
			majorPart := strings.Split(versionPart, ".")[0]

			return strconv.Atoi(majorPart)
		}
	}
	return 0, fmt.Errorf("could not find 'System Firmware Version' in system_profiler output")
}
