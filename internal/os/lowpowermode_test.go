package os

import (
	"context"
	"errors"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func resetLPMCache() {
	lpmMu.Lock()
	lpmValid = false
	lpmValue = false
	lpmCachedAt = time.Time{}
	lpmMu.Unlock()
}

func TestGetLowPowerModeEnabledParsing(t *testing.T) {
	oldRead := lowPowerModeReadFn
	t.Cleanup(func() { lowPowerModeReadFn = oldRead })

	resetLPMCache()

	lowPowerModeReadFn = func() (bool, bool, error) {
		return true, true, nil
	}

	enabled, available, err := GetLowPowerModeEnabled()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !available || !enabled {
		t.Fatalf("expected low power mode to be available and enabled")
	}

	resetLPMCache()
	lowPowerModeReadFn = func() (bool, bool, error) {
		return false, false, nil
	}

	enabled, available, err = GetLowPowerModeEnabled()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if available || enabled {
		t.Fatalf("expected unavailable low power mode to report false/false")
	}
}

func TestLowPowerModeCacheUsesCachedValue(t *testing.T) {
	oldRead := lowPowerModeReadFn
	origTTL := lpmTTL
	t.Cleanup(func() {
		lowPowerModeReadFn = oldRead
		lpmTTL = origTTL
	})

	resetLPMCache()

	var runCalls int32
	lowPowerModeReadFn = func() (bool, bool, error) {
		atomic.AddInt32(&runCalls, 1)
		return false, true, nil
	}

	lpmTTL = time.Minute

	enabled, _, err := GetLowPowerModeEnabled()
	if err != nil {
		t.Fatalf("unexpected error on first read: %v", err)
	}
	if enabled {
		t.Fatalf("expected first call to report disabled")
	}
	if got := atomic.LoadInt32(&runCalls); got != 1 {
		t.Fatalf("expected pmset to run once, got %d", got)
	}

	enabled, _, err = GetLowPowerModeEnabled()
	if err != nil {
		t.Fatalf("unexpected error on cached read: %v", err)
	}
	if enabled {
		t.Fatalf("expected cached value to remain disabled")
	}
	if got := atomic.LoadInt32(&runCalls); got != 1 {
		t.Fatalf("expected cached read to avoid pmset run, got %d", got)
	}
}

func TestLowPowerModeCacheExpires(t *testing.T) {
	oldRead := lowPowerModeReadFn
	origTTL := lpmTTL
	t.Cleanup(func() {
		lowPowerModeReadFn = oldRead
		lpmTTL = origTTL
	})

	resetLPMCache()

	var runCalls int32
	lowPowerModeReadFn = func() (bool, bool, error) {
		call := atomic.AddInt32(&runCalls, 1)
		if call == 1 {
			return false, true, nil
		}
		return true, true, nil
	}

	lpmTTL = time.Millisecond

	enabled, _, err := GetLowPowerModeEnabled()
	if err != nil {
		t.Fatalf("unexpected error on prime read: %v", err)
	}
	if enabled {
		t.Fatalf("expected initial read to be disabled")
	}

	lpmMu.Lock()
	lpmCachedAt = time.Now().Add(-lpmTTL - time.Millisecond)
	lpmMu.Unlock()

	enabled, _, err = GetLowPowerModeEnabled()
	if err != nil {
		t.Fatalf("unexpected error after ttl expiry: %v", err)
	}
	if !enabled {
		t.Fatalf("expected refreshed read to report enabled")
	}
	if got := atomic.LoadInt32(&runCalls); got != 2 {
		t.Fatalf("expected pmset to run twice, got %d", got)
	}
}

func TestSetLowPowerModeContextCanceledBeforePmset(t *testing.T) {
	oldExecContext := pmsetExecContextFn
	t.Cleanup(func() { pmsetExecContextFn = oldExecContext })

	var calls int32
	pmsetExecContextFn = func(context.Context, ...string) error {
		atomic.AddInt32(&calls, 1)
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := SetLowPowerModeContext(ctx, true)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 0 {
		t.Fatalf("expected canceled context to skip pmset, got %d calls", got)
	}
}

func TestSetLowPowerModeContextRunsPmsetAndInvalidatesCache(t *testing.T) {
	oldExecContext := pmsetExecContextFn
	t.Cleanup(func() { pmsetExecContextFn = oldExecContext })

	tests := []struct {
		name   string
		enable bool
		want   []string
	}{
		{
			name:   "enable",
			enable: true,
			want:   []string{"-a", "lowpowermode", "1"},
		},
		{
			name:   "disable",
			enable: false,
			want:   []string{"-a", "lowpowermode", "0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetLPMCache()
			lpmMu.Lock()
			lpmValue = !tt.enable
			lpmCachedAt = time.Now()
			lpmValid = true
			lpmMu.Unlock()

			var gotArgs []string
			pmsetExecContextFn = func(ctx context.Context, args ...string) error {
				if err := ctx.Err(); err != nil {
					return err
				}
				gotArgs = append([]string(nil), args...)
				return nil
			}

			if err := SetLowPowerModeContext(context.Background(), tt.enable); err != nil {
				t.Fatalf("SetLowPowerModeContext returned error: %v", err)
			}
			if !reflect.DeepEqual(gotArgs, tt.want) {
				t.Fatalf("pmset args = %v, want %v", gotArgs, tt.want)
			}

			lpmMu.Lock()
			valid := lpmValid
			lpmMu.Unlock()
			if valid {
				t.Fatal("expected successful write to invalidate low power mode cache")
			}
		})
	}
}

func TestSetLowPowerModeContextKeepsCacheOnPmsetError(t *testing.T) {
	oldExecContext := pmsetExecContextFn
	t.Cleanup(func() { pmsetExecContextFn = oldExecContext })

	resetLPMCache()
	lpmMu.Lock()
	lpmValue = true
	lpmCachedAt = time.Now()
	lpmValid = true
	lpmMu.Unlock()

	wantErr := errors.New("pmset failed")
	pmsetExecContextFn = func(context.Context, ...string) error {
		return wantErr
	}

	err := SetLowPowerModeContext(context.Background(), false)
	if !errors.Is(err, wantErr) {
		t.Fatalf("SetLowPowerModeContext error = %v, want %v", err, wantErr)
	}

	lpmMu.Lock()
	valid := lpmValid
	lpmMu.Unlock()
	if !valid {
		t.Fatal("expected failed write to leave low power mode cache untouched")
	}
}

func TestToggleLowPowerModeContextWritesOppositeState(t *testing.T) {
	oldRead := lowPowerModeReadFn
	oldExecContext := pmsetExecContextFn
	t.Cleanup(func() {
		lowPowerModeReadFn = oldRead
		pmsetExecContextFn = oldExecContext
	})

	tests := []struct {
		name          string
		current       bool
		wantPmsetArgs []string
	}{
		{
			name:          "enabled to disabled",
			current:       true,
			wantPmsetArgs: []string{"-a", "lowpowermode", "0"},
		},
		{
			name:          "disabled to enabled",
			current:       false,
			wantPmsetArgs: []string{"-a", "lowpowermode", "1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetLPMCache()

			lowPowerModeReadFn = func() (bool, bool, error) {
				return tt.current, true, nil
			}

			var gotArgs []string
			pmsetExecContextFn = func(ctx context.Context, args ...string) error {
				if err := ctx.Err(); err != nil {
					return err
				}
				gotArgs = append([]string(nil), args...)
				return nil
			}

			if err := ToggleLowPowerModeContext(context.Background()); err != nil {
				t.Fatalf("ToggleLowPowerModeContext returned error: %v", err)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantPmsetArgs) {
				t.Fatalf("pmset args = %v, want %v", gotArgs, tt.wantPmsetArgs)
			}
		})
	}
}

func TestToggleLowPowerModeContextUnavailableSkipsPmset(t *testing.T) {
	oldRead := lowPowerModeReadFn
	oldExecContext := pmsetExecContextFn
	t.Cleanup(func() {
		lowPowerModeReadFn = oldRead
		pmsetExecContextFn = oldExecContext
	})

	resetLPMCache()

	lowPowerModeReadFn = func() (bool, bool, error) {
		return false, false, nil
	}

	var calls int32
	pmsetExecContextFn = func(context.Context, ...string) error {
		atomic.AddInt32(&calls, 1)
		return nil
	}

	err := ToggleLowPowerModeContext(context.Background())
	if err == nil {
		t.Fatal("expected unavailable low power mode to return error")
	}
	if got := atomic.LoadInt32(&calls); got != 0 {
		t.Fatalf("expected unavailable low power mode to skip pmset, got %d calls", got)
	}
}

func TestToggleLowPowerModeContextReadErrorSkipsPmset(t *testing.T) {
	oldRead := lowPowerModeReadFn
	oldExecContext := pmsetExecContextFn
	t.Cleanup(func() {
		lowPowerModeReadFn = oldRead
		pmsetExecContextFn = oldExecContext
	})

	resetLPMCache()

	wantErr := errors.New("read failed")
	lowPowerModeReadFn = func() (bool, bool, error) {
		return false, false, wantErr
	}

	var calls int32
	pmsetExecContextFn = func(context.Context, ...string) error {
		atomic.AddInt32(&calls, 1)
		return nil
	}

	err := ToggleLowPowerModeContext(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("ToggleLowPowerModeContext error = %v, want %v", err, wantErr)
	}
	if got := atomic.LoadInt32(&calls); got != 0 {
		t.Fatalf("expected read error to skip pmset, got %d calls", got)
	}
}

func TestToggleLowPowerModeContextCanceledBeforeRead(t *testing.T) {
	oldRead := lowPowerModeReadFn
	oldExecContext := pmsetExecContextFn
	t.Cleanup(func() {
		lowPowerModeReadFn = oldRead
		pmsetExecContextFn = oldExecContext
	})

	var readCalls int32
	lowPowerModeReadFn = func() (bool, bool, error) {
		atomic.AddInt32(&readCalls, 1)
		return false, true, nil
	}
	pmsetExecContextFn = func(context.Context, ...string) error {
		t.Fatal("pmset must not run when context is already canceled")
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := ToggleLowPowerModeContext(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ToggleLowPowerModeContext error = %v, want %v", err, context.Canceled)
	}
	if got := atomic.LoadInt32(&readCalls); got != 0 {
		t.Fatalf("expected canceled context to skip read, got %d calls", got)
	}
}
