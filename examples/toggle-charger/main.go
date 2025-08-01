// This example demonstrates how to toggle the charger's connection to the battery.
// It reads the current charger state from the SMC and then performs the opposite action.
package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/peterneutron/powerkit-go/pkg/powerkit"
	// We import the internal smc package to get access to the key constants
	"github.com/peterneutron/powerkit-go/internal/smc"
)

func main() {
	// --- 1. Define the Key and Expected "Disabled" State ---
	chargerControlKey := smc.KeyChargerControl
	disabledBytes := []byte{0x08} // This is the byte sequence for "charger disabled"

	// --- 2. Read the Current State from the SMC ---
	fmt.Printf("Reading current state of SMC key '%s'...\n", chargerControlKey)
	rawValues, err := powerkit.GetRawSMCValues([]string{chargerControlKey})
	if err != nil {
		log.Fatalf("Error reading SMC key '%s': %v", chargerControlKey, err)
	}

	chargerValue, ok := rawValues[chargerControlKey]
	if !ok {
		log.Fatalf("Error: Could not find SMC key '%s' on this system.", chargerControlKey)
	}

	// --- 3. Check if the Charger is Currently Disabled ---
	isChargerDisabled := bytes.Equal(chargerValue.Data, disabledBytes)

	// --- 4. Perform the Toggle Action ---
	if isChargerDisabled {
		fmt.Printf("Charger is currently DISABLED. Enabling now...\n")
		// Call the library function to enable charging.
		err = powerkit.EnableCharger()
		if err != nil {
			log.Fatalf("Failed to enable charger: %v", err)
		}
		fmt.Println("Successfully enabled charger. The battery will now charge normally.")
	} else {
		fmt.Printf("Charger is currently ENABLED. Disabling now...\n")
		// Call the library function to disable charging.
		err = powerkit.DisableCharger()
		if err != nil {
			log.Fatalf("Failed to disable charger: %v", err)
		}
		fmt.Println("Successfully disabled charger. The battery is now disconnected from the charger.")
	}
}
