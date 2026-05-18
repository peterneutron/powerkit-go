//go:build darwin

package powerkit

import (
	"context"
	"errors"
	"os"
	"testing"
)

func TestSetAdapterStateContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := SetAdapterStateContext(ctx, AdapterActionOn)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

func TestSetLowPowerModeContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := SetLowPowerModeContext(ctx, true)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

func TestRequireRootErrorType(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("test expects non-root execution")
	}

	err := SetChargingState(ChargingActionOn)
	if !errors.Is(err, ErrPermissionRequired) {
		t.Fatalf("expected ErrPermissionRequired, got: %v", err)
	}
}

func TestSMCWriteRequiresDetectedControlProfile(t *testing.T) {
	oldFirmwareInfo := currentFirmwareInfo
	t.Cleanup(func() {
		currentFirmwareInfo = oldFirmwareInfo
	})

	currentFirmwareInfo.Major = 0

	for name, err := range map[string]error{
		"adapter":  SetAdapterState(AdapterActionOn),
		"charging": SetChargingState(ChargingActionOn),
		"magsafe":  SetMagsafeLEDState(LEDSystem),
	} {
		if !errors.Is(err, ErrNotSupported) {
			t.Fatalf("%s write expected ErrNotSupported, got %v", name, err)
		}
	}
}
