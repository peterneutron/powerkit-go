package main

import (
	"fmt"
	"log"

	"github.com/peterneutron/powerkit-go/pkg/powerkit"
)

func handleWriteCommand(group string, args []string) {
	checkRoot()

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
