// This program demonstrates how to use the powerkit-go library's real-time
// streaming feature to monitor for changes in the battery's state.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/peterneutron/powerkit-go/pkg/powerkit"
)

func main() {
	fmt.Println("Starting IOKit power event monitor...")
	fmt.Println("This program will report changes to charging status and battery percentage.")
	fmt.Println("Note: Updates are event-driven and may not be instantaneous.")
	fmt.Println("Press Ctrl+C to exit gracefully.")

	// 1. Initiate the streaming process.
	// StreamSystemInfo returns a read-only channel that will receive
	// a *powerkit.SystemInfo object whenever IOKit posts a notification.
	infoChan, err := powerkit.StreamSystemInfo()
	if err != nil {
		log.Fatalf("Fatal: Could not start powerkit stream: %v", err)
	}

	// 2. Keep track of the previous state to detect actual changes.
	// IOKit can send multiple notifications for a single logical event.
	// We initialize these to values that will ensure the first event is always reported.
	var (
		lastKnownCharge   = -1
		lastKnownCharging = false
		firstEvent        = true
	)

	// 3. Set up a channel to listen for OS signals for graceful shutdown.
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	// 4. Run the main event processing loop in a separate goroutine.
	// This prevents the main thread from blocking.
	go func() {
		for info := range infoChan {
			// The stream provides IOKit data. Always check for nil to be safe.
			if info == nil || info.IOKit == nil {
				continue
			}

			// Extract the values we're interested in.
			currentCharge := info.IOKit.Battery.CurrentCharge
			isCharging := info.IOKit.State.IsCharging

			// On the very first event, print the initial state.
			if firstEvent {
				fmt.Printf("\n--- Initial State ---\n")
				fmt.Printf("  Charge: %d%%\n", currentCharge)
				fmt.Printf("  Charging: %v\n", isCharging)
				fmt.Printf("---------------------\n\n")
				firstEvent = false
			}

			// Now, check if our specific values of interest have changed.
			chargeChanged := currentCharge != lastKnownCharge
			chargingStatusChanged := isCharging != lastKnownCharging

			if chargeChanged {
				fmt.Printf("[%s] Battery percentage changed to: %d%%\n", time.Now().Format(time.RFC3339), currentCharge)
			}
			if chargingStatusChanged {
				status := "STOPPED charging"
				if isCharging {
					status = "STARTED charging"
				}
				fmt.Printf("[%s] Power state changed: %s\n", time.Now().Format(time.RFC3339), status)
			}

			// Update the state for the next event.
			if chargeChanged {
				lastKnownCharge = currentCharge
			}
			if chargingStatusChanged {
				lastKnownCharging = isCharging
			}
		}
	}()

	// 5. Block the main function until a shutdown signal is received.
	<-shutdownChan
	fmt.Println("\nShutdown signal received. Exiting.")
}
