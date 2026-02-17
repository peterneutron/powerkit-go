//go:build darwin

package powerkit

import "testing"

func TestFirmwareCompatStatus(t *testing.T) {
	tests := []struct {
		name  string
		major int
		want  string
	}{
		{name: "tested", major: FirmwareMajorVersionThreshold, want: firmwareCompatTested},
		{name: "untested newer", major: FirmwareMajorVersionThreshold + 1, want: firmwareCompatUntestedNew},
		{name: "untested older", major: FirmwareMajorVersionThreshold - 1, want: firmwareCompatUntestedOld},
		{name: "unknown zero", major: 0, want: firmwareCompatUnknown},
		{name: "unknown negative", major: -1, want: firmwareCompatUnknown},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := firmwareCompatStatus(tc.major); got != tc.want {
				t.Fatalf("firmwareCompatStatus(%d) = %q, want %q", tc.major, got, tc.want)
			}
		})
	}
}
