package smc

// List of keys to read from the System Management Controller (SMC).
var KeysToRead = []string{
	"VD0R", // DC In Voltage
	"ID0R", // DC In Current
	"PDTR", // DC In Power (Watts)
	"B0AV", // Battery Voltage
	"B0AC", // Battery Current
	"PPBR", // Battery Power (Watts)
	"PSTR", // System Power (Watts)
}
