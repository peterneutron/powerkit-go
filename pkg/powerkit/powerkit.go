//go:build darwin
// +build darwin

package power

import (
	"log"
	"math"

	"github.com/peterneutron/powerkit-go/internal/iokit"
	"github.com/peterneutron/powerkit-go/internal/smc"
)

// GetBatteryInfo is the primary public entrypoint to the library.
func GetBatteryInfo() (*BatteryInfo, error) {
	// Phase 1: Fetch data from IOKit.
	rawData, err := iokit.FetchData()
	if err != nil {
		return nil, err
	}

	// Convert from the internal IOKit struct to the public Go struct.
	// This is where unit conversions happen (e.g., mV -> V).
	info := &BatteryInfo{
		State: State{
			IsCharging:   rawData.IsCharging,
			IsConnected:  rawData.IsConnected,
			FullyCharged: rawData.IsFullyCharged,
		},
		Battery: Battery{
			SerialNumber:           rawData.SerialNumber,
			DeviceName:             rawData.DeviceName,
			CycleCount:             rawData.CycleCount,
			DesignCapacity:         rawData.DesignCapacity,
			MaxCapacity:            rawData.MaxCapacity,
			NominalCapacity:        rawData.NominalCapacity,
			CurrentCapacity:        rawData.CurrentCapacity,
			TimeToEmpty:            rawData.TimeToEmpty,
			TimeToFull:             rawData.TimeToFull,
			Temperature:            float64(rawData.Temperature) / 100.0,
			Voltage:                float64(rawData.Voltage) / 1000.0,
			Amperage:               float64(rawData.Amperage) / 1000.0,
			IndividualCellVoltages: rawData.CellVoltages,
		},
		Adapter: Adapter{
			Description:   rawData.AdapterDesc,
			MaxWatts:      rawData.AdapterWatts,
			MaxVoltage:    float64(rawData.AdapterVoltage) / 1000.0,
			MaxAmperage:   float64(rawData.AdapterAmperage) / 1000.0,
			InputVoltage:  float64(rawData.SourceVoltage) / 1000.0,
			InputAmperage: float64(rawData.SourceAmperage) / 1000.0,
		},
	}

	// Fetch and populate real-time SMC data into its own struct.
	if smcResults, err := smc.FetchData(smc.KeysToRead); err == nil {
		info.SMC = &SMC{}

		if val, ok := smcResults["VD0R"]; ok {
			roundedVal := math.Round(val*100) / 100
			info.SMC.InputVoltage = roundedVal
		}
		if val, ok := smcResults["ID0R"]; ok {
			roundedVal := math.Round(val*100) / 100
			info.SMC.InputAmperage = roundedVal
		}
		if val, ok := smcResults["PDTR"]; ok {
			info.SMC.InputPower = math.Round(val*100) / 100
		}
		if val, ok := smcResults["B0AV"]; ok {
			roundedVal := math.Round((val/1000.0)*100) / 100
			info.SMC.BatteryVoltage = roundedVal
		}
		if val, ok := smcResults["B0AC"]; ok {
			roundedVal := math.Round((val/1000.0)*100) / 100
			info.SMC.BatteryAmperage = roundedVal
		}
		if val, ok := smcResults["PPBR"]; ok {
			info.SMC.BatteryPower = math.Round(val*100) / 100
		}
		if val, ok := smcResults["PSTR"]; ok {
			info.SMC.SystemPower = math.Round(val*100) / 100
		}
	} else {
		log.Printf("Could not fetch SMC data, falling back to IOKit: %v", err)
	}

	// Populate derived calculations using the best available data.
	calculateDerivedMetrics(info)

	return info, nil
}
