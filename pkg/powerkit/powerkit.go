//go:build darwin
// +build darwin

package powerkit

import (
	"fmt"
	"log"

	"github.com/peterneutron/powerkit-go/internal/iokit"
	"github.com/peterneutron/powerkit-go/internal/smc"
)

// StreamSystemInfo starts monitoring IOKit for power-related events and returns
// a read-only channel that will receive a SystemInfo object whenever a
// change is detected. The returned object will only contain IOKit data.
func StreamSystemInfo() (<-chan *SystemInfo, error) {
	// This channel will be returned to the caller.
	systemInfoChan := make(chan *SystemInfo)

	// Start the low-level IOKit monitor. This spawns a goroutine with a C RunLoop.
	iokit.StartMonitor()

	// Launch a processor goroutine. This goroutine bridges the raw notification
	// signal to a high-level, IOKit-focused SystemInfo object.
	go func() {
		// Ensure the channel is closed when the loop exits, signaling the end of the stream.
		defer close(systemInfoChan)

		// The first event should be sent immediately to provide the initial state.
		// A non-blocking send ensures we don't stall if the consumer isn't ready.
		select {
		case iokit.Updates <- struct{}{}:
		default:
		}

		for range iokit.Updates {
			// When a signal is received, fetch ONLY the fresh IOKit data.
			// Do NOT call the generic GetSystemInfo() here.
			iokitRawData, err := iokit.FetchData()
			if err != nil {
				log.Printf("Error fetching IOKit data in stream: %v", err)
				continue // Skip this update on error
			}

			// Construct a new SystemInfo object containing only the data
			// relevant to this event stream.
			info := &SystemInfo{
				OS: OSInfo{
					Firmware: currentSMCConfig.Firmware,
				},
				IOKit: newIOKitData(iokitRawData),
				SMC:   nil, // Explicitly set SMC to nil for clarity.
			}

			// Populate derived calculations for the IOKit part.
			// This function is safe as it checks for nil pointers.
			calculateDerivedMetrics(info)

			// Send the processed, IOKit-only info object to the consumer.
			systemInfoChan <- info
		}
	}()

	return systemInfoChan, nil
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

	info := &SystemInfo{
		OS: OSInfo{Firmware: currentSMCConfig.Firmware},
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
