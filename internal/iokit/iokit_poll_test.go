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

	data := &RawData{SourceVoltage: 0, SourceAmperage: 0, IsConnected: true}
	applyAdapterTelemetryFallback(data, AdapterTelemetryReasonMissingIOKit)

	if calls != 1 {
		t.Fatalf("expected fallback to call SMC once, got %d", calls)
	}
	if data.SourceVoltage != 19500 || data.SourceAmperage != 3250 {
		t.Fatalf("expected source values 19500/3250, got %d/%d", data.SourceVoltage, data.SourceAmperage)
	}
	if data.TelemetrySource != AdapterTelemetrySourceSMCFallback || data.TelemetryReason != AdapterTelemetryReasonMissingIOKit || !data.TelemetryAvailable {
		t.Fatalf("unexpected telemetry metadata: source=%s reason=%s available=%v", data.TelemetrySource, data.TelemetryReason, data.TelemetryAvailable)
	}
}

func TestEvaluateAdapterTelemetryNoOpWhenAvailable(t *testing.T) {
	oldSMCFetch := smcFetchDataFn
	t.Cleanup(func() { smcFetchDataFn = oldSMCFetch })

	smcFetchDataFn = func(_ []string) (map[string]float64, error) {
		t.Fatalf("unexpected fallback call when telemetry is available")
		return nil, nil
	}

	data := &RawData{SourceVoltage: 20000, SourceAmperage: 3000, IsConnected: true}
	evaluateAdapterTelemetry(data, true, false)

	if data.SourceVoltage != 20000 || data.SourceAmperage != 3000 {
		t.Fatalf("expected values to remain unchanged when telemetry available")
	}
	if data.TelemetrySource != AdapterTelemetrySourceIOKit || data.TelemetryReason != AdapterTelemetryReasonNone || !data.TelemetryAvailable {
		t.Fatalf("unexpected telemetry metadata: source=%s reason=%s available=%v", data.TelemetrySource, data.TelemetryReason, data.TelemetryAvailable)
	}
}

func TestEvaluateAdapterTelemetrySkipsWhenDisconnected(t *testing.T) {
	oldSMCFetch := smcFetchDataFn
	t.Cleanup(func() { smcFetchDataFn = oldSMCFetch })

	smcFetchDataFn = func(_ []string) (map[string]float64, error) {
		t.Fatalf("unexpected fallback call when adapter is disconnected")
		return nil, nil
	}

	data := &RawData{SourceVoltage: 0, SourceAmperage: 0, IsConnected: false}
	evaluateAdapterTelemetry(data, false, false)

	if data.TelemetrySource != AdapterTelemetrySourceUnavailable || data.TelemetryReason != AdapterTelemetryReasonNoAdapter || data.TelemetryAvailable {
		t.Fatalf("unexpected telemetry metadata: source=%s reason=%s available=%v", data.TelemetrySource, data.TelemetryReason, data.TelemetryAvailable)
	}
}

func TestEvaluateAdapterTelemetryUsesFallbackForInvalidInput(t *testing.T) {
	oldSMCFetch := smcFetchDataFn
	t.Cleanup(func() { smcFetchDataFn = oldSMCFetch })

	smcFetchDataFn = func(_ []string) (map[string]float64, error) {
		return map[string]float64{
			smc.KeyAdapterVoltage: 20.2,
			smc.KeyAdapterCurrent: 3.1,
		}, nil
	}

	data := &RawData{SourceVoltage: 0, SourceAmperage: 2500, IsConnected: true}
	evaluateAdapterTelemetry(data, true, false)

	if data.TelemetrySource != AdapterTelemetrySourceSMCFallback || data.TelemetryReason != AdapterTelemetryReasonInvalidIOKit || !data.TelemetryAvailable {
		t.Fatalf("unexpected telemetry metadata: source=%s reason=%s available=%v", data.TelemetrySource, data.TelemetryReason, data.TelemetryAvailable)
	}
	if data.SourceVoltage != 20200 || data.SourceAmperage != 3100 {
		t.Fatalf("expected fallback values 20200/3100, got %d/%d", data.SourceVoltage, data.SourceAmperage)
	}
}
