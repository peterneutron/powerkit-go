package main

import (
	"fmt"
	"log"
	"os"

	"github.com/peterneutron/powerkit-go/pkg/powerkit"
)

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
