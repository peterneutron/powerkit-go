//go:build darwin

// Package os provides an internal wrapper for OS-specific queries, like getting
// the macOS version number.
package os

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Foundation -framework IOKit -framework CoreFoundation -lobjc
#import <Foundation/Foundation.h>
#import <IOKit/IOKitLib.h>
#import <CoreFoundation/CoreFoundation.h>
#import <stdlib.h>
#import <string.h>

// Returns the major version as an int (e.g. 26)
int get_os_major_version() {
    NSOperatingSystemVersion v = [[NSProcessInfo processInfo] operatingSystemVersion];
    return (int)v.majorVersion;
}

static char* copy_cfstring(CFStringRef s) {
    if (!s) {
        return NULL;
    }
    CFIndex len = CFStringGetLength(s);
    CFIndex maxSize = CFStringGetMaximumSizeForEncoding(len, kCFStringEncodingUTF8) + 1;
    char* buffer = (char*)malloc((size_t)maxSize);
    if (!buffer) {
        return NULL;
    }
    if (!CFStringGetCString(s, buffer, maxSize, kCFStringEncodingUTF8)) {
        free(buffer);
        return NULL;
    }
    return buffer;
}

static char* copy_cfdata_best_effort(CFDataRef dataRef) {
    if (!dataRef) {
        return NULL;
    }
    CFIndex len = CFDataGetLength(dataRef);
    if (len <= 0) {
        return NULL;
    }
    const UInt8* bytes = CFDataGetBytePtr(dataRef);
    char* out = (char*)calloc((size_t)len + 1, 1);
    if (!out) {
        return NULL;
    }
    for (CFIndex i = 0; i < len; i++) {
        unsigned char c = bytes[i];
        if (c == 0) {
            out[i] = ' ';
            continue;
        }
        if (c >= 32 && c <= 126) {
            out[i] = (char)c;
        } else {
            out[i] = ' ';
        }
    }
    out[len] = '\0';
    return out;
}

// Returns malloc-allocated string with firmware version from IODeviceTree,
// or NULL if no suitable value is found.
char* get_firmware_version_from_ioreg() {
    const char* paths[] = {"IODeviceTree:/chosen", "IODeviceTree:/"};
    const char* keys[] = {"system-firmware-version", "firmware-version"};

    for (int p = 0; p < 2; p++) {
        io_registry_entry_t entry = IORegistryEntryFromPath(kIOMainPortDefault, paths[p]);
        if (entry == MACH_PORT_NULL) {
            continue;
        }

        for (int k = 0; k < 2; k++) {
            CFStringRef keyRef = CFStringCreateWithCString(kCFAllocatorDefault, keys[k], kCFStringEncodingUTF8);
            if (!keyRef) {
                continue;
            }
            CFTypeRef value = IORegistryEntryCreateCFProperty(entry, keyRef, kCFAllocatorDefault, 0);
            CFRelease(keyRef);
            if (!value) {
                continue;
            }

            char* out = NULL;
            CFTypeID type = CFGetTypeID(value);
            if (type == CFStringGetTypeID()) {
                out = copy_cfstring((CFStringRef)value);
            } else if (type == CFDataGetTypeID()) {
                out = copy_cfdata_best_effort((CFDataRef)value);
            }
            CFRelease(value);

            if (out && out[0] != '\0') {
                IOObjectRelease(entry);
                return out;
            }
            if (out) {
                free(out);
            }
        }

        IOObjectRelease(entry);
    }

    return NULL;
}
*/
import "C"

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unsafe"
)

var (
	majorVersion      int
	versionCheckOnce  sync.Once
	firmwareCheckOnce sync.Once
	firmwareInfo      FirmwareInfo
)

// FirmwareInfo describes detected firmware metadata used for resolver gating.
type FirmwareInfo struct {
	Major   int
	Version string
	Source  string
}

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

// GetFirmwareInfo retrieves firmware version metadata and caches the result.
// It first tries IORegistry DeviceTree keys and falls back to system_profiler.
func GetFirmwareInfo() FirmwareInfo {
	firmwareCheckOnce.Do(func() {
		version := getFirmwareVersionFromIORegistry()
		if major, err := extractFirmwareMajor(version); err == nil && major > 0 {
			firmwareInfo = FirmwareInfo{
				Major:   major,
				Version: normalizeFirmwareValue(version),
				Source:  "ioreg_device_tree",
			}
			return
		}

		version, major, err := getFirmwareFromSystemProfiler()
		if err == nil && major > 0 {
			firmwareInfo = FirmwareInfo{
				Major:   major,
				Version: normalizeFirmwareValue(version),
				Source:  "system_profiler",
			}
			return
		}

		firmwareInfo = FirmwareInfo{
			Major:   0,
			Version: "",
			Source:  "unknown",
		}
	})
	return firmwareInfo
}

// GetFirmwareMajorVersion is kept as a compatibility wrapper for older callsites.
func GetFirmwareMajorVersion() int {
	return GetFirmwareInfo().Major
}

func getFirmwareVersionFromIORegistry() string {
	raw := C.get_firmware_version_from_ioreg()
	if raw == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(raw))
	return C.GoString(raw)
}

func getFirmwareFromSystemProfiler() (string, int, error) {
	cmd := exec.Command("system_profiler", "SPHardwareDataType")
	out, err := cmd.Output()
	if err != nil {
		return "", 0, err
	}
	version, major, err := parseFirmwareRecordFromSystemProfiler(string(out))
	if err != nil {
		return "", 0, err
	}
	return version, major, nil
}

// parseFirmwareVersion is retained for backwards-compatible tests/callers.
func parseFirmwareVersion(output string) (int, error) {
	_, major, err := parseFirmwareRecordFromSystemProfiler(output)
	return major, err
}

// parseFirmwareRecordFromSystemProfiler scans the text output of system_profiler
// and returns the raw firmware value and its parsed major version.
func parseFirmwareRecordFromSystemProfiler(output string) (string, int, error) {
	labels := []string{
		"System Firmware Version:",
		"Boot ROM Version:",
		"OS Loader Version:",
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		for _, label := range labels {
			if !strings.Contains(line, label) {
				continue
			}
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			valStr := strings.TrimSpace(parts[1])
			if valStr == "" {
				continue
			}
			major, err := extractFirmwareMajor(valStr)
			if err != nil {
				continue
			}
			return normalizeFirmwareValue(valStr), major, nil
		}
	}
	return "", 0, fmt.Errorf("could not find firmware version in system_profiler output")
}

var firmwareVersionRE = regexp.MustCompile(`([0-9]{4,})(?:\.[0-9]+)+`)
var firmwareMajorFallbackRE = regexp.MustCompile(`\b([0-9]{4,})\b`)

func extractFirmwareMajor(value string) (int, error) {
	if m := firmwareVersionRE.FindStringSubmatch(value); len(m) == 2 {
		return strconv.Atoi(m[1])
	}
	if m := firmwareMajorFallbackRE.FindStringSubmatch(value); len(m) == 2 {
		return strconv.Atoi(m[1])
	}
	return 0, fmt.Errorf("no firmware major version found in %q", value)
}

func normalizeFirmwareValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "\"")
	value = strings.Trim(value, "<>")
	value = strings.Join(strings.Fields(value), " ")
	return value
}
