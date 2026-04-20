package iokit

import (
	"testing"
	"time"
)

func expectQueuedEventType(t *testing.T, want InternalEventType) {
	t.Helper()

	select {
	case event := <-Events:
		if event.Type != want {
			t.Fatalf("expected %v event, got %v", want, event.Type)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %v event", want)
	}
}

func expectNoSignalWithin(t *testing.T, ch <-chan struct{}, wait time.Duration, message string) {
	t.Helper()

	select {
	case <-ch:
		t.Fatal(message)
	case <-time.After(wait):
	}
}

func expectSignalWithin(t *testing.T, ch <-chan struct{}, wait time.Duration, message string) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(wait):
		t.Fatal(message)
	}
}

func TestProcessWillSleepNotificationRunsHookBeforeAck(t *testing.T) {
	oldEvents := Events
	oldHook := beforeSleepHook
	Events = make(chan InternalEvent, 1)
	beforeSleepHook = nil
	t.Cleanup(func() {
		Events = oldEvents
		beforeSleepHook = oldHook
	})

	ackCalled := false
	hookCalled := false
	setBeforeSleepHook(func() {
		if ackCalled {
			t.Fatalf("before-sleep hook ran after sleep acknowledgement")
		}
		hookCalled = true
	})

	processWillSleepNotification(func() {
		if !hookCalled {
			t.Fatalf("sleep acknowledgement ran before before-sleep hook")
		}
		ackCalled = true
	})

	if !hookCalled {
		t.Fatalf("expected before-sleep hook to run")
	}
	if !ackCalled {
		t.Fatalf("expected acknowledgement to run")
	}

	expectQueuedEventType(t, SystemWillSleep)
}

func TestPushDidWakeReliableUnderQueuePressureWhileBatteryUpdatesStayLossy(t *testing.T) {
	oldEvents := Events
	Events = make(chan InternalEvent, 1)
	t.Cleanup(func() {
		Events = oldEvents
	})

	pushBatteryUpdate()
	pushBatteryUpdate()

	if got := len(Events); got != 1 {
		t.Fatalf("expected lossy battery queue to cap at 1 event, got %d", got)
	}

	delivered := make(chan struct{})
	go func() {
		pushDidWake()
		close(delivered)
	}()

	expectNoSignalWithin(t, delivered, 50*time.Millisecond, "expected wake delivery to wait for queue space instead of dropping")
	expectQueuedEventType(t, BatteryUpdate)
	expectSignalWithin(t, delivered, time.Second, "wake delivery did not complete after queue space became available")
	expectQueuedEventType(t, SystemDidWake)
}
