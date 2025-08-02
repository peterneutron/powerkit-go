// internal/smc/keys.go
package smc

// SMC Key Constants
const (
	// Adapter / DC In
	KeyAdapterVoltage = "VD0R"
	KeyAdapterCurrent = "ID0R"

	// Battery
	KeyBatteryVoltage = "B0AV"
	KeyBatteryCurrent = "B0AC"

	// Adapter Control
	KeyIsAdapterEnabled = "CHIE" // macOS >= 26.x only

	// Charge Control
	KeyIsChargingEnabled = "CHTE" // macOS >= 26.x only

	// Magsafe LED Control
	KeyMagsafeLED = "ACLC"
)

// KeysToRead is the standard list of keys fetched by the main GetSystemInfo function.
var KeysToRead = []string{
	KeyAdapterVoltage,
	KeyAdapterCurrent,
	KeyBatteryVoltage,
	KeyBatteryCurrent,
	KeyIsAdapterEnabled,
	KeyIsChargingEnabled,
	KeyMagsafeLED,
}
