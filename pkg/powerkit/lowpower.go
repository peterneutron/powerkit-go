//go:build darwin

package powerkit

import (
	sysos "github.com/peterneutron/powerkit-go/internal/os"
)

// GetLowPowerModeEnabled reports whether macOS Low Power Mode is enabled.
// The second return value indicates availability on this system.
func GetLowPowerModeEnabled() (bool, bool, error) {
	return sysos.GetLowPowerModeEnabled()
}

// SetLowPowerMode enables or disables macOS Low Power Mode.
// Requires root privileges; callers should handle privilege escalation at the CLI layer.
func SetLowPowerMode(enable bool) error {
	if err := requireRoot("set low power mode"); err != nil {
		return err
	}
	return sysos.SetLowPowerMode(enable)
}

// ToggleLowPowerMode toggles the current Low Power Mode setting.
func ToggleLowPowerMode() error {
	if err := requireRoot("toggle low power mode"); err != nil {
		return err
	}
	return sysos.ToggleLowPowerMode()
}
