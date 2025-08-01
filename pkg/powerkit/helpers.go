// pkg/powerkit/helpers.go
package powerkit

import (
	"math"

	"github.com/peterneutron/powerkit-go/internal/iokit"
	"github.com/peterneutron/powerkit-go/internal/smc"
)

// truncate rounds a float down to two decimal places. This is used
// consistently across the library for formatting final values.
func truncate(f float64) float64 {
	return math.Trunc(f*100) / 100
}

// findMinMax finds the min and max values in a slice of integers.
// (This is also a helper, so it's good to move it here from calculations.go)
func findMinMax(a []int) (min int, max int) {
	if len(a) == 0 {
		return 0, 0
	}
	min = a[0]
	max = a[0]
	for _, value := range a {
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}
	return min, max
}

// newSMCData is a private helper that transforms raw SMC key-value data
// into the public SMCData struct.
func newSMCData(results map[string]float64) *SMCData {
	data := &SMCData{
		Battery: SMCBattery{},
		Adapter: SMCAdapter{},
	}

	if val, ok := results[smc.KeyDCInVoltage]; ok {
		data.Adapter.InputVoltage = truncate(val)
	}
	if val, ok := results[smc.KeyDCInCurrent]; ok {
		data.Adapter.InputAmperage = truncate(val)
	}
	if val, ok := results[smc.KeyBatteryVoltage]; ok {
		data.Battery.Voltage = truncate(val / 1000.0)
	}
	if val, ok := results[smc.KeyBatteryCurrent]; ok {
		data.Battery.Amperage = truncate(val / 1000.0)
	}
	return data
}

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
