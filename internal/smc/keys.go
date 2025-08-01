// internal/smc/keys.go
package smc

// SMC Key Constants
const (
	// Adapter / DC In
	KeyDCInVoltage = "VD0R"
	KeyDCInCurrent = "ID0R"

	// Battery
	KeyBatteryVoltage = "B0AV"
	KeyBatteryCurrent = "B0AC"

	// Adapter Control
	KeyChargerControl = "CHIE" // macOS >= 26.x only

	// Charge Control
	KeyChargeInhibit = "CHTE" // macOS >= 26.x only

	// Magsafe LED Control
	KeyMagsafeLED = "ACLC"
)

// KeysToRead is the standard list of keys fetched by the main GetSystemInfo function.
var KeysToRead = []string{
	KeyDCInVoltage,
	KeyDCInCurrent,
	KeyBatteryVoltage,
	KeyBatteryCurrent,
}
