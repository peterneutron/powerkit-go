// The powerkit-cli tool provides a simple way to dump all hardware
// and power information from the local macOS system as a JSON object.
package main

import (
	"encoding/json"
	"fmt"
	"log"

	// Import your new, clean, public package
	"github.com/peterneutron/powerkit-go/pkg/powerkit"
)

func main() {
	// 1. Call the library's main function to get the data.
	info, err := powerkit.GetBatteryInfo()
	if err != nil {
		// If there's an error (e.g., SMC not found), print it to stderr
		// and exit with a non-zero status code. log.Fatalf does this automatically.
		log.Fatalf("Error getting hardware info: %v", err)
	}

	// 2. Marshal the returned struct into a nicely formatted JSON byte slice.
	// The "" prefix means no prefix per line.
	// The "  " indent means use two spaces for indentation.
	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		// This error is rare but should be handled.
		log.Fatalf("Error formatting data to JSON: %v", err)
	}

	// 3. Print the final JSON string to standard output.
	fmt.Println(string(jsonData))
}
