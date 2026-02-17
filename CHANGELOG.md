# Changelog

All notable changes to this project are documented in this file.

## v0.9.0

### Added
- Stable JSON v1 contract serialization via `(*SystemInfo).ToJSON()` with domain-first snake_case structure.
- Firmware metadata fields in JSON/typed output:
  - `os.firmware_version`
  - `os.firmware_source`
  - `os.firmware_major`
  - `os.firmware_compat_status`
  - `os.firmware_profile_id`
  - `os.firmware_profile_version`
- Adapter telemetry provenance contract:
  - `sources.adapter_telemetry.source`
  - `sources.adapter_telemetry.reason`
  - `sources.adapter_telemetry.available`
  - `sources.adapter_telemetry.force_fallback`
- Battery drift diagnostics:
  - `battery.health.voltage_drift_mv`
  - `battery.health.balance_state`

### Changed
- Firmware detection now prioritizes IORegistry DeviceTree and falls back to `system_profiler`.
- Adapter telemetry fallback is connection-aware and reason-coded (`no_adapter`, `missing_iokit`, `invalid_iokit`, `forced`, `smc_error`).
- Documentation now defines a `0.9.x` additive-compatibility policy and forward-compatible enum handling guidance.

### Breaking and Compatibility Notes
- Legacy PascalCase CLI JSON shape is removed; v1 snake_case JSON is the supported contract.
- Consumers that require the old shape should pin to a pre-v1-json release.

### Release Flow
1. Merge `dev` into `master`.
2. Tag `v0.9.0` on `master`.
3. Continue with `0.9.x` additive hardening patches toward `v1.0.0`.
