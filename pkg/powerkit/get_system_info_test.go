package powerkit

import (
	"errors"
	"testing"

	"github.com/peterneutron/powerkit-go/internal/iokit"
	"github.com/peterneutron/powerkit-go/internal/powerd"
	"github.com/peterneutron/powerkit-go/internal/smc"
)

func setupSystemInfoFixture(t *testing.T) (*SystemInfo, bool) {
	t.Helper()

	oldFetchIOKit := fetchIOKitData
	oldFetchSMCFloat := fetchSMCFloatData
	oldFetchSMCRaw := fetchSMCRawData
	oldGlobalSleep := globalSleepStatusFn
	oldPowerdActive := powerdIsActiveFn
	oldGetLPM := getLowPowerModeFn
	oldConfig := currentSMCConfig

	t.Cleanup(func() {
		fetchIOKitData = oldFetchIOKit
		fetchSMCFloatData = oldFetchSMCFloat
		fetchSMCRawData = oldFetchSMCRaw
		globalSleepStatusFn = oldGlobalSleep
		powerdIsActiveFn = oldPowerdActive
		getLowPowerModeFn = oldGetLPM
		currentSMCConfig = oldConfig
	})

	currentSMCConfig = smcControlConfig{
		Firmware:             "Test",
		AdapterKey:           smc.KeyIsAdapterEnabled,
		AdapterEnableBytes:   []byte{0x00},
		AdapterDisableBytes:  []byte{0x08},
		IsLegacyCharging:     false,
		ChargingKeyModern:    smc.KeyIsChargingEnabled,
		ChargingEnableBytes:  []byte{0x00, 0x00, 0x00, 0x00},
		ChargingDisableBytes: []byte{0x01, 0x00, 0x00, 0x00},
	}

	var observedForce bool
	fetchIOKitData = func(force bool) (*iokit.RawData, error) {
		observedForce = force
		return &iokit.RawData{
			IsCharging:         true,
			IsConnected:        true,
			AdapterWatts:       70,
			AdapterVoltage:     20000,
			AdapterAmperage:    3500,
			AdapterDesc:        "USB-C 70W",
			SourceVoltage:      20000,
			SourceAmperage:     3000,
			Voltage:            12000,
			Amperage:           2000,
			Temperature:        3000,
			CycleCount:         500,
			DesignCapacity:     8000,
			MaxCapacity:        7500,
			NominalCapacity:    7200,
			CurrentCharge:      80,
			CurrentChargeRaw:   800,
			CurrentCapacityRaw: 6000,
			TimeToEmpty:        90,
			TimeToFull:         45,
			SerialNumber:       "TEST123",
			DeviceName:         "TestPack",
			TelemetryAvailable: false,
			CellVoltages:       []int{4000, 4005, 4010},
		}, nil
	}

	fetchSMCFloatData = func(_ []string) (map[string]float64, error) {
		return map[string]float64{
			smc.KeyAdapterVoltage: 19.8,
			smc.KeyAdapterCurrent: 2.5,
			smc.KeyBatteryVoltage: 12000,
			smc.KeyBatteryCurrent: 2000,
		}, nil
	}

	fetchSMCRawData = func(_ []string) (map[string]smc.RawSMCValue, error) {
		return map[string]smc.RawSMCValue{
			smc.KeyIsChargingEnabled: {
				Data: []byte{0x00, 0x00, 0x00, 0x00},
			},
			smc.KeyIsAdapterEnabled: {
				Data: []byte{0x00},
			},
		}, nil
	}

	globalSleepStatusFn = func() (bool, bool, error) {
		return false, false, errors.New("status unavailable")
	}

	powerdIsActiveFn = func(a powerd.AssertionType) bool {
		return a == powerd.PreventDisplaySleep
	}

	getLowPowerModeFn = func() (bool, bool, error) {
		return true, true, nil
	}

	info, err := GetSystemInfo(FetchOptions{QueryIOKit: true, QuerySMC: true, ForceTelemetryFallback: true})
	if err != nil {
		t.Fatalf("GetSystemInfo returned error: %v", err)
	}

	return info, observedForce
}

func TestGetSystemInfoPropagatesForceTelemetryFallback(t *testing.T) {
	_, observed := setupSystemInfoFixture(t)
	if !observed {
		t.Fatalf("expected ForceTelemetryFallback to be propagated")
	}
}

func TestGetSystemInfoPopulatesSources(t *testing.T) {
	info, _ := setupSystemInfoFixture(t)
	if info.IOKit == nil || info.SMC == nil {
		t.Fatalf("expected both IOKit and SMC data, got %#v", info)
	}
	if info.IOKit.Adapter.TelemetryAvailable {
		t.Fatalf("expected telemetry to be marked unavailable when fallback forced")
	}
}

func TestGetSystemInfoCalculatesPower(t *testing.T) {
	info, _ := setupSystemInfoFixture(t)
	if got := info.IOKit.Calculations.SystemPower; got != 36 {
		t.Fatalf("expected IOKit system power 36, got %.2f", got)
	}
	expectedSMC := 19.8*2.5 - 12*2.0
	if got := info.SMC.Calculations.SystemPower; got != expectedSMC {
		t.Fatalf("unexpected SMC system power: %.2f", got)
	}
}

func TestGetSystemInfoOSFallbackBehavior(t *testing.T) {
	info, _ := setupSystemInfoFixture(t)
	if !info.OS.LowPowerMode.Enabled || !info.OS.LowPowerMode.Available {
		t.Fatalf("expected low power mode enabled and available")
	}
	if info.OS.GlobalDisplaySleepAllowed != info.OS.AppDisplaySleepAllowed {
		t.Fatalf("expected global display allowance to mirror app state on fallback")
	}
	if info.OS.AppDisplaySleepAllowed {
		t.Fatalf("expected app display sleep disallowed when PreventDisplaySleep active")
	}
	if info.OS.AppSystemSleepAllowed {
		t.Fatalf("expected app system sleep disallowed when display assertion active")
	}
}

func TestGetSystemInfoRequiresSource(t *testing.T) {
	if _, err := GetSystemInfo(FetchOptions{QueryIOKit: false, QuerySMC: false}); err == nil {
		t.Fatalf("expected error when no sources requested")
	}
}
