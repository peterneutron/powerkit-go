//go:build darwin && powerkit_integration

package powerkit

import (
	"errors"
	"os"
	"testing"

	"github.com/peterneutron/powerkit-go/internal/smc"
)

func TestIntegrationReadSystemInfo(t *testing.T) {
	info, err := GetSystemInfo()
	if err != nil {
		t.Fatalf("GetSystemInfo returned error: %v", err)
	}
	if info == nil {
		t.Fatal("GetSystemInfo returned nil info")
	}
	if info.OS.FirmwareProfileID == "" {
		t.Fatal("expected firmware profile metadata to be populated")
	}
	if info.IOKit == nil && info.SMC == nil {
		t.Fatal("expected at least one telemetry source to be available")
	}
}

func TestIntegrationReadLowPowerMode(t *testing.T) {
	_, _, err := GetLowPowerModeEnabled()
	if err != nil {
		t.Fatalf("GetLowPowerModeEnabled returned error: %v", err)
	}
}

func TestIntegrationReadRawSMC(t *testing.T) {
	values, err := GetRawSMCValues([]string{smc.KeyMagsafeLED})
	if err != nil {
		t.Fatalf("GetRawSMCValues returned error: %v", err)
	}
	if values == nil {
		t.Fatal("GetRawSMCValues returned nil map")
	}
}

func TestIntegrationMagsafeStatusContract(t *testing.T) {
	status, err := GetMagsafeStatus()
	if errors.Is(err, ErrNotSupported) {
		if status.Available {
			t.Fatalf("unsupported MagSafe status reported available: %+v", status)
		}
		return
	}
	if err != nil {
		t.Fatalf("GetMagsafeStatus returned error: %v", err)
	}
	if !status.Available {
		t.Fatalf("MagSafe status returned no error but unavailable status: %+v", status)
	}
}

func TestIntegrationLowPowerModeWriteRoundTrip(t *testing.T) {
	if os.Getenv("POWERKIT_WRITE_INTEGRATION") != "1" {
		t.Skip("set POWERKIT_WRITE_INTEGRATION=1 and run as root to test Low Power Mode writes")
	}

	enabled, available, err := GetLowPowerModeEnabled()
	if err != nil {
		t.Fatalf("initial GetLowPowerModeEnabled returned error: %v", err)
	}
	if !available {
		t.Skip("Low Power Mode is not available on this system")
	}

	t.Cleanup(func() {
		if err := SetLowPowerMode(enabled); err != nil {
			t.Logf("failed to restore Low Power Mode state to %v: %v", enabled, err)
		}
	})

	if err := SetLowPowerMode(!enabled); err != nil {
		t.Fatalf("SetLowPowerMode(%v) returned error: %v", !enabled, err)
	}
	got, available, err := GetLowPowerModeEnabled()
	if err != nil {
		t.Fatalf("GetLowPowerModeEnabled after write returned error: %v", err)
	}
	if !available {
		t.Fatal("Low Power Mode became unavailable after write")
	}
	if got != !enabled {
		t.Fatalf("Low Power Mode state = %v, want %v", got, !enabled)
	}
}
