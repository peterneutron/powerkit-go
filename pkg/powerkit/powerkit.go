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
		smcResults, err := smc.FetchData(smc.KeysToRead)
		if err != nil {
			if !options.QueryIOKit {
				return nil, fmt.Errorf("failed to fetch required SMC data: %w", err)
			}
			log.Printf("Warning: could not fetch SMC data: %v", err)
		} else {
			// The logic is clean here too.
			info.SMC = newSMCData(smcResults)
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
