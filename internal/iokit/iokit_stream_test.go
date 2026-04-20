package iokit

import (
	"testing"
	"time"
)

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

	select {
	case event := <-Events:
		if event.Type != SystemWillSleep {
			t.Fatalf("expected SystemWillSleep event, got %v", event.Type)
		}
	default:
		t.Fatalf("expected SystemWillSleep event to be emitted")
	}
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

	select {
	case <-delivered:
		t.Fatalf("expected wake delivery to wait for queue space instead of dropping")
	case <-time.After(50 * time.Millisecond):
	}

	select {
	case event := <-Events:
		if event.Type != BatteryUpdate {
			t.Fatalf("expected first queued event to remain BatteryUpdate, got %v", event.Type)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out draining battery update")
	}

	select {
	case <-delivered:
	case <-time.After(time.Second):
		t.Fatalf("wake delivery did not complete after queue space became available")
	}

	select {
	case event := <-Events:
		if event.Type != SystemDidWake {
			t.Fatalf("expected SystemDidWake event, got %v", event.Type)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out receiving wake event")
	}
}
