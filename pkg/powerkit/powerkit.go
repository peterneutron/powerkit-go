//go:build darwin

package powerkit

import (
	"fmt"

	sysos "github.com/peterneutron/powerkit-go/internal/os"
	"github.com/peterneutron/powerkit-go/internal/powerd"
	"github.com/peterneutron/powerkit-go/internal/smc"
)

var (
	globalSleepStatusFn = powerd.GlobalSleepStatus
	powerdIsActiveFn    = powerd.IsActive
	getLowPowerModeFn   = sysos.GetLowPowerModeEnabled
)

// GetSystemInfo is the primary entrypoint to the library.
// It acts as a high-level coordinator for fetching and processing data.
func GetSystemInfo(opts ...FetchOptions) (*SystemInfo, error) {
	options := FetchOptions{QueryIOKit: true, QuerySMC: true}
	if len(opts) > 0 {
		options = opts[0]
	}
	if !options.QueryIOKit && !options.QuerySMC {
		return nil, fmt.Errorf("FetchOptions must specify at least one data source")
	}

	// Determine assertion-based sleep allowances (global + app-local)
	sysAllowedGlobal, dspAllowedGlobal, gErr := globalSleepStatusFn()
	dspActiveApp := powerdIsActiveFn(powerd.PreventDisplaySleep)
	sysActiveApp := powerdIsActiveFn(powerd.PreventSystemSleep)
	dspAllowedApp := !dspActiveApp
	sysAllowedApp := !sysActiveApp && !dspActiveApp
	// If global query failed, mirror app-local into global as a fallback
	if gErr != nil {
		dspAllowedGlobal = dspAllowedApp
		sysAllowedGlobal = sysAllowedApp
	}

	// Read Low Power Mode state (cached)
	lpmEnabled, lpmAvailable, _ := getLowPowerModeFn()

	info := &SystemInfo{
		OS: OSInfo{
			Firmware:               currentSMCConfig.Firmware,
			FirmwareVersion:        currentFirmwareInfo.Version,
			FirmwareSource:         currentFirmwareInfo.Source,
			FirmwareMajor:          currentFirmwareInfo.Major,
			FirmwareCompatStatus:   firmwareCompatStatus(currentFirmwareInfo.Major),
			FirmwareProfileID:      currentSMCConfig.FirmwareProfileID,
			FirmwareProfileVersion: currentSMCConfig.FirmwareProfileVersion,
			// Back-compat: mirror global values
			GlobalSystemSleepAllowed:  sysAllowedGlobal,
			GlobalDisplaySleepAllowed: dspAllowedGlobal,
			AppSystemSleepAllowed:     sysAllowedApp,
			AppDisplaySleepAllowed:    dspAllowedApp,
			LowPowerMode:              LowPowerModeInfo{Enabled: lpmEnabled, Available: lpmAvailable},
		},
	}
	initSystemInfoMetadata(info)

	if options.QueryIOKit {
		getIOKitInfo(info, options)
	}
	if options.QuerySMC {
		getSMCInfo(info)
	}

	// Final check: if both failed, we have nothing to return.
	if info.IOKit == nil && info.SMC == nil {
		return nil, fmt.Errorf("failed to fetch data from all sources")
	}

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
	results := make(map[string]RawSMCValue, len(rawResults))
	for key := range rawResults {
		val := rawResults[key]
		results[key] = RawSMCValue{
			DataType: val.DataType,
			DataSize: val.DataSize,
			Data:     val.Data,
		}
	}

	return results, nil
}
