//go:build darwin
// +build darwin

package powerkit

import (
	"bytes"

	"github.com/peterneutron/powerkit-go/internal/iokit"
	"github.com/peterneutron/powerkit-go/internal/smc"
)

// newSMCData is a private helper that transforms raw SMC key-value data
// into the public SMCData struct.
func newSMCData(floatResults map[string]float64, rawResults map[string]smc.RawSMCValue) *SMCData {
	data := &SMCData{
		State:   SMCState{},
		Battery: SMCBattery{},
		Adapter: SMCAdapter{},
	}

	if val, ok := floatResults[smc.KeyAdapterVoltage]; ok {
		data.Adapter.InputVoltage = truncate(val)
	}
	if val, ok := floatResults[smc.KeyAdapterCurrent]; ok {
		data.Adapter.InputAmperage = truncate(val)
	}
	if val, ok := floatResults[smc.KeyBatteryVoltage]; ok {
		data.Battery.Voltage = truncate(val / 1000.0)
	}
	if val, ok := floatResults[smc.KeyBatteryCurrent]; ok {
		data.Battery.Amperage = truncate(val / 1000.0)
	}

	// --- Populate the State struct from the rawResults, respecting OS version ---

	// Check for IsChargingEnabled state
	var chargingKeyToCheck string
	if currentSMCConfig.IsLegacyCharging {
		chargingKeyToCheck = smc.KeyIsChargingEnabledLegacyBCLM
	} else {
		chargingKeyToCheck = smc.KeyIsChargingEnabled
	}
	if val, ok := rawResults[chargingKeyToCheck]; ok {
		// Enabled is the default; we check for the disabled bytes.
		// A value not equal to the 'disabled' state is considered 'enabled'.
		if !bytes.Equal(val.Data, currentSMCConfig.ChargingDisableBytes) {
			data.State.IsChargingEnabled = true
		}
	}

	// Check for IsAdapterEnabled state
	adapterKeyToCheck := currentSMCConfig.AdapterKey
	if val, ok := rawResults[adapterKeyToCheck]; ok {
		// Enabled is the default; we check for the disabled bytes.
		if !bytes.Equal(val.Data, currentSMCConfig.AdapterDisableBytes) {
			data.State.IsAdapterEnabled = true
		}
	}

	return data
}

// newIOKitData is a private helper that transforms raw IOKit data
// into the public IOKitData struct. This is its only job.
func newIOKitData(raw *iokit.RawData) *IOKitData {
	return &IOKitData{
		State: IOKitState{
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
			CurrentCharge:          raw.CurrentCharge,
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
