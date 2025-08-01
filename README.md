# PowerKit-Go

[![Go Reference](https://pkg.go.dev/badge/github.com/peterneutron/powerkit-go.svg)](https://pkg.go.dev/github.com/peterneutron/powerkit-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/peterneutron/powerkit-go)](https://goreportcard.com/report/github.com/peterneutron/powerkit-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A comprehensive Go library for monitoring macOS hardware, with a current focus on power and battery sensors. It provides detailed, source-separated information from both IOKit and the System Management Controller (SMC).

> ### ⚠️ Pre-Release Software Notice ⚠️
> This library is in its initial development phase (v0.x.x). The API is not yet stable and is subject to breaking changes in future releases. Please use with caution and consider pinning to a specific version in your project.

## Features

*   **Dual Source Data:** Access both the high-level `IOKit` registry and the low-level `SMC` for a complete hardware picture.
*   **Source-Centric API:** The primary API returns a clean, structured `SystemInfo` object that strictly separates data by its source, eliminating ambiguity and data overwriting.
*   **Flexible Queries:** Choose to query IOKit, the SMC, or both, for maximum efficiency.
*   **Raw SMC Access:** A dedicated, powerful API for advanced users to query custom SMC keys and receive raw, undecoded values for their own interpretation.
*   **Command-Line Tool:** Includes `powerkit-cli`, a user-friendly tool to dump hardware info or perform raw SMC queries directly from your terminal.

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
      "IsConnected": true,
      "FullyCharged": false
    },
    "Battery": {
      "SerialNumber": "...",
      "DeviceName": "...",
      "CycleCount": 183,
      "DesignCapacity": 8579,
      "MaxCapacity": 7649,
      "NominalCapacity": 7893,
      "CurrentCapacity": 6050,
      "TimeToEmpty": 65535,
      "TimeToFull": 65535,
      "Temperature": 30.13,
      "Voltage": 12.27,
      "Amperage": 0,
      "IndividualCellVoltages": [
        4091,
        4092,
        4091
      ]
    },
    "Adapter": {
      "Description": "...",
      "MaxWatts": 60,
      "MaxVoltage": 20,
      "MaxAmperage": 3,
      "InputVoltage": 19.32,
      "InputAmperage": 0.13
    },
    "Calculations": {
      "HealthByMaxCapacity": 89,
      "HealthByNominalCapacity": 92,
      "ConditionAdjustedHealth": 95,
      "ACPower": 2.51,
      "BatteryPower": 0,
      "SystemPower": 2.51
    }
  },
  "SMC": {
    "Battery": {
      "Voltage": 12.27,
      "Amperage": 0
    },
    "Adapter": {
      "InputVoltage": 19.32,
      "InputAmperage": 0.13
    },
    "Calculations": {
      "ACPower": 2.51,
      "BatteryPower": 0,
      "SystemPower": 2.51
    }
  }
}
```

### 2. Advanced Usage: Querying a Single Source

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

### 3. Power User: Raw SMC Key Queries

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

The library includes a simple CLI tool.

### Installation
```bash
go install github.com/peterneutron/powerkit-go/cmd/powerkit-cli@latest
```

### Usage

**Get Help (default):**
```bash
powerkit-cli
```

**Dump all curated data:**
```bash
powerkit-cli all
```

**Dump only IOKit or SMC data:**
```bash
powerkit-cli iokit
powerkit-cli smc
```

**Perform a raw query for custom SMC keys:**
```bash
powerkit-cli raw FNum TC0D PPBR```
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

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.