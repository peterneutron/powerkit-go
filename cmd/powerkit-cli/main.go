// The powerkit-cli tool provides a simple way to dump, query, and control
// macOS hardware sensors.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/peterneutron/powerkit-go/pkg/powerkit"
)

const (
	actionOn  = "on"
	actionOff = "off"
	// assertion
	assertionCmd         = "assertion"
	assertionCreate      = "create"
	assertionRelease     = "release"
	assertionStatus      = "status"
	assertionTypeSystem  = "system"
	assertionTypeDisplay = "display"
	colorOff             = "off"
	colorAmber           = "amber"
	colorGreen           = "green"
	colorSystem          = "system"
	colorErrOnce         = "error-once"
	colorErrPermSlow     = "error-perm-slow"
	colorErrPermFast     = "error-perm-fast"
	colorErrPermOff      = "error-perm-off"
	cmdGetColor          = "get-color"
	cmdSetColor          = "set-color"
	// low power mode
	cmdLowPower = "lowpower"
	lpmGet      = "get"
	lpmSet      = "set"
	lpmToggle   = "toggle"
)

// EventTypeToString provides a human-readable name for an event type.
func EventTypeToString(t powerkit.EventType) string {
	switch t {
	case powerkit.EventTypeBatteryUpdate:
		return "Battery Update"
	case powerkit.EventTypeSystemWillSleep:
		return "System Will Sleep"
	case powerkit.EventTypeSystemDidWake:
		return "System Did Wake"
	default:
		return "Unknown Event"
	}
}

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
	case "watch": // New command
		handleWatchCommand()

		// Write commands (two words)
	case "adapter", "charging":
		// Delegate arg validation to handler
		handleWriteCommand(commandGroup, os.Args[2:])

	case "magsafe":
		// Delegate arg validation to handler
		subcommand := os.Args[2]
		args := os.Args[3:]
		handleMagsafeCommand(subcommand, args)

	case assertionCmd:
		// Delegate arg validation to handler
		handleAssertionCommand(os.Args[2:])

	case cmdLowPower:
		handleLowPowerCommand(os.Args[2:])

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
	fmt.Println("  watch        Stream real-time power events as they happen")
	fmt.Println("  magsafe get-color               Get the current Magsafe LED state")
	fmt.Println("  lowpower get                    Get macOS Low Power Mode state")
	fmt.Println("  lowpower set <on|off>           Set Low Power Mode (requires sudo)")
	fmt.Println("  lowpower toggle                 Toggle Low Power Mode (requires sudo)")
	fmt.Println("  assertion status <system|display>   Show whether the assertion is active and its ID")

	fmt.Println("\nControl Commands:")
	fmt.Println("  adapter <on|off>                Enable or disable the adapter connection (requires sudo)")
	fmt.Println("  charging <on|off>               Enable or disable battery charging (requires sudo)")
	fmt.Println("  magsafe set-color <state>       Set the Magsafe LED state (system, off, amber, green, error-once, error-perm-slow, error-perm-fast, error-perm-off) (requires sudo)")
	fmt.Println("  assertion create <system|display> [reason...]   Create a sleep assertion with optional reason")
	fmt.Println("  assertion release <system|display>              Release a sleep assertion of the given type")

	fmt.Println("\nOther Commands:")
	fmt.Println("  help         Show this help message")
}

// --- NEW: Watch Command Handler ---
func handleWatchCommand() {
	fmt.Println("Watching for system events... Press Ctrl+C to exit.")

	// Subscribe to the new, unified event stream.
	eventChan, err := powerkit.StreamSystemEvents()
	if err != nil {
		log.Fatalf("Error starting event stream: %v", err)
	}

	for event := range eventChan {
		// Clear screen for a cleaner live view.
		fmt.Print("\033[H\033[2J")
		fmt.Printf("--- Event Received at %s ---\n", time.Now().Format(time.RFC3339))
		fmt.Printf("Event Type: %s\n\n", EventTypeToString(event.Type))

		// The Info payload is only present for battery updates.
		if event.Info != nil {
			jsonData, err := json.MarshalIndent(event.Info, "", "  ")
			if err != nil {
				log.Printf("Error formatting data to JSON: %v", err)
				continue
			}
			fmt.Println(string(jsonData))
		} else {
			fmt.Println("This event type does not have an info payload.")
		}
	}
}

// --- Universal Write Command Handler ---
func handleWriteCommand(group string, args []string) {
	checkRoot() // All write commands require the root check.

	var err error
	var successMsg string
	if len(args) < 1 {
		log.Fatalf("Error: '%s' command requires an action ('on' or 'off').\nUsage: powerkit-cli %s on", group, group)
	}
	action := args[0]

	switch group {
	case "adapter":
		switch action {
		case actionOn:
			fmt.Println("Attempting to enable the charger...")
			err = powerkit.SetAdapterState(powerkit.AdapterActionOn)
			successMsg = "Successfully enabled the charger."
		case actionOff:
			fmt.Println("Attempting to disable the charger...")
			err = powerkit.SetAdapterState(powerkit.AdapterActionOff)
			successMsg = "Successfully disabled the charger."
		default:
			log.Fatalf("Error: invalid action '%s' for 'adapter' command. Use 'on' or 'off'.", action)
		}
	case "charging":
		switch action {
		case actionOn:
			fmt.Println("Attempting to enable charging")
			err = powerkit.SetChargingState(powerkit.ChargingActionOn)
			successMsg = "Successfully enabled charging."
		case actionOff:
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

// MagsafeStateToString converts the enum to a human-readable string for printing.
func MagsafeStateToString(s powerkit.MagsafeLEDState) string {
	switch s {
	case powerkit.LEDSystem:
		return "System"
	case powerkit.LEDOff:
		return "Off"
	case powerkit.LEDAmber:
		return "Amber"
	case powerkit.LEDGreen:
		return "Green"
	case powerkit.LEDErrorOnce:
		return "Error (Once)"
	case powerkit.LEDErrorPermSlow:
		return "Error (Perm Slow)"
	case powerkit.LEDErrorPermFast:
		return "Error (Perm Fast)"
	case powerkit.LEDErrorPermOff:
		return "Error (Perm Off)"
	default:
		return fmt.Sprintf("Unknown (0x%02x)", byte(s))
	}
}

// --- Magsafe helpers to keep cyclomatic complexity low ---
func parseMagsafeStateArg(val string) (powerkit.MagsafeLEDState, bool) {
	switch val {
	case colorSystem:
		return powerkit.LEDSystem, true
	case colorOff:
		return powerkit.LEDOff, true
	case colorAmber:
		return powerkit.LEDAmber, true
	case colorGreen:
		return powerkit.LEDGreen, true
	case colorErrOnce:
		return powerkit.LEDErrorOnce, true
	case colorErrPermSlow:
		return powerkit.LEDErrorPermSlow, true
	case colorErrPermFast:
		return powerkit.LEDErrorPermFast, true
	case colorErrPermOff:
		return powerkit.LEDErrorPermOff, true
	default:
		return 0, false
	}
}

func doMagsafeGet() {
	state, available, err := powerkit.GetMagsafeLEDState()
	if err != nil {
		log.Fatalf("Error getting Magsafe LED state: %v", err)
	}
	if !available {
		fmt.Println("MagSafe LED: Not available")
		return
	}
	fmt.Printf("Current Magsafe LED state: %s\n", MagsafeStateToString(state))
}

func doMagsafeSet(args []string) {
	checkRoot()
	if len(args) < 1 {
		log.Fatalf("Error: 'set-color' requires a state argument (system, off, amber, green, error-once, error-perm-slow, error-perm-fast, error-perm-off).")
	}
	state, ok := parseMagsafeStateArg(args[0])
	if !ok {
		log.Fatalf("Error: invalid state '%s'. Use one of: system, off, amber, green, error-once, error-perm-slow, error-perm-fast, error-perm-off.", args[0])
	}
	fmt.Printf("Attempting to set Magsafe LED to %s...\n", MagsafeStateToString(state))
	if err := powerkit.SetMagsafeLEDState(state); err != nil {
		log.Fatalf("Command failed: %v", err)
	}
	fmt.Printf("Successfully set Magsafe LED state to %s.\n", MagsafeStateToString(state))
}

func handleMagsafeCommand(subcommand string, args []string) {
	switch subcommand {
	case cmdGetColor:
		doMagsafeGet()

	case cmdSetColor:
		doMagsafeSet(args)

	default:
		fmt.Printf("Error: unknown subcommand '%s' for 'magsafe'.\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

// --- Low Power Mode Command Handler ---
func handleLowPowerCommand(args []string) {
	if len(args) < 1 {
		log.Fatalf("Error: 'lowpower' requires a subcommand ('get', 'set', 'toggle').")
	}
	switch args[0] {
	case lpmGet:
		doLowPowerGet()
	case lpmSet:
		if len(args) < 2 {
			log.Fatalf("Error: 'lowpower set' requires 'on' or 'off'.")
		}
		doLowPowerSet(args[1])
	case lpmToggle:
		doLowPowerToggle()
	default:
		log.Fatalf("Error: unknown subcommand '%s' for 'lowpower'.", args[0])
	}
}

func doLowPowerGet() {
	enabled, available, err := powerkit.GetLowPowerModeEnabled()
	if err != nil {
		log.Fatalf("Error reading Low Power Mode: %v", err)
	}
	if !available {
		fmt.Println("Low Power Mode: Not available")
		return
	}
	if enabled {
		fmt.Println("Low Power Mode: Enabled")
	} else {
		fmt.Println("Low Power Mode: Disabled")
	}
}

func doLowPowerSet(val string) {
	checkRoot()
	switch val {
	case actionOn:
		if err := powerkit.SetLowPowerMode(true); err != nil {
			log.Fatalf("Error enabling Low Power Mode: %v", err)
		}
		fmt.Println("Low Power Mode enabled.")
	case actionOff:
		if err := powerkit.SetLowPowerMode(false); err != nil {
			log.Fatalf("Error disabling Low Power Mode: %v", err)
		}
		fmt.Println("Low Power Mode disabled.")
	default:
		log.Fatalf("Error: invalid argument '%s'. Use 'on' or 'off'.", val)
	}
}

func doLowPowerToggle() {
	checkRoot()
	if err := powerkit.ToggleLowPowerMode(); err != nil {
		log.Fatalf("Error toggling Low Power Mode: %v", err)
	}
	fmt.Println("Low Power Mode toggled.")
}

// --- Assertion Command Handler ---
func parseAssertionType(val string) (powerkit.AssertionType, bool) {
	switch val {
	case assertionTypeSystem:
		return powerkit.AssertionTypePreventSystemSleep, true
	case assertionTypeDisplay:
		return powerkit.AssertionTypePreventDisplaySleep, true
	default:
		return 0, false
	}
}

func doAssertionCreate(args []string) {
	if len(args) < 1 {
		log.Fatalf("Error: 'assertion create' requires a type ('system' or 'display').")
	}
	at, ok := parseAssertionType(args[0])
	if !ok {
		log.Fatalf("Error: invalid assertion type '%s'. Use 'system' or 'display'.", args[0])
	}
	reason := "powerkit-cli"
	if len(args) > 1 {
		reason = strings.Join(args[1:], " ")
	}
	id, err := powerkit.CreateAssertion(at, reason)
	if err != nil {
		log.Fatalf("Error creating assertion: %v", err)
	}
	fmt.Printf("Created %s sleep assertion with ID %d.\n", args[0], id)
}

func doAssertionRelease(args []string) {
	if len(args) < 1 {
		log.Fatalf("Error: 'assertion release' requires a type ('system' or 'display').")
	}
	at, ok := parseAssertionType(args[0])
	if !ok {
		log.Fatalf("Error: invalid assertion type '%s'. Use 'system' or 'display'.", args[0])
	}
	powerkit.ReleaseAssertion(at)
	fmt.Printf("Released %s sleep assertion.\n", args[0])
}

func doAssertionStatus(args []string) {
	if len(args) < 1 {
		log.Fatalf("Error: 'assertion status' requires a type ('system' or 'display').")
	}
	at, ok := parseAssertionType(args[0])
	if !ok {
		log.Fatalf("Error: invalid assertion type '%s'. Use 'system' or 'display'.", args[0])
	}
	active := powerkit.IsAssertionActive(at)
	if active {
		id, _ := powerkit.GetAssertionID(at)
		fmt.Printf("%s assertion is ACTIVE (ID %d).\n", args[0], id)
	} else {
		fmt.Printf("%s assertion is not active.\n", args[0])
	}
}

func handleAssertionCommand(args []string) {
	sub := args[0]
	rest := []string{}
	if len(args) > 1 {
		rest = args[1:]
	}
	switch sub {
	case assertionCreate:
		doAssertionCreate(rest)
	case assertionRelease:
		doAssertionRelease(rest)
	case assertionStatus:
		doAssertionStatus(rest)
	default:
		fmt.Printf("Error: unknown subcommand '%s' for 'assertion'.\n\n", sub)
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
	for key := range rawValues { // Iterate over keys only
		val := rawValues[key] // Get the value once inside the loop
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
