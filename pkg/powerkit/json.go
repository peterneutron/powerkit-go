//go:build darwin

package powerkit

import "time"

const jsonSchemaVersion = "1.0.0"

// SystemInfoJSON is the stable JSON projection of SystemInfo.
type SystemInfoJSON struct {
	SchemaVersion string       `json:"schema_version"`
	CollectedAt   string       `json:"collected_at"`
	OS            OSJSON       `json:"os"`
	Battery       BatteryJSON  `json:"battery"`
	Adapter       AdapterJSON  `json:"adapter"`
	Power         PowerJSON    `json:"power"`
	Controls      ControlsJSON `json:"controls"`
	Sources       SourcesJSON  `json:"sources"`
}

// OSJSON contains OS-level status and firmware metadata.
type OSJSON struct {
	Firmware               string              `json:"firmware"`
	FirmwareVersion        string              `json:"firmware_version"`
	FirmwareSource         string              `json:"firmware_source"`
	FirmwareMajor          int                 `json:"firmware_major"`
	FirmwareCompatStatus   string              `json:"firmware_compat_status"`
	FirmwareProfileID      string              `json:"firmware_profile_id"`
	FirmwareProfileVersion int                 `json:"firmware_profile_version"`
	LowPowerMode           LowPowerModeJSON    `json:"low_power_mode"`
	SleepAssertions        SleepAssertionsJSON `json:"sleep_assertions"`
}

// LowPowerModeJSON models Low Power Mode state and availability.
type LowPowerModeJSON struct {
	Enabled   bool `json:"enabled"`
	Available bool `json:"available"`
}

// SleepAssertionsJSON groups global and app-local sleep assertion views.
type SleepAssertionsJSON struct {
	Global SleepAssertionStateJSON `json:"global"`
	App    SleepAssertionStateJSON `json:"app"`
}

// SleepAssertionStateJSON reports whether system/display sleep are allowed.
type SleepAssertionStateJSON struct {
	SystemSleepAllowed  bool `json:"system_sleep_allowed"`
	DisplaySleepAllowed bool `json:"display_sleep_allowed"`
}

// BatteryJSON contains battery state, identity, capacity, health, sensors, and time data.
type BatteryJSON struct {
	State    BatteryStateJSON    `json:"state"`
	Identity BatteryIdentityJSON `json:"identity"`
	Capacity BatteryCapacityJSON `json:"capacity"`
	Health   BatteryHealthJSON   `json:"health"`
	Sensors  BatterySensorsJSON  `json:"sensors"`
	Time     BatteryTimeJSON     `json:"time"`
}

// BatteryStateJSON reports battery connection and charging state flags.
type BatteryStateJSON struct {
	IsCharging   bool `json:"is_charging"`
	IsConnected  bool `json:"is_connected"`
	FullyCharged bool `json:"fully_charged"`
}

// BatteryIdentityJSON identifies the installed battery pack.
type BatteryIdentityJSON struct {
	SerialNumber string `json:"serial_number"`
	DeviceName   string `json:"device_name"`
}

// BatteryCapacityJSON exposes key battery capacity measurements.
type BatteryCapacityJSON struct {
	CurrentPercent int `json:"current_percent"`
	CurrentRaw     int `json:"current_raw"`
	Design         int `json:"design"`
	Max            int `json:"max"`
	Nominal        int `json:"nominal"`
}

// BatteryHealthJSON exposes normalized battery health indicators.
type BatteryHealthJSON struct {
	ByMaxCapacityPercent     int    `json:"by_max_capacity_percent"`
	ByNominalCapacityPercent int    `json:"by_nominal_capacity_percent"`
	ConditionAdjustedPercent int    `json:"condition_adjusted_percent"`
	VoltageDriftMV           int    `json:"voltage_drift_mv"`
	BalanceState             string `json:"balance_state"`
}

// BatterySensorsJSON contains live battery sensor readings.
type BatterySensorsJSON struct {
	VoltageMV      int     `json:"voltage_mv"`
	AmperageMA     int     `json:"amperage_ma"`
	TemperatureC   float64 `json:"temperature_c"`
	CellVoltagesMV []int   `json:"cell_voltages_mv"`
}

// BatteryTimeJSON reports estimated charge/discharge durations in minutes.
type BatteryTimeJSON struct {
	ToEmptyMin int `json:"to_empty_min"`
	ToFullMin  int `json:"to_full_min"`
}

// AdapterJSON contains adapter identity, ratings, input telemetry, and computed power.
type AdapterJSON struct {
	Description string            `json:"description"`
	Rating      AdapterRatingJSON `json:"rating"`
	Input       AdapterInputJSON  `json:"input"`
	PowerW      float64           `json:"power_w"`
}

// AdapterRatingJSON describes nominal adapter capabilities.
type AdapterRatingJSON struct {
	MaxWatts      int `json:"max_watts"`
	MaxVoltageMV  int `json:"max_voltage_mv"`
	MaxAmperageMA int `json:"max_amperage_ma"`
}

// AdapterInputJSON contains live adapter input telemetry values.
type AdapterInputJSON struct {
	VoltageMV          int  `json:"voltage_mv"`
	AmperageMA         int  `json:"amperage_ma"`
	TelemetryAvailable bool `json:"telemetry_available"`
}

// PowerJSON contains computed adapter, battery, and system power values.
type PowerJSON struct {
	AdapterW float64 `json:"adapter_w"`
	BatteryW float64 `json:"battery_w"`
	SystemW  float64 `json:"system_w"`
}

// ControlsJSON contains writable control state and capability flags.
type ControlsJSON struct {
	SMC          ControlsSMCJSON          `json:"smc"`
	Capabilities ControlsCapabilitiesJSON `json:"capabilities"`
}

// ControlsSMCJSON reports writable SMC state as observed by the library.
type ControlsSMCJSON struct {
	ChargingEnabled bool `json:"charging_enabled"`
	AdapterEnabled  bool `json:"adapter_enabled"`
}

// ControlsCapabilitiesJSON reports what query/write capabilities are available.
type ControlsCapabilitiesJSON struct {
	CanQueryIOKit bool `json:"can_query_iokit"`
	CanQuerySMC   bool `json:"can_query_smc"`
	CanWriteSMC   bool `json:"can_write_smc"`
}

// SourcesJSON records source availability and telemetry provenance.
type SourcesJSON struct {
	IOKit            SourceStatusJSON           `json:"iokit"`
	SMC              SourceStatusJSON           `json:"smc"`
	AdapterTelemetry AdapterTelemetrySourceJSON `json:"adapter_telemetry"`
}

// SourceStatusJSON reports whether a source was queried and had data available.
type SourceStatusJSON struct {
	Queried   bool `json:"queried"`
	Available bool `json:"available"`
}

// AdapterTelemetrySourceJSON describes adapter telemetry provenance decisions.
type AdapterTelemetrySourceJSON struct {
	Source        string `json:"source"`
	Available     bool   `json:"available"`
	Reason        string `json:"reason"`
	ForceFallback bool   `json:"force_fallback"`
}

// ToJSON projects SystemInfo into the stable v1 JSON contract.
func (s *SystemInfo) ToJSON() SystemInfoJSON {
	if s == nil {
		return SystemInfoJSON{SchemaVersion: jsonSchemaVersion}
	}

	collectedAt := s.collectedAt
	if collectedAt.IsZero() {
		collectedAt = time.Now().UTC()
	}

	out := SystemInfoJSON{
		SchemaVersion: jsonSchemaVersion,
		CollectedAt:   collectedAt.UTC().Format(time.RFC3339),
		OS: OSJSON{
			Firmware:               s.OS.Firmware,
			FirmwareVersion:        s.OS.FirmwareVersion,
			FirmwareSource:         s.OS.FirmwareSource,
			FirmwareMajor:          s.OS.FirmwareMajor,
			FirmwareCompatStatus:   s.OS.FirmwareCompatStatus,
			FirmwareProfileID:      s.OS.FirmwareProfileID,
			FirmwareProfileVersion: s.OS.FirmwareProfileVersion,
			LowPowerMode: LowPowerModeJSON{
				Enabled:   s.OS.LowPowerMode.Enabled,
				Available: s.OS.LowPowerMode.Available,
			},
			SleepAssertions: SleepAssertionsJSON{
				Global: SleepAssertionStateJSON{
					SystemSleepAllowed:  s.OS.GlobalSystemSleepAllowed,
					DisplaySleepAllowed: s.OS.GlobalDisplaySleepAllowed,
				},
				App: SleepAssertionStateJSON{
					SystemSleepAllowed:  s.OS.AppSystemSleepAllowed,
					DisplaySleepAllowed: s.OS.AppDisplaySleepAllowed,
				},
			},
		},
		Sources: SourcesJSON{
			IOKit: SourceStatusJSON{Queried: s.iokitQueried, Available: s.iokitAvailable},
			SMC:   SourceStatusJSON{Queried: s.smcQueried, Available: s.smcAvailable},
			AdapterTelemetry: AdapterTelemetrySourceJSON{
				Source:        s.adapterTelemetrySource,
				Reason:        s.adapterTelemetryReason,
				ForceFallback: s.forceTelemetryFallback,
			},
		},
		Controls: ControlsJSON{
			Capabilities: ControlsCapabilitiesJSON{
				CanQueryIOKit: true,
				CanQuerySMC:   true,
				CanWriteSMC:   true,
			},
		},
	}
	out.Battery.Health.BalanceState = string(BatteryBalanceUnknown)

	if s.IOKit != nil {
		out.Battery.State.IsCharging = s.IOKit.State.IsCharging
		out.Battery.State.IsConnected = s.IOKit.State.IsConnected
		out.Battery.State.FullyCharged = s.IOKit.State.FullyCharged
		out.Battery.Identity.SerialNumber = s.IOKit.Battery.SerialNumber
		out.Battery.Identity.DeviceName = s.IOKit.Battery.DeviceName
		out.Battery.Capacity.CurrentPercent = s.IOKit.Battery.CurrentCharge
		out.Battery.Capacity.CurrentRaw = s.IOKit.Battery.CurrentChargeRaw
		out.Battery.Capacity.Design = s.IOKit.Battery.DesignCapacity
		out.Battery.Capacity.Max = s.IOKit.Battery.MaxCapacity
		out.Battery.Capacity.Nominal = s.IOKit.Battery.NominalCapacity
		out.Battery.Health.ByMaxCapacityPercent = s.IOKit.Calculations.HealthByMaxCapacity
		out.Battery.Health.ByNominalCapacityPercent = s.IOKit.Calculations.HealthByNominalCapacity
		out.Battery.Health.ConditionAdjustedPercent = s.IOKit.Calculations.ConditionAdjustedHealth
		out.Battery.Health.VoltageDriftMV = s.IOKit.Calculations.VoltageDriftMV
		out.Battery.Health.BalanceState = string(s.IOKit.Calculations.BalanceState)
		out.Battery.Sensors.VoltageMV = int(s.IOKit.Battery.Voltage * 1000)
		out.Battery.Sensors.AmperageMA = int(s.IOKit.Battery.Amperage * 1000)
		out.Battery.Sensors.TemperatureC = s.IOKit.Battery.Temperature
		out.Battery.Sensors.CellVoltagesMV = append([]int(nil), s.IOKit.Battery.IndividualCellVoltages...)
		out.Battery.Time.ToEmptyMin = s.IOKit.Battery.TimeToEmpty
		out.Battery.Time.ToFullMin = s.IOKit.Battery.TimeToFull

		out.Adapter.Description = s.IOKit.Adapter.Description
		out.Adapter.Rating.MaxWatts = s.IOKit.Adapter.MaxWatts
		out.Adapter.Rating.MaxVoltageMV = int(s.IOKit.Adapter.MaxVoltage * 1000)
		out.Adapter.Rating.MaxAmperageMA = int(s.IOKit.Adapter.MaxAmperage * 1000)
		out.Adapter.Input.VoltageMV = int(s.IOKit.Adapter.InputVoltage * 1000)
		out.Adapter.Input.AmperageMA = int(s.IOKit.Adapter.InputAmperage * 1000)
		out.Adapter.Input.TelemetryAvailable = s.IOKit.Adapter.TelemetryAvailable

		out.Power.AdapterW = s.IOKit.Calculations.AdapterPower
		out.Power.BatteryW = s.IOKit.Calculations.BatteryPower
		out.Power.SystemW = s.IOKit.Calculations.SystemPower
		out.Adapter.PowerW = s.IOKit.Calculations.AdapterPower
	}

	if s.SMC != nil {
		out.Controls.SMC.ChargingEnabled = s.SMC.State.IsChargingEnabled
		out.Controls.SMC.AdapterEnabled = s.SMC.State.IsAdapterEnabled
		// Prefer SMC-derived power when available to expose fallback-consistent values.
		out.Power.AdapterW = s.SMC.Calculations.AdapterPower
		out.Power.BatteryW = s.SMC.Calculations.BatteryPower
		out.Power.SystemW = s.SMC.Calculations.SystemPower
		out.Adapter.PowerW = s.SMC.Calculations.AdapterPower
		if out.Adapter.Input.VoltageMV == 0 {
			out.Adapter.Input.VoltageMV = int(s.SMC.Adapter.InputVoltage * 1000)
		}
		if out.Adapter.Input.AmperageMA == 0 {
			out.Adapter.Input.AmperageMA = int(s.SMC.Adapter.InputAmperage * 1000)
		}
	}

	out.Sources.AdapterTelemetry.Available = out.Adapter.Input.TelemetryAvailable
	if out.Sources.AdapterTelemetry.Source == "" {
		out.Sources.AdapterTelemetry.Source = "unavailable"
	}
	if out.Sources.AdapterTelemetry.Reason == "" {
		out.Sources.AdapterTelemetry.Reason = "none"
	}

	return out
}
