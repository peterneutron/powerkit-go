package powerkit

import (
	"fmt"
	"testing"
)

func TestCalculateIOKitPower(t *testing.T) {
	tests := []struct {
		name            string
		adapterVoltage  float64
		adapterCurrent  float64
		batteryVoltage  float64
		batteryCurrent  float64
		wantAdapter     float64
		wantBattery     float64
		wantSystemPower float64
	}{
		{
			name:            "ChargingBattery",
			adapterVoltage:  20,
			adapterCurrent:  3,
			batteryVoltage:  12,
			batteryCurrent:  2.5,
			wantAdapter:     60,
			wantBattery:     30,
			wantSystemPower: 30,
		},
		{
			name:            "DischargingOnBattery",
			adapterVoltage:  0,
			adapterCurrent:  0,
			batteryVoltage:  12,
			batteryCurrent:  -2.5,
			wantAdapter:     0,
			wantBattery:     -30,
			wantSystemPower: 30,
		},
		{
			name:            "AdapterAndBatteryAssist",
			adapterVoltage:  20,
			adapterCurrent:  1,
			batteryVoltage:  12,
			batteryCurrent:  -1,
			wantAdapter:     20,
			wantBattery:     -12,
			wantSystemPower: 32,
		},
		{
			name:            "LargeChargeClamped",
			adapterVoltage:  20,
			adapterCurrent:  1,
			batteryVoltage:  15,
			batteryCurrent:  2,
			wantAdapter:     20,
			wantBattery:     30,
			wantSystemPower: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := &IOKitData{
				Adapter: IOKitAdapter{
					InputVoltage:  tc.adapterVoltage,
					InputAmperage: tc.adapterCurrent,
				},
				Battery: IOKitBattery{
					Voltage:  tc.batteryVoltage,
					Amperage: tc.batteryCurrent,
				},
			}

			calculateIOKitPower(data)

			fmt.Printf("IOKit/%s -> adapter=%.2fW battery=%.2fW system=%.2fW\n", tc.name, data.Calculations.AdapterPower, data.Calculations.BatteryPower, data.Calculations.SystemPower)

			if data.Calculations.AdapterPower != tc.wantAdapter {
				t.Fatalf("Adapter power mismatch: want %.2f, got %.2f", tc.wantAdapter, data.Calculations.AdapterPower)
			}
			if data.Calculations.BatteryPower != tc.wantBattery {
				t.Fatalf("Battery power mismatch: want %.2f, got %.2f", tc.wantBattery, data.Calculations.BatteryPower)
			}
			if data.Calculations.SystemPower != tc.wantSystemPower {
				t.Fatalf("System power mismatch: want %.2f, got %.2f", tc.wantSystemPower, data.Calculations.SystemPower)
			}
		})
	}
}

func TestCalculateSMCMetrics(t *testing.T) {
	tests := []struct {
		name            string
		adapterVoltage  float64
		adapterCurrent  float64
		batteryVoltage  float64
		batteryCurrent  float64
		wantAdapter     float64
		wantBattery     float64
		wantSystemPower float64
	}{
		{
			name:            "ChargingBattery",
			adapterVoltage:  20,
			adapterCurrent:  3,
			batteryVoltage:  12,
			batteryCurrent:  2.5,
			wantAdapter:     60,
			wantBattery:     30,
			wantSystemPower: 30,
		},
		{
			name:            "DischargingOnBattery",
			adapterVoltage:  0,
			adapterCurrent:  0,
			batteryVoltage:  12,
			batteryCurrent:  -2.5,
			wantAdapter:     0,
			wantBattery:     -30,
			wantSystemPower: 30,
		},
		{
			name:            "AdapterAndBatteryAssist",
			adapterVoltage:  20,
			adapterCurrent:  1,
			batteryVoltage:  12,
			batteryCurrent:  -1,
			wantAdapter:     20,
			wantBattery:     -12,
			wantSystemPower: 32,
		},
		{
			name:            "LargeChargeClamped",
			adapterVoltage:  20,
			adapterCurrent:  1,
			batteryVoltage:  15,
			batteryCurrent:  2,
			wantAdapter:     20,
			wantBattery:     30,
			wantSystemPower: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := &SMCData{
				Adapter: SMCAdapter{
					InputVoltage:  tc.adapterVoltage,
					InputAmperage: tc.adapterCurrent,
				},
				Battery: SMCBattery{
					Voltage:  tc.batteryVoltage,
					Amperage: tc.batteryCurrent,
				},
			}

			calculateSMCMetrics(data)

			fmt.Printf("SMC/%s -> adapter=%.2fW battery=%.2fW system=%.2fW\n", tc.name, data.Calculations.AdapterPower, data.Calculations.BatteryPower, data.Calculations.SystemPower)

			if data.Calculations.AdapterPower != tc.wantAdapter {
				t.Fatalf("Adapter power mismatch: want %.2f, got %.2f", tc.wantAdapter, data.Calculations.AdapterPower)
			}
			if data.Calculations.BatteryPower != tc.wantBattery {
				t.Fatalf("Battery power mismatch: want %.2f, got %.2f", tc.wantBattery, data.Calculations.BatteryPower)
			}
			if data.Calculations.SystemPower != tc.wantSystemPower {
				t.Fatalf("System power mismatch: want %.2f, got %.2f", tc.wantSystemPower, data.Calculations.SystemPower)
			}
		})
	}
}
