//go:build darwin

package powerkit

import "time"

// --- Event Streaming Structs ---

// EventType defines the kind of system event that occurred.
type EventType int

const (
	// EventTypeBatteryUpdate signifies a change in battery state (e.g., percentage).
	// The `Info` field of the SystemEvent will be populated.
	EventTypeBatteryUpdate EventType = iota

	// EventTypeSystemWillSleep signifies the system is about to enter sleep.
	// The `Info` field of the SystemEvent will be nil.
	EventTypeSystemWillSleep

	// EventTypeSystemDidWake signifies the system has just woken from sleep.
	// The `Info` field of the SystemEvent will be nil.
	EventTypeSystemDidWake
)

// SystemEvent is the unified structure delivered by the event stream. It contains
// the type of event and, if applicable, the associated system information.
type SystemEvent struct {
	Type EventType   `json:"Type"`
	Info *SystemInfo `json:"Info,omitempty"` // Populated only for EventTypeBatteryUpdate
}

// --- Configuration Structs ---

// FetchOptions allows the user to specify which data sources to query.
// By default, both sources are enabled.
type FetchOptions struct {
	QueryIOKit             bool
	QuerySMC               bool
	ForceTelemetryFallback bool
}

// --- Top-Level Container Struct ---

// SystemInfo is the new top-level struct that holds all hardware information,
// cleanly separated by its source (IOKit or SMC).
type SystemInfo struct {
	OS    OSInfo     `json:"OS"`
	IOKit *IOKitData `json:"IOKit,omitempty"`
	SMC   *SMCData   `json:"SMC,omitempty"`

	collectedAt            time.Time
	iokitQueried           bool
	smcQueried             bool
	iokitAvailable         bool
	smcAvailable           bool
	adapterTelemetrySource string
	adapterTelemetryReason string
	forceTelemetryFallback bool
}

// --- OS-Specific Data Structures ---

// OSInfo holds information about the operating system environment.
type OSInfo struct {
	Firmware string `json:"Firmware"` // "Supported", "Legacy" or "Unknown"
	// FirmwareVersion is the normalized detected firmware version string.
	FirmwareVersion string `json:"FirmwareVersion"`
	// FirmwareSource identifies where FirmwareVersion/Firmware major were sourced.
	// Values: ioreg_device_tree | system_profiler | unknown
	FirmwareSource string `json:"FirmwareSource"`
	// FirmwareMajor is the parsed major firmware version used for resolver gating.
	FirmwareMajor int `json:"FirmwareMajor"`
	// FirmwareCompatStatus indicates how current firmware relates to tested threshold.
	// Values: tested | untested_newer | untested_older | unknown
	FirmwareCompatStatus string `json:"FirmwareCompatStatus"`
	// FirmwareProfileID is a stable identifier for the selected SMC control profile.
	FirmwareProfileID string `json:"FirmwareProfileID"`
	// FirmwareProfileVersion is an independent profile revision number.
	FirmwareProfileVersion int `json:"FirmwareProfileVersion"`

	// Global: reflect system-wide assertion state (any process)
	GlobalSystemSleepAllowed  bool `json:"GlobalSystemSleepAllowed"`
	GlobalDisplaySleepAllowed bool `json:"GlobalDisplaySleepAllowed"`

	// App: reflect assertions created by this process (library client)
	AppSystemSleepAllowed  bool `json:"AppSystemSleepAllowed"`
	AppDisplaySleepAllowed bool `json:"AppDisplaySleepAllowed"`

	// LowPowerMode reports macOS Low Power Mode state and availability
	LowPowerMode LowPowerModeInfo `json:"LowPowerMode"`
}

// LowPowerModeInfo contains the LPM state and availability flag
type LowPowerModeInfo struct {
	Enabled   bool `json:"Enabled"`
	Available bool `json:"Available"`
}

// --- IOKit-Specific Data Structures ---

// IOKitData is a container for all data and calculations derived from the IOKit registry.
type IOKitData struct {
	State        IOKitState        `json:"State"`
	Battery      IOKitBattery      `json:"Battery"`
	Adapter      IOKitAdapter      `json:"Adapter"`
	Calculations IOKitCalculations `json:"Calculations"`
}

// IOKitState holds booleans describing the current charging status, sourced from IOKit.
type IOKitState struct {
	IsCharging   bool `json:"IsCharging"`
	IsConnected  bool `json:"IsConnected"`
	FullyCharged bool `json:"FullyCharged"`
}

// IOKitBattery contains all data points related to the battery itself, as reported by IOKit.
type IOKitBattery struct {
	SerialNumber           string  `json:"SerialNumber"`
	DeviceName             string  `json:"DeviceName"`
	CycleCount             int     `json:"CycleCount"`
	DesignCapacity         int     `json:"DesignCapacity"`
	MaxCapacity            int     `json:"MaxCapacity"`
	NominalCapacity        int     `json:"NominalCapacity"`
	CurrentCapacityRaw     int     `json:"CurrentCapacityRaw"`
	TimeToEmpty            int     `json:"TimeToEmpty"`
	TimeToFull             int     `json:"TimeToFull"`
	Temperature            float64 `json:"Temperature"`
	Voltage                float64 `json:"Voltage"`
	Amperage               float64 `json:"Amperage"`
	CurrentCharge          int     `json:"CurrentCharge"`
	CurrentChargeRaw       int     `json:"CurrentChargeRaw"`
	IndividualCellVoltages []int   `json:"IndividualCellVoltages"`
}

// IOKitAdapter contains all data points related to the adapter, as reported by IOKit.
type IOKitAdapter struct {
	Description        string  `json:"Description"`
	MaxWatts           int     `json:"MaxWatts"`
	MaxVoltage         float64 `json:"MaxVoltage"`
	MaxAmperage        float64 `json:"MaxAmperage"`
	InputVoltage       float64 `json:"InputVoltage"`
	InputAmperage      float64 `json:"InputAmperage"`
	TelemetryAvailable bool    `json:"TelemetryAvailable"`
}

// IOKitCalculations holds all health and power metrics derived from IOKit data.
type IOKitCalculations struct {
	HealthByMaxCapacity     int                 `json:"HealthByMaxCapacity"`
	HealthByNominalCapacity int                 `json:"HealthByNominalCapacity"`
	ConditionAdjustedHealth int                 `json:"ConditionAdjustedHealth"`
	VoltageDriftMV          int                 `json:"VoltageDriftMV"`
	BalanceState            BatteryBalanceState `json:"BalanceState"`
	AdapterPower            float64             `json:"AdapterPower"`
	BatteryPower            float64             `json:"BatteryPower"`
	SystemPower             float64             `json:"SystemPower"`
}

type BatteryBalanceState string

const (
	BatteryBalanceUnknown         BatteryBalanceState = "unknown"
	BatteryBalanceBalanced        BatteryBalanceState = "balanced"
	BatteryBalanceSlightImbalance BatteryBalanceState = "slight_imbalance"
	BatteryBalanceHighImbalance   BatteryBalanceState = "high_imbalance"
)

// --- SMC-Specific Data Structures ---

// SMCData is a container for all raw sensor data and calculations derived from the SMC.
type SMCData struct {
	State        SMCState        `json:"State"`
	Battery      SMCBattery      `json:"Battery"`
	Adapter      SMCAdapter      `json:"Adapter"`
	Calculations SMCCalculations `json:"Calculations"`
}

// SMCState holds booleans describing the adapter and charging enable/disable state.
type SMCState struct {
	IsChargingEnabled bool `json:"IsChargingEnabled"` // was IsChargingEnabled
	IsAdapterEnabled  bool `json:"IsAdapterEnabled"`  // was IsAdapterEnabled
}

// SMCBattery holds raw battery-related sensor readings from the SMC.
type SMCBattery struct {
	Voltage  float64 `json:"Voltage"`
	Amperage float64 `json:"Amperage"`
}

// SMCAdapter holds raw adapter-related sensor readings from the SMC.
type SMCAdapter struct {
	InputVoltage  float64 `json:"InputVoltage"`
	InputAmperage float64 `json:"InputAmperage"`
}

// SMCCalculations holds power metrics derived purely from SMC sensor readings.
type SMCCalculations struct {
	AdapterPower float64 `json:"AdapterPower"`
	BatteryPower float64 `json:"BatteryPower"`
	SystemPower  float64 `json:"SystemPower"`
}

// RawSMCValue holds the raw, undecoded result of a custom SMC query.
// It is the responsibility of the caller to decode the Data bytes
// based on the DataType and DataSize.
type RawSMCValue struct {
	DataType string `json:"DataType"`
	DataSize int    `json:"DataSize"`
	Data     []byte `json:"Data"` // This will be base64-encoded in JSON for readability
}
