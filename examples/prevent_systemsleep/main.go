// This program demonstrates how to use the powerassert package to prevent
// the display from sleeping for a short period.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/peterneutron/powerkit-go/internal/powerd" // Note: direct internal import for example purposes
)

func main() {
	fmt.Println("This program will prevent the display from sleeping for 30 seconds.")
	fmt.Println("You can verify this by checking 'pmset -g assertions' in another terminal.")
	fmt.Println("Press Ctrl+C to exit early.")

	// A channel to listen for OS signals (like Ctrl+C) for graceful shutdown.
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	// --- Core Logic ---

	// 1. Create the power assertion to prevent the display from sleeping.
	// We give it a clear reason that will appear in system utilities.
	reason := "PowerKit-GO: prevent_systemsleep is running"
	fmt.Printf("\n[1/3] Creating power assertion: '%s'\n", reason)
	assertionID, err := powerd.PreventSleep(powerd.PreventDisplaySleep, reason)
	if err != nil {
		log.Fatalf("Failed to create power assertion: %v", err)
	}

	// Print the assertion ID we received from the OS!
	fmt.Printf("      -> Successfully created assertion with ID: %d\n", assertionID)

	// 2. Defer the cleanup. This is crucial!
	// This function will run when main() exits for any reason (completion or signal).
	// We use AllowAllSleep() as a robust way to clean up everything.
	defer func() {
		fmt.Println("\n[3/3] Releasing power assertion.")
		powerd.AllowAllSleep()
		fmt.Println("System can now sleep normally.")
	}()

	fmt.Println("[2/3] Assertion active. Waiting for 30 seconds...")

	// 3. Wait for either 30 seconds to pass or for a shutdown signal.
	select {
	case <-time.After(30 * time.Second):
		fmt.Println("\n30-second timer finished.")
	case <-shutdownChan:
		fmt.Println("\nShutdown signal received, exiting early.")
	}
}
