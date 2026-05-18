//go:build darwin

package powerkit

import (
	"context"

	sysos "github.com/peterneutron/powerkit-go/internal/os"
)

func checkContext(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// GetSystemInfoContext is the context-aware variant of GetSystemInfo. The
// context is checked before and after the underlying IOKit/SMC calls; those
// cgo calls are not interrupted once started.
func GetSystemInfoContext(ctx context.Context, opts ...FetchOptions) (*SystemInfo, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	info, err := GetSystemInfo(opts...)
	if err != nil {
		return nil, err
	}
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	return info, nil
}

// SetAdapterStateContext is the context-aware variant of SetAdapterState.
// Context cancellation is checked before the SMC write path starts.
func SetAdapterStateContext(ctx context.Context, action AdapterAction) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	return SetAdapterState(action)
}

// SetChargingStateContext is the context-aware variant of SetChargingState.
// Context cancellation is checked before the SMC write path starts.
func SetChargingStateContext(ctx context.Context, action ChargingAction) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	return SetChargingState(action)
}

// SetMagsafeLEDStateContext is the context-aware variant of SetMagsafeLEDState.
// Context cancellation is checked before the SMC write path starts.
func SetMagsafeLEDStateContext(ctx context.Context, state MagsafeLEDState) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	return SetMagsafeLEDState(state)
}

// SetLowPowerModeContext is the context-aware variant of SetLowPowerMode.
func SetLowPowerModeContext(ctx context.Context, enable bool) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	if err := requireRoot("set low power mode"); err != nil {
		return err
	}
	return sysos.SetLowPowerModeContext(ctx, enable)
}

// ToggleLowPowerModeContext is the context-aware variant of ToggleLowPowerMode.
func ToggleLowPowerModeContext(ctx context.Context) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	if err := requireRoot("toggle low power mode"); err != nil {
		return err
	}
	return sysos.ToggleLowPowerModeContext(ctx)
}
