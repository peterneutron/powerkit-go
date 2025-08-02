// The powerkit-cli tool provides a simple way to dump, query, and control
// macOS hardware sensors.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"

	"github.com/peterneutron/powerkit-go/pkg/powerkit"
)

func main() {
	// --- Command Dispatching ---
	// We need at least one argument for a command.
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	commandGroup := os.Args[1]

	// Route to the correct handler based on the command group.
	switch commandGroup {
	// Read commands (single word)
	case "all", "smc", "iokit":
		handleDumpCommand(commandGroup)
	case "raw":
		handleRawCommand(os.Args[2:])

	// Write commands (two words)
	case "adapter", "charging":
		// These commands require a second argument ('on' or 'off')
		if len(os.Args) < 3 {
			log.Fatalf("Error: '%s' command requires an action ('on' or 'off').\nUsage: powerkit-cli %s on", commandGroup, commandGroup)
		}
		action := os.Args[2]
		handleWriteCommand(commandGroup, action)

	// Help and default
	case "help":
		printUsage()
	default:
		fmt.Printf("Error: unknown command '%s'\n\n", commandGroup)
		printUsage()
		os.Exit(1)
	}
}

// printUsage prints the main help message for the tool.
func printUsage() {
	fmt.Println("powerkit-cli: A tool to dump, query, and control macOS hardware sensors.")
	fmt.Println("\nUsage:")
	fmt.Println("  powerkit-cli <command> [subcommand/arguments]")
	fmt.Println("\nRead Commands:")
	fmt.Println("  all          Dump curated SystemInfo from both IOKit and SMC")
	fmt.Println("  iokit        Dump curated SystemInfo from IOKit only")
	fmt.Println("  smc          Dump curated SystemInfo from SMC only")
	fmt.Println("  raw [keys...] Query for custom SMC keys (e.g., 'powerkit-cli raw FNum')")

	fmt.Println("\nWrite Commands (may require sudo):")
	fmt.Println("  adapter on   Connect the battery to the charger")
	fmt.Println("  adapter off  Disconnect the battery from the charger (using CHIE)")
	fmt.Println("  charging on  Allow battery to charge to 100% (using BCLM)")
	fmt.Println("  charging off Set max charge level to current level (using BCLM)")

	fmt.Println("\nOther Commands:")
	fmt.Println("  help         Show this help message")
}

// --- NEW: Universal Write Command Handler ---
func handleWriteCommand(group, action string) {
	checkRoot() // All write commands require the root check.

	var err error
	var successMsg string

	if group == "adapter" {
		if action == "on" {
			fmt.Println("Attempting to enable the charger...")
			err = powerkit.SetAdapterState(powerkit.AdapterActionOn)
			successMsg = "Successfully enabled the charger."
		} else if action == "off" {
			fmt.Println("Attempting to disable the charger...")
			err = powerkit.SetAdapterState(powerkit.AdapterActionOff)
			successMsg = "Successfully disabled the charger."
		} else {
			log.Fatalf("Error: invalid action '%s' for 'adapter' command. Use 'on' or 'off'.", action)
		}
	} else if group == "charging" {
		if action == "on" {
			fmt.Println("Attempting to enable charging")
			err = powerkit.SetChargingState(powerkit.ChargingActionOn)
			successMsg = "Successfully enabled charging."
		} else if action == "off" {
			fmt.Println("Attempting to disable charging")
			err = powerkit.SetChargingState(powerkit.ChargingActionOff)
			successMsg = "Successfully disabled charging."
		} else {
			log.Fatalf("Error: invalid action '%s' for 'charging' command. Use 'on' or 'off'.", action)
		}
	}

	if err != nil {
		log.Fatalf("Command failed: %v", err)
	}
	fmt.Println(successMsg)
}

// checkRoot checks for root privileges.
func checkRoot() {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("Fatal: Could not determine current user: %v", err)
	}
	if currentUser.Uid != "0" {
		log.Fatalf("Error: This command requires root privileges to write to the SMC.\nPlease run with 'sudo'.")
	}
}

// Read command handlers
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
