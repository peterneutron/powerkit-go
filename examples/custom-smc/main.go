// This program demonstrates how to query for custom SMC keys.
package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"

	"github.com/peterneutron/powerkit-go/pkg/powerkit"
)

func main() {
	// Let's query for the raw power keys we decided to ignore in the main API.
	keys := []string{
		"PPBR", // The raw Battery Power sensor (float)
		"PSTR", // The raw System Total Power sensor (float)
	}

	rawValues, err := powerkit.GetRawSMCValues(keys)
	if err != nil {
		log.Fatalf("Failed to get raw SMC values: %v", err)
	}

	fmt.Println("Querying for raw, direct power sensor readings from the SMC:")

	// The user is responsible for decoding the raw bytes.
	for key := range rawValues {
		val := rawValues[key] // Get the value once inside the loop.

		fmt.Printf("\n--- Key: %s ---\n", key)
		fmt.Printf("  - Type:      %s\n", val.DataType)
		fmt.Printf("  - Size:      %d bytes\n", val.DataSize)
		fmt.Printf("  - Raw Bytes: %x\n", val.Data) // Print bytes as hex for debugging

		// --- User-Side Decoding Logic ---
		// This is the logic a user would write to interpret the raw data.
		var decodedValue interface{}

		if val.DataType == "flt " && val.DataSize == 4 {
			// This is how you decode a little-endian 32-bit float in Go.
			bits := binary.LittleEndian.Uint32(val.Data)
			floatVal := math.Float32frombits(bits)
			decodedValue = floatVal
		} else {
			// If we get a type we don't know how to decode, we just say so.
			decodedValue = "Unhandled data type"
		}

		fmt.Printf("  - Decoded Value: %v\n", decodedValue)
	}

	fmt.Println("\nThis demonstrates how a user can access the raw sensor data directly,")
	fmt.Println("bypassing the library's internal calculations.")
}
