//go:build darwin
// +build darwin

package iokit

/*
#cgo CFLAGS: -mmacosx-version-min=15.0
#cgo LDFLAGS: -framework CoreFoundation -framework IOKit
#include <CoreFoundation/CoreFoundation.h>
#include <IOKit/IOKitLib.h>
#include <IOKit/IOMessage.h>
#include <IOKit/pwr_mgt/IOPMLib.h>

// --- Battery Globals ---
static IONotificationPortRef   gNotifyPort;
static io_object_t             gNotifier;
static io_service_t            gBattery;

// --- Power Management Globals ---
static IONotificationPortRef   gPowerNotifyPort;
static io_object_t             gPowerNotifier;
static io_connect_t            gRootPort; // Root Power Domain connection

// --- Go-side callbacks ---
extern void pushBatteryUpdate();
extern void pushWillSleep();
extern void pushDidWake();

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
    pushBatteryUpdate();

    // Tear down last notifier and re-arm for the next event
    IOObjectRelease(gNotifier);
    registerInterest();
}

// This is called on the CFRunLoop thread when the system sleeps or wakes.
static void powerCallback(
    void* refCon,
    io_service_t service,
    natural_t messageType,
    void* messageArgument
) {
    switch (messageType) {
        case kIOMessageSystemWillSleep:
            pushWillSleep();
            // Acknowledge the notification to allow the sleep process to continue.
            // This is a required step.
            IOAllowPowerChange(gRootPort, (long)messageArgument);
            break;
        case kIOMessageCanSystemSleep:
            // This message asks if we will allow sleep. We must allow it.
            IOAllowPowerChange(gRootPort, (long)messageArgument);
            break;
        case kIOMessageSystemHasPoweredOn:
            pushDidWake();
            break;
    }
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

// startNotifications sets up all IOKit listeners on a single run loop.
static void startNotifications() {
    // --- Setup Power Management Notifications ---
    // Get a connection to the root power domain to listen for sleep/wake.
    gRootPort = IORegisterForSystemPower(
        NULL,               // refCon
        &gPowerNotifyPort,  // notification port
        powerCallback,      // C callback function
        &gPowerNotifier     // notifier object
    );
    if (gRootPort) {
        // Hook the power notifications into the main run loop.
        CFRunLoopAddSource(CFRunLoopGetCurrent(),
                           IONotificationPortGetRunLoopSource(gPowerNotifyPort),
                           kCFRunLoopDefaultMode);
    }


    // --- Setup Battery Change Notifications ---
    gBattery = IOServiceGetMatchingService(
        kIOMainPortDefault,
        IOServiceMatching("AppleSmartBattery")
    );
    // Note: It's okay if there's no battery. The power events will still work.
    if (gBattery != IO_OBJECT_NULL) {
        // Create the notification port for battery events
        gNotifyPort = IONotificationPortCreate(kIOMainPortDefault);
        if (gNotifyPort) {
            // Arm for the very first battery property change
            registerInterest();
            // Hook the battery notifications into the same run loop.
            CFRunLoopSourceRef src = IONotificationPortGetRunLoopSource(gNotifyPort);
            CFRunLoopAddSource(CFRunLoopGetCurrent(), src, kCFRunLoopDefaultMode);
        }
    }


    // Blocks forever; Go’s goroutine will keep running and service both
    // battery and power management events from the single loop.
    CFRunLoopRun();
    }
*/
import "C"

//export pushBatteryUpdate
func pushBatteryUpdate() {
	select {
	case Events <- InternalEvent{Type: BatteryUpdate}:
	default:
	}
}

//export pushWillSleep
func pushWillSleep() {
	select {
	case Events <- InternalEvent{Type: SystemWillSleep}:
	default:
	}
}

//export pushDidWake
func pushDidWake() {
	select {
	case Events <- InternalEvent{Type: SystemDidWake}:
	default:
	}
}

// StartMonitor initializes the unified IOKit notification system.
// It starts a dedicated goroutine to run the C RunLoop.
func StartMonitor() {
	startOnce.Do(func() {
		// Call the new unified C function
		go C.startNotifications()
	})
}
