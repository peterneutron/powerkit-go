//go:build darwin
// +build darwin

package smc

/*
#cgo LDFLAGS: -framework CoreFoundation -framework IOKit

#include <CoreFoundation/CoreFoundation.h>
#include <IOKit/IOKitLib.h>
#include <string.h>
#include <libkern/OSByteOrder.h> // For OSSwap... functions

#define KERNEL_INDEX_SMC 2
#define SMC_CMD_READ_BYTES 5
#define SMC_CMD_READ_KEYINFO 9

// --- Data Structures ---
typedef struct {
    char major;
    char minor;
    char build;
    char reserved[1];
    UInt16 release;
} SMCKeyData_vers_t;

typedef struct {
    UInt16 version;
    UInt16 length;
    UInt32 cpuPLimit;
    UInt32 gpuPLimit;
    UInt32 memPLimit;
} SMCKeyData_pLimitData_t;

typedef struct {
    UInt32 dataSize;
    UInt32 dataType;
    char dataAttributes;
} SMCKeyData_keyInfo_t;

typedef unsigned char SMCBytes_t[32];

typedef struct {
    UInt32 key;
    SMCKeyData_vers_t vers;
    SMCKeyData_pLimitData_t pLimitData;
    SMCKeyData_keyInfo_t keyInfo;
    char result;
    char status;
    char data8;
    UInt32 data32;
    SMCBytes_t bytes;
} SMCKeyData_t;

// --- C Helper Function Declarations ---
kern_return_t smc_read_key(io_connect_t conn, const char* key, float *value);

// --- SMC C-Side Implementation ---
static UInt32 str_to_key(const char *str) {
    return (UInt32)((str[0] << 24) | (str[1] << 16) | (str[2] << 8) | str[3]);
}

static void key_to_str(UInt32 key, char *str) {
    str[0] = (key >> 24) & 0xFF;
    str[1] = (key >> 16) & 0xFF;
    str[2] = (key >> 8) & 0xFF;
    str[3] = key & 0xFF;
    str[4] = '\0';
}

static kern_return_t smc_open(io_connect_t *conn) {
    io_service_t service = IOServiceGetMatchingService(kIOMainPortDefault, IOServiceMatching("AppleSMC"));
    if (service == IO_OBJECT_NULL) return KERN_FAILURE;
    kern_return_t kr = IOServiceOpen(service, mach_task_self(), 0, conn);
    IOObjectRelease(service);
    return kr;
}

static kern_return_t smc_close(io_connect_t conn) {
    return IOServiceClose(conn);
}

static kern_return_t smc_call(io_connect_t conn, SMCKeyData_t *input, SMCKeyData_t *output) {
    size_t structSize = sizeof(SMCKeyData_t);
    return IOConnectCallStructMethod(conn, KERNEL_INDEX_SMC, input, structSize, output, &structSize);
}

// --- SMC_READ_KEY Function ---
kern_return_t smc_read_key(io_connect_t conn, const char* key, float *value) {
    SMCKeyData_t input;
    SMCKeyData_t output;
    size_t structSize = sizeof(SMCKeyData_t);

    memset(&input, 0, structSize);
    memset(&output, 0, structSize);

    input.key = str_to_key(key);
    input.data8 = SMC_CMD_READ_KEYINFO;

    kern_return_t kr = smc_call(conn, &input, &output);
    if (kr != KERN_SUCCESS || output.result != 0) {
        return KERN_FAILURE;
    }

    // Prepare for the second call
    input.keyInfo = output.keyInfo; // Save the metadata
    input.data8 = SMC_CMD_READ_BYTES;

    // *** CRITICAL FIX #1: Reset the output struct before the second call. ***
    memset(&output, 0, structSize);

    kr = smc_call(conn, &input, &output);
    if (kr != KERN_SUCCESS || output.result != 0) {
        return KERN_FAILURE;
    }

    // --- Universal Translator Logic ---
    char dataTypeStr[5];
    // *** CRITICAL FIX #2: Use the saved metadata from `input`, not `output`. ***
    key_to_str(input.keyInfo.dataType, dataTypeStr);
    UInt32 dataSize = input.keyInfo.dataSize;

    if (strcmp(dataTypeStr, "flt ") == 0 && dataSize == 4) {
        union { uint32_t i; float f; } val;
        memcpy(&val.i, output.bytes, 4);
        val.i = OSSwapLittleToHostInt32(val.i);
        *value = val.f;
    } else if (strcmp(dataTypeStr, "sp78") == 0 && dataSize == 2) {
        int16_t raw_value;
        memcpy(&raw_value, output.bytes, 2);
        raw_value = OSSwapLittleToHostInt16(raw_value);
        *value = (float)raw_value / 256.0f;
    } else if (strcmp(dataTypeStr, "fpe2") == 0 && dataSize == 2) {
        uint16_t raw_value;
        memcpy(&raw_value, output.bytes, 2);
        raw_value = OSSwapLittleToHostInt16(raw_value);
        *value = (float)raw_value / 4.0f;
    } else if (strcmp(dataTypeStr, "ui8 ") == 0) {
        *value = (float)output.bytes[0];
    } else if (strcmp(dataTypeStr, "ui16") == 0) {
        uint16_t raw_value; memcpy(&raw_value, output.bytes, 2);
        raw_value = OSSwapLittleToHostInt16(raw_value);
        *value = (float)raw_value;
    } else if (strcmp(dataTypeStr, "ui32") == 0) {
        uint32_t raw_value; memcpy(&raw_value, output.bytes, 4);
        raw_value = OSSwapLittleToHostInt32(raw_value);
        *value = (float)raw_value;
    } else if (strcmp(dataTypeStr, "si8 ") == 0) {
        *value = (float)((int8_t)output.bytes[0]);
    } else if (strcmp(dataTypeStr, "si16") == 0) {
        int16_t raw_value; memcpy(&raw_value, output.bytes, 2);
        raw_value = OSSwapLittleToHostInt16(raw_value);
        *value = (float)raw_value;
    } else if (strcmp(dataTypeStr, "flag") == 0) {
        *value = (float)output.bytes[0];
    } else {
        return kIOReturnUnsupported;
    }

    return KERN_SUCCESS;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// FetchData retrieves a map of SMC keys and their float values.
func FetchData(keys []string) (map[string]float64, error) {
	var conn C.io_connect_t
	kr := C.smc_open(&conn)
	if kr != C.KERN_SUCCESS {
		return nil, fmt.Errorf("failed to open SMC connection: %d", kr)
	}
	defer C.smc_close(conn)

	results := make(map[string]float64)

	for _, key := range keys {
		var value C.float
		ckey := C.CString(key)
		defer C.free(unsafe.Pointer(ckey))

		kr = C.smc_read_key(conn, ckey, &value)
		if kr == C.KERN_SUCCESS {
			results[key] = float64(value)
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("SMC read succeeded but no keys returned values")
	}

	return results, nil
}
