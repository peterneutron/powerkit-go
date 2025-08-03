//go:build darwin
// +build darwin

// Package powerd provides a simple, high-level wrapper around macOS power
// assertions to prevent the system or display from sleeping using IOPMLib.
package powerd

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation
#include <IOKit/pwr_mgt/IOPMLib.h>

// This C function creates a power assertion and returns its ID.
// It returns 0 on failure.
// We now pass the assertionTypeCFString, which is one of the kIOPMAssertionType* constants.
IOPMAssertionID create_assertion(CFStringRef assertionTypeCFString, const char* reasonStr) {
    IOPMAssertionID assertionID;
    CFStringRef reason = CFStringCreateWithCString(NULL, reasonStr, kCFStringEncodingUTF8);

    IOReturn success = IOPMAssertionCreateWithName(
        assertionTypeCFString,
        kIOPMAssertionLevelOn,
        reason,
        &assertionID
    );

    CFRelease(reason);

    if (success == kIOReturnSuccess) {
        return assertionID;
    }

    return 0; // Return 0 to indicate failure
}

// This C function releases the power assertion using its ID.
void release_assertion(IOPMAssertionID assertionID) {
    IOPMAssertionRelease(assertionID);
}

// We need a C-side helper to get a reference to the constant CFStringRef values.
// Trying to get these directly from Go is difficult and fragile.
CFStringRef get_assertion_type_system_sleep() {
    return kIOPMAssertionTypePreventUserIdleSystemSleep;
}

CFStringRef get_assertion_type_display_sleep() {
    return kIOPMAssertionTypePreventUserIdleDisplaySleep;
}
*/
import "C"

import (
	"fmt"
	"sync"
	"unsafe"
)

// AssertionType defines the type of sleep to prevent.
type AssertionType int

const (
	// PreventSystemSleep prevents the system from sleeping due to user inactivity.
	// This is for long-running background tasks. The display may still sleep.
	PreventSystemSleep AssertionType = iota

	// PreventDisplaySleep prevents the display from sleeping. This implies the
	// system will also not sleep. This is for presentation or video playback.
	PreventDisplaySleep
)

// AssertionID is a Go-native type for the underlying C IOPMAssertionID.
type AssertionID uint32

var (
	// mu protects access to the global assertion IDs.
	mu sync.Mutex

	// We now store IDs for both assertion types separately.
	systemSleepAssertionID  AssertionID
	displaySleepAssertionID AssertionID
)

// PreventSleep creates a power assertion of the specified type.
// It is safe to call this function multiple times for the same type; it will
// only create one assertion of each type. The reason is shown in Activity Monitor.
func PreventSleep(assertionType AssertionType, reason string) (AssertionID, error) {
	mu.Lock()
	defer mu.Unlock()

	var assertionTypeCFString C.CFStringRef
	var currentIDPtr *AssertionID // This is now a pointer to our Go type.

	switch assertionType {
	case PreventSystemSleep:
		assertionTypeCFString = C.get_assertion_type_system_sleep()
		currentIDPtr = &systemSleepAssertionID // Correctly get the address of the variable.
	case PreventDisplaySleep:
		assertionTypeCFString = C.get_assertion_type_display_sleep()
		currentIDPtr = &displaySleepAssertionID // Correctly get the address of the variable.
	default:
		return 0, fmt.Errorf("unknown assertion type: %d", assertionType)
	}

	// If an assertion of this specific type is already active, return its existing ID.
	if *currentIDPtr != 0 {
		return *currentIDPtr, nil
	}

	cReason := C.CString(reason)
	defer C.free(unsafe.Pointer(cReason))

	// The C function returns a C.IOPMAssertionID
	newID := C.create_assertion(assertionTypeCFString, cReason)

	if newID == 0 {
		return 0, fmt.Errorf("IOPMAssertionCreateWithName failed for type %d", assertionType)
	}

	// Store the new ID by dereferencing the pointer and casting the C type to our Go type.
	*currentIDPtr = AssertionID(newID)
	return *currentIDPtr, nil
}

// AllowSleep releases an active power assertion of the specified type.
// It is safe to call this function even if no assertion of that type is active.
func AllowSleep(assertionType AssertionType) {
	mu.Lock()
	defer mu.Unlock()

	var currentIDPtr *AssertionID // Pointer to our Go type

	switch assertionType {
	case PreventSystemSleep:
		currentIDPtr = &systemSleepAssertionID
	case PreventDisplaySleep:
		currentIDPtr = &displaySleepAssertionID
	default:
		return
	}

	if *currentIDPtr == 0 {
		return
	}

	// Release the assertion by casting our Go ID back to the C type.
	C.release_assertion(C.IOPMAssertionID(*currentIDPtr))
	*currentIDPtr = 0 // Reset our state.
}

// AllowAllSleep is a convenience function to release all active assertions
// created by this package. This is useful for cleanup on application exit.
func AllowAllSleep() {
	AllowSleep(PreventSystemSleep)
	AllowSleep(PreventDisplaySleep)
}
