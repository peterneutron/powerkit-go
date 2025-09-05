//go:build darwin
// +build darwin

// Package os provides internal OS helpers such as Low Power Mode control.
package os

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// cached state for low power mode reads to avoid frequent process spawns
var (
	lpmMu       sync.Mutex
	lpmCachedAt time.Time
	lpmTTL      = 2 * time.Second
	lpmValue    bool
	lpmValid    bool
)

// GetLowPowerModeEnabled returns whether Low Power Mode is enabled.
// available=false indicates the system does not report the key (older macOS) or output could not be parsed.
func GetLowPowerModeEnabled() (enabled bool, available bool, err error) {
	// Quick cache path
	lpmMu.Lock()
	if lpmValid && time.Since(lpmCachedAt) < lpmTTL {
		v := lpmValue
		lpmMu.Unlock()
		return v, true, nil
	}
	lpmMu.Unlock()

	cmd := exec.Command("/usr/bin/pmset", "-g")
	var out bytes.Buffer
	cmd.Stdout = &out
	if runErr := cmd.Run(); runErr != nil {
		// Do not update cache on failure to allow quick retry next tick
		return false, false, runErr
	}

	enabled = false
	available = false
	for _, line := range strings.Split(out.String(), "\n") {
		s := strings.TrimSpace(strings.ToLower(line))
		if strings.HasPrefix(s, "lowpowermode") {
			available = true
			// Example formats:
			//  lowpowermode         1
			//  lowpowermode = 1
			if strings.Contains(s, " 1") || strings.HasSuffix(s, "1") || strings.Contains(s, "= 1") {
				enabled = true
			}
			break
		}
	}

	// Update cache only when available and parsed
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
	target := "0"
	if enable {
		target = "1"
	}

	// Currently write all power sources: -a
	// Foundation for future finer control using -b (battery) or -c (charger):
	//   source := "-a" // or "-b" / "-c" in the future when API supports per-source control
	//   cmd := exec.Command("/usr/bin/pmset", source, "lowpowermode", target)
	// For now, keep -a to avoid dead code and ensure consistent behavior.
	cmd := exec.Command("/usr/bin/pmset", "-a", "lowpowermode", target)
	if err := cmd.Run(); err != nil {
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
	enabled, available, err := GetLowPowerModeEnabled()
	if err != nil {
		return err
	}
	if !available {
		return errors.New("low power mode not available on this system")
	}
	return SetLowPowerMode(!enabled)
}
