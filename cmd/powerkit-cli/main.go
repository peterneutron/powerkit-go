// The powerkit-cli tool provides a simple way to dump and query macOS hardware
// sensor data as a JSON object.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/peterneutron/powerkit-go/pkg/powerkit"
)

func main() {
	// --- Command Dispatching ---
	// os.Args[0] is the command name itself ("powerkit-cli")
	// The subcommands start at os.Args[1]

	var command string
	if len(os.Args) < 2 {
		// If no arguments are provided, default to the "all" command.
		command = "help"
	} else {
		command = os.Args[1]
	}

	// Route to the correct handler based on the command.
	switch command {
	case "all", "smc", "iokit":
		handleDumpCommand(command)
	case "raw":
		// For the "raw" command, the rest of the arguments are the keys.
		handleRawCommand(os.Args[2:])
	case "help":
		printUsage()
	default:
		fmt.Printf("Error: unknown command '%s'\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

// printUsage prints the main help message for the tool.
func printUsage() {
	fmt.Println("powerkit-cli: A tool to dump and query macOS hardware sensor data.")
	fmt.Println("\nUsage:")
	fmt.Println("  powerkit-cli [command]")
	fmt.Println("\nCommands:")
	fmt.Println("  all      Dump curated SystemInfo from both IOKit and SMC")
	fmt.Println("  iokit    Dump curated SystemInfo from IOKit only")
	fmt.Println("  smc      Dump curated SystemInfo from SMC only")
	fmt.Println("  raw      Query for custom SMC keys and get raw, undecoded values")
	fmt.Println("           (e.g., 'powerkit-cli raw FNum TC0P')")
	fmt.Println("  help     (default) Show this help message")
}

// handleDumpCommand runs the logic for the "dump" commands.
func handleDumpCommand(source string) {
	var options powerkit.FetchOptions

	switch source {
	case "all":
		options.QueryIOKit, options.QuerySMC = true, true
	case "smc":
		options.QueryIOKit, options.QuerySMC = false, true
	case "iokit":
		options.QueryIOKit, options.QuerySMC = true, false
	}

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

func handleRawCommand(keys []string) {
	if len(keys) == 0 {
		log.Fatalf("Error: 'raw' command requires at least one SMC key to query.\nUsage: powerkit-cli raw FNum TC0P")
	}

	rawValues, err := powerkit.GetRawSMCValues(keys)
	if err != nil {
		log.Fatalf("Error getting raw SMC values: %v", err)
	}

	// We do this to convert the []byte slice into a hex string before marshaling.
	formattedOutput := make(map[string]interface{})
	for key, val := range rawValues {
		formattedOutput[key] = struct {
			DataType string `json:"DataType"`
			DataSize int    `json:"DataSize"`
			Data     string `json:"Data"` // The key change: Data is now a string
		}{
			DataType: val.DataType,
			DataSize: val.DataSize,
			// Use fmt.Sprintf with the "%x" verb to format the byte slice as a hex string.
			Data: fmt.Sprintf("%x", val.Data),
		}
	}

	// Marshal the new formatted map, not the original rawValues.
	jsonData, err := json.MarshalIndent(formattedOutput, "", "  ")
	if err != nil {
		log.Fatalf("Error formatting data to JSON: %v", err)
	}
	fmt.Println(string(jsonData))
}
