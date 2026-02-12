package powerkit

import (
	"encoding/json"
	"testing"
)

func TestToJSONIncludesSchemaAndSources(t *testing.T) {
	info, _ := setupSystemInfoFixture(t)
	j := info.ToJSON()

	if j.SchemaVersion == "" {
		t.Fatalf("expected schema_version")
	}
	if j.CollectedAt == "" {
		t.Fatalf("expected collected_at")
	}
	if !j.Sources.IOKit.Queried || !j.Sources.IOKit.Available {
		t.Fatalf("unexpected iokit source status: %+v", j.Sources.IOKit)
	}
	if !j.Sources.SMC.Queried || !j.Sources.SMC.Available {
		t.Fatalf("unexpected smc source status: %+v", j.Sources.SMC)
	}
	if j.Sources.AdapterTelemetry.Source == "" || j.Sources.AdapterTelemetry.Reason == "" {
		t.Fatalf("expected adapter telemetry provenance, got %+v", j.Sources.AdapterTelemetry)
	}
}

func TestToJSONUsesSnakeCaseKeys(t *testing.T) {
	info, _ := setupSystemInfoFixture(t)
	payload, err := json.Marshal(info.ToJSON())
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	required := []string{"schema_version", "collected_at", "os", "battery", "adapter", "power", "controls", "sources"}
	for _, key := range required {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("expected top-level key %q in payload", key)
		}
	}
	if _, ok := decoded["OS"]; ok {
		t.Fatalf("unexpected legacy PascalCase key found")
	}
}
