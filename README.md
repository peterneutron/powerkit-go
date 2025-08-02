# PowerKit-Go

[![Go Reference](https://pkg.go.dev/badge/github.com/peterneutron/powerkit-go.svg)](https://pkg.go.dev/github.com/peterneutron/powerkit-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/peterneutron/powerkit-go)](https://goreportcard.com/report/github.com/peterneutron/powerkit-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A comprehensive Go library for monitoring and controlling macOS power features. It provides detailed, source-separated information from both IOKit and the System Management Controller (SMC), and offers functions to control charging behavior and the MagSafe LED.

> ### ⚠️ Pre-Release Software Notice ⚠️
> This library is in its initial development phase. The API is not yet stable and is subject to breaking changes in future releases. Please use with caution and consider pinning to a specific version in your project.

## Features

*   **Dual Source Data:** Access both the high-level `IOKit` registry and the low-level `SMC` for a complete power profile.
*   **Hardware Control:** Enable/disable charging, connect/disconnect the AC adapter, and change the MagSafe LED color (requires root privileges).
*   **Source-Centric API:** The primary API returns a clean, structured `SystemInfo` object that strictly separates data by its source, eliminating ambiguity.
*   **Flexible Queries:** Choose to query IOKit, the SMC, or both, for maximum efficiency.
*   **Raw SMC Access:** A dedicated API for advanced users to query custom SMC keys.
*   **Command-Line Tool:** Includes `powerkit-cli` to dump hardware info, query SMC keys, and control power states from your terminal.

## Installation

```bash
go get github.com/peterneutron/powerkit-go
```

## Library Usage

### 1. Basic Usage: Getting All Information

The primary entrypoint `powerkit.GetSystemInfo()` fetches data from all available sources by default and returns a clean, structured object.

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/peterneutron/powerkit-go/pkg/powerkit"
)

func main() {
	info, err := powerkit.GetSystemInfo()
	if err != nil {
		log.Fatalf("Error getting hardware info: %v", err)
	}

	jsonData, _ := json.MarshalIndent(info, "", "  ")
	fmt.Println(string(jsonData))
}
```

**Example Output:**
```json
{
  "IOKit": {
    "State": {
      "IsCharging": false,
      "IsConnected": false,
      "FullyCharged": false
    },
    "Battery": {
      "SerialNumber": "...",
      "DeviceName": "...",
      "CycleCount": 183,
      "DesignCapacity": 8579,
      "MaxCapacity": 7647,
      "NominalCapacity": 7891,
      "CurrentCapacity": 2620,
      "TimeToEmpty": 229,
      "TimeToFull": 65535,
      "Temperature": 30.19,
      "Voltage": 11.36,
      "Amperage": -0.48,
      "CurrentCharge": 35,
      "IndividualCellVoltages": [
        3787,
        3788,
        3787
      ]
    },
    "Adapter": {
      "Description": "...",
      "MaxWatts": 0,
      "MaxVoltage": 0,
      "MaxAmperage": 0,
      "InputVoltage": 0,
      "InputAmperage": 0
    },
    "Calculations": {
      "HealthByMaxCapacity": 89,
      "HealthByNominalCapacity": 92,
      "ConditionAdjustedHealth": 95,
      "ACPower": 0,
      "BatteryPower": -5.45,
      "SystemPower": 5.45
    }
  },
  "SMC": {
    "State": {
      "IsChargingEnabled": true,
      "IsAdapterEnabled": true
    },
    "Battery": {
      "Voltage": 11.36,
      "Amperage": -0.42
    },
    "Adapter": {
      "InputVoltage": 0,
      "InputAmperage": 0
    },
    "Calculations": {
      "ACPower": 0,
      "BatteryPower": -4.77,
      "SystemPower": 4.77
    }
  }
}
```

### 2. Hardware Control (Requires Root)

> **⚠️ WARNING:** The following functions write directly to the System Management Controller (SMC). They require root privileges to run and can potentially impact your hardware. Use with caution.

#### Controlling Charging

You can enable, disable, or toggle the battery charging state.

```go
// Disable charging
err := powerkit.SetChargingState(powerkit.ChargingActionOff)
if err != nil {
    log.Fatal(err)
}

// Enable charging
err = powerkit.SetChargingState(powerkit.ChargingActionOn)
if err != nil {
    log.Fatal(err)
}
```

#### Controlling the Adapter

You can programmatically connect or disconnect the AC adapter.

```go
// Disable the adapter
err := powerkit.SetAdapterState(powerkit.AdapterActionOff)
if err != nil {
    log.Fatal(err)
}

// Enable the adapter
err = powerkit.SetAdapterState(powerkit.AdapterActionOn)
if err != nil {
    log.Fatal(err)
}
```

#### Controlling the MagSafe LED

You can change the color of the MagSafe charging LED.

```go
// Set LED to Amber
err := powerkit.SetMagsafeLEDColor(powerkit.LEDAmber)
if err != nil {
    log.Fatal(err)
}

// Set LED to Green
err = powerkit.SetMagsafeLEDColor(powerkit.LEDGreen)
if err != nil {
    log.Fatal(err)
}

// Turn LED Off
err = powerkit.SetMagsafeLEDColor(powerkit.LEDOff)
if err != nil {
    log.Fatal(err)
}
```

### 3. Advanced Usage: Querying a Single Source

For efficiency, you can provide `FetchOptions` to query only the data source you need. The resulting JSON will cleanly omit the unused source.

```go
// Query for SMC data only
options := powerkit.FetchOptions{
    QueryIOKit: false,
    QuerySMC:   true,
}
info, err := powerkit.GetSystemInfo(options)

// The output JSON will have no "IOKit" key.
```

### 4. Power User: Raw SMC Key Queries

For maximum flexibility, `GetRawSMCValues()` allows you to query any custom SMC key and receive the raw, undecoded data. You are responsible for interpreting the bytes.

```go
// Query for the number of fans (FNum) and a CPU temperature (TC0D)
keys := []string{"FNum", "TC0D"}
rawValues, err := powerkit.GetRawSMCValues(keys)
if err != nil {
    log.Fatal(err)
}

// The user is responsible for decoding the raw bytes based on the DataType
for key, val := range rawValues {
    fmt.Printf("Key: %s, Type: %s, Raw Bytes: %x\n", key, val.DataType, val.Data)
}
```

## Command-Line Tool (`powerkit-cli`)

The library includes a simple CLI tool for reading sensors and controlling power states.

### Installation
```bash
go install github.com/peterneutron/powerkit-go/cmd/powerkit-cli@latest
```

### Usage

**Read Commands:**
```bash
# Dump all curated data as JSON
powerkit-cli all

# Dump only IOKit or SMC data
powerkit-cli iokit
powerkit-cli smc

# Perform a raw query for custom SMC keys
powerkit-cli raw FNum TC0D PPBR
```
**Example Output for `raw`:**
```json
{
  "FNum": {
    "DataType": "ui8 ",
    "DataSize": 1,
    "Data": "02"
  },
  "PPBR": {
    "DataType": "flt ",
    "DataSize": 4,
    "Data": "7359d640"
  },
  "TC0D": {
    "DataType": "sp78",
    "DataSize": 2,
    "Data": "34a0"
  }
}
```

**Write Commands (Requires Root):**
```bash
# Disable the AC adapter
sudo powerkit-cli adapter off

# Enable the AC adapter
sudo powerkit-cli adapter on

# Disable battery charging
sudo powerkit-cli charging off

# Enable battery charging
sudo powerkit-cli charging on
```

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.