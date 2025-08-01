//go:build darwin
// +build darwin

package powerkit

import (
	"fmt"
	"log"

	"github.com/peterneutron/powerkit-go/internal/iokit"
	"github.com/peterneutron/powerkit-go/internal/smc"
)

// GetBatteryInfo is the primary public entrypoint to the library.
// It accepts optional FetchOptions to control which data sources are queried.
func GetBatteryInfo(opts ...FetchOptions) (*BatteryInfo, error) {
	// --- Configuration Handling ---
	// Set default options: query both sources.
	options := FetchOptions{
		QueryIOKit: true,
		QuerySMC:   true,
	}
	// If the user provided options, use them instead.
	if len(opts) > 0 {
		options = opts[0]
	}
	// If the user wants neither, that's an error.
	if !options.QueryIOKit && !options.QuerySMC {
		return nil, fmt.Errorf("FetchOptions must specify at least one data source (IOKit or SMC)")
	}

	// --- Main Data Gathering ---
	// Create a new, empty info struct. We will populate it conditionally.
	info := &BatteryInfo{}

	// Phase 1: Fetch and populate from IOKit ONLY IF requested.
	if options.QueryIOKit {
		rawData, err := iokit.FetchData()
		if err != nil {
			// If IOKit is the ONLY source requested, this is a fatal error.
			if !options.QuerySMC {
				return nil, fmt.Errorf("failed to fetch required IOKit data: %w", err)
			}
			// Otherwise, just log a warning and proceed.
			log.Printf("Warning: IOKit data fetch failed, continuing with SMC: %v", err)
		} else {
			// Populate all the IOKit-related structs.
			info.State = State{
				IsCharging:   rawData.IsCharging,
				IsConnected:  rawData.IsConnected,
				FullyCharged: rawData.IsFullyCharged,
			}
			info.Battery = Battery{
				SerialNumber:           rawData.SerialNumber,
				DeviceName:             rawData.DeviceName,
				CycleCount:             rawData.CycleCount,
				DesignCapacity:         rawData.DesignCapacity,
				MaxCapacity:            rawData.MaxCapacity,
				NominalCapacity:        rawData.NominalCapacity,
				CurrentCapacity:        rawData.CurrentCapacity,
				TimeToEmpty:            rawData.TimeToEmpty,
				TimeToFull:             rawData.TimeToFull,
				Temperature:            truncate(float64(rawData.Temperature) / 100.0),
				Voltage:                truncate(float64(rawData.Voltage) / 1000.0),
				Amperage:               truncate(float64(rawData.Amperage) / 1000.0),
				IndividualCellVoltages: rawData.CellVoltages,
			}
			info.Adapter = Adapter{
				Description:   rawData.AdapterDesc,
				MaxWatts:      rawData.AdapterWatts,
				MaxVoltage:    truncate(float64(rawData.AdapterVoltage) / 1000.0),
				MaxAmperage:   truncate(float64(rawData.AdapterAmperage) / 1000.0),
				InputVoltage:  truncate(float64(rawData.SourceVoltage) / 1000.0),
				InputAmperage: truncate(float64(rawData.SourceAmperage) / 1000.0),
			}
		}
	}

	// Phase 2: Fetch and populate real-time SMC data ONLY IF requested.
	if options.QuerySMC {
		smcResults, err := smc.FetchData(smc.KeysToRead)
		if err != nil {
			// If SMC is the ONLY source requested, this is a fatal error.
			if !options.QueryIOKit {
				return nil, fmt.Errorf("failed to fetch required SMC data: %w", err)
			}
			// Otherwise, log it and proceed with only the IOKit data (if any).
			log.Printf("Warning: could not fetch SMC data: %v", err)
		} else {
			// Populate the SMC struct.
			info.SMC = &SMC{}
			if val, ok := smcResults["VD0R"]; ok {
				info.SMC.InputVoltage = truncate(val)
			}
			if val, ok := smcResults["ID0R"]; ok {
				info.SMC.InputAmperage = truncate(val)
			}
			if val, ok := smcResults["PDTR"]; ok {
				info.SMC.InputPower = truncate(val)
			}
			if val, ok := smcResults["B0AV"]; ok {
				info.SMC.BatteryVoltage = truncate(val / 1000.0)
			}
			if val, ok := smcResults["B0AC"]; ok {
				info.SMC.BatteryAmperage = truncate(val / 1000.0)
			}
			if val, ok := smcResults["PPBR"]; ok {
				info.SMC.BatteryPower = truncate(val)
			}
			if val, ok := smcResults["PSTR"]; ok {
				info.SMC.SystemPower = truncate(val)
			}

			// Overwrite less-accurate values ONLY if IOKit was also queried.
			// This prevents populating fields on an empty Battery struct.
			if options.QueryIOKit {
				if info.SMC.BatteryVoltage > 0 {
					info.Battery.Voltage = info.SMC.BatteryVoltage
				}
				if info.SMC.BatteryAmperage != 0 {
					info.Battery.Amperage = info.SMC.BatteryAmperage
				}
				if info.SMC.InputVoltage > 0 {
					info.Adapter.InputVoltage = info.SMC.InputVoltage
				}
				if info.SMC.InputAmperage != 0 {
					info.Adapter.InputAmperage = info.SMC.InputAmperage
				}
			}
		}
	}

	// Phase 3: Populate derived calculations using the best available data.
	calculateDerivedMetrics(info)

	return info, nil
}
