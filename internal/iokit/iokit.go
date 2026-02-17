//go:build darwin

// Package iokit provides internal access to IOKit for both polling
// and streaming power source data.
package iokit

import "sync"

// InternalEventType is a private enum for distinguishing event sources.
type InternalEventType int

const (
	// BatteryUpdate indicates a change in the battery's properties.
	BatteryUpdate InternalEventType = iota
	// SystemWillSleep indicates the system is about to sleep.
	SystemWillSleep
	// SystemDidWake indicates the system has woken up.
	SystemDidWake
)

// InternalEvent is a private struct used to pass notifications from the
// C RunLoop to the Go dispatcher.
type InternalEvent struct {
	Type InternalEventType
}

// Events is the single, unified channel for all notifications from the C layer.
var Events = make(chan InternalEvent, 2) // Buffer of 2 is safe

// AdapterTelemetrySource identifies which subsystem produced adapter input telemetry.
type AdapterTelemetrySource string

const (
	// AdapterTelemetrySourceIOKit indicates adapter input telemetry came from IOKit.
	AdapterTelemetrySourceIOKit AdapterTelemetrySource = "iokit"
	// AdapterTelemetrySourceSMCFallback indicates telemetry came from SMC fallback reads.
	AdapterTelemetrySourceSMCFallback AdapterTelemetrySource = "smc_fallback"
	// AdapterTelemetrySourceUnavailable indicates telemetry was not available from any source.
	AdapterTelemetrySourceUnavailable AdapterTelemetrySource = "unavailable"
)

// AdapterTelemetryReason describes why a telemetry source decision was made.
type AdapterTelemetryReason string

const (
	// AdapterTelemetryReasonNone indicates no exceptional reason was needed.
	AdapterTelemetryReasonNone AdapterTelemetryReason = "none"
	// AdapterTelemetryReasonNoAdapter indicates no power adapter was connected.
	AdapterTelemetryReasonNoAdapter AdapterTelemetryReason = "no_adapter"
	// AdapterTelemetryReasonMissingIOKit indicates IOKit did not expose required telemetry.
	AdapterTelemetryReasonMissingIOKit AdapterTelemetryReason = "missing_iokit"
	// AdapterTelemetryReasonInvalidIOKit indicates IOKit values were present but invalid.
	AdapterTelemetryReasonInvalidIOKit AdapterTelemetryReason = "invalid_iokit"
	// AdapterTelemetryReasonForced indicates fallback was explicitly requested.
	AdapterTelemetryReasonForced AdapterTelemetryReason = "forced"
	// AdapterTelemetryReasonSMCError indicates SMC fallback failed.
	AdapterTelemetryReasonSMCError AdapterTelemetryReason = "smc_error"
)

// RawData holds the unprocessed data returned from the IOKit C functions.
// This struct is intended for internal use within the library.
type RawData struct {
	CurrentCharge      int
	CurrentChargeRaw   int
	CurrentCapacityRaw int

	IsCharging         bool
	IsConnected        bool
	IsFullyCharged     bool
	CycleCount         int
	DesignCapacity     int
	MaxCapacity        int
	NominalCapacity    int
	TimeToEmpty        int
	TimeToFull         int
	Temperature        int
	Voltage            int
	Amperage           int
	SerialNumber       string
	DeviceName         string
	AdapterWatts       int
	AdapterVoltage     int
	AdapterAmperage    int
	AdapterDesc        string
	SourceVoltage      int
	SourceAmperage     int
	CellVoltages       []int
	TelemetryAvailable bool
	TelemetrySource    AdapterTelemetrySource
	TelemetryReason    AdapterTelemetryReason
	ForceFallback      bool
}

// --- Streaming Globals ---
var startOnce sync.Once
