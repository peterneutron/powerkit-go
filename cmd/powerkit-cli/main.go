// The powerkit-cli tool provides a simple way to dump all hardware
// and power information from the local macOS system as a JSON object.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/peterneutron/powerkit-go/pkg/powerkit"
)

func main() {
	// --- 1. Define SIMPLE Command-Line Flags ---
	// We define one canonical name for each flag. The Go toolchain lets the
	// user call them with either a single dash (-) or a double dash (--).

	querySource := flag.String("q", "all", "Data source to query: all, smc, or iokit")
	help := flag.Bool("h", false, "Show this help message")

	// After defining all flags, parse them from the command-line arguments.
	flag.Parse()

	// If the user requested help, print the usage message and exit.
	if *help {
		fmt.Println("powerkit-cli: A tool to dump macOS hardware sensor data.")
		flag.PrintDefaults()
		os.Exit(0)
	}

	// --- 2. Build the FetchOptions from the Parsed Flag ---
	var options powerkit.FetchOptions

	// This logic is now much simpler because we only have one variable to check.
	switch *querySource {
	case "all":
		options.QueryIOKit = true
		options.QuerySMC = true
	case "smc":
		options.QueryIOKit = false
		options.QuerySMC = true
	case "iokit":
		options.QueryIOKit = true
		options.QuerySMC = false
	default:
		// If the user provides an invalid value, log an error and show help.
		log.Printf("Error: invalid value for -q: '%s'. Must be 'all', 'smc', or 'iokit'.\n", *querySource)
		flag.PrintDefaults()
		os.Exit(1) // Exit with a non-zero code to indicate an error.
	}

	// --- 3. Call the Library and Print the Result ---
	info, err := powerkit.GetSystemInfo(options)
	if err != nil {
		log.Fatalf("Error getting hardware info: %v", err)
	}

	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		log.Fatalf("Error formatting data to JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}
