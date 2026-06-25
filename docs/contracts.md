# Contract Details

This document holds the durable compatibility and contract detail for `powerkit-go`.

## Public Package

Primary package: `github.com/peterneutron/powerkit-go/pkg/powerkit`

### Telemetry

- `GetSystemInfo(opts ...FetchOptions) (*SystemInfo, error)`
- `GetSystemInfoContext(ctx context.Context, opts ...FetchOptions) (*SystemInfo, error)`
- `GetRawSMCValues(keys []string) (map[string]RawSMCValue, error)`
- `StreamSystemEvents() (<-chan SystemEvent, error)`
- `StreamSystemEventsWithHooks(StreamHooks) (<-chan SystemEvent, error)`
- `StreamSystemEventsContext(context.Context, StreamHooks) (<-chan SystemEvent, error)`
- `(*SystemInfo).ToJSON() SystemInfoJSON`

### Control APIs

- `SetChargingState(ChargingAction) error`
- `SetAdapterState(AdapterAction) error`
- `SetMagsafeLEDState(MagsafeLEDState) error`
- `GetMagsafeLEDState() (state MagsafeLEDState, available bool, err error)`
- `GetMagsafeStatus() (MagsafeStatus, error)`
- `GetLowPowerModeEnabled() (enabled bool, available bool, err error)`
- `SetLowPowerMode(enable bool) error`
- `ToggleLowPowerMode() error`
- `SetChargeLimit(percent int) error`

Charging and adapter state reads expose separate availability flags on
`SMCState`:

- `ChargingControlAvailable`
- `AdapterControlAvailable`

When an SMC control key is absent, the corresponding state defaults to enabled
and the availability flag is false. Consumers must check availability before
interpreting `IsChargingEnabled == false` as an active limiter or before issuing
charge-control writes.

Charge-limit behavior is exposed separately from the SMC firmware profile through
`SystemInfo.Controls.ChargeLimit`:

- `Backend`: `native_macos`, `smc_inhibit`, or `unavailable`
- `Available` / `Writable`
- `MinPercent`, `MaxPercent`, `StepPercent`
- `AllowedPercents` when the backend exposes a discrete set

`smc_inhibit` means callers must enforce the requested percentage by toggling
SMC charging state with `SetChargingState`; its step is a UI hint, not a hard
validation set. `native_macos` means callers may use `SetChargeLimit` directly
and must honor the returned allowed values.

Context-aware variants exist for mutating APIs:

- `SetAdapterStateContext`
- `SetChargingStateContext`
- `SetMagsafeLEDStateContext`
- `SetLowPowerModeContext`
- `ToggleLowPowerModeContext`

SMC-backed context variants check cancellation before starting the SMC read/write
path; the underlying cgo calls are not interruptible once started. Low Power Mode
reads use Foundation, while Low Power Mode writes pass cancellation to the
underlying `pmset` process.

### Sleep Assertions

- `CreateAssertion(AssertionType, reason string) (AssertionID, error)`
- `ReleaseAssertion(AssertionType)`
- `AllowAllSleep()`
- `IsAssertionActive(AssertionType) bool`
- `GetAssertionID(AssertionType) (AssertionID, bool)`

### Typed Errors

- `ErrPermissionRequired`
- `ErrNotSupported`
- `ErrTransientIO`

## JSON Contract

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

### Canonical Example

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
    "capacity": {
      "current_percent": 67,
      "current_raw": 65,
      "hardware_percent": 65,
      "hardware_percent_precise": 64.09,
      "hardware_percent_available": true
    },
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
  "controls": {
    "smc": {
      "charging_enabled": true,
      "adapter_enabled": true,
      "charging_control_available": true,
      "adapter_control_available": true
    },
    "charge_limit": {
      "available": true,
      "writable": true,
      "backend": "smc_inhibit",
      "min_percent": 60,
      "max_percent": 100,
      "step_percent": 10,
      "reason": "smc_control"
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

`battery.capacity.current_percent` is the macOS-facing displayed battery
percentage. `battery.capacity.current_raw` is the legacy raw smart-battery
`BatteryData.StateOfCharge` percentage. New consumers should prefer
`hardware_percent` / `hardware_percent_precise` when they need the battery
management system percentage; `hardware_percent_available` indicates whether
the raw smart-battery inputs were present.

`controls.smc.charging_control_available` and
`controls.smc.adapter_control_available` report whether the relevant SMC control
keys were observed. These fields are capability flags, not user preferences.
`controls.charge_limit` reports the selected charge-limit backend and range.

## macOS 27 Battery Notes

Observed on macOS 27.0 Developer Beta 2:

- Top-level `AppleSmartBattery` no longer exposes all legacy capacity,
  temperature, and cell-voltage values. `powerkit-go` preserves old top-level
  reads when present and fills missing values from nested `AppleSmartBatteryPack`
  and `AppleSmartBatteryBank` `BatteryData`.
- The modern SMC charging control key `CHTE` and legacy keys `BCLM`, `BCDS`, and
  `CH0B` were not available. The adapter control key `CHIE` remained readable.
- Missing charge-control keys are represented as
  `ChargingControlAvailable=false` with `IsChargingEnabled=true`, so callers do
  not confuse missing data with an active charging inhibit.
- Apple's native manual charge limit path is private. Runtime probing of
  `PowerUISmartChargeClient` reported available manual charge limits of
  `80, 85, 90, 95, 100`; attempts to set `60` failed with
  `PowerUISmartChargingErrorDomain Code=4`. `powerkit-go` models this as the
  `native_macos` charge-limit backend when the runtime probe succeeds.
- Direct `IOPSCopyBatteryLevelLimits` / `IOPSLimitBatteryLevel*` access is gated
  by Apple private entitlements such as `com.apple.private.iokit.soc-limit`.

## Firmware Profile Model

Resolver-related OS fields:

- `os.firmware`: resolver mode (`Supported` | `Legacy` | `Unknown`)
- `os.firmware_version`: normalized detected firmware string
- `os.firmware_source`: `ioreg_device_tree` | `system_profiler` | `unknown`
- `os.firmware_major`: parsed major used for profile selection
- `os.firmware_compat_status`: `tested` | `untested_newer` | `untested_older` | `unknown`
- `os.firmware_profile_id`: stable profile ID (`smc_profile_modern` | `smc_profile_legacy`)
- `os.firmware_profile_version`: independent numeric profile revision

Detection order:

1. IORegistry DeviceTree (`system-firmware-version`, then `firmware-version`)
2. `system_profiler SPHardwareDataType` fallback
3. `unknown` if both paths fail

## Adapter Telemetry Fallback

Telemetry provenance is explicit in `sources.adapter_telemetry`:

- `source`: `iokit` | `smc_fallback` | `unavailable`
- `reason`: `none` | `no_adapter` | `missing_iokit` | `invalid_iokit` | `forced` | `smc_error`
- `available`: boolean
- `force_fallback`: boolean

Connection-aware behavior:

- adapter disconnected: fallback is skipped and `reason=no_adapter`
- adapter connected plus missing or invalid IOKit telemetry: SMC fallback attempted
- `ForceTelemetryFallback` forces fallback while connected

## Compatibility Policy

For the `0.9.x` line:

- public Go API is additive-only unless explicitly documented as breaking
- JSON contract with `schema_version = "1.0.0"` is stable across `0.9.x` patches
- breaking JSON changes require a schema version bump and release notes
- additive fields are allowed; consumers should ignore unknown fields

Enum forward-compat policy:

- string enums are open sets
- consumers must handle unknown enum values gracefully

Module semver and JSON `schema_version` are separate concerns. `schema_version` describes serialization contract shape, not module release number.
