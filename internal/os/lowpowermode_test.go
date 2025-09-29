package os

import (
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
	oldRun := pmsetRunFn
	t.Cleanup(func() { pmsetRunFn = oldRun })

	resetLPMCache()

	pmsetRunFn = func(_ ...string) ([]byte, error) {
		return []byte("Battery Status\n lowpowermode = 1\n"), nil
	}

	enabled, available, err := GetLowPowerModeEnabled()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !available || !enabled {
		t.Fatalf("expected low power mode to be available and enabled")
	}

	resetLPMCache()
	pmsetRunFn = func(_ ...string) ([]byte, error) {
		return []byte("some other key\n"), nil
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
	oldRun := pmsetRunFn
	origTTL := lpmTTL
	t.Cleanup(func() {
		pmsetRunFn = oldRun
		lpmTTL = origTTL
	})

	resetLPMCache()

	var runCalls int32
	pmsetRunFn = func(_ ...string) ([]byte, error) {
		atomic.AddInt32(&runCalls, 1)
		return []byte(" lowpowermode\t0\n"), nil
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
	oldRun := pmsetRunFn
	origTTL := lpmTTL
	t.Cleanup(func() {
		pmsetRunFn = oldRun
		lpmTTL = origTTL
	})

	resetLPMCache()

	var runCalls int32
	pmsetRunFn = func(_ ...string) ([]byte, error) {
		call := atomic.AddInt32(&runCalls, 1)
		if call == 1 {
			return []byte(" lowpowermode\t0\n"), nil
		}
		return []byte(" lowpowermode\t1\n"), nil
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
