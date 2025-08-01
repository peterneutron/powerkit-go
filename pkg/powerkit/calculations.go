package powerkit

import "math"

// calculateDerivedMetrics populates the Calculations struct with health
// percentages and source-specific power flow data in Watts.
func calculateDerivedMetrics(info *SystemInfo) {
	if info.IOKit != nil {

		// --- Health Percentage Calculations ---
		if info.IOKit.Battery.DesignCapacity > 0 {
			designCapF := float64(info.IOKit.Battery.DesignCapacity)

			healthByMax := (float64(info.IOKit.Battery.MaxCapacity) / designCapF) * 100.0
			info.IOKit.Calculations.HealthByMaxCapacity = int(math.Round(healthByMax))

			healthByNominal := (float64(info.IOKit.Battery.NominalCapacity) / designCapF) * 100.0
			info.IOKit.Calculations.HealthByNominalCapacity = int(math.Round(healthByNominal))

			var conditionModifier float64
			if len(info.IOKit.Battery.IndividualCellVoltages) > 1 {
				minV, maxV := findMinMax(info.IOKit.Battery.IndividualCellVoltages)
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
			info.IOKit.Calculations.ConditionAdjustedHealth = int(math.Round(healthByNominal + conditionModifier))
		}

		// --- IOKit Based Calculations ---
		// Uses ONLY values from the IOKit-populated structs.
		iokitACPower := info.IOKit.Adapter.InputVoltage * info.IOKit.Adapter.InputAmperage
		// IOKit Amperage is negative on discharge, so this product will be correctly negative.
		iokitBatteryPower := info.IOKit.Battery.Voltage * info.IOKit.Battery.Amperage
		// SystemPower is the absolute difference.
		iokitSystemPower := math.Abs(iokitACPower - iokitBatteryPower)

		info.IOKit.Calculations.ACPower = truncate(iokitACPower)
		info.IOKit.Calculations.BatteryPower = truncate(iokitBatteryPower)
		info.IOKit.Calculations.SystemPower = truncate(iokitSystemPower)
	}

	// --- SMC Based Calculations ---
	// These are only performed if SMC data was successfully fetched.
	if info.SMC != nil {
		smcACPower := info.SMC.Adapter.InputVoltage * info.SMC.Adapter.InputAmperage
		// SMC Amperage is also negative on discharge, so this product is also correctly negative.
		smcBatteryPower := info.SMC.Battery.Voltage * info.SMC.Battery.Amperage
		// SystemPower is the absolute difference.
		smcSystemPower := math.Abs(smcACPower - smcBatteryPower)

		info.SMC.Calculations = SMCCalculations{
			ACPower:      truncate(smcACPower),
			BatteryPower: truncate(smcBatteryPower),
			SystemPower:  truncate(smcSystemPower),
		}
	}
}
