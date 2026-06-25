//go:build darwin

package powerkit

import (
	"fmt"
	"slices"
)

const (
	nativeMacOSChargeLimitMajor = 27

	chargeLimitReasonNativePowerUI     = "powerui_probe"
	chargeLimitReasonNativeUnavailable = "powerui_unavailable"
	chargeLimitReasonSMCControl        = "smc_control"
	chargeLimitReasonUnavailable       = "no_charge_limit_backend"
)

var (
	nativeChargeLimitProbeFn = probeNativeChargeLimit
	setNativeChargeLimitFn   = setNativeChargeLimit
)

type nativeChargeLimitProbeResult struct {
	Available       bool
	Writable        bool
	AllowedPercents []int
	Reason          string
}

func resolveChargeLimitCapability(info *SystemInfo) ChargeLimitCapability {
	native := nativeMacOSChargeLimitCapability()
	if native.Available {
		return native
	}

	if info != nil && info.SMC != nil && info.SMC.State.ChargingControlAvailable {
		return smcInhibitChargeLimitCapability()
	}

	return unavailableChargeLimitCapability(chargeLimitReasonUnavailable)
}

func nativeMacOSChargeLimitCapability() ChargeLimitCapability {
	if getOSMajorVersionFn() < nativeMacOSChargeLimitMajor {
		return unavailableChargeLimitCapability(chargeLimitReasonNativeUnavailable)
	}

	probe := nativeChargeLimitProbeFn()
	if !probe.Available {
		reason := probe.Reason
		if reason == "" {
			reason = chargeLimitReasonNativeUnavailable
		}
		return unavailableChargeLimitCapability(reason)
	}

	allowed := normalizedAllowedPercents(probe.AllowedPercents)
	if len(allowed) == 0 {
		return unavailableChargeLimitCapability(chargeLimitReasonNativeUnavailable)
	}

	return ChargeLimitCapability{
		Available:       true,
		Writable:        probe.Writable,
		Backend:         ChargeLimitBackendNativeMacOS,
		MinPercent:      allowed[0],
		MaxPercent:      allowed[len(allowed)-1],
		StepPercent:     inferredStep(allowed),
		AllowedPercents: allowed,
		Reason:          chargeLimitReasonNativePowerUI,
	}
}

func smcInhibitChargeLimitCapability() ChargeLimitCapability {
	return ChargeLimitCapability{
		Available:   true,
		Writable:    true,
		Backend:     ChargeLimitBackendSMCInhibit,
		MinPercent:  60,
		MaxPercent:  100,
		StepPercent: 10,
		Reason:      chargeLimitReasonSMCControl,
	}
}

func unavailableChargeLimitCapability(reason string) ChargeLimitCapability {
	return ChargeLimitCapability{
		Available: false,
		Writable:  false,
		Backend:   ChargeLimitBackendUnavailable,
		Reason:    reason,
	}
}

func normalizedAllowedPercents(values []int) []int {
	allowed := append([]int(nil), values...)
	slices.Sort(allowed)
	allowed = slices.Compact(allowed)

	out := allowed[:0]
	for _, value := range allowed {
		if value < 0 || value > 100 {
			continue
		}
		out = append(out, value)
	}
	return out
}

func inferredStep(values []int) int {
	if len(values) < 2 {
		return 0
	}

	step := values[1] - values[0]
	for i := 2; i < len(values); i++ {
		diff := values[i] - values[i-1]
		if diff <= 0 || diff != step {
			return 0
		}
	}
	return step
}

func (c ChargeLimitCapability) Allows(percent int) bool {
	if percent < c.MinPercent || percent > c.MaxPercent {
		return false
	}
	if len(c.AllowedPercents) > 0 {
		return slices.Contains(c.AllowedPercents, percent)
	}
	return true
}

// SetChargeLimit writes a native macOS manual charge limit when that backend is available.
func SetChargeLimit(percent int) error {
	capability := nativeMacOSChargeLimitCapability()
	if !capability.Available || !capability.Writable || capability.Backend != ChargeLimitBackendNativeMacOS {
		return fmt.Errorf("%w: native macOS charge limit is unavailable", ErrNotSupported)
	}
	if !capability.Allows(percent) {
		return fmt.Errorf("charge limit %d is not allowed for backend %s", percent, capability.Backend)
	}
	return setNativeChargeLimitFn(percent)
}
