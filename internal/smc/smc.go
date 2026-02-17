//go:build darwin

// Package smc provides internal access to the System Management Controller.
package smc

/*
#cgo LDFLAGS: -framework CoreFoundation -framework IOKit

#include <CoreFoundation/CoreFoundation.h>
#include <IOKit/IOKitLib.h>
#include <string.h>

#define KERNEL_INDEX_SMC 2
#define SMC_CMD_READ_BYTES 5
#define SMC_CMD_WRITE_BYTES 6
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

// Reads the raw metadata and byte value for a given SMC key.
kern_return_t smc_read_key_raw(
    io_connect_t conn,
    const char* key,
    char* dataTypeResult,
    unsigned char* bytesResult,
    UInt32* dataSizeResult
);

// Writes raw bytes to a given SMC key.
kern_return_t smc_write_key(
    io_connect_t conn,
    const char* key,
    const unsigned char* bytes,
    UInt32 dataSize
);

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

// --- SMC_READ_KEY_RAW Function ---
kern_return_t smc_read_key_raw(
    io_connect_t conn,
    const char* key,
    char* dataTypeResult,
    unsigned char* bytesResult,
    UInt32* dataSizeResult
) {
    SMCKeyData_t input;
    SMCKeyData_t output;
    size_t structSize = sizeof(SMCKeyData_t);

    memset(&input, 0, structSize);
    memset(&output, 0, structSize);
    input.key = str_to_key(key);
    input.data8 = SMC_CMD_READ_KEYINFO;
    kern_return_t kr = smc_call(conn, &input, &output);
    if (kr != KERN_SUCCESS || output.result != 0) return KERN_FAILURE;

    input.keyInfo = output.keyInfo;
    input.data8 = SMC_CMD_READ_BYTES;
    memset(&output, 0, structSize);
    kr = smc_call(conn, &input, &output);
    if (kr != KERN_SUCCESS || output.result != 0) return KERN_FAILURE;

    // Export the raw results instead of decoding
    key_to_str(input.keyInfo.dataType, dataTypeResult);
    *dataSizeResult = input.keyInfo.dataSize;
    if (*dataSizeResult > 32) {
        return kIOReturnBadArgument;
    }
    memcpy(bytesResult, output.bytes, *dataSizeResult);

    return KERN_SUCCESS;
}

// --- SMC_WRITE_KEY Function ---
kern_return_t smc_write_key(
    io_connect_t conn,
    const char* key,
    const unsigned char* bytesToWrite,
    UInt32 dataSize
) {
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

    SMCKeyData_keyInfo_t keyInfo = output.keyInfo;
    if (keyInfo.dataSize > 32) {
        return kIOReturnBadArgument;
    }

    if (dataSize != keyInfo.dataSize) {
        return kIOReturnBadArgument;
    }

    memset(&input, 0, structSize);
    memset(&output, 0, structSize);

    input.key = str_to_key(key);
    input.data8 = SMC_CMD_WRITE_BYTES;
    input.keyInfo = keyInfo;
    memcpy(input.bytes, bytesToWrite, dataSize);

    kr = smc_call(conn, &input, &output);
    if (kr != KERN_SUCCESS || output.result != 0) {
        if (output.result != 0) {
            return (kern_return_t)output.result;
        }
        return kr;
    }

    return KERN_SUCCESS;
}

*/
import "C"

import (
	"fmt"
	"log"
	"sync"
	"unsafe"
)

var (
	smcMu       sync.Mutex
	smcConn     C.io_connect_t
	smcInitOnce sync.Once
	smcInitErr  error
)

// initSMC is the internal function that performs the connection.
func initSMC() {
	kr := C.smc_open(&smcConn)
	if kr != C.KERN_SUCCESS {
		smcInitErr = fmt.Errorf("failed to open SMC connection with kern_return: %d", kr)
	}
}

// getSMCConnection ensures the SMC connection is open and returns it.
func getSMCConnection() (C.io_connect_t, error) {
	smcInitOnce.Do(initSMC)
	return smcConn, smcInitErr
}

// CloseConnection explicitly closes the shared SMC connection.
func CloseConnection() {
	smcMu.Lock()
	defer smcMu.Unlock()

	if smcConn != 0 {
		C.smc_close(smcConn)
		smcConn = 0
		smcInitOnce = sync.Once{}
		smcInitErr = nil
	}
}

// RawSMCValue holds the raw, undecoded result of an SMC query.
type RawSMCValue struct {
	DataType string
	DataSize int
	Data     []byte
}

// FetchData retrieves a map of SMC keys and their decoded float values.
// It works by first fetching the raw data and then decoding it.
func FetchData(keys []string) (map[string]float64, error) {
	rawResults, err := FetchRawData(keys)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch raw SMC data: %w", err)
	}

	decodedResults := make(map[string]float64, len(rawResults))
	for key, rawValue := range rawResults {
		// Use the Go-side translator to decode the value.
		decodedValue, err := decodeSMCValue(rawValue.DataType, rawValue.Data)
		if err == nil {
			decodedResults[key] = decodedValue
		} else {
			// Provide visibility when a key cannot be decoded.
			log.Printf("powerkit/smc: failed to decode key %s (type %s): %v", key, rawValue.DataType, err)
		}
		// Note: Keys with unsupported data types are silently ignored.
	}

	if len(decodedResults) == 0 {
		return nil, fmt.Errorf("SMC read succeeded but no keys could be decoded to a float value")
	}

	return decodedResults, nil
}

// FetchRawData retrieves the raw, undecoded metadata and byte values for a given list of SMC keys.
func FetchRawData(keys []string) (map[string]RawSMCValue, error) {
	smcMu.Lock()
	defer smcMu.Unlock()

	conn, err := getSMCConnection()
	if err != nil {
		return nil, err
	}
	if conn == 0 {
		return nil, fmt.Errorf("SMC connection is not valid")
	}

	results := make(map[string]RawSMCValue, len(keys))
	for _, key := range keys {
		ckey := C.CString(key)

		var dataTypeResult [5]C.char
		var bytesResult [32]C.uchar
		var dataSizeResult C.UInt32

		kr := C.smc_read_key_raw(
			conn,
			ckey,
			&dataTypeResult[0],
			&bytesResult[0],
			&dataSizeResult,
		)

		if kr == C.KERN_SUCCESS {
			dataSize := int(dataSizeResult)
			goBytes := C.GoBytes(unsafe.Pointer(&bytesResult[0]), C.int(dataSize))

			results[key] = RawSMCValue{
				DataType: C.GoString(&dataTypeResult[0]),
				DataSize: dataSize,
				Data:     goBytes,
			}
		}
		C.free(unsafe.Pointer(ckey))
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("SMC raw read failed for all requested keys")
	}
	return results, nil
}

// WriteData writes raw bytes to a given SMC key.
func WriteData(key string, data []byte) error {
	// This function opens its own connection to ensure write operations
	// are atomic and don't interfere with concurrent reads on the shared connection.
	var conn C.io_connect_t
	kr := C.smc_open(&conn)
	if kr != C.KERN_SUCCESS {
		return fmt.Errorf("failed to open SMC connection for writing: %d", kr)
	}
	defer C.smc_close(conn)

	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))

	if len(data) == 0 {
		return fmt.Errorf("cannot write empty data slice")
	}
	cBytes := (*C.uchar)(unsafe.Pointer(&data[0]))
	cDataSize := C.UInt32(len(data))

	kr = C.smc_write_key(conn, ckey, cBytes, cDataSize)

	if kr != C.KERN_SUCCESS {
		if kr == C.kIOReturnBadArgument {
			return fmt.Errorf("SMC write failed for key '%s': provided data size does not match key's expected size", key)
		}
		return fmt.Errorf("SMC write failed for key '%s' with kern_return code: %d", key, kr)
	}

	return nil
}
