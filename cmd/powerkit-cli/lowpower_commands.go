package main

import (
	"fmt"
	"log"

	"github.com/peterneutron/powerkit-go/pkg/powerkit"
)

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
