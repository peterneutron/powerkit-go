// The powerkit-cli tool provides a simple way to dump, query, and control
// macOS hardware sensors.
package main

import (
	"fmt"
	"os"
)

func main() {
	run(os.Args)
}

func run(argv []string) {
	if len(argv) < 2 {
		printUsage()
		os.Exit(0)
	}

	commandGroup := argv[1]
	args := argv[2:]
	if handleReadCommands(commandGroup, args) {
		return
	}
	if handleWriteCommands(commandGroup, args) {
		return
	}

	handleControlCommands(commandGroup, args)
}

func handleReadCommands(commandGroup string, args []string) bool {
	switch commandGroup {
	case "all", "smc", "iokit":
		handleDumpCommand(commandGroup, args)
		return true
	case "raw":
		handleRawCommand(args)
		return true
	case "watch":
		handleWatchCommand()
		return true
	}
	return false
}

func handleWriteCommands(commandGroup string, args []string) bool {
	switch commandGroup {
	case "adapter", "charging":
		handleWriteCommand(commandGroup, args)
		return true
	}
	return false
}

func handleControlCommands(commandGroup string, args []string) {
	switch commandGroup {
	case "magsafe":
		if len(args) < 1 {
			fmt.Println("Error: 'magsafe' requires a subcommand ('get-color' or 'set-color').")
			printUsage()
			os.Exit(1)
		}
		handleMagsafeCommand(args[0], args[1:])
	case assertionCmd:
		handleAssertionCommand(args)
	case cmdLowPower:
		handleLowPowerCommand(args)
	case "help":
		printUsage()
	default:
		fmt.Printf("Error: unknown command '%s'\n\n", commandGroup)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("powerkit-cli: A tool to dump, query, and control macOS hardware sensors.")
	fmt.Println("\nUsage:")
	fmt.Println("  powerkit-cli <command> [subcommand/arguments]")
	fmt.Println("\nRead Commands:")
	fmt.Println("  all [fallback] Dump curated SystemInfo; append 'fallback' to force SMC adapter telemetry")
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
