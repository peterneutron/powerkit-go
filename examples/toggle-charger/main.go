// This example demonstrates the high-level API for reading and writing charger state.
// It reads the current charger status and then uses the 'Toggle' action
// to flip the state.
package main

import (
	"fmt"
	"log"

	"github.com/peterneutron/powerkit-go/pkg/powerkit"
)

func main() {
	// --- 1. Read the Current State using the Main API ---
	// We only need SMC data, so we'll use FetchOptions for efficiency.
	fmt.Println("Reading current charger state...")
	options := powerkit.FetchOptions{
		QueryIOKit: false,
		QuerySMC:   true,
	}
	info, err := powerkit.GetSystemInfo(options)
	if err != nil {
		log.Fatalf("Error reading current state: %v", err)
	}

	// The GetSystemInfo function provides a clean, pre-calculated boolean.
	// No more raw byte checking is needed here!
	if info.SMC == nil {
		log.Fatal("Could not retrieve SMC data.")
	}
	isCurrentlyDisabled := !info.SMC.State.IsAdapterEnabled

	// Report the current state to the user.
	if isCurrentlyDisabled {
		fmt.Println("Charger is currently: DISABLED")
	} else {
		fmt.Println("Charger is currently: ENABLED")
	}

	// --- 2. Perform the Toggle Action using the New Consolidated Function ---
	fmt.Println("\nSending toggle command...")
	// We call the single, powerful function with the 'Toggle' action.
	// The library itself will handle the read-then-write logic internally.
	err = powerkit.SetAdapterState(powerkit.AdapterActionToggle)
	if err != nil {
		log.Fatalf("Failed to toggle charger state: %v", err)
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
	isNowDisabled := !infoAfter.SMC.State.IsAdapterEnabled

	// Report the new state.
	if isNowDisabled {
		fmt.Println("Charger is now: DISABLED")
	} else {
		fmt.Println("Charger is now: ENABLED")
	}
}
