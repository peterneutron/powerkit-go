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

// --- Constructor Functions ---

// newIOKitData is a private helper that transforms raw IOKit data
// into the public IOKitData struct. This is its only job.
func newIOKitData(raw *iokit.RawData) *IOKitData {
	return &IOKitData{
		State: State{
			IsCharging:   raw.IsCharging,
			IsConnected:  raw.IsConnected,
			FullyCharged: raw.IsFullyCharged,
		},
		Battery: IOKitBattery{
			SerialNumber:           raw.SerialNumber,
			DeviceName:             raw.DeviceName,
			CycleCount:             raw.CycleCount,
			DesignCapacity:         raw.DesignCapacity,
			MaxCapacity:            raw.MaxCapacity,
			NominalCapacity:        raw.NominalCapacity,
			CurrentCapacity:        raw.CurrentCapacity,
			TimeToEmpty:            raw.TimeToEmpty,
			TimeToFull:             raw.TimeToFull,
			Temperature:            truncate(float64(raw.Temperature) / 100.0),
			Voltage:                truncate(float64(raw.Voltage) / 1000.0),
			Amperage:               truncate(float64(raw.Amperage) / 1000.0),
			IndividualCellVoltages: raw.CellVoltages,
		},
		Adapter: IOKitAdapter{
			Description:   raw.AdapterDesc,
			MaxWatts:      raw.AdapterWatts,
			MaxVoltage:    truncate(float64(raw.AdapterVoltage) / 1000.0),
			MaxAmperage:   truncate(float64(raw.AdapterAmperage) / 1000.0),
			InputVoltage:  truncate(float64(raw.SourceVoltage) / 1000.0),
			InputAmperage: truncate(float64(raw.SourceAmperage) / 1000.0),
		},
	}
}

// newSMCData is a private helper that transforms raw SMC key-value data
// into the public SMCData struct.
func newSMCData(results map[string]float64) *SMCData {
	data := &SMCData{
		Battery: SMCBattery{},
		Adapter: SMCAdapter{},
	}

	if val, ok := results["VD0R"]; ok {
		data.Adapter.InputVoltage = truncate(val)
	}
	if val, ok := results["ID0R"]; ok {
		data.Adapter.InputAmperage = truncate(val)
	}
	if val, ok := results["B0AV"]; ok {
		data.Battery.Voltage = truncate(val / 1000.0)
	}
	if val, ok := results["B0AC"]; ok {
		data.Battery.Amperage = truncate(val / 1000.0)
	}
	return data
}
