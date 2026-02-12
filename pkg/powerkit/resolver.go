//go:build darwin

package powerkit

import (
	"github.com/peterneutron/powerkit-go/internal/os"
	"github.com/peterneutron/powerkit-go/internal/smc"
)

// FirmwareMajorVersionThreshold is the major version of the System Firmware
// where the new SMC keys were introduced. As of August 2025, this is 13822.
const FirmwareMajorVersionThreshold = 13822

// smcControlConfig holds the dynamically resolved keys and byte values for SMC write operations.
type smcControlConfig struct {
	// General
	Firmware string
	// Stable firmware profile identifier and independent version number.
	FirmwareProfileID      string
	FirmwareProfileVersion int

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
var currentFirmwareInfo os.FirmwareInfo

const (
	firmwareCompatTested      = "tested"
	firmwareCompatUntestedNew = "untested_newer"
	firmwareCompatUntestedOld = "untested_older"
	firmwareCompatUnknown     = "unknown"
	profileModernID           = "smc_profile_modern"
	profileLegacyID           = "smc_profile_legacy"
	defaultProfileVersion     = 1
)

// The init function runs once when the package is imported.
// It resolves which set of SMC keys and values to use based on the firmware version.
func init() {
	currentFirmwareInfo = os.GetFirmwareInfo()
	firmwareVersion := currentFirmwareInfo.Major

	switch {
	case firmwareVersion == FirmwareMajorVersionThreshold:
		// --- Supported Configuration ---
		// This is the specific firmware version we have tested and know works.
		currentSMCConfig = smcControlConfig{
			Firmware:               "Supported",
			FirmwareProfileID:      profileModernID,
			FirmwareProfileVersion: defaultProfileVersion,
			AdapterKey:             smc.KeyIsAdapterEnabled,
			AdapterEnableBytes:     []byte{0x00},
			AdapterDisableBytes:    []byte{0x08},
			IsLegacyCharging:       false,
			ChargingKeyModern:      smc.KeyIsChargingEnabled,
			ChargingEnableBytes:    []byte{0x00, 0x00, 0x00, 0x00},
			ChargingDisableBytes:   []byte{0x01, 0x00, 0x00, 0x00},
		}

	case firmwareVersion > 0 && firmwareVersion < FirmwareMajorVersionThreshold:
		// --- Legacy Configuration ---
		// This applies to all known firmwares before the threshold.
		currentSMCConfig = smcControlConfig{
			Firmware:               "Legacy",
			FirmwareProfileID:      profileLegacyID,
			FirmwareProfileVersion: defaultProfileVersion,
			AdapterKey:             smc.KeyIsAdapterEnabledLegacy,
			AdapterEnableBytes:     []byte{0x00},
			AdapterDisableBytes:    []byte{0x01},
			IsLegacyCharging:       true,
			ChargingKeysLegacy:     []string{smc.KeyIsChargingEnabledLegacyBCLM, smc.KeyIsChargingEnabledLegacyBCDS},
			ChargingEnableBytes:    []byte{0x00},
			ChargingDisableBytes:   []byte{0x02},
		}

	default:
		// --- Unknown Configuration ---
		// We set the Mode to "Unknown" but use the modern keys as a safe, forward-looking guess.
		currentSMCConfig = smcControlConfig{
			Firmware:               "Unknown (using latest known behavior)",
			FirmwareProfileID:      profileModernID,
			FirmwareProfileVersion: defaultProfileVersion,
			AdapterKey:             smc.KeyIsAdapterEnabled,
			AdapterEnableBytes:     []byte{0x00},
			AdapterDisableBytes:    []byte{0x08},
			IsLegacyCharging:       false,
			ChargingKeyModern:      smc.KeyIsChargingEnabled,
			ChargingEnableBytes:    []byte{0x00, 0x00, 0x00, 0x00},
			ChargingDisableBytes:   []byte{0x01, 0x00, 0x00, 0x00},
		}
	}
}

func firmwareCompatStatus(major int) string {
	switch {
	case major == FirmwareMajorVersionThreshold:
		return firmwareCompatTested
	case major > FirmwareMajorVersionThreshold:
		return firmwareCompatUntestedNew
	case major > 0 && major < FirmwareMajorVersionThreshold:
		return firmwareCompatUntestedOld
	default:
		return firmwareCompatUnknown
	}
}
