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

Context-aware variants exist for mutating APIs:

- `SetAdapterStateContext`
- `SetChargingStateContext`
- `SetMagsafeLEDStateContext`
- `SetLowPowerModeContext`
- `ToggleLowPowerModeContext`

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
