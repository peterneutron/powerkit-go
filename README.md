# PowerKit-Go

`powerkit-go` is a Darwin-only Go library for reading and controlling macOS power-related state via IOKit/SMC and OS power interfaces.

## Status
Pre-release. Public APIs can still change.

## Platform
- macOS only (`//go:build darwin`)
- cgo required (links CoreFoundation/IOKit/Foundation)

## Install
```bash
go get github.com/peterneutron/powerkit-go
```

## Core API
Public package: `github.com/peterneutron/powerkit-go/pkg/powerkit`

### Telemetry
- `GetSystemInfo(opts ...FetchOptions) (*SystemInfo, error)`
- `GetSystemInfoContext(ctx, opts ...FetchOptions) (*SystemInfo, error)`
- `GetRawSMCValues(keys []string) (map[string]RawSMCValue, error)`
- `StreamSystemEvents() (<-chan SystemEvent, error)`
- `(*SystemInfo).ToJSON() SystemInfoJSON`

### Control APIs (privileged)
These return `ErrPermissionRequired` when not root.

- Charging/adapter:
  - `SetChargingState(ChargingAction)`
  - `SetAdapterState(AdapterAction)`
- MagSafe LED:
  - `SetMagsafeLEDState(MagsafeLEDState)`
  - `GetMagsafeLEDState() (state MagsafeLEDState, available bool, err error)`
  - `GetMagsafeStatus() (MagsafeStatus, error)`
- Low Power Mode:
  - `GetLowPowerModeEnabled() (enabled bool, available bool, err error)`
  - `SetLowPowerMode(enable bool)`
  - `ToggleLowPowerMode()`

Context variants are available for mutating calls (`*Context` methods).

### Sleep assertions (no root required)
- `CreateAssertion(AssertionType, reason string) (AssertionID, error)`
- `ReleaseAssertion(AssertionType)`
- `AllowAllSleep()`
- `IsAssertionActive(AssertionType) bool`
- `GetAssertionID(AssertionType) (AssertionID, bool)`

## Typed Errors
- `ErrPermissionRequired`
- `ErrNotSupported`
- `ErrTransientIO`

## JSON Contract (v1)
- JSON output is now a single stable domain-first schema with snake_case keys.
- Versioning is explicit via `schema_version` (no type/root suffixes).
- Legacy PascalCase JSON is removed; pin an older version if you need the previous shape.

Top-level keys:
- `schema_version`
- `collected_at`
- `os`
- `battery`
- `adapter`
- `power`
- `controls`
- `sources`

OS firmware fields:
- `os.firmware`: resolver mode (`Supported` | `Legacy` | `Unknown`)
- `os.firmware_version`: normalized detected firmware string (for diagnostics)
- `os.firmware_source`: `ioreg_device_tree` | `system_profiler` | `unknown`

Firmware detection order:
1. IORegistry DeviceTree keys (`system-firmware-version`, then `firmware-version`)
2. Fallback to `system_profiler SPHardwareDataType`
3. If both fail, source is `unknown` and resolver uses unknown-mode behavior

Battery drift metrics:
- `battery.health.voltage_drift_mv`: max(cell_mv) - min(cell_mv)
- `battery.health.balance_state`: `balanced` | `slight_imbalance` | `high_imbalance` | `unknown`
- Thresholds:
  - `balanced`: `<= 10 mV`
  - `slight_imbalance`: `11..30 mV`
  - `high_imbalance`: `> 30 mV`
  - `unknown`: fewer than 2 cell voltages

Telemetry provenance:
- `sources.adapter_telemetry.source`: `iokit` | `smc_fallback` | `unavailable`
- `sources.adapter_telemetry.reason`: `none` | `no_adapter` | `missing_iokit` | `invalid_iokit` | `forced` | `smc_error`
- `sources.adapter_telemetry.available`: boolean
- `sources.adapter_telemetry.force_fallback`: boolean

Connection-aware fallback behavior:
- When adapter is disconnected, telemetry fallback is skipped and `reason=no_adapter`.
- When connected, invalid/missing IOKit telemetry falls back to SMC.
- `ForceTelemetryFallback` still forces SMC adapter telemetry when connected.

## Minimal Example
```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/peterneutron/powerkit-go/pkg/powerkit"
)

func main() {
	info, err := powerkit.GetSystemInfoContext(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(info.IOKit.Battery.CurrentCharge)
}
```

## CLI (`powerkit-cli`)
Build:
```bash
go build -o powerkit-cli ./cmd/powerkit-cli
```

Examples:
```bash
./powerkit-cli all
./powerkit-cli watch
./powerkit-cli raw FNum ACLC
sudo ./powerkit-cli charging off
sudo ./powerkit-cli magsafe set-color amber
./powerkit-cli lowpower get
```

## Development
From `powerkit-go/`:

- `make tests`
- `make vet`
- `make lint`
- `make verify`

## Safety Notes
- SMC writes can affect charging/adapter behavior. Use write APIs carefully.
- Prefer read APIs and assertions unless you specifically need SMC control.
