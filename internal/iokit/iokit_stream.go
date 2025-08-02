//go:build darwin
// +build darwin

package iokit

/*
#cgo CFLAGS: -mmacosx-version-min=15.0
#cgo LDFLAGS: -framework CoreFoundation -framework IOKit
#include <CoreFoundation/CoreFoundation.h>
#include <IOKit/IOKitLib.h>
#include <IOKit/IOMessage.h>

static IONotificationPortRef   gNotifyPort;
static io_object_t             gNotifier;
static io_service_t            gBattery;

extern void pushUpdate();

// Forward-declare so both functions can see it
static void registerInterest();

// This is called on the CFRunLoop thread when battery props change.
static void batteryChanged(
    void *refCon,
    io_service_t service,
    natural_t messageType,
    void *messageArgument
) {
    // Tell Go there’s new data
    pushUpdate();

    // Tear down last notifier and re-arm for the next event
    IOObjectRelease(gNotifier);
    registerInterest();
}

static void registerInterest() {
    kern_return_t kr = IOServiceAddInterestNotification(
        gNotifyPort,
        gBattery,
        kIOGeneralInterest,
        batteryChanged,
        NULL,
        &gNotifier
    );
    if (kr != KERN_SUCCESS) {
        // you might log or handle error here
    }
}

static void startBatteryNotifications() {
    // Look up service once
    gBattery = IOServiceGetMatchingService(
        kIOMainPortDefault,
        IOServiceMatching("AppleSmartBattery")
    );
    if (gBattery == IO_OBJECT_NULL) return;

    // Create the notification port
    gNotifyPort = IONotificationPortCreate(kIOMainPortDefault);
    if (!gNotifyPort) {
        IOObjectRelease(gBattery);
        return;
    }

    // Arm for the very first change
    registerInterest();

    // Hook into CFRunLoop
    CFRunLoopSourceRef src = IONotificationPortGetRunLoopSource(gNotifyPort);
    CFRunLoopAddSource(CFRunLoopGetCurrent(), src, kCFRunLoopDefaultMode);

    // Blocks forever; Go’s goroutine will keep running
    CFRunLoopRun();
}
*/
import "C"

//export pushUpdate
func pushUpdate() {
	// This function is called from a C thread. The non-blocking send ensures
	// that we never block the C RunLoop, even if the Go channel is full.
	select {
	case Updates <- struct{}{}:
	default:
	}
}

// StartMonitor initializes the IOKit notification system.
// It starts a dedicated goroutine to run the C RunLoop.
func StartMonitor() {
	startOnce.Do(func() {
		go C.startBatteryNotifications()
	})
}
