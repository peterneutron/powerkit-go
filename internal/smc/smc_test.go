//go:build darwin

package smc

import (
	"encoding/binary"
	"math"
	"strings"
	"testing"
)

func TestValidateSMCKeyAcceptsFourByteKeys(t *testing.T) {
	for _, key := range []string{"CH0B", "ACLC", "BCLM", "BCDS"} {
		if err := validateSMCKey(key); err != nil {
			t.Fatalf("validateSMCKey(%q) returned error: %v", key, err)
		}
	}
}

func TestValidateSMCKeyRejectsInvalidKeys(t *testing.T) {
	for _, key := range []string{"", "A", "ABC", "ABCDE", "AB\x00D"} {
		if err := validateSMCKey(key); err == nil {
			t.Fatalf("validateSMCKey(%q) returned nil, want error", key)
		}
	}
}

func TestDecodeSMCValueDecodesSupportedTypes(t *testing.T) {
	flt := make([]byte, 4)
	binary.LittleEndian.PutUint32(flt, math.Float32bits(12.5))

	tests := []struct {
		name     string
		dataType string
		data     []byte
		want     float64
	}{
		{
			name:     "flt",
			dataType: "flt ",
			data:     flt,
			want:     12.5,
		},
		{
			name:     "sp78 positive",
			dataType: "sp78",
			data:     littleEndianUint16(0x1980),
			want:     25.5,
		},
		{
			name:     "sp78 negative",
			dataType: "sp78",
			data:     littleEndianInt16(-384),
			want:     -1.5,
		},
		{
			name:     "fpe2",
			dataType: "fpe2",
			data:     littleEndianUint16(401),
			want:     100.25,
		},
		{
			name:     "ui8",
			dataType: "ui8 ",
			data:     []byte{42, 99},
			want:     42,
		},
		{
			name:     "ui16",
			dataType: "ui16",
			data:     littleEndianUint16(513),
			want:     513,
		},
		{
			name:     "ui32",
			dataType: "ui32",
			data:     littleEndianUint32(66051),
			want:     66051,
		},
		{
			name:     "si8 negative",
			dataType: "si8 ",
			data:     []byte{0xfe},
			want:     -2,
		},
		{
			name:     "si16 negative",
			dataType: "si16",
			data:     littleEndianInt16(-513),
			want:     -513,
		},
		{
			name:     "flag",
			dataType: "flag",
			data:     []byte{1},
			want:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeSMCValue(tt.dataType, tt.data)
			if err != nil {
				t.Fatalf("decodeSMCValue(%q, %v) returned error: %v", tt.dataType, tt.data, err)
			}
			if got != tt.want {
				t.Fatalf("decodeSMCValue(%q, %v) = %v, want %v", tt.dataType, tt.data, got, tt.want)
			}
		})
	}
}

func TestDecodeSMCValueRejectsUnsupportedType(t *testing.T) {
	_, err := decodeSMCValue("bad!", []byte{1, 2, 3, 4})
	if err == nil {
		t.Fatal("decodeSMCValue returned nil error for unsupported type")
	}
	if !strings.Contains(err.Error(), "unsupported SMC data type") {
		t.Fatalf("decodeSMCValue error = %q, want unsupported type error", err)
	}
}

func TestDecodeSMCValueRejectsInvalidSizes(t *testing.T) {
	tests := []struct {
		name     string
		dataType string
		data     []byte
	}{
		{name: "flt short", dataType: "flt ", data: []byte{1, 2, 3}},
		{name: "flt long", dataType: "flt ", data: []byte{1, 2, 3, 4, 5}},
		{name: "sp78 short", dataType: "sp78", data: []byte{1}},
		{name: "sp78 long", dataType: "sp78", data: []byte{1, 2, 3}},
		{name: "fpe2 short", dataType: "fpe2", data: []byte{1}},
		{name: "fpe2 long", dataType: "fpe2", data: []byte{1, 2, 3}},
		{name: "ui8 empty", dataType: "ui8 ", data: nil},
		{name: "ui16 short", dataType: "ui16", data: []byte{1}},
		{name: "ui16 long", dataType: "ui16", data: []byte{1, 2, 3}},
		{name: "ui32 short", dataType: "ui32", data: []byte{1, 2, 3}},
		{name: "ui32 long", dataType: "ui32", data: []byte{1, 2, 3, 4, 5}},
		{name: "si8 empty", dataType: "si8 ", data: nil},
		{name: "si16 short", dataType: "si16", data: []byte{1}},
		{name: "si16 long", dataType: "si16", data: []byte{1, 2, 3}},
		{name: "flag empty", dataType: "flag", data: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decodeSMCValue(tt.dataType, tt.data)
			if err == nil {
				t.Fatalf("decodeSMCValue(%q, %v) returned nil error", tt.dataType, tt.data)
			}
			if !strings.Contains(err.Error(), "invalid data size") {
				t.Fatalf("decodeSMCValue error = %q, want invalid data size error", err)
			}
		})
	}
}

func littleEndianUint16(value uint16) []byte {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, value)
	return data
}

func littleEndianInt16(value int16) []byte {
	return littleEndianUint16(uint16(value))
}

func littleEndianUint32(value uint32) []byte {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, value)
	return data
}
