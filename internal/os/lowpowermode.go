//go:build darwin

// Package os provides internal OS helpers such as Low Power Mode control.
package os

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Foundation
#import <Foundation/Foundation.h>

int get_low_power_mode_enabled(int* available, int* enabled) {
    if (!available || !enabled) {
        return 1;
    }

    NSProcessInfo* processInfo = [NSProcessInfo processInfo];
    if (![processInfo respondsToSelector:@selector(isLowPowerModeEnabled)]) {
        *available = 0;
        *enabled = 0;
        return 0;
    }

    *available = 1;
    *enabled = [processInfo isLowPowerModeEnabled] ? 1 : 0;
    return 0;
}
*/
import "C"

import (
	"context"
	"errors"
	"os/exec"
	"sync"
	"time"
)

var (
	lowPowerModeReadFn = func() (enabled bool, available bool, err error) {
		var cAvailable C.int
		var cEnabled C.int
		if rc := C.get_low_power_mode_enabled(&cAvailable, &cEnabled); rc != 0 {
			return false, false, errors.New("failed to read low power mode state")
		}
		return cEnabled != 0, cAvailable != 0, nil
	}
	pmsetExecContextFn = func(ctx context.Context, args ...string) error {
		cmd := exec.CommandContext(ctx, "/usr/bin/pmset", args...)
		return cmd.Run()
	}
)

// cached state for low power mode reads
var (
	lpmMu       sync.Mutex
	lpmCachedAt time.Time
	lpmTTL      = 2 * time.Second
	lpmValue    bool
	lpmValid    bool
)

// GetLowPowerModeEnabled returns whether Low Power Mode is enabled. available=false
// indicates the system does not expose the Foundation low power mode property.
func GetLowPowerModeEnabled() (enabled bool, available bool, err error) {
	// Quick cache path
	lpmMu.Lock()
	if lpmValid && time.Since(lpmCachedAt) < lpmTTL {
		v := lpmValue
		lpmMu.Unlock()
		return v, true, nil
	}
	lpmMu.Unlock()

	enabled, available, err = lowPowerModeReadFn()
	if err != nil {
		// Do not update cache on failure to allow quick retry next tick.
		return false, false, err
	}

	// Update cache only when the OS exposes the value.
	if available {
		lpmMu.Lock()
		lpmValue = enabled
		lpmCachedAt = time.Now()
		lpmValid = true
		lpmMu.Unlock()
		return enabled, true, nil
	}
	return false, false, nil
}

// SetLowPowerMode sets Low Power Mode using pmset.
// Requires root privileges.
func SetLowPowerMode(enable bool) error {
	return SetLowPowerModeContext(context.Background(), enable)
}

// SetLowPowerModeContext sets Low Power Mode using pmset.
// Requires root privileges.
func SetLowPowerModeContext(ctx context.Context, enable bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	target := "0"
	if enable {
		target = "1"
	}

	// Currently write all power sources: -a
	// Foundation for future finer control using -b (battery) or -c (charger):
	//   source := "-a" // or "-b" / "-c" in the future when API supports per-source control
	//   cmd := exec.Command("/usr/bin/pmset", source, "lowpowermode", target)
	// For now, keep -a to avoid dead code and ensure consistent behavior.
	if err := pmsetExecContextFn(ctx, "-a", "lowpowermode", target); err != nil {
		return err
	}

	// Invalidate cache immediately so next read reflects the change.
	lpmMu.Lock()
	lpmValid = false
	lpmMu.Unlock()
	return nil
}

// ToggleLowPowerMode toggles the current LPM state.
func ToggleLowPowerMode() error {
	return ToggleLowPowerModeContext(context.Background())
}

// ToggleLowPowerModeContext toggles the current LPM state.
func ToggleLowPowerModeContext(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	enabled, available, err := GetLowPowerModeEnabled()
	if err != nil {
		return err
	}
	if !available {
		return errors.New("low power mode not available on this system")
	}
	return SetLowPowerModeContext(ctx, !enabled)
}
