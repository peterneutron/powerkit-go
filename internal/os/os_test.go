//go:build darwin

package os

import "testing"

func TestExtractFirmwareMajor(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{name: "iboot format", input: "iBoot-13822.81.10", want: 13822},
		{name: "numeric dotted", input: "13822.0.233.0.0", want: 13822},
		{name: "with wrapper chars", input: "<\"iBoot-13822.81.10\">", want: 13822},
		{name: "malformed", input: "unknown", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := extractFirmwareMajor(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got major=%d", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, got)
			}
		})
	}
}

func TestParseFirmwareRecordFromSystemProfiler(t *testing.T) {
	output := `
Hardware:

    Hardware Overview:

      Model Name: MacBook Pro
      System Firmware Version: 13822.0.233.0.0
      OS Loader Version: 13822.81.10
`

	version, major, err := parseFirmwareRecordFromSystemProfiler(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != "13822.0.233.0.0" {
		t.Fatalf("unexpected version: %q", version)
	}
	if major != 13822 {
		t.Fatalf("expected major 13822, got %d", major)
	}
}

func TestParseFirmwareVersionFallbackLabels(t *testing.T) {
	output := `
      Boot ROM Version: iBoot-13822.81.10
      OS Loader Version: 13822.81.10
`
	major, err := parseFirmwareVersion(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if major != 13822 {
		t.Fatalf("expected major 13822, got %d", major)
	}
}
