//go:build darwin

package powerkit

import (
	"errors"
	"log"
	"reflect"
	"sync"

	"github.com/peterneutron/powerkit-go/internal/iokit"
	"github.com/peterneutron/powerkit-go/internal/powerd"
)

var (
	errSystemEventStreamActive  = errors.New("powerkit system event stream already active")
	errConflictingStreamHooks   = errors.New("powerkit system event stream already active with different hooks")
	streamMu                    sync.Mutex
	streamActive                bool
	activeStreamHooks           StreamHooks
	startMonitorFn              = iokit.StartMonitor
	setBeforeSleepHookFn        = iokit.SetBeforeSleepHook
	internalEventSource         = func() <-chan iokit.InternalEvent { return iokit.Events }
	enqueueInitialBatteryUpdate = func() {
		select {
		case iokit.Events <- iokit.InternalEvent{Type: iokit.BatteryUpdate}:
		default:
		}
	}
)

// StreamSystemEvents starts monitoring IOKit for all relevant power and
// battery events. It returns a single, read-only channel that delivers a
// unified SystemEvent for any change.
func StreamSystemEvents() (<-chan SystemEvent, error) {
	return StreamSystemEventsWithHooks(StreamHooks{})
}

// StreamSystemEventsWithHooks starts the singleton system event stream and
// installs synchronous lifecycle hooks such as BeforeSleep.
func StreamSystemEventsWithHooks(hooks StreamHooks) (<-chan SystemEvent, error) {
	streamMu.Lock()
	defer streamMu.Unlock()

	if streamActive {
		if !sameStreamHooks(activeStreamHooks, hooks) {
			return nil, errConflictingStreamHooks
		}
		return nil, errSystemEventStreamActive
	}

	setBeforeSleepHookFn(hooks.BeforeSleep)
	streamActive = true
	activeStreamHooks = hooks

	systemEventChan := make(chan SystemEvent, 16)
	go func(source <-chan iokit.InternalEvent) {
		defer close(systemEventChan)
		defer releaseStreamRegistration()

		for internalEvent := range source {
			publicEvent, ok := translateInternalEvent(internalEvent)
			if !ok {
				continue
			}

			if publicEvent.Type == EventTypeBatteryUpdate {
				select {
				case systemEventChan <- publicEvent:
				default:
				}
				continue
			}

			systemEventChan <- publicEvent
		}
	}(internalEventSource())

	startMonitorFn()
	enqueueInitialBatteryUpdate()

	return systemEventChan, nil
}

func releaseStreamRegistration() {
	setBeforeSleepHookFn(nil)
	streamMu.Lock()
	defer streamMu.Unlock()
	streamActive = false
	activeStreamHooks = StreamHooks{}
}

func sameStreamHooks(a, b StreamHooks) bool {
	return sameFunc(a.BeforeSleep, b.BeforeSleep)
}

func sameFunc(a, b func()) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	default:
		return reflect.ValueOf(a).Pointer() == reflect.ValueOf(b).Pointer()
	}
}

func translateInternalEvent(internalEvent iokit.InternalEvent) (SystemEvent, bool) {
	switch internalEvent.Type {
	case iokit.BatteryUpdate:
		info, err := buildBatteryUpdateInfo()
		if err != nil {
			log.Printf("Error fetching IOKit data in stream: %v", err)
			return SystemEvent{}, false
		}
		return SystemEvent{Type: EventTypeBatteryUpdate, Info: info}, true
	case iokit.SystemWillSleep:
		return SystemEvent{Type: EventTypeSystemWillSleep}, true
	case iokit.SystemDidWake:
		return SystemEvent{Type: EventTypeSystemDidWake}, true
	default:
		log.Printf("Warning: Received unknown internal event type: %d", internalEvent.Type)
		return SystemEvent{}, false
	}
}

func buildBatteryUpdateInfo() (*SystemInfo, error) {
	iokitRawData, err := fetchIOKitData(false)
	if err != nil {
		return nil, err
	}

	sysAllowedGlobal, dspAllowedGlobal, gErr := globalSleepStatusFn()
	dspActiveApp := powerdIsActiveFn(powerd.PreventDisplaySleep)
	sysActiveApp := powerdIsActiveFn(powerd.PreventSystemSleep)
	dspAllowedApp := !dspActiveApp
	sysAllowedApp := !sysActiveApp && !dspActiveApp
	if gErr != nil {
		dspAllowedGlobal = dspAllowedApp
		sysAllowedGlobal = sysAllowedApp
	}

	lpmEnabled, lpmAvailable, _ := getLowPowerModeFn()

	info := &SystemInfo{
		OS: OSInfo{
			Firmware:                  currentSMCConfig.Firmware,
			FirmwareVersion:           currentFirmwareInfo.Version,
			FirmwareSource:            currentFirmwareInfo.Source,
			FirmwareMajor:             currentFirmwareInfo.Major,
			FirmwareCompatStatus:      firmwareCompatStatus(currentFirmwareInfo.Major),
			FirmwareProfileID:         currentSMCConfig.FirmwareProfileID,
			FirmwareProfileVersion:    currentSMCConfig.FirmwareProfileVersion,
			GlobalSystemSleepAllowed:  sysAllowedGlobal,
			GlobalDisplaySleepAllowed: dspAllowedGlobal,
			AppSystemSleepAllowed:     sysAllowedApp,
			AppDisplaySleepAllowed:    dspAllowedApp,
			LowPowerMode:              LowPowerModeInfo{Enabled: lpmEnabled, Available: lpmAvailable},
		},
		IOKit: newIOKitData(iokitRawData),
		SMC:   nil,
	}
	initSystemInfoMetadata(info)
	info.iokitQueried = true
	info.iokitAvailable = true
	info.smcQueried = false
	info.smcAvailable = false
	info.adapterTelemetrySource = string(iokitRawData.TelemetrySource)
	info.adapterTelemetryReason = string(iokitRawData.TelemetryReason)
	info.forceTelemetryFallback = iokitRawData.ForceFallback
	calculateDerivedMetrics(info)

	return info, nil
}
