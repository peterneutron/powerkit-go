//go:build darwin
// +build darwin

package powerkit

import (
	"bytes"
	"fmt"

	"github.com/peterneutron/powerkit-go/internal/smc"
)

// ---------------  Public Write API  -------------- //
// WARNING: These functions require root privileges. //

var (
	// For SetAdapterState (uses KeyIsAdapterEnabled)
	adapterEnabledBytes  = []byte{0x00}
	adapterDisabledBytes = []byte{0x08}

	// For SetChargingState (uses KeyIsChargingEnabled)
	chargingEnabledBytes  = []byte{0x00, 0x00, 0x00, 0x00}
	chargingDisabledBytes = []byte{0x01, 0x00, 0x00, 0x00}
)

// AdapterAction sets the desired Adapter state
type AdapterAction int

const (
	// AdapterActionToggle will read the current state and perform the opposite action.
	AdapterActionToggle = iota // 0
	// AdapterActionOn will force the charger to be enabled.
	AdapterActionOn // 1
	// AdapterActionOff will force the charger to be disabled.
	AdapterActionOff // 2
)

// ChargingAction sets the desired charging state
type ChargingAction int

const (
	// ChargingActionToggle will read the current state and perform the opposite action.
	ChargingActionToggle = iota // 0
	// ChargingActionOn will force the charger to be enabled.
	ChargingActionOn // 1
	// ChargingActionOff will force the charger to be disabled.
	ChargingActionOff // 2
)

// MagsafeColor defines the possible states for the charging LED.
type MagsafeColor int

const (
	// LEDOff represents the 'Off' state for the Magsafe LED.
	LEDOff MagsafeColor = iota // 0
	// LEDAmber represents the 'Amber' (charging) state for the Magsafe LED.
	LEDAmber // 1
	// LEDGreen represents the 'Green' (fully charged) state for the Magsafe LED.
	LEDGreen // 2
)

func SetAdapterState(action AdapterAction) error {
	switch action {
	case AdapterActionOn:
		fmt.Println("Forcing charger ON...")
		return smc.WriteData(smc.KeyIsAdapterEnabled, adapterEnabledBytes)

	case AdapterActionOff:
		fmt.Println("Forcing charger OFF...")
		return smc.WriteData(smc.KeyIsAdapterEnabled, adapterDisabledBytes)

	case AdapterActionToggle:
		fmt.Println("Reading current charger state to perform toggle...")
		rawValues, err := GetRawSMCValues([]string{smc.KeyIsAdapterEnabled})
		if err != nil {
			return fmt.Errorf("could not read current charger state: %w", err)
		}

		chargerValue, ok := rawValues[smc.KeyIsAdapterEnabled]
		if !ok {
			return fmt.Errorf("could not find key '%s' on this system", smc.KeyIsAdapterEnabled)
		}

		isChargerDisabled := bytes.Equal(chargerValue.Data, adapterDisabledBytes)

		if isChargerDisabled {
			fmt.Println("Charger is currently OFF. Toggling ON...")
			return SetAdapterState(AdapterActionOn)
		} else {
			fmt.Println("Charger is currently ON. Toggling OFF...")
			return SetAdapterState(AdapterActionOff)
		}

	default:
		return fmt.Errorf("invalid AdapterAction provided")
	}
}

func SetChargingState(action ChargingAction) error {
	switch action {
	case ChargingActionOn:
		fmt.Println("Forcing charging ON...")
		return smc.WriteData(smc.KeyIsChargingEnabled, chargingEnabledBytes)

	case ChargingActionOff:
		fmt.Println("Forcing charging OFF...")
		return smc.WriteData(smc.KeyIsChargingEnabled, chargingDisabledBytes)

	case ChargingActionToggle:
		fmt.Println("Reading current charging state to perform toggle...")
		rawValues, err := GetRawSMCValues([]string{smc.KeyIsChargingEnabled})
		if err != nil {
			return fmt.Errorf("could not read current charging state: %w", err)
		}

		chargerValue, ok := rawValues[smc.KeyIsChargingEnabled]
		if !ok {
			return fmt.Errorf("could not find key '%s' on this system", smc.KeyIsChargingEnabled)
		}

		isChargingDisabled := bytes.Equal(chargerValue.Data, chargingDisabledBytes)

		if isChargingDisabled {
			fmt.Println("Charging is currently OFF. Toggling ON...")
			return SetChargingState(ChargingActionOn)
		} else {
			fmt.Println("Charging is currently ON. Toggling OFF...")
			return SetChargingState(ChargingActionOff)
		}

	default:
		return fmt.Errorf("invalid ChargingAction provided")
	}
}

// SetMagsafeLEDColor sets the color of the Magsafe charging LED.
func SetMagsafeLEDColor(color MagsafeColor) error {
	// The ACLC key expects two bytes:
	// Byte 0: LED ID (0 for Magsafe)
	// Byte 1: Color code (0=Off, 1=Amber, 2=Green)
	var colorCode byte
	switch color {
	case LEDAmber:
		colorCode = 0x01
	case LEDGreen:
		colorCode = 0x02
	case LEDOff:
		colorCode = 0x00
	default:
		return fmt.Errorf("invalid MagsafeColor provided: %d", color)
	}

	// Prepare the 2-byte slice to write to the SMC.
	data := []byte{0x00, colorCode}

	// Call the internal, generic write function.
	return smc.WriteData(smc.KeyMagsafeLED, data)
}
