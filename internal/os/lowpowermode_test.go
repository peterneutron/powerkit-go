package os

import (
	"context"
	"errors"
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
