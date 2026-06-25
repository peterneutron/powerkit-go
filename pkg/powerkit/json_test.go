package powerkit

import (
	"encoding/json"
	"testing"
)

func TestToJSONIncludesSchemaAndSources(t *testing.T) {
	info, _ := setupSystemInfoFixture(t)
	j := info.ToJSON()

	assertBaseJSONFields(t, &j)
	assertSourceFields(t, &j)
	assertHealthAndFirmwareFields(t, &j)
	assertHardwareChargeFields(t, &j)
	assertSMCControlFields(t, &j)
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

	assertTopLevelSnakeCaseKeys(t, decoded)
	osPayload := assertOSPayload(t, decoded)
	assertOSFirmwareJSONKeys(t, osPayload)
	batteryPayload := assertBatteryPayload(t, decoded)
	assertBatteryCapacityJSONKeys(t, batteryPayload)
}

func assertBaseJSONFields(t *testing.T, j *SystemInfoJSON) {
	t.Helper()
	if j.SchemaVersion == "" {
		t.Fatalf("expected schema_version")
	}
	if j.CollectedAt == "" {
		t.Fatalf("expected collected_at")
	}
}

func assertSourceFields(t *testing.T, j *SystemInfoJSON) {
	t.Helper()
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

func assertHealthAndFirmwareFields(t *testing.T, j *SystemInfoJSON) {
	t.Helper()
	if j.Battery.Health.BalanceState == "" {
		t.Fatalf("expected balance_state to be set")
	}
	if j.OS.FirmwareVersion == "" {
		t.Fatalf("expected firmware_version to be set")
	}
	if j.OS.FirmwareSource == "" {
		t.Fatalf("expected firmware_source to be set")
	}
	if j.OS.FirmwareMajor <= 0 {
		t.Fatalf("expected firmware_major to be set")
	}
	if j.OS.FirmwareCompatStatus == "" {
		t.Fatalf("expected firmware_compat_status to be set")
	}
	if j.OS.FirmwareProfileID == "" {
		t.Fatalf("expected firmware_profile_id to be set")
	}
	if j.OS.FirmwareProfileVersion <= 0 {
		t.Fatalf("expected firmware_profile_version to be set")
	}
}

func assertHardwareChargeFields(t *testing.T, j *SystemInfoJSON) {
	t.Helper()
	if !j.Battery.Capacity.HardwarePercentAvailable {
		t.Fatalf("expected hardware_percent_available")
	}
	if j.Battery.Capacity.HardwarePercent != 79 {
		t.Fatalf("hardware_percent = %d, want 79", j.Battery.Capacity.HardwarePercent)
	}
	if j.Battery.Capacity.HardwarePercentPrecise != 80 {
		t.Fatalf("hardware_percent_precise = %.2f, want 80.00", j.Battery.Capacity.HardwarePercentPrecise)
	}
	if j.Battery.Capacity.CurrentRaw != 79 {
		t.Fatalf("current_raw = %d, want legacy raw StateOfCharge 79", j.Battery.Capacity.CurrentRaw)
	}
}

func assertSMCControlFields(t *testing.T, j *SystemInfoJSON) {
	t.Helper()
	if !j.Controls.SMC.ChargingControlAvailable {
		t.Fatalf("expected charging_control_available")
	}
	if !j.Controls.SMC.AdapterControlAvailable {
		t.Fatalf("expected adapter_control_available")
	}
	if !j.Controls.SMC.ChargingEnabled {
		t.Fatalf("expected charging_enabled")
	}
	if !j.Controls.SMC.AdapterEnabled {
		t.Fatalf("expected adapter_enabled")
	}
}

func assertTopLevelSnakeCaseKeys(t *testing.T, decoded map[string]any) {
	t.Helper()
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

func assertOSPayload(t *testing.T, decoded map[string]any) map[string]any {
	t.Helper()
	osPayload, ok := decoded["os"].(map[string]any)
	if !ok {
		t.Fatalf("expected os payload to be object")
	}
	return osPayload
}

func assertBatteryPayload(t *testing.T, decoded map[string]any) map[string]any {
	t.Helper()
	batteryPayload, ok := decoded["battery"].(map[string]any)
	if !ok {
		t.Fatalf("expected battery payload to be object")
	}
	return batteryPayload
}

func assertBatteryCapacityJSONKeys(t *testing.T, batteryPayload map[string]any) {
	t.Helper()
	capacityPayload, ok := batteryPayload["capacity"].(map[string]any)
	if !ok {
		t.Fatalf("expected battery.capacity payload to be object")
	}
	required := []string{
		"current_percent",
		"current_raw",
		"hardware_percent",
		"hardware_percent_precise",
		"hardware_percent_available",
	}
	for _, key := range required {
		if _, ok := capacityPayload[key]; !ok {
			t.Fatalf("expected battery.capacity.%s key", key)
		}
	}
}

func assertOSFirmwareJSONKeys(t *testing.T, osPayload map[string]any) {
	t.Helper()
	required := []string{
		"firmware_version",
		"firmware_source",
		"firmware_major",
		"firmware_compat_status",
		"firmware_profile_id",
		"firmware_profile_version",
	}
	for _, key := range required {
		if _, ok := osPayload[key]; !ok {
			t.Fatalf("expected os.%s key", key)
		}
	}
}
