package powerkit

import (
	"testing"
	"time"

	"github.com/peterneutron/powerkit-go/internal/iokit"
)

func resetStreamStateForTest() {
	streamMu.Lock()
	streamActive = false
	activeStreamHooks = StreamHooks{}
	streamMu.Unlock()
}

func TestStreamSystemEventsCompatibilityPath(t *testing.T) {
	oldStartMonitorFn := startMonitorFn
	oldSetBeforeSleepHookFn := setBeforeSleepHookFn
	oldInternalEventSource := internalEventSource
	oldEnqueueInitialBatteryUpdate := enqueueInitialBatteryUpdate
	resetStreamStateForTest()
	t.Cleanup(func() {
		startMonitorFn = oldStartMonitorFn
		setBeforeSleepHookFn = oldSetBeforeSleepHookFn
		internalEventSource = oldInternalEventSource
		enqueueInitialBatteryUpdate = oldEnqueueInitialBatteryUpdate
		resetStreamStateForTest()
	})

	source := make(chan iokit.InternalEvent, 1)
	var installedHook func()
	startMonitorFn = func() {}
	setBeforeSleepHookFn = func(fn func()) { installedHook = fn }
	internalEventSource = func() <-chan iokit.InternalEvent { return source }
	enqueueInitialBatteryUpdate = func() {}

	eventChan, err := StreamSystemEvents()
	if err != nil {
		t.Fatalf("StreamSystemEvents returned error: %v", err)
	}
	if installedHook != nil {
		t.Fatalf("expected compatibility path to install no before-sleep hook")
	}

	source <- iokit.InternalEvent{Type: iokit.SystemDidWake}

	select {
	case event := <-eventChan:
		if event.Type != EventTypeSystemDidWake {
			t.Fatalf("expected SystemDidWake event, got %v", event.Type)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for streamed wake event")
	}

	close(source)
}

func TestStreamSystemEventsWithHooksRejectsConflictingRegistration(t *testing.T) {
	oldStartMonitorFn := startMonitorFn
	oldSetBeforeSleepHookFn := setBeforeSleepHookFn
	oldInternalEventSource := internalEventSource
	oldEnqueueInitialBatteryUpdate := enqueueInitialBatteryUpdate
	resetStreamStateForTest()
	t.Cleanup(func() {
		startMonitorFn = oldStartMonitorFn
		setBeforeSleepHookFn = oldSetBeforeSleepHookFn
		internalEventSource = oldInternalEventSource
		enqueueInitialBatteryUpdate = oldEnqueueInitialBatteryUpdate
		resetStreamStateForTest()
	})

	source := make(chan iokit.InternalEvent)
	startMonitorFn = func() {}
	setBeforeSleepHookFn = func(func()) {}
	internalEventSource = func() <-chan iokit.InternalEvent { return source }
	enqueueInitialBatteryUpdate = func() {}

	stream, err := StreamSystemEventsWithHooks(StreamHooks{BeforeSleep: func() {}})
	if err != nil {
		t.Fatalf("first stream registration failed: %v", err)
	}

	if _, err := StreamSystemEventsWithHooks(StreamHooks{}); err != errConflictingStreamHooks {
		t.Fatalf("expected conflicting hook registration error, got %v", err)
	}

	close(source)
	for event := range stream {
		_ = event
	}
}
