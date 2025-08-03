//go:build darwin
// +build darwin

package powerkit

// --- Configuration Structs ---

// FetchOptions allows the user to specify which data sources to query.
// By default, both sources are enabled.
type FetchOptions struct {
	QueryIOKit bool
	QuerySMC   bool
}

// --- Top-Level Container Struct ---

// SystemInfo is the new top-level struct that holds all hardware information,
// cleanly separated by its source (IOKit or SMC).
type SystemInfo struct {
	OS    OSInfo     `json:"OS"`
	IOKit *IOKitData `json:"IOKit,omitempty"`
	SMC   *SMCData   `json:"SMC,omitempty"`
}

// --- OS-Specific Data Structures ---

// OSInfo holds information about the operating system environment.
type OSInfo struct {
	Mode string `json:"Mode"` // "Modern" or "Legacy"
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
	CurrentCapacity        int     `json:"CurrentCapacity"`
	TimeToEmpty            int     `json:"TimeToEmpty"`
	TimeToFull             int     `json:"TimeToFull"`
	Temperature            float64 `json:"Temperature"`
	Voltage                float64 `json:"Voltage"`
	Amperage               float64 `json:"Amperage"`
	CurrentCharge          int     `json:"CurrentCharge"`
	IndividualCellVoltages []int   `json:"IndividualCellVoltages"`
}

// IOKitAdapter contains all data points related to the adapter, as reported by IOKit.
type IOKitAdapter struct {
	Description   string  `json:"Description"`
	MaxWatts      int     `json:"MaxWatts"`
	MaxVoltage    float64 `json:"MaxVoltage"`
	MaxAmperage   float64 `json:"MaxAmperage"`
	InputVoltage  float64 `json:"InputVoltage"`
	InputAmperage float64 `json:"InputAmperage"`
}

// IOKitCalculations holds all health and power metrics derived from IOKit data.
type IOKitCalculations struct {
	HealthByMaxCapacity     int     `json:"HealthByMaxCapacity"`
	HealthByNominalCapacity int     `json:"HealthByNominalCapacity"`
	ConditionAdjustedHealth int     `json:"ConditionAdjustedHealth"`
	AdapterPower            float64 `json:"AdapterPower"`
	BatteryPower            float64 `json:"BatteryPower"`
	SystemPower             float64 `json:"SystemPower"`
}

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
