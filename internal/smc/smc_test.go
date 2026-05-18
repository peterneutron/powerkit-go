//go:build darwin

package smc

import "testing"

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
