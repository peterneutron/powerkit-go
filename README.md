# PowerKit-Go

`powerkit-go` is a Darwin-only Go library for reading and controlling macOS power state via IOKit, SMC, and OS power interfaces.

## Status
Release target: `v0.9.0` (prepared on `dev`, tagged after merge to `master`).

## Platform and Support Matrix
- OS: macOS only (`//go:build darwin`)
- Architecture target: Apple Silicon (project target), Intel behavior is not guaranteed
- Tooling: cgo required (CoreFoundation/IOKit/Foundation)

## Install
```bash
go get github.com/peterneutron/powerkit-go
```

## Privilege Model
- Read telemetry APIs: no root required.
- Sleep assertion APIs: no root required.
- Mutating power control APIs (charging/adapter/MagSafe/low-power): require root and return `ErrPermissionRequired` when denied.

## Quick Start
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

	payload := info.ToJSON()
	fmt.Println(payload.SchemaVersion, payload.Battery.Capacity.CurrentPercent)
}
```

## Public API Surface
Public package: `github.com/peterneutron/powerkit-go/pkg/powerkit`

### Telemetry
- `GetSystemInfo(opts ...FetchOptions) (*SystemInfo, error)`
- `GetSystemInfoContext(ctx context.Context, opts ...FetchOptions) (*SystemInfo, error)`
- `GetRawSMCValues(keys []string) (map[string]RawSMCValue, error)`
- `StreamSystemEvents() (<-chan SystemEvent, error)`
- `(*SystemInfo).ToJSON() SystemInfoJSON`

### Control APIs (root required)
- Charging and adapter:
  - `SetChargingState(ChargingAction) error`
  - `SetAdapterState(AdapterAction) error`
- MagSafe LED:
  - `SetMagsafeLEDState(MagsafeLEDState) error`
  - `GetMagsafeLEDState() (state MagsafeLEDState, available bool, err error)`
  - `GetMagsafeStatus() (MagsafeStatus, error)`
- Low Power Mode:
  - `GetLowPowerModeEnabled() (enabled bool, available bool, err error)`
  - `SetLowPowerMode(enable bool) error`
  - `ToggleLowPowerMode() error`

Context variants are available for mutating calls:
- `SetAdapterStateContext`
- `SetChargingStateContext`
- `SetMagsafeLEDStateContext`
- `SetLowPowerModeContext`
- `ToggleLowPowerModeContext`

### Sleep Assertions (no root required)
- `CreateAssertion(AssertionType, reason string) (AssertionID, error)`
- `ReleaseAssertion(AssertionType)`
- `AllowAllSleep()`
- `IsAssertionActive(AssertionType) bool`
- `GetAssertionID(AssertionType) (AssertionID, bool)`

### Typed Errors
- `ErrPermissionRequired`
- `ErrNotSupported`
- `ErrTransientIO`

## JSON Contract (v1)
JSON serialization uses `(*SystemInfo).ToJSON()` and follows a single stable, domain-first schema:
- snake_case keys
- explicit schema contract marker: `schema_version`
- no legacy PascalCase compatibility mode

Top-level keys:
- `schema_version`
- `collected_at`
- `os`
- `battery`
- `adapter`
- `power`
- `controls`
- `sources`

### Canonical JSON Example
```json
{
  "schema_version": "1.0.0",
  "collected_at": "2026-02-17T11:30:00Z",
  "os": {
    "firmware": "Supported",
    "firmware_version": "iBoot-13822.81.10",
    "firmware_source": "ioreg_device_tree",
    "firmware_major": 13822,
    "firmware_compat_status": "tested",
    "firmware_profile_id": "smc_profile_modern",
    "firmware_profile_version": 1,
    "low_power_mode": { "enabled": false, "available": true },
    "sleep_assertions": {
      "global": { "system_sleep_allowed": true, "display_sleep_allowed": true },
      "app": { "system_sleep_allowed": true, "display_sleep_allowed": true }
    }
  },
  "battery": {
    "health": {
      "voltage_drift_mv": 18,
      "balance_state": "slight_imbalance"
    }
  },
  "adapter": {
    "input": {
      "telemetry_available": true
    }
  },
  "sources": {
    "adapter_telemetry": {
      "source": "iokit",
      "available": true,
      "reason": "none",
      "force_fallback": false
    }
  }
}
```

## Firmware Profile Model
Resolver-related OS fields:
- `os.firmware`: resolver mode (`Supported` | `Legacy` | `Unknown`)
- `os.firmware_version`: normalized detected firmware string
- `os.firmware_source`: `ioreg_device_tree` | `system_profiler` | `unknown`
- `os.firmware_major`: parsed major used for profile selection
- `os.firmware_compat_status`: `tested` | `untested_newer` | `untested_older` | `unknown`
- `os.firmware_profile_id`: stable profile ID (`smc_profile_modern` | `smc_profile_legacy`)
- `os.firmware_profile_version`: independent numeric profile revision (currently `1`)

Detection order:
1. IORegistry DeviceTree (`system-firmware-version`, then `firmware-version`)
2. `system_profiler SPHardwareDataType` fallback
3. `unknown` if both paths fail

## Adapter Telemetry Fallback and Gating
Telemetry provenance is explicit in `sources.adapter_telemetry`:
- `source`: `iokit` | `smc_fallback` | `unavailable`
- `reason`: `none` | `no_adapter` | `missing_iokit` | `invalid_iokit` | `forced` | `smc_error`
- `available`: boolean
- `force_fallback`: boolean

Connection-aware behavior:
- Adapter disconnected: fallback is skipped and `reason=no_adapter`.
- Adapter connected + missing/invalid IOKit telemetry: SMC fallback attempted.
- `ForceTelemetryFallback` forces fallback while connected.

## 0.9 Stability Policy
For the `0.9.x` line:
- Public Go API is additive-only unless explicitly documented as breaking.
- JSON contract with `schema_version = "1.0.0"` is stable across `0.9.x` patches.
- Breaking JSON changes require a schema version bump and release notes.
- Additive fields are allowed; consumers should ignore unknown fields.

Enum forward-compat policy:
- String enums are open sets (`balance_state`, `firmware_compat_status`, adapter telemetry `source`/`reason`).
- Consumers must handle unknown enum values gracefully (display fallback, non-fatal parsing).

## Versioning and Release Workflow
- Module semver (e.g., `v0.9.0`) and JSON `schema_version` are separate concerns.
- `schema_version` describes serialization contract shape, not module release number.

Release workflow for `v0.9.0`:
1. Finalize on `dev`.
2. Merge `dev` -> `master`.
3. Tag `v0.9.0` on `master`.
4. Use `0.9.x` patches to harden toward `v1.0.0`.

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
- SMC writes can alter charging/adapter behavior.
- Prefer read APIs/assertions unless you explicitly need hardware control.
