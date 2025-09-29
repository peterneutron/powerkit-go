package iokit

import (
	"testing"

	"github.com/peterneutron/powerkit-go/internal/smc"
)

func TestApplyAdapterTelemetryFallbackUsesSMC(t *testing.T) {
	oldSMCFetch := smcFetchDataFn
	t.Cleanup(func() { smcFetchDataFn = oldSMCFetch })

	calls := 0
	smcFetchDataFn = func(_ []string) (map[string]float64, error) {
		calls++
		return map[string]float64{
			smc.KeyAdapterVoltage: 19.5,
			smc.KeyAdapterCurrent: 3.25,
		}, nil
	}

	data := &RawData{SourceVoltage: 0, SourceAmperage: 0}
	applyAdapterTelemetryFallback(data, false)

	if calls != 1 {
		t.Fatalf("expected fallback to call SMC once, got %d", calls)
	}
	if data.SourceVoltage != 19500 || data.SourceAmperage != 3250 {
		t.Fatalf("expected source values 19500/3250, got %d/%d", data.SourceVoltage, data.SourceAmperage)
	}
}

func TestApplyAdapterTelemetryFallbackNoOpWhenAvailable(t *testing.T) {
	oldSMCFetch := smcFetchDataFn
	t.Cleanup(func() { smcFetchDataFn = oldSMCFetch })

	smcFetchDataFn = func(_ []string) (map[string]float64, error) {
		t.Fatalf("unexpected fallback call when telemetry is available")
		return nil, nil
	}

	data := &RawData{SourceVoltage: 123, SourceAmperage: 456}
	applyAdapterTelemetryFallback(data, true)

	if data.SourceVoltage != 123 || data.SourceAmperage != 456 {
		t.Fatalf("expected values to remain unchanged when telemetry available")
	}
}
