//go:build darwin
// +build darwin

package powerkit

import (
	"fmt"
	"log"

	"github.com/peterneutron/powerkit-go/internal/iokit"
	"github.com/peterneutron/powerkit-go/internal/powerd"
	"github.com/peterneutron/powerkit-go/internal/smc"
)

// StreamSystemEvents starts monitoring IOKit for all relevant power and
// battery events. It returns a single, read-only channel that delivers a
// unified SystemEvent for any change.
func StreamSystemEvents() (<-chan SystemEvent, error) {
	// The public-facing channel that the consumer will receive.
	systemEventChan := make(chan SystemEvent, 2)

	// Start the low-level IOKit monitor. This is idempotent and safe to call.
	// It spawns one C RunLoop to handle all event sources.
	iokit.StartMonitor()

	// Launch the dispatcher goroutine. This is the core of the unified model.
	// It bridges the internal, low-level events to the public, high-level API.
	go func() {
		// Ensure the public channel is closed when this goroutine exits.
		defer close(systemEventChan)

		// On startup, immediately trigger a battery update to provide the
		// consumer with an initial state, just like a real event.
		select {
		case iokit.Events <- iokit.InternalEvent{Type: iokit.BatteryUpdate}:
		default:
		}

		// Main event processing loop.
		for internalEvent := range iokit.Events {
			var publicEvent SystemEvent

			// Translate the internal event type to the public one.
			switch internalEvent.Type {
			case iokit.BatteryUpdate:
				// For a battery update, we need to fetch the full IOKit data.
				iokitRawData, err := iokit.FetchData()
				if err != nil {
					log.Printf("Error fetching IOKit data in stream: %v", err)
					continue // Skip this event on error
				}

				// Construct the SystemInfo object, which will be the payload.
				// Determine assertion-based sleep allowances (global + app-local)
				sysAllowedGlobal, dspAllowedGlobal, gErr := powerd.GlobalSleepStatus()
				dspActiveApp := powerd.IsActive(powerd.PreventDisplaySleep)
				sysActiveApp := powerd.IsActive(powerd.PreventSystemSleep)
				dspAllowedApp := !dspActiveApp
				sysAllowedApp := !sysActiveApp && !dspActiveApp
				// If global query failed, mirror app-local into global as a fallback
				if gErr != nil {
					dspAllowedGlobal = dspAllowedApp
					sysAllowedGlobal = sysAllowedApp
				}

				info := &SystemInfo{
					OS: OSInfo{
						Firmware: currentSMCConfig.Firmware,
						// Back-compat: mirror global values
						GlobalSystemSleepAllowed:  sysAllowedGlobal,
						GlobalDisplaySleepAllowed: dspAllowedGlobal,
						AppSystemSleepAllowed:     sysAllowedApp,
						AppDisplaySleepAllowed:    dspAllowedApp,
					},
					IOKit: newIOKitData(iokitRawData),
					SMC:   nil, // SMC data is not queried in event streams.
				}
				calculateDerivedMetrics(info) // Populate calculated fields.

				// Build the final public event.
				publicEvent = SystemEvent{
					Type: EventTypeBatteryUpdate,
					Info: info,
				}

			case iokit.SystemWillSleep:
				publicEvent = SystemEvent{Type: EventTypeSystemWillSleep, Info: nil}

			case iokit.SystemDidWake:
				publicEvent = SystemEvent{Type: EventTypeSystemDidWake, Info: nil}

			default:
				// Should not happen, but good to have a fallback.
				log.Printf("Warning: Received unknown internal event type: %d", internalEvent.Type)
				continue
			}

			// Send the fully constructed event to the consumer.
			systemEventChan <- publicEvent
		}
	}()

	return systemEventChan, nil
}

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
	sysAllowedGlobal, dspAllowedGlobal, gErr := powerd.GlobalSleepStatus()
	dspActiveApp := powerd.IsActive(powerd.PreventDisplaySleep)
	sysActiveApp := powerd.IsActive(powerd.PreventSystemSleep)
	dspAllowedApp := !dspActiveApp
	sysAllowedApp := !sysActiveApp && !dspActiveApp
	// If global query failed, mirror app-local into global as a fallback
	if gErr != nil {
		dspAllowedGlobal = dspAllowedApp
		sysAllowedGlobal = sysAllowedApp
	}

	info := &SystemInfo{
		OS: OSInfo{
			Firmware: currentSMCConfig.Firmware,
			// Back-compat: mirror global values
			GlobalSystemSleepAllowed:  sysAllowedGlobal,
			GlobalDisplaySleepAllowed: dspAllowedGlobal,
			AppSystemSleepAllowed:     sysAllowedApp,
			AppDisplaySleepAllowed:    dspAllowedApp,
		},
	}

	if options.QueryIOKit {
		getIOKitInfo(info)
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
