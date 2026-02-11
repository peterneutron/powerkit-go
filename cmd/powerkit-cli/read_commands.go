package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/peterneutron/powerkit-go/pkg/powerkit"
)

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

func handleWatchCommand() {
	fmt.Println("Watching for system events... Press Ctrl+C to exit.")

	eventChan, err := powerkit.StreamSystemEvents()
	if err != nil {
		log.Fatalf("Error starting event stream: %v", err)
	}

	for event := range eventChan {
		fmt.Print("\033[H\033[2J")
		fmt.Printf("--- Event Received at %s ---\n", time.Now().Format(time.RFC3339))
		fmt.Printf("Event Type: %s\n\n", EventTypeToString(event.Type))
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

func handleDumpCommand(source string, args []string) {
	options, err := resolveDumpOptions(source, args)
	if err != nil {
		log.Fatalf("%v", err)
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

func resolveDumpOptions(source string, args []string) (powerkit.FetchOptions, error) {
	switch source {
	case "all":
		return optionsForAll(args)
	case "smc":
		if len(args) > 0 {
			return powerkit.FetchOptions{}, fmt.Errorf("command 'smc' does not accept additional arguments")
		}
		return powerkit.FetchOptions{QuerySMC: true}, nil
	case "iokit":
		if len(args) > 0 {
			return powerkit.FetchOptions{}, fmt.Errorf("command 'iokit' does not accept additional arguments")
		}
		return powerkit.FetchOptions{QueryIOKit: true}, nil
	default:
		return powerkit.FetchOptions{}, fmt.Errorf("unknown dump source '%s'", source)
	}
}

func optionsForAll(args []string) (powerkit.FetchOptions, error) {
	options := powerkit.FetchOptions{QueryIOKit: true, QuerySMC: true}
	if len(args) == 0 {
		return options, nil
	}
	if len(args) > 1 {
		return powerkit.FetchOptions{}, fmt.Errorf("too many arguments for 'all' (expected optional 'fallback')")
	}
	switch strings.ToLower(args[0]) {
	case "fallback", "--fallback":
		options.ForceTelemetryFallback = true
		return options, nil
	default:
		return powerkit.FetchOptions{}, fmt.Errorf("unknown argument '%s' for 'all'. did you mean 'fallback'?", args[0])
	}
}

func handleRawCommand(keys []string) {
	if len(keys) == 0 {
		log.Fatalf("Error: 'raw' command requires at least one SMC key to query.\nUsage: powerkit-cli raw FNum TC0P")
	}
	rawValues, err := powerkit.GetRawSMCValues(keys)
	if err != nil {
		log.Fatalf("Error getting raw SMC values: %v", err)
	}

	formattedOutput := make(map[string]interface{})
	for key := range rawValues {
		val := rawValues[key]
		formattedOutput[key] = struct {
			DataType string `json:"DataType"`
			DataSize int    `json:"DataSize"`
			Data     string `json:"Data"`
		}{
			DataType: val.DataType,
			DataSize: val.DataSize,
			Data:     fmt.Sprintf("%x", val.Data),
		}
	}

	jsonData, err := json.MarshalIndent(formattedOutput, "", "  ")
	if err != nil {
		log.Fatalf("Error formatting data to JSON: %v", err)
	}
	fmt.Println(string(jsonData))
}
