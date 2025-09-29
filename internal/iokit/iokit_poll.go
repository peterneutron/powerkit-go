//go:build darwin
// +build darwin

package iokit

/*
#cgo CFLAGS: -mmacosx-version-min=15.0
#cgo LDFLAGS: -framework CoreFoundation -framework IOKit
#include <CoreFoundation/CoreFoundation.h>
#include <IOKit/IOKitLib.h>
#include <string.h>

typedef struct {
    int is_charging;
    int is_connected;
    int is_fully_charged;
    long cycle_count;
    long design_capacity;
    long max_capacity;
    long nominal_capacity;

    //long state_of_charge;

    long current_capacity_raw;
    long current_charge_raw;
    long current_charge;

    long time_to_empty;
    long time_to_full;
    long temperature;
    long voltage;
    long amperage;
    char serial_number[256];
    char device_name[256];
    long adapter_watts;
    long adapter_voltage;
    long adapter_amperage;
    char adapter_description[256];
    long source_voltage;
    long source_amperage;
    long cell_voltages[16];
    int  cell_voltage_count;
    int  has_power_telemetry;
} c_battery_info;

static long get_long_prop(CFDictionaryRef dict, const char *key) {
    CFStringRef key_ref = CFStringCreateWithCString(NULL, key, kCFStringEncodingUTF8);
    if (!key_ref) return 0;
    long value = 0;
    CFNumberRef num_ref = (CFNumberRef)CFDictionaryGetValue(dict, key_ref);
    if (num_ref != NULL && CFGetTypeID(num_ref) == CFNumberGetTypeID()) {
        CFNumberGetValue(num_ref, kCFNumberSInt64Type, &value);
    }
    CFRelease(key_ref);
    return value;
}

static int get_bool_prop(CFDictionaryRef dict, const char *key) {
    CFStringRef key_ref = CFStringCreateWithCString(NULL, key, kCFStringEncodingUTF8);
    if (!key_ref) return 0;
    int value = 0;
    CFBooleanRef bool_ref = (CFBooleanRef)CFDictionaryGetValue(dict, key_ref);
    if (bool_ref != NULL && CFGetTypeID(bool_ref) == CFBooleanGetTypeID()) {
        value = CFBooleanGetValue(bool_ref);
    }
    CFRelease(key_ref);
    return value;
}

static void get_string_prop(CFDictionaryRef dict, const char *key, char *buffer, int buffer_size) {
    CFStringRef key_ref = CFStringCreateWithCString(NULL, key, kCFStringEncodingUTF8);
    if (!key_ref) { buffer[0] = '\0'; return; }
    CFStringRef str_ref = (CFStringRef)CFDictionaryGetValue(dict, key_ref);
    if (str_ref != NULL && CFGetTypeID(str_ref) == CFStringGetTypeID()) {
        CFStringGetCString(str_ref, buffer, buffer_size, kCFStringEncodingUTF8);
    } else {
        buffer[0] = '\0';
    }
    CFRelease(key_ref);
}

static CFDictionaryRef get_dict_prop(CFDictionaryRef dict, const char *key) {
    CFStringRef key_ref = CFStringCreateWithCString(NULL, key, kCFStringEncodingUTF8);
    if (!key_ref) return NULL;
    CFDictionaryRef value = (CFDictionaryRef)CFDictionaryGetValue(dict, key_ref);
    CFRelease(key_ref);
    if (value != NULL && CFGetTypeID(value) == CFDictionaryGetTypeID()) {
        return value;
    }
    return NULL;
}

static void get_long_array_prop(CFDictionaryRef dict, const char *key, long *out_array, int max_count, int *final_count) {
    *final_count = 0;
    CFStringRef key_ref = CFStringCreateWithCString(NULL, key, kCFStringEncodingUTF8);
    if (!key_ref) return;
    CFTypeRef value_ref = CFDictionaryGetValue(dict, key_ref);
    CFRelease(key_ref);
    if (value_ref != NULL && CFGetTypeID(value_ref) == CFArrayGetTypeID()) {
        CFArrayRef array_ref = (CFArrayRef)value_ref;
        CFIndex count = CFArrayGetCount(array_ref);
        if (count > max_count) {
            count = max_count;
        }
        *final_count = (int)count;
        for (CFIndex i = 0; i < count; i++) {
            CFNumberRef num_ref = (CFNumberRef)CFArrayGetValueAtIndex(array_ref, i);
            if (num_ref != NULL && CFGetTypeID(num_ref) == CFNumberGetTypeID()) {
                CFNumberGetValue(num_ref, kCFNumberSInt64Type, &out_array[i]);
            } else {
                out_array[i] = 0;
            }
        }
    }
}

int get_all_battery_info(c_battery_info *info) {
    CFMutableDictionaryRef matching = IOServiceMatching("AppleSmartBattery");
    if (matching == NULL) return 1;
    io_iterator_t iterator;
    if (IOServiceGetMatchingServices(kIOMainPortDefault, matching, &iterator) != KERN_SUCCESS) {
        return 2;
    }
    io_service_t battery = IOIteratorNext(iterator);
    IOObjectRelease(iterator);
    if (battery == IO_OBJECT_NULL) return 3;
    CFMutableDictionaryRef properties = NULL;
    kern_return_t result = IORegistryEntryCreateCFProperties(battery, &properties, kCFAllocatorDefault, 0);
    IOObjectRelease(battery);
    if (result != KERN_SUCCESS || properties == NULL) return 4;
    info->is_charging = get_bool_prop(properties, "IsCharging");
    info->is_connected = get_bool_prop(properties, "ExternalConnected");
    info->is_fully_charged = get_bool_prop(properties, "FullyCharged");
    info->cycle_count = get_long_prop(properties, "CycleCount");
    info->design_capacity = get_long_prop(properties, "DesignCapacity");
    info->max_capacity = get_long_prop(properties, "AppleRawMaxCapacity");
    info->nominal_capacity = get_long_prop(properties, "NominalChargeCapacity");

    info->current_capacity_raw = get_long_prop(properties, "AppleRawCurrentCapacity");
    info->current_charge = get_long_prop(properties, "CurrentCapacity");

    info->time_to_empty = get_long_prop(properties, "AvgTimeToEmpty");
    info->time_to_full = get_long_prop(properties, "AvgTimeToFull");
    info->temperature = get_long_prop(properties, "Temperature");
    info->voltage = get_long_prop(properties, "Voltage");
    info->amperage = get_long_prop(properties, "Amperage");
    get_string_prop(properties, "Serial", info->serial_number, 256);
    get_string_prop(properties, "DeviceName", info->device_name, 256);
    CFDictionaryRef adapter_details = get_dict_prop(properties, "AdapterDetails");
    if (adapter_details) {
        info->adapter_watts = get_long_prop(adapter_details, "Watts");
        info->adapter_voltage = get_long_prop(adapter_details, "AdapterVoltage");
        info->adapter_amperage = get_long_prop(adapter_details, "Current");
        get_string_prop(adapter_details, "Description", info->adapter_description, 256);
    }
    CFDictionaryRef power_telemetry = get_dict_prop(properties, "PowerTelemetryData");
    if (power_telemetry) {
        info->source_voltage = get_long_prop(power_telemetry, "SystemVoltageIn");
        info->source_amperage = get_long_prop(power_telemetry, "SystemCurrentIn");
        info->has_power_telemetry = 1;
    }
    CFDictionaryRef battery_data = get_dict_prop(properties, "BatteryData");
    if (battery_data) {
        get_long_array_prop(battery_data, "CellVoltage", info->cell_voltages, 16, &info->cell_voltage_count);
    }
    CFDictionaryRef battery_data_dict = get_dict_prop(properties, "BatteryData");
    if (battery_data_dict != NULL) {
        info->current_charge_raw = get_long_prop(battery_data_dict, "StateOfCharge");
    } else {
        info->current_charge_raw = 0;
    }
    CFRelease(properties);
    return 0;
}
*/
import "C"

import (
	"fmt"
	"math"

	"github.com/peterneutron/powerkit-go/internal/smc"
)

var (
	getAllBatteryInfoFn = func(info *C.c_battery_info) C.int {
		return C.get_all_battery_info(info)
	}
	smcFetchDataFn = smc.FetchData
)

// FetchData retrieves the raw battery and power data from IOKit.
// When forceFallback is true, SMC data will be used for adapter telemetry even if
// PowerTelemetryData is present.
func FetchData(forceFallback bool) (*RawData, error) {
	var cInfo C.c_battery_info

	ret := getAllBatteryInfoFn(&cInfo)
	if ret != 0 {
		return nil, fmt.Errorf("iokit query failed with C error code: %d", ret)
	}

	telemetryAvailable := cInfo.has_power_telemetry != 0 && !forceFallback

	data := &RawData{

		CurrentCharge:      int(cInfo.current_charge),       // %
		CurrentChargeRaw:   int(cInfo.current_charge_raw),   // %
		CurrentCapacityRaw: int(cInfo.current_capacity_raw), // mAh
		IsCharging:         cInfo.is_charging != 0,
		IsConnected:        cInfo.is_connected != 0,
		IsFullyCharged:     cInfo.is_fully_charged != 0,
		CycleCount:         int(cInfo.cycle_count),
		DesignCapacity:     int(cInfo.design_capacity),
		MaxCapacity:        int(cInfo.max_capacity),
		NominalCapacity:    int(cInfo.nominal_capacity),
		TimeToEmpty:        int(cInfo.time_to_empty),
		TimeToFull:         int(cInfo.time_to_full),
		Temperature:        int(cInfo.temperature),
		Voltage:            int(cInfo.voltage),
		Amperage:           int(cInfo.amperage),
		SerialNumber:       C.GoString(&cInfo.serial_number[0]),
		DeviceName:         C.GoString(&cInfo.device_name[0]),
		AdapterWatts:       int(cInfo.adapter_watts),
		AdapterVoltage:     int(cInfo.adapter_voltage),
		AdapterAmperage:    int(cInfo.adapter_amperage),
		AdapterDesc:        C.GoString(&cInfo.adapter_description[0]),
		SourceVoltage:      int(cInfo.source_voltage),
		SourceAmperage:     int(cInfo.source_amperage),
		TelemetryAvailable: telemetryAvailable,
	}

	applyAdapterTelemetryFallback(data, telemetryAvailable)

	if cInfo.cell_voltage_count > 0 {
		data.CellVoltages = make([]int, cInfo.cell_voltage_count)
		cVoltagesPtr := &cInfo.cell_voltages
		for i := 0; i < int(cInfo.cell_voltage_count); i++ {
			data.CellVoltages[i] = int(cVoltagesPtr[i])
		}
	}
	return data, nil
}

func applyAdapterTelemetryFallback(data *RawData, telemetryAvailable bool) {
	if telemetryAvailable {
		return
	}
	fallback, err := smcFetchDataFn([]string{smc.KeyAdapterVoltage, smc.KeyAdapterCurrent})
	if err != nil {
		return
	}
	if v, ok := fallback[smc.KeyAdapterVoltage]; ok {
		data.SourceVoltage = int(math.Round(v * 1000.0))
	}
	if a, ok := fallback[smc.KeyAdapterCurrent]; ok {
		data.SourceAmperage = int(math.Round(a * 1000.0))
	}
}
