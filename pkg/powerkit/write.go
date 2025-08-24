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

// MagsafeLEDState is the full state enum for the MagSafe LED.
type MagsafeLEDState uint8

const (
	// LEDSystem lets the system control the LED automatically.
	LEDSystem MagsafeLEDState = 0x00
	// LEDOff turns the LED off.
	LEDOff MagsafeLEDState = 0x01
	// LEDGreen shows green (often fully charged).
	LEDGreen MagsafeLEDState = 0x03
	// LEDAmber shows amber/orange (charging).
	LEDAmber MagsafeLEDState = 0x04
	// LEDErrorOnce flashes an error once.
	LEDErrorOnce MagsafeLEDState = 0x05
	// LEDErrorPermSlow indicates a persistent slow error blink.
	LEDErrorPermSlow MagsafeLEDState = 0x06
	// LEDErrorPermFast indicates a persistent fast error blink.
	LEDErrorPermFast MagsafeLEDState = 0x07
	// LEDErrorPermOff indicates a persistent error-off state.
	LEDErrorPermOff MagsafeLEDState = 0x19
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
		return setCharging(true)

	case ChargingActionOff:
		return setCharging(false)

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

// SetMagsafeLEDState writes the single-byte LED state to the SMC.
// This uses the common 1-byte ACLC format.
func SetMagsafeLEDState(state MagsafeLEDState) error {
	return smc.WriteData(smc.KeyMagsafeLED, []byte{byte(state)})
}

// GetMagsafeLEDState reads the single-byte LED state. It returns:
// - state: the raw state value (mapped to constants when known)
// - available: false when the key exists but contains no data (or cannot be read)
// - err: transport or read error; unknown state is not an error
func GetMagsafeLEDState() (state MagsafeLEDState, available bool, err error) {
	rawValues, err := GetRawSMCValues([]string{smc.KeyMagsafeLED})
	if err != nil {
		return LEDAmber, false, fmt.Errorf("could not read Magsafe LED state from SMC: %w", err)
	}

	ledValue, ok := rawValues[smc.KeyMagsafeLED]
	if !ok {
		// Key not present – treat as unavailable
		return LEDAmber, false, nil
	}
	if ledValue.DataSize == 0 || len(ledValue.Data) == 0 {
		// Explicitly signal not available when key contains no data
		return LEDAmber, false, nil
	}

	// Interpret the first byte as the state. If the device uses a 2-byte
	// format, byte 0 is typically 0x00 which maps to LEDSystem.
	b := ledValue.Data[0]

	// Map 0x02 to green (seen on some machines)
	if b == 0x02 {
		return LEDGreen, true, nil
	}
	return MagsafeLEDState(b), true, nil
}

// IsMagsafeCharging checks if the Magsafe LED indicates a charging state (Amber).
func IsMagsafeCharging() (bool, error) {
	state, available, err := GetMagsafeLEDState()
	if err != nil {
		return false, fmt.Errorf("could not determine Magsafe charging state: %w", err)
	}
	if !available {
		return false, fmt.Errorf("magsafe LED not available")
	}
	return state == LEDAmber, nil
}

// IsMagsafeAvailable returns true if the ACLC key is present and has data.
func IsMagsafeAvailable() bool {
	_, ok, _ := GetMagsafeLEDState()
	return ok
}
