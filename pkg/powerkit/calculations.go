package powerkit

import "math"

// calculateDerivedMetrics is the top-level function that orchestrates the
// calculation of all derived data, such as power metrics and battery health.
func calculateDerivedMetrics(info *SystemInfo) {
	if info.IOKit != nil {
		calculateIOKitMetrics(info.IOKit)
	}
	if info.SMC != nil {
		calculateSMCMetrics(info.SMC)
	}
}

// calculateIOKitMetrics populates all calculated fields for the IOKitData struct.
func calculateIOKitMetrics(d *IOKitData) {
	calculateHealthMetrics(&d.Battery, &d.Calculations)
	calculateIOKitPower(d)
}

// calculateHealthMetrics computes various battery health percentages.
func calculateHealthMetrics(b *IOKitBattery, c *IOKitCalculations) {
	if b.DesignCapacity <= 0 {
		return
	}

	designCapF := float64(b.DesignCapacity)
	c.HealthByMaxCapacity = int(math.Round((float64(b.MaxCapacity) / designCapF) * 100.0))
	c.HealthByNominalCapacity = int(math.Round((float64(b.NominalCapacity) / designCapF) * 100.0))

	var conditionModifier float64
	if len(b.IndividualCellVoltages) > 1 {
		minV, maxV := findMinMax(b.IndividualCellVoltages)
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
	c.ConditionAdjustedHealth = int(math.Round(float64(c.HealthByNominalCapacity) + conditionModifier))
}

// calculateIOKitPower computes power metrics based on IOKit data.
func calculateIOKitPower(d *IOKitData) {
	AdapterPower := d.Adapter.InputVoltage * d.Adapter.InputAmperage
	batteryPower := d.Battery.Voltage * d.Battery.Amperage
	systemPower := AdapterPower - batteryPower
	if systemPower < 0 {
		systemPower = 0
	}

	d.Calculations.AdapterPower = truncate(AdapterPower)
	d.Calculations.BatteryPower = truncate(batteryPower)
	d.Calculations.SystemPower = truncate(systemPower)
}

// calculateSMCMetrics populates all calculated fields for the SMCData struct.
func calculateSMCMetrics(d *SMCData) {
	AdapterPower := d.Adapter.InputVoltage * d.Adapter.InputAmperage
	batteryPower := d.Battery.Voltage * d.Battery.Amperage
	systemPower := AdapterPower - batteryPower
	if systemPower < 0 {
		systemPower = 0
	}

	d.Calculations = SMCCalculations{
		AdapterPower: truncate(AdapterPower),
		BatteryPower: truncate(batteryPower),
		SystemPower:  truncate(systemPower),
	}
}
