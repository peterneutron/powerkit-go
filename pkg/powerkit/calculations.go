package power

import "math"

// calculateDerivedMetrics populates the Calculations struct with health
// percentages and source-specific power flow data in Watts.
func calculateDerivedMetrics(info *BatteryInfo) {
	// --- Health Percentage Calculations ---
	if info.Battery.DesignCapacity > 0 {
		designCapF := float64(info.Battery.DesignCapacity)

		healthByMax := (float64(info.Battery.MaxCapacity) / designCapF) * 100.0
		info.Calculations.HealthByMaxCapacity = int(math.Round(healthByMax))

		healthByNominal := (float64(info.Battery.NominalCapacity) / designCapF) * 100.0
		info.Calculations.HealthByNominalCapacity = int(math.Round(healthByNominal))

		var conditionModifier float64
		if len(info.Battery.IndividualCellVoltages) > 1 {
			minV, maxV := findMinMax(info.Battery.IndividualCellVoltages)
			drift := maxV - minV
			switch {
			case drift <= 5:
				conditionModifier = 2.5
			case drift <= 15:
				conditionModifier = 1.0
			case drift <= 30:
				conditionModifier = 0.0
			case drift <= 50:
				conditionModifier = -2.0
			default:
				conditionModifier = -10.0
			}
		}
		info.Calculations.ConditionAdjustedHealth = int(math.Round(healthByNominal + conditionModifier))
	}

	// --- IOKit Based Calculations ---
	// Uses ONLY values from the IOKit-populated structs.
	iokitACPower := info.Adapter.InputVoltage * info.Adapter.InputAmperage
	// IOKit Amperage is negative on discharge, so this product will be correctly negative.
	iokitBatteryPower := info.Battery.Voltage * info.Battery.Amperage
	// SystemPower is the absolute difference.
	iokitSystemPower := math.Abs(iokitBatteryPower - iokitACPower)

	info.Calculations.IOKit = PowerCalculation{
		ACPower:      truncate(iokitACPower),
		BatteryPower: truncate(iokitBatteryPower),
		SystemPower:  truncate(iokitSystemPower),
	}

	// --- SMC Based Calculations ---
	// These are only performed if SMC data was successfully fetched.
	if info.SMC != nil {
		smcACPower := info.SMC.InputVoltage * info.SMC.InputAmperage
		// SMC Amperage is also negative on discharge, so this product is also correctly negative.
		smcBatteryPower := info.SMC.BatteryVoltage * info.SMC.BatteryAmperage
		// SystemPower is the absolute difference.
		smcSystemPower := math.Abs(smcBatteryPower - smcACPower)

		info.Calculations.SMC = PowerCalculation{
			ACPower:      truncate(smcACPower),
			BatteryPower: truncate(smcBatteryPower),
			SystemPower:  truncate(smcSystemPower),
		}
	}
}

// --- Helper Functions ---

// Find min/max in a slice
func findMinMax(a []int) (min int, max int) {
	if len(a) == 0 {
		return 0, 0
	}
	min = a[0]
	max = a[0]
	for _, value := range a {
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}
	return min, max
}

// Truncate rounds a float down to two decimal places.
func truncate(f float64) float64 {
	return math.Trunc(f*100) / 100
}
