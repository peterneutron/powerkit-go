package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/peterneutron/powerkit-go/pkg/power"
)

func main() {
	info, err := power.GetBatteryInfo()
	if err != nil {
		log.Fatalf("Error getting battery info: %v", err)
	}

	// Print the data as a nicely formatted JSON object.
	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling to JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}
