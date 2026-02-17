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
