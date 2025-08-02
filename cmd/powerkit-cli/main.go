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

	case "magsafe":
		if len(os.Args) < 3 {
			log.Fatalf("Error: 'magsafe' command requires a subcommand ('get-color' or 'set-color').")
		}
		subcommand := os.Args[2]
		args := os.Args[3:]
		handleMagsafeCommand(subcommand, args)

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

	fmt.Println("\nControl Commands:")
	fmt.Println("  adapter <on|off>                Enable or disable the adapter connection (requires sudo)")
	fmt.Println("  charging <on|off>               Enable or disable battery charging (requires sudo)")
	fmt.Println("  magsafe get-color               Get the current Magsafe LED color")
	fmt.Println("  magsafe set-color <color>       Set the Magsafe LED color (off, amber, green) (requires sudo)")

	fmt.Println("\nOther Commands:")
	fmt.Println("  help         Show this help message")
}

// --- Universal Write Command Handler ---
func handleWriteCommand(group, action string) {
	checkRoot() // All write commands require the root check.

	var err error
	var successMsg string

	switch group {
	case "adapter":
		switch action {
		case "on":
			fmt.Println("Attempting to enable the charger...")
			err = powerkit.SetAdapterState(powerkit.AdapterActionOn)
			successMsg = "Successfully enabled the charger."
		case "off":
			fmt.Println("Attempting to disable the charger...")
			err = powerkit.SetAdapterState(powerkit.AdapterActionOff)
			successMsg = "Successfully disabled the charger."
		default:
			log.Fatalf("Error: invalid action '%s' for 'adapter' command. Use 'on' or 'off'.", action)
		}
	case "charging":
		switch action {
		case "on":
			fmt.Println("Attempting to enable charging")
			err = powerkit.SetChargingState(powerkit.ChargingActionOn)
			successMsg = "Successfully enabled charging."
		case "off":
			fmt.Println("Attempting to disable charging")
			err = powerkit.SetChargingState(powerkit.ChargingActionOff)
			successMsg = "Successfully disabled charging."
		default:
			log.Fatalf("Error: invalid action '%s' for 'charging' command. Use 'on' or 'off'.", action)
		}
	}

	if err != nil {
		log.Fatalf("Command failed: %v", err)
	}
	fmt.Println(successMsg)
}

// --- Magsafe Command Handler ---

// MagsafeColorToString converts the enum to a human-readable string for printing.
func MagsafeColorToString(c powerkit.MagsafeColor) string {
	switch c {
	case powerkit.LEDOff:
		return "Off"
	case powerkit.LEDAmber:
		return "Amber"
	case powerkit.LEDGreen:
		return "Green"
	default:
		return "Unknown"
	}
}

func handleMagsafeCommand(subcommand string, args []string) {
	switch subcommand {
	case "get-color":
		color, err := powerkit.GetMagsafeLEDColor()
		if err != nil {
			log.Fatalf("Error getting Magsafe LED color: %v", err)
		}
		fmt.Printf("Current Magsafe LED color: %s\n", MagsafeColorToString(color))

	case "set-color":
		checkRoot()
		if len(args) < 1 {
			log.Fatalf("Error: 'set-color' requires a color argument (off, amber, green).")
		}
		colorStr := args[0]
		var color powerkit.MagsafeColor
		var validColor = true

		switch colorStr {
		case "off":
			color = powerkit.LEDOff
		case "amber":
			color = powerkit.LEDAmber
		case "green":
			color = powerkit.LEDGreen
		default:
			validColor = false
		}

		if !validColor {
			log.Fatalf("Error: invalid color '%s'. Use 'off', 'amber', or 'green'.", colorStr)
		}

		fmt.Printf("Attempting to set Magsafe LED to %s...\n", colorStr)
		err := powerkit.SetMagsafeLEDColor(color)
		if err != nil {
			log.Fatalf("Command failed: %v", err)
		}
		fmt.Println("Successfully set Magsafe LED color.")

	default:
		fmt.Printf("Error: unknown subcommand '%s' for 'magsafe'.\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}
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
