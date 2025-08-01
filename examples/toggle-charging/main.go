// This example demonstrates a complete read-then-write workflow.
// It reads the current state of the charge inhibit setting and toggles it,
// effectively toggling the charger's state.
package main

import (
	"bytes" // Import the bytes package for safe slice comparison
	"fmt"
	"log"

	"github.com/peterneutron/powerkit-go/pkg/powerkit"
	// We import the internal smc package to get access to the key constants
	"github.com/peterneutron/powerkit-go/internal/smc"
)

func main() {
	// --- 1. Define the Key and Expected Values using Constants ---
	// Using constants from the smc package makes this code safe from typos.
	chargeInhibitKey := smc.KeyChargeControl
	enabledBytes := []byte{0x00} // The byte sequence for "inhibit enabled" (charging disabled)

	// --- 2. Read the Current State ---
	fmt.Printf("Reading current state of SMC key '%s'...\n", chargeInhibitKey)
	rawValues, err := powerkit.GetRawSMCValues([]string{chargeInhibitKey})
	if err != nil {
		log.Fatalf("Error reading SMC key '%s': %v", chargeInhibitKey, err)
	}

	inhibitValue, ok := rawValues[chargeInhibitKey]
	if !ok {
		log.Fatalf("Error: Could not find SMC key '%s' on this system.", chargeInhibitKey)
	}

	// --- 3. Check the State Correctly ---
	// We use bytes.Equal to safely compare the entire slice. This is robust.
	isChargingDisabled := bytes.Equal(inhibitValue.Data, enabledBytes)

	// --- 4. Perform the Toggle Action ---
	if isChargingDisabled {
		fmt.Printf("Charging is currently DISABLED. Enabling now...\n")
		// To enable charging, we must disable the inhibit.
		err = powerkit.DisableChargeInhibit()
		if err != nil {
			log.Fatalf("Failed to enable charging: %v", err)
		}
		fmt.Println("Successfully enabled charging. The battery will now charge normally.")
	} else {
		fmt.Printf("Charging is currently ENABLED. Disabling now...\n")
		// To disable charging, we must enable the inhibit.
		err = powerkit.EnableChargeInhibit()
		if err != nil {
			log.Fatalf("Failed to disable charging: %v", err)
		}
		fmt.Println("Successfully disabled charging. The battery will not charge past its current level.")
	}
}
