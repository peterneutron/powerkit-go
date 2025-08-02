// This example demonstrates the high-level API for controlling the charge level limit.
// It reads the current status of the charge inhibit flag and then uses the 'Toggle' action to flip the state.
package main

import (
	"fmt"
	"log"

	"github.com/peterneutron/powerkit-go/pkg/powerkit"
)

func main() {
	// --- 1. Read the Current State using the Main API ---
	// We only need SMC data for this, so we'll use FetchOptions for efficiency.
	fmt.Println("Reading current charge level limit state...")
	options := powerkit.FetchOptions{
		QueryIOKit: false,
		QuerySMC:   true,
	}
	info, err := powerkit.GetSystemInfo(options)
	if err != nil {
		log.Fatalf("Error reading current state: %v", err)
	}

	// The GetSystemInfo function provides a clean, pre-calculated boolean.
	if info.SMC == nil {
		log.Fatal("Could not retrieve SMC data.")
	}
	isCurrentlyInhibited := !info.SMC.State.IsChargingEnabled

	// Report the current state to the user.
	if isCurrentlyInhibited {
		fmt.Println("Charge level limit is currently: ENABLED (Charging is inhibited)")
	} else {
		fmt.Println("Charge level limit is currently: DISABLED (Charging is allowed to 100%)")
	}

	// --- 2. Perform the Toggle Action using the New Consolidated Function ---
	fmt.Println("\nSending toggle command...")
	// We call the single, powerful function with the 'Toggle' action.
	// The library will handle the read-then-write logic internally.
	err = powerkit.SetChargingState(powerkit.ChargingActionToggle)
	if err != nil {
		log.Fatalf("Failed to toggle charging state: %v", err)
	}
	fmt.Println("Toggle command sent successfully.")

	// --- 3. (Optional but good practice) Read the State Again to Confirm ---
	fmt.Println("\nReading new state to confirm change...")
	infoAfter, err := powerkit.GetSystemInfo(options)
	if err != nil {
		log.Fatalf("Error reading state after toggle: %v", err)
	}
	if infoAfter.SMC == nil {
		log.Fatal("Could not retrieve SMC data after toggle.")
	}
	isNowInhibited := !infoAfter.SMC.State.IsChargingEnabled

	// Report the new state.
	if isNowInhibited {
		fmt.Println("Charge level limit is now: ENABLED")
	} else {
		fmt.Println("Charge level limit is now: DISABLED")
	}
}
