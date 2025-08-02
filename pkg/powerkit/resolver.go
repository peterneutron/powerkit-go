//go:build darwin
// +build darwin

package powerkit

import (
	"github.com/peterneutron/powerkit-go/internal/os"
	"github.com/peterneutron/powerkit-go/internal/smc"
)

// macOSMajorVersionThreshold is the major version where the new SMC keys were introduced.
// The user mentioned "26.x", so we will use 26 as the threshold.
const macOSMajorVersionThreshold = 26

// smcControlConfig holds the dynamically resolved keys and byte values for SMC write operations.
type smcControlConfig struct {
	// General
	Mode string

	// Adapter Control
	AdapterKey          string
	AdapterEnableBytes  []byte
	AdapterDisableBytes []byte

	// Charging Control
	IsLegacyCharging     bool
	ChargingKeyModern    string
	ChargingKeysLegacy   []string
	ChargingEnableBytes  []byte
	ChargingDisableBytes []byte
}

// currentSMCConfig is a package-level variable holding the correct configuration for the running OS.
var currentSMCConfig smcControlConfig

// The init function runs once when the package is imported.
// It resolves which set of SMC keys and values to use based on the OS version.
func init() {
	majorVersion := os.GetMajorVersion()

	if majorVersion >= macOSMajorVersionThreshold {
		// --- Modern Configuration (macOS 26 "Tahoe" and newer) ---
		currentSMCConfig = smcControlConfig{
			Mode:                "Modern",
			AdapterKey:          smc.KeyIsAdapterEnabled,
			AdapterEnableBytes:  []byte{0x00},
			AdapterDisableBytes: []byte{0x08},

			IsLegacyCharging:     false,
			ChargingKeyModern:    smc.KeyIsChargingEnabled,
			ChargingEnableBytes:  []byte{0x00, 0x00, 0x00, 0x00},
			ChargingDisableBytes: []byte{0x01, 0x00, 0x00, 0x00},
		}
	} else {
		// --- Legacy Configuration (macOS 15 "Sequoia" and older) ---
		currentSMCConfig = smcControlConfig{
			Mode:                "Legacy",
			AdapterKey:          smc.KeyIsAdapterEnabled_Legacy,
			AdapterEnableBytes:  []byte{0x00},
			AdapterDisableBytes: []byte{0x01}, // Note the different value

			IsLegacyCharging:     true,
			ChargingKeysLegacy:   []string{smc.KeyIsChargingEnabled_Legacy_BCLM, smc.KeyIsChargingEnabled_Legacy_BCDS},
			ChargingEnableBytes:  []byte{0x00},
			ChargingDisableBytes: []byte{0x02}, // Note the different value
		}
	}
}
