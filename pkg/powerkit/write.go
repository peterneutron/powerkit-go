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

// SetAdapterState sets the desired adapter state (On, Off, or Toggle).
// This function requires root privileges.
func SetAdapterState(action AdapterAction) error {
	key := currentSMCConfig.AdapterKey

	switch action {
	case AdapterActionOn:
		fmt.Println("Forcing adapter ON...")
		return smc.WriteData(key, currentSMCConfig.AdapterEnableBytes)

	case AdapterActionOff:
		fmt.Println("Forcing adapter OFF...")
		return smc.WriteData(key, currentSMCConfig.AdapterDisableBytes)

	case AdapterActionToggle:
		fmt.Println("Reading current adapter state to perform toggle...")
		rawValues, err := GetRawSMCValues([]string{key})
		if err != nil {
			return fmt.Errorf("could not read current adapter state: %w", err)
		}

		adapterValue, ok := rawValues[key]
		if !ok {
			return fmt.Errorf("could not find key '%s' on this system", key)
		}

		isAdapterDisabled := bytes.Equal(adapterValue.Data, currentSMCConfig.AdapterDisableBytes)

		if isAdapterDisabled {
			fmt.Println("Adapter is currently OFF. Toggling ON...")
			return SetAdapterState(AdapterActionOn)
		}
		fmt.Println("Adapter is currently ON. Toggling OFF...")
		return SetAdapterState(AdapterActionOff)

	default:
		return fmt.Errorf("invalid AdapterAction provided")
	}
}

// SetChargingState sets the desired charging state (On, Off, or Toggle).
// This function requires root privileges.
func SetChargingState(action ChargingAction) error {
	switch action {
	case ChargingActionOn:
		fmt.Println("Forcing charging ON...")
		if currentSMCConfig.IsLegacyCharging {
			for _, key := range currentSMCConfig.ChargingKeysLegacy {
				if err := smc.WriteData(key, currentSMCConfig.ChargingEnableBytes); err != nil {
					return fmt.Errorf("failed to write to legacy charging key '%s': %w", key, err)
				}
			}
			return nil
		}
		return smc.WriteData(currentSMCConfig.ChargingKeyModern, currentSMCConfig.ChargingEnableBytes)

	case ChargingActionOff:
		fmt.Println("Forcing charging OFF...")
		if currentSMCConfig.IsLegacyCharging {
			for _, key := range currentSMCConfig.ChargingKeysLegacy {
				if err := smc.WriteData(key, currentSMCConfig.ChargingDisableBytes); err != nil {
					return fmt.Errorf("failed to write to legacy charging key '%s': %w", key, err)
				}
			}
			return nil
		}
		return smc.WriteData(currentSMCConfig.ChargingKeyModern, currentSMCConfig.ChargingDisableBytes)

	case ChargingActionToggle:
		fmt.Println("Reading current charging state to perform toggle...")
		// For toggle, we only need to read one key to determine the state.
		keyToRead := currentSMCConfig.ChargingKeyModern
		if currentSMCConfig.IsLegacyCharging {
			keyToRead = currentSMCConfig.ChargingKeysLegacy[0] // BCLM is sufficient
		}

		rawValues, err := GetRawSMCValues([]string{keyToRead})
		if err != nil {
			return fmt.Errorf("could not read current charging state: %w", err)
		}

		chargerValue, ok := rawValues[keyToRead]
		if !ok {
			return fmt.Errorf("could not find key '%s' on this system", keyToRead)
		}

		isChargingDisabled := bytes.Equal(chargerValue.Data, currentSMCConfig.ChargingDisableBytes)

		if isChargingDisabled {
			fmt.Println("Charging is currently OFF. Toggling ON...")
			return SetChargingState(ChargingActionOn)
		}

		fmt.Println("Charging is currently ON. Toggling OFF...")
		return SetChargingState(ChargingActionOff)

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

// GetMagsafeLEDColor reads the current color of the Magsafe charging LED.
func GetMagsafeLEDColor() (MagsafeColor, error) {
	rawValues, err := GetRawSMCValues([]string{smc.KeyMagsafeLED})
	if err != nil {
		return LEDOff, fmt.Errorf("could not read Magsafe LED state from SMC: %w", err)
	}

	ledValue, ok := rawValues[smc.KeyMagsafeLED]
	if !ok {
		return LEDOff, fmt.Errorf("could not find Magsafe LED key '%s' on this system", smc.KeyMagsafeLED)
	}

	// The ACLC key's data is 2 bytes. Byte 0 is the LED ID, Byte 1 is the color code.
	if ledValue.DataSize < 2 {
		return LEDOff, fmt.Errorf("unexpected data size for Magsafe LED key: got %d bytes, want at least 2", ledValue.DataSize)
	}

	colorCode := ledValue.Data[1]
	switch colorCode {
	case 0x00:
		return LEDOff, nil
	case 0x01:
		return LEDAmber, nil
	case 0x02:
		return LEDGreen, nil
	default:
		// Other states exist but are not mapped to our public enum.
		// Returning an error for unhandled states is the safest approach.
		return LEDOff, fmt.Errorf("unknown Magsafe LED color code received: 0x%02x", colorCode)
	}
}

// IsMagsafeCharging checks if the Magsafe LED indicates a charging state (Amber).
func IsMagsafeCharging() (bool, error) {
	color, err := GetMagsafeLEDColor()
	if err != nil {
		return false, fmt.Errorf("could not determine Magsafe charging state: %w", err)
	}
	return color == LEDAmber, nil
}
