//go:build darwin
// +build darwin

package powerkit

// FetchOptions allows the user to specify which data sources to query.
// By default, both sources are enabled.
type FetchOptions struct {
	QueryIOKit bool
	QuerySMC   bool
}

// BatteryInfo holds a comprehensive snapshot of all data points.
// It separates data from its source (IOKit vs. SMC) for full transparency.
type BatteryInfo struct {
	State        State        `json:"State"`
	Battery      Battery      `json:"Battery"`
	Adapter      Adapter      `json:"Adapter"`
	SMC          *SMC         `json:"SMC,omitempty"` // omitempty will hide this field if SMC data is not available
	Calculations Calculations `json:"Calculations"`
}

// State holds booleans describing the current charging status.
// as reported by the high-level IOKit service.
type State struct {
	IsCharging   bool `json:"IsCharging"`
	IsConnected  bool `json:"IsConnected"`
	FullyCharged bool `json:"FullyCharged"`
}

// Battery contains all data points directly related to the battery itself,
// as reported by the high-level IOKit service.
type Battery struct {
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
	IndividualCellVoltages []int   `json:"IndividualCellVoltages"`
}

// Adapter contains all data points directly related to the adapter itself,
// as reported by the high-level IOKit service.
type Adapter struct {
	Description   string  `json:"Description"`
	MaxWatts      int     `json:"MaxWatts"`
	MaxVoltage    float64 `json:"MaxVoltage"`
	MaxAmperage   float64 `json:"MaxAmperage"`
	InputVoltage  float64 `json:"InputVoltage"`
	InputAmperage float64 `json:"InputAmperage"`
}

// SMC holds real-time, low-level sensor readings directly from the hardware.
type SMC struct {
	InputVoltage    float64 `json:"InputVoltage"`
	InputAmperage   float64 `json:"InputAmperage"`
	InputPower      float64 `json:"InputPower"`
	BatteryVoltage  float64 `json:"BatteryVoltage"`
	BatteryAmperage float64 `json:"BatteryAmperage"`
	BatteryPower    float64 `json:"BatteryPower"`
	SystemPower     float64 `json:"SystemPower"`
}

// PowerCalculation holds a set of power metrics derived from a single source.
type PowerCalculation struct {
	ACPower      float64 `json:"ACPower"`
	BatteryPower float64 `json:"BatteryPower"`
	SystemPower  float64 `json:"SystemPower"`
}

// Calculations holds health metrics and distinct sub-calculations for each data source.
type Calculations struct {
	HealthByMaxCapacity     int              `json:"HealthByMaxCapacity"`
	HealthByNominalCapacity int              `json:"HealthByNominalCapacity"`
	ConditionAdjustedHealth int              `json:"ConditionAdjustedHealth"`
	IOKit                   PowerCalculation `json:"IOKit"`
	SMC                     PowerCalculation `json:"SMC,omitempty"` // omitempty will hide this if no SMC data exists
}
