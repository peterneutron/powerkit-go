package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/peterneutron/powerkit-go/pkg/powerkit"
)

func main() {
	// 1. Define the options to query only the SMC.
	// We explicitly disable the IOKit query.
	smcOnlyOptions := powerkit.FetchOptions{
		QueryIOKit: true,  // Set to false to skip IOKit
		QuerySMC:   false, // Set to true to fetch SMC
	}

	// 2. Call the library's main function, passing in our custom options.
	info, err := powerkit.GetBatteryInfo(smcOnlyOptions)
	if err != nil {
		// If the SMC read fails, this will now correctly be a fatal error.
		log.Fatalf("Error getting SMC info: %v", err)
	}

	// 3. Marshal and print the result.
	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		log.Fatalf("Error formatting data to JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}
