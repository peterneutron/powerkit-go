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
    long state_of_charge;
    int is_charging;
    int is_connected;
    int is_fully_charged;
    long cycle_count;
    long design_capacity;
    long max_capacity;
    long nominal_capacity;
    long current_capacity;
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
    info->current_capacity = get_long_prop(properties, "AppleRawCurrentCapacity");
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
    }
    CFDictionaryRef battery_data = get_dict_prop(properties, "BatteryData");
    if (battery_data) {
        get_long_array_prop(battery_data, "CellVoltage", info->cell_voltages, 16, &info->cell_voltage_count);
    }
    CFDictionaryRef battery_data_dict = get_dict_prop(properties, "BatteryData");
    if (battery_data_dict != NULL) {
        info->state_of_charge = get_long_prop(battery_data_dict, "StateOfCharge");
    } else {
        info->state_of_charge = 0;
    }
    CFRelease(properties);
    return 0;
}
*/
import "C"

import "fmt"

// FetchData retrieves the raw battery and power data from IOKit.
func FetchData() (*RawData, error) {
	var c_info C.c_battery_info

	ret := C.get_all_battery_info(&c_info)
	if ret != 0 {
		return nil, fmt.Errorf("iokit query failed with C error code: %d", ret)
	}

	data := &RawData{
		CurrentCharge:   int(c_info.state_of_charge),
		IsCharging:      c_info.is_charging != 0,
		IsConnected:     c_info.is_connected != 0,
		IsFullyCharged:  c_info.is_fully_charged != 0,
		CycleCount:      int(c_info.cycle_count),
		DesignCapacity:  int(c_info.design_capacity),
		MaxCapacity:     int(c_info.max_capacity),
		NominalCapacity: int(c_info.nominal_capacity),
		CurrentCapacity: int(c_info.current_capacity),
		TimeToEmpty:     int(c_info.time_to_empty),
		TimeToFull:      int(c_info.time_to_full),
		Temperature:     int(c_info.temperature),
		Voltage:         int(c_info.voltage),
		Amperage:        int(c_info.amperage),
		SerialNumber:    C.GoString(&c_info.serial_number[0]),
		DeviceName:      C.GoString(&c_info.device_name[0]),
		AdapterWatts:    int(c_info.adapter_watts),
		AdapterVoltage:  int(c_info.adapter_voltage),
		AdapterAmperage: int(c_info.adapter_amperage),
		AdapterDesc:     C.GoString(&c_info.adapter_description[0]),
		SourceVoltage:   int(c_info.source_voltage),
		SourceAmperage:  int(c_info.source_amperage),
	}

	if c_info.cell_voltage_count > 0 {
		data.CellVoltages = make([]int, c_info.cell_voltage_count)
		c_voltages_ptr := &c_info.cell_voltages
		for i := 0; i < int(c_info.cell_voltage_count); i++ {
			data.CellVoltages[i] = int(c_voltages_ptr[i])
		}
	}
	return data, nil
}
