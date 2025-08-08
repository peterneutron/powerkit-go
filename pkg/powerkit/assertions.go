//go:build darwin
// +build darwin

package powerkit

import (
	"github.com/peterneutron/powerkit-go/internal/powerd"
)

// AssertionType defines the type of sleep to prevent.
type AssertionType int

const (
	// AssertionTypePreventSystemSleep prevents the system from sleeping due to user inactivity.
	AssertionTypePreventSystemSleep = AssertionType(powerd.PreventSystemSleep)
	// AssertionTypePreventDisplaySleep prevents the display from sleeping, which also prevents system sleep.
	AssertionTypePreventDisplaySleep = AssertionType(powerd.PreventDisplaySleep)
)

// AssertionID is a Go-native type for the underlying C IOPMAssertionID.
type AssertionID uint32

// CreateAssertion creates a power assertion of the specified type.
// It is safe to call this function multiple times; it will only create one
// assertion of each type. The reason is shown in Activity Monitor.
func CreateAssertion(assertionType AssertionType, reason string) (AssertionID, error) {
	id, err := powerd.PreventSleep(powerd.AssertionType(assertionType), reason)
	return AssertionID(id), err
}

// ReleaseAssertion releases an active power assertion of the specified type.
// It is safe to call this function even if no assertion of that type is active.
func ReleaseAssertion(assertionType AssertionType) {
	powerd.AllowSleep(powerd.AssertionType(assertionType))
}

// AllowAllSleep is a convenience function to release all active assertions
// created by this package. This is useful for cleanup on application exit.
func AllowAllSleep() {
	powerd.AllowAllSleep()
}

// IsAssertionActive reports whether an assertion of the given type is active
// (created and not released) by this process.
func IsAssertionActive(assertionType AssertionType) bool {
	return powerd.IsActive(powerd.AssertionType(assertionType))
}

// GetAssertionID returns the active assertion ID for the given type, if any.
// The boolean return indicates whether an assertion is active.
func GetAssertionID(assertionType AssertionType) (AssertionID, bool) {
	id, ok := powerd.GetAssertionID(powerd.AssertionType(assertionType))
	return AssertionID(id), ok
}
