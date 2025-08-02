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
	KeyIsAdapterEnabled        = "CHIE" // macOS >= 26.x only
	KeyIsAdapterEnabled_Legacy = "CH0B" // macOS < 26.x only

	// Charge Control
	KeyIsChargingEnabled             = "CHTE" // macOS >= 26.x only
	KeyIsChargingEnabled_Legacy_BCLM = "BCLM" // macOS < 26.x only
	KeyIsChargingEnabled_Legacy_BCDS = "BCDS" // macOS < 26.x only

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
	KeyIsAdapterEnabled,
	KeyIsAdapterEnabled_Legacy,
	KeyIsChargingEnabled,
	KeyIsChargingEnabled_Legacy_BCLM,
	KeyIsChargingEnabled_Legacy_BCDS,
}
