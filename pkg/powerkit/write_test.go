package powerkit

import (
	"errors"
	"reflect"
	"testing"

	"github.com/peterneutron/powerkit-go/internal/smc"
)

type smcWriteCall struct {
	key  string
	data []byte
}

func setupSMCControlTest(t *testing.T) *[]smcWriteCall {
	t.Helper()

	oldConfig := currentSMCConfig
	oldFirmwareInfo := currentFirmwareInfo
	oldGeteuid := geteuidFn
	oldFetchSMCRaw := fetchSMCRawData
	oldWriteSMC := writeSMCData

	var writes []smcWriteCall

	t.Cleanup(func() {
		currentSMCConfig = oldConfig
		currentFirmwareInfo = oldFirmwareInfo
		geteuidFn = oldGeteuid
		fetchSMCRawData = oldFetchSMCRaw
		writeSMCData = oldWriteSMC
	})

	currentSMCConfig = smcControlConfig{
		Firmware:               "Test",
		FirmwareProfileID:      profileModernID,
		FirmwareProfileVersion: defaultProfileVersion,
		AdapterKey:             smc.KeyIsAdapterEnabled,
		AdapterEnableBytes:     []byte{0x00},
		AdapterDisableBytes:    []byte{0x08},
		IsLegacyCharging:       false,
		ChargingKeyModern:      smc.KeyIsChargingEnabled,
		ChargingEnableBytes:    []byte{0x00, 0x00, 0x00, 0x00},
		ChargingDisableBytes:   []byte{0x01, 0x00, 0x00, 0x00},
	}
	currentFirmwareInfo.Major = FirmwareMajorVersionThreshold
	geteuidFn = func() int { return 0 }
	fetchSMCRawData = func([]string) (map[string]smc.RawSMCValue, error) {
		return map[string]smc.RawSMCValue{}, nil
	}
	writeSMCData = func(key string, data []byte) error {
		writes = append(writes, smcWriteCall{
			key:  key,
			data: append([]byte(nil), data...),
		})
		return nil
	}

	return &writes
}

func TestGetRawSMCValuesUsesRawFetcher(t *testing.T) {
	oldFetchSMCRaw := fetchSMCRawData
	t.Cleanup(func() { fetchSMCRawData = oldFetchSMCRaw })

	wantKeys := []string{"ACLC", "CH0B"}
	fetchSMCRawData = func(keys []string) (map[string]smc.RawSMCValue, error) {
		if !reflect.DeepEqual(keys, wantKeys) {
			t.Fatalf("fetchSMCRawData keys = %v, want %v", keys, wantKeys)
		}
		return map[string]smc.RawSMCValue{
			"ACLC": {
				DataType: "ui8 ",
				DataSize: 1,
				Data:     []byte{0x04},
			},
		}, nil
	}

	got, err := GetRawSMCValues(wantKeys)
	if err != nil {
		t.Fatalf("GetRawSMCValues returned error: %v", err)
	}
	want := map[string]RawSMCValue{
		"ACLC": {
			DataType: "ui8 ",
			DataSize: 1,
			Data:     []byte{0x04},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetRawSMCValues = %#v, want %#v", got, want)
	}
}

func TestGetRawSMCValuesPropagatesFetchError(t *testing.T) {
	oldFetchSMCRaw := fetchSMCRawData
	t.Cleanup(func() { fetchSMCRawData = oldFetchSMCRaw })

	wantErr := errors.New("smc unavailable")
	fetchSMCRawData = func([]string) (map[string]smc.RawSMCValue, error) {
		return nil, wantErr
	}

	got, err := GetRawSMCValues([]string{"ACLC"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("GetRawSMCValues error = %v, want %v", err, wantErr)
	}
	if got != nil {
		t.Fatalf("GetRawSMCValues = %#v, want nil", got)
	}
}

func TestGetMagsafeLEDStateMapsRawValues(t *testing.T) {
	oldFetchSMCRaw := fetchSMCRawData
	t.Cleanup(func() { fetchSMCRawData = oldFetchSMCRaw })

	tests := []struct {
		name          string
		raw           map[string]smc.RawSMCValue
		wantState     MagsafeLEDState
		wantAvailable bool
	}{
		{
			name: "amber",
			raw: map[string]smc.RawSMCValue{
				smc.KeyMagsafeLED: {DataSize: 1, Data: []byte{byte(LEDAmber)}},
			},
			wantState:     LEDAmber,
			wantAvailable: true,
		},
		{
			name: "green alias",
			raw: map[string]smc.RawSMCValue{
				smc.KeyMagsafeLED: {DataSize: 1, Data: []byte{0x02}},
			},
			wantState:     LEDGreen,
			wantAvailable: true,
		},
		{
			name:          "missing key",
			raw:           map[string]smc.RawSMCValue{},
			wantState:     LEDAmber,
			wantAvailable: false,
		},
		{
			name: "empty data",
			raw: map[string]smc.RawSMCValue{
				smc.KeyMagsafeLED: {DataSize: 0, Data: nil},
			},
			wantState:     LEDAmber,
			wantAvailable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetchSMCRawData = func(keys []string) (map[string]smc.RawSMCValue, error) {
				if !reflect.DeepEqual(keys, []string{smc.KeyMagsafeLED}) {
					t.Fatalf("fetchSMCRawData keys = %v, want [%s]", keys, smc.KeyMagsafeLED)
				}
				return tt.raw, nil
			}

			gotState, gotAvailable, err := GetMagsafeLEDState()
			if err != nil {
				t.Fatalf("GetMagsafeLEDState returned error: %v", err)
			}
			if gotState != tt.wantState || gotAvailable != tt.wantAvailable {
				t.Fatalf("GetMagsafeLEDState = (%v, %v), want (%v, %v)", gotState, gotAvailable, tt.wantState, tt.wantAvailable)
			}
		})
	}
}

func TestGetMagsafeStatusUnavailableReturnsNotSupported(t *testing.T) {
	oldFetchSMCRaw := fetchSMCRawData
	t.Cleanup(func() { fetchSMCRawData = oldFetchSMCRaw })

	fetchSMCRawData = func([]string) (map[string]smc.RawSMCValue, error) {
		return map[string]smc.RawSMCValue{}, nil
	}

	status, err := GetMagsafeStatus()
	if !errors.Is(err, ErrNotSupported) {
		t.Fatalf("GetMagsafeStatus error = %v, want %v", err, ErrNotSupported)
	}
	if status.Available {
		t.Fatalf("GetMagsafeStatus returned available status: %+v", status)
	}
}

func TestSetMagsafeLEDStateWritesSingleByteState(t *testing.T) {
	writes := setupSMCControlTest(t)

	if err := SetMagsafeLEDState(LEDGreen); err != nil {
		t.Fatalf("SetMagsafeLEDState returned error: %v", err)
	}

	want := []smcWriteCall{{key: smc.KeyMagsafeLED, data: []byte{byte(LEDGreen)}}}
	if !reflect.DeepEqual(*writes, want) {
		t.Fatalf("SMC writes = %#v, want %#v", *writes, want)
	}
}

func TestSetAdapterStateToggleReadsCurrentStateAndWritesOpposite(t *testing.T) {
	writes := setupSMCControlTest(t)

	fetchSMCRawData = func(keys []string) (map[string]smc.RawSMCValue, error) {
		if !reflect.DeepEqual(keys, []string{smc.KeyIsAdapterEnabled}) {
			t.Fatalf("fetchSMCRawData keys = %v, want [%s]", keys, smc.KeyIsAdapterEnabled)
		}
		return map[string]smc.RawSMCValue{
			smc.KeyIsAdapterEnabled: {Data: []byte{0x08}},
		}, nil
	}

	if err := SetAdapterState(AdapterActionToggle); err != nil {
		t.Fatalf("SetAdapterState returned error: %v", err)
	}

	want := []smcWriteCall{{key: smc.KeyIsAdapterEnabled, data: []byte{0x00}}}
	if !reflect.DeepEqual(*writes, want) {
		t.Fatalf("SMC writes = %#v, want %#v", *writes, want)
	}
}

func TestSetChargingStateWritesModernAndLegacyKeys(t *testing.T) {
	t.Run("modern", func(t *testing.T) {
		writes := setupSMCControlTest(t)

		if err := SetChargingState(ChargingActionOff); err != nil {
			t.Fatalf("SetChargingState returned error: %v", err)
		}

		want := []smcWriteCall{{key: smc.KeyIsChargingEnabled, data: []byte{0x01, 0x00, 0x00, 0x00}}}
		if !reflect.DeepEqual(*writes, want) {
			t.Fatalf("SMC writes = %#v, want %#v", *writes, want)
		}
	})

	t.Run("legacy", func(t *testing.T) {
		writes := setupSMCControlTest(t)
		currentSMCConfig.IsLegacyCharging = true
		currentSMCConfig.ChargingKeysLegacy = []string{
			smc.KeyIsChargingEnabledLegacyBCLM,
			smc.KeyIsChargingEnabledLegacyBCDS,
		}
		currentSMCConfig.ChargingEnableBytes = []byte{0x00}

		if err := SetChargingState(ChargingActionOn); err != nil {
			t.Fatalf("SetChargingState returned error: %v", err)
		}

		want := []smcWriteCall{
			{key: smc.KeyIsChargingEnabledLegacyBCLM, data: []byte{0x00}},
			{key: smc.KeyIsChargingEnabledLegacyBCDS, data: []byte{0x00}},
		}
		if !reflect.DeepEqual(*writes, want) {
			t.Fatalf("SMC writes = %#v, want %#v", *writes, want)
		}
	})
}

func TestWriteAPIsRequireDetectedSMCProfileBeforeWriting(t *testing.T) {
	writes := setupSMCControlTest(t)
	currentFirmwareInfo.Major = 0

	err := SetMagsafeLEDState(LEDAmber)
	if !errors.Is(err, ErrNotSupported) {
		t.Fatalf("SetMagsafeLEDState error = %v, want %v", err, ErrNotSupported)
	}
	if len(*writes) != 0 {
		t.Fatalf("unexpected SMC writes: %#v", *writes)
	}
}

func TestWriteAPIsRequireRootBeforeWriting(t *testing.T) {
	writes := setupSMCControlTest(t)
	geteuidFn = func() int { return 501 }

	err := SetAdapterState(AdapterActionOn)
	if !errors.Is(err, ErrPermissionRequired) {
		t.Fatalf("SetAdapterState error = %v, want %v", err, ErrPermissionRequired)
	}
	if len(*writes) != 0 {
		t.Fatalf("unexpected SMC writes: %#v", *writes)
	}
}
