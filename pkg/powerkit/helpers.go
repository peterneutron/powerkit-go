// Package powerkit provides a high-level API for querying and controlling
// macOS power management features.
package powerkit

import (
	"fmt"
	"math"

	"log"

	"github.com/peterneutron/powerkit-go/internal/iokit"
	"github.com/peterneutron/powerkit-go/internal/smc"
)

// truncate rounds a float down to two decimal places. This is used
// consistently across the library for formatting final values.
func truncate(f float64) float64 {
	return math.Trunc(f*100) / 100
}

// findMinMax finds the minVal and maxVal values in a slice of integers.
// (This is also a helper, so it's good to move it here from calculations.go)
func findMinMax(a []int) (minVal int, maxVal int) {
	if len(a) == 0 {
		return 0, 0
	}
	minVal = a[0]
	maxVal = a[0]
	for _, value := range a {
		if value < minVal {
			minVal = value
		}
		if value > maxVal {
			maxVal = value
		}
	}
	return minVal, maxVal
}

// We'll create a helper function for the "On" and "Off" logic.
func setCharging(enable bool) error {
	var bytesToWrite []byte
	if enable {
		bytesToWrite = currentSMCConfig.ChargingEnableBytes
	} else {
		bytesToWrite = currentSMCConfig.ChargingDisableBytes
	}

	if currentSMCConfig.IsLegacyCharging {
		for _, key := range currentSMCConfig.ChargingKeysLegacy {
			if err := smc.WriteData(key, bytesToWrite); err != nil {
				return fmt.Errorf("failed to write to legacy charging key '%s': %w", key, err)
			}
		}
		return nil
	}
	return smc.WriteData(currentSMCConfig.ChargingKeyModern, bytesToWrite)
}

// Create a helper for fetching IOKit data
func getIOKitInfo(info *SystemInfo, options FetchOptions) {
	iokitRawData, err := iokit.FetchData(options.ForceTelemetryFallback)
	if err != nil {
		log.Printf("Warning: IOKit data fetch failed, continuing without it: %v", err)
		return
	}
	info.IOKit = newIOKitData(iokitRawData)
}

// Create a helper for fetching SMC data
func getSMCInfo(info *SystemInfo) {
	// Build SMC key lists: separate numeric sensor keys (for float decode)
	// from control/state keys that are not decodable to float.
	floatKeys := []string{
		smc.KeyAdapterVoltage,
		smc.KeyAdapterCurrent,
		smc.KeyBatteryVoltage,
		smc.KeyBatteryCurrent,
	}
	rawKeys := append([]string{}, floatKeys...)
	// Adapter enable state key depends on firmware
	rawKeys = append(rawKeys, currentSMCConfig.AdapterKey)
	// Charging enable state keys depend on legacy vs modern
	if currentSMCConfig.IsLegacyCharging {
		rawKeys = append(rawKeys, currentSMCConfig.ChargingKeysLegacy...)
	} else {
		rawKeys = append(rawKeys, currentSMCConfig.ChargingKeyModern)
	}

	smcFloatResults, err1 := smc.FetchData(floatKeys)
	smcRawResults, err2 := smc.FetchRawData(rawKeys)
	if err1 != nil || err2 != nil {
		log.Printf("Warning: SMC data fetch failed, continuing without it. FltErr: %v, RawErr: %v", err1, err2)
		return
	}
	info.SMC = newSMCData(smcFloatResults, smcRawResults)
}
