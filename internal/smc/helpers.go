//go:build darwin

package smc

import (
	"encoding/binary"
	"fmt"
	"math"
)

// decoderFunc defines the signature for any function that can decode a raw SMC byte slice.
type decoderFunc func(data []byte) (float64, error)

// smcDecoders is a dispatch map that links a data type string to its corresponding decoder function.
// This pattern avoids a large switch statement, reducing cyclomatic complexity.
var smcDecoders = map[string]decoderFunc{
	"flt ": decodeFlt,
	"sp78": decodeSp78,
	"fpe2": decodeFpe2,
	"ui8 ": decodeUI8,
	"ui16": decodeUI16,
	"ui32": decodeUI32,
	"si8 ": decodeSi8,
	"si16": decodeSi16,
	"flag": decodeFlag,
}

// decodeSMCValue is the Go "universal translator". It uses the smcDecoders dispatch map
// to find the correct decoder for the given data type and executes it.
func decodeSMCValue(dataType string, data []byte) (float64, error) {
	decoder, found := smcDecoders[dataType]
	if !found {
		return 0, fmt.Errorf("unsupported SMC data type: '%s'", dataType)
	}
	return decoder(data)
}

// --- Individual Decoder Functions ---

func decodeFlt(data []byte) (float64, error) {
	if len(data) != 4 {
		return 0, fmt.Errorf("invalid data size for type 'flt ': expected 4, got %d", len(data))
	}
	bits := binary.LittleEndian.Uint32(data)
	return float64(math.Float32frombits(bits)), nil
}

func decodeSp78(data []byte) (float64, error) {
	if len(data) != 2 {
		return 0, fmt.Errorf("invalid data size for type 'sp78': expected 2, got %d", len(data))
	}
	val := int16(binary.LittleEndian.Uint16(data))
	return float64(val) / 256.0, nil
}

func decodeFpe2(data []byte) (float64, error) {
	if len(data) != 2 {
		return 0, fmt.Errorf("invalid data size for type 'fpe2': expected 2, got %d", len(data))
	}
	val := binary.LittleEndian.Uint16(data)
	return float64(val) / 4.0, nil
}

func decodeUI8(data []byte) (float64, error) {
	if len(data) < 1 {
		return 0, fmt.Errorf("invalid data size for type 'ui8 ': expected at least 1, got %d", len(data))
	}
	return float64(data[0]), nil
}

func decodeUI16(data []byte) (float64, error) {
	if len(data) != 2 {
		return 0, fmt.Errorf("invalid data size for type 'ui16': expected 2, got %d", len(data))
	}
	return float64(binary.LittleEndian.Uint16(data)), nil
}

func decodeUI32(data []byte) (float64, error) {
	if len(data) != 4 {
		return 0, fmt.Errorf("invalid data size for type 'ui32': expected 4, got %d", len(data))
	}
	return float64(binary.LittleEndian.Uint32(data)), nil
}

func decodeSi8(data []byte) (float64, error) {
	if len(data) < 1 {
		return 0, fmt.Errorf("invalid data size for type 'si8 ': expected at least 1, got %d", len(data))
	}
	return float64(int8(data[0])), nil
}

func decodeSi16(data []byte) (float64, error) {
	if len(data) != 2 {
		return 0, fmt.Errorf("invalid data size for type 'si16': expected 2, got %d", len(data))
	}
	return float64(int16(binary.LittleEndian.Uint16(data))), nil
}

func decodeFlag(data []byte) (float64, error) {
	if len(data) < 1 {
		return 0, fmt.Errorf("invalid data size for type 'flag': expected at least 1, got %d", len(data))
	}
	return float64(data[0]), nil
}
