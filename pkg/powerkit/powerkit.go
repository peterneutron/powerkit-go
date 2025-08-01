//go:build darwin
// +build darwin

package powerkit

import (
	"fmt"
	"log"

	"github.com/peterneutron/powerkit-go/internal/iokit"
	"github.com/peterneutron/powerkit-go/internal/smc"
)

// GetSystemInfo is the new primary entrypoint to the library.
// It acts as a high-level coordinator for fetching and processing data.
func GetSystemInfo(opts ...FetchOptions) (*SystemInfo, error) {
	// --- Configuration Handling ---
	options := FetchOptions{
		QueryIOKit: true,
		QuerySMC:   true,
	}
	if len(opts) > 0 {
		options = opts[0]
	}
	if !options.QueryIOKit && !options.QuerySMC {
		return nil, fmt.Errorf("FetchOptions must specify at least one data source")
	}

	// --- Main Data Gathering ---
	info := &SystemInfo{}

	// Phase 1: Fetch and populate IOKit data if requested.
	if options.QueryIOKit {
		iokitRawData, err := iokit.FetchData()
		if err != nil {
			if !options.QuerySMC {
				return nil, fmt.Errorf("failed to fetch required IOKit data: %w", err)
			}
			log.Printf("Warning: IOKit data fetch failed, continuing with SMC: %v", err)
		} else {
			// The main function's logic is now clean. It just calls the constructor.
			info.IOKit = newIOKitData(iokitRawData)
		}
	}

	// Phase 2: Fetch and populate SMC data if requested.
	if options.QuerySMC {
		// 1. Fetch the standard float values
		smcFloatResults, err := smc.FetchData(smc.KeysToRead)
		// 2. Fetch the specific raw values we need for the state
		smcRawResults, rawErr := smc.FetchRawData([]string{smc.KeyChargeControl, smc.KeyAdapterControl})
		if err != nil || rawErr != nil {
			if !options.QueryIOKit {
				return nil, fmt.Errorf("failed to fetch required SMC data: %w", err)
			}
			log.Printf("Warning: could not fetch SMC data: %v", err)
		} else {
			// The logic is clean here too.
			info.SMC = newSMCData(smcFloatResults, smcRawResults)
		}
	}

	// Phase 3: Populate derived calculations.
	calculateDerivedMetrics(info)

	return info, nil
}

// GetRawSMCValues allows advanced users to query custom SMC keys.
// It returns a map of raw, undecoded values. The caller is responsible for
// interpreting the bytes in the 'Data' field based on the 'DataType'.
func GetRawSMCValues(keys []string) (map[string]RawSMCValue, error) {
	// Call the new raw fetcher in our internal smc package
	rawResults, err := smc.FetchRawData(keys)
	if err != nil {
		return nil, err
	}

	// The internal smc.RawSMCValue struct is identical to our public one,
	// but it's a best practice to convert between them to keep the packages decoupled.
	// This also makes it easy to change the public API later without breaking the internal code.
	results := make(map[string]RawSMCValue)
	for key, val := range rawResults {
		results[key] = RawSMCValue{
			DataType: val.DataType,
			DataSize: val.DataSize,
			Data:     val.Data,
		}
	}

	return results, nil
}

// ---------------  Public Write API  -------------- //
// WARNING: These functions require root privileges. //

// MagsafeColor defines the possible states for the charging LED.
type MagsafeColor int

const (
	// LEDOff represents the 'Off' state for the Magsafe LED.
	LEDOff MagsafeColor = iota // 0
	// LEDAmber represents the 'Amber' (charging) state for the Magsafe LED.
	LEDAmber // 1
	// LEDGreen represents the 'Green' (fully charged) state for the Magsafe LED.
	LEDGreen // 2
)

// SetMagsafeLEDColor sets the color of the Magsafe charging LED.
func SetMagsafeLEDColor(color MagsafeColor) error {
	// The ACLC key expects two bytes:
	// Byte 0: LED ID (0 for Magsafe)
	// Byte 1: Color code (0=Off, 1=Amber, 2=Green)
	var colorCode byte
	switch color {
	case LEDAmber:
		colorCode = 0x01
	case LEDGreen:
		colorCode = 0x02
	case LEDOff:
		colorCode = 0x00
	default:
		return fmt.Errorf("invalid MagsafeColor provided: %d", color)
	}

	// Prepare the 2-byte slice to write to the SMC.
	data := []byte{0x00, colorCode}

	// Call the internal, generic write function.
	return smc.WriteData(smc.KeyMagsafeLED, data)
}

// EnableCharger enables the charger.
func EnableCharger() error {
	// The CHIE key expects a single byte: 0x0 to enable the charger.
	data := []byte{0x0}
	return smc.WriteData(smc.KeyAdapterControl, data)
}

// DisableCharger disables the charger.
func DisableCharger() error {
	// The CHIE key expects a single byte: 0x8 to disable the charger.
	data := []byte{0x8}
	return smc.WriteData(smc.KeyAdapterControl, data)
}

// EnableChargeInhibit prevents the battery from charging even when the AC
// adapter is connected.
func EnableChargeInhibit() error {
	// The CHTE key expects 4 byte: 0x01, 0x00, 0x00, 0x00 to enable inhibit.
	data := []byte{0x01, 0x00, 0x00, 0x00}
	return smc.WriteData(smc.KeyChargeControl, data)
}

// DisableChargeInhibit allows the battery to resume normal charging.
func DisableChargeInhibit() error {
	// The CHTE key expects 4 byte: 0x00, 0x00, 0x00, 0x00 to disable inhibit.
	data := []byte{0x00, 0x00, 0x00, 0x00}
	return smc.WriteData(smc.KeyChargeControl, data)
}
