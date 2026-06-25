//go:build darwin

package powerkit

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Foundation -lobjc
#import <Foundation/Foundation.h>
#import <objc/message.h>
#import <objc/runtime.h>
#import <dlfcn.h>
#include <stdlib.h>

static char* pg_copy_cstring(NSString *s) {
	if (!s) {
		return NULL;
	}
	const char *utf8 = [s UTF8String];
	if (!utf8) {
		return NULL;
	}
	return strdup(utf8);
}

static id pg_powerui_smart_charge_client(void) {
	dlopen("/System/Library/PrivateFrameworks/PowerUI.framework/PowerUI", RTLD_LAZY | RTLD_LOCAL);

	Class cls = NSClassFromString(@"PowerUISmartChargeClient");
	if (!cls) {
		return nil;
	}

	SEL sharedSelectors[] = {
		@selector(sharedClient),
		@selector(sharedInstance),
		@selector(defaultClient),
	};
	for (NSUInteger i = 0; i < sizeof(sharedSelectors) / sizeof(sharedSelectors[0]); i++) {
		SEL sel = sharedSelectors[i];
		if ([cls respondsToSelector:sel]) {
			id (*send)(id, SEL) = (id (*)(id, SEL))objc_msgSend;
			id client = send((id)cls, sel);
			if (client) {
				return client;
			}
		}
	}

	return [[cls alloc] init];
}

static int pg_powerui_available_charge_limits(int *limits, int max_count, int *writable, char **error_message) {
	@autoreleasepool {
		if (writable) {
			*writable = 0;
		}
		if (error_message) {
			*error_message = NULL;
		}

		id client = pg_powerui_smart_charge_client();
		if (!client) {
			if (error_message) {
				*error_message = strdup("PowerUISmartChargeClient unavailable");
			}
			return -1;
		}

		SEL availableSel = @selector(availableChargeLimitsWithError:);
		SEL setSel = @selector(setMCLLimit:error:);
		if (![client respondsToSelector:availableSel]) {
			if (error_message) {
				*error_message = strdup("availableChargeLimitsWithError: unavailable");
			}
			return -1;
		}
		if (writable && [client respondsToSelector:setSel]) {
			*writable = 1;
		}

		NSError *error = nil;
		NSArray *(*send)(id, SEL, NSError **) = (NSArray *(*)(id, SEL, NSError **))objc_msgSend;
		NSArray *values = send(client, availableSel, &error);
		if (!values) {
			if (error_message) {
				*error_message = pg_copy_cstring(error ? [error localizedDescription] : @"availableChargeLimitsWithError: returned nil");
			}
			return -1;
		}

		int count = 0;
		for (id value in values) {
			if (![value respondsToSelector:@selector(integerValue)]) {
				continue;
			}
			if (count >= max_count) {
				break;
			}
			limits[count] = (int)[value integerValue];
			count++;
		}
		return count;
	}
}

static int pg_powerui_set_charge_limit(int limit, char **error_message) {
	@autoreleasepool {
		if (error_message) {
			*error_message = NULL;
		}

		id client = pg_powerui_smart_charge_client();
		if (!client) {
			if (error_message) {
				*error_message = strdup("PowerUISmartChargeClient unavailable");
			}
			return 0;
		}

		SEL setSel = @selector(setMCLLimit:error:);
		if (![client respondsToSelector:setSel]) {
			if (error_message) {
				*error_message = strdup("setMCLLimit:error: unavailable");
			}
			return 0;
		}

		NSError *error = nil;
		BOOL (*send)(id, SEL, NSInteger, NSError **) = (BOOL (*)(id, SEL, NSInteger, NSError **))objc_msgSend;
		BOOL ok = send(client, setSel, (NSInteger)limit, &error);
		if (!ok && error_message) {
			*error_message = pg_copy_cstring(error ? [error localizedDescription] : @"setMCLLimit:error: returned false");
		}
		return ok ? 1 : 0;
	}
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func probeNativeChargeLimit() nativeChargeLimitProbeResult {
	limits := make([]C.int, 16)
	var writable C.int
	var errorMessage *C.char

	count := C.pg_powerui_available_charge_limits(&limits[0], C.int(len(limits)), &writable, &errorMessage)
	if errorMessage != nil {
		defer C.free(unsafe.Pointer(errorMessage))
	}
	if count <= 0 {
		return nativeChargeLimitProbeResult{
			Available: false,
			Writable:  false,
			Reason:    cStringOr(errorMessage, chargeLimitReasonNativeUnavailable),
		}
	}

	allowed := make([]int, 0, int(count))
	for i := 0; i < int(count); i++ {
		allowed = append(allowed, int(limits[i]))
	}
	return nativeChargeLimitProbeResult{
		Available:       true,
		Writable:        writable != 0,
		AllowedPercents: allowed,
		Reason:          chargeLimitReasonNativePowerUI,
	}
}

func setNativeChargeLimit(percent int) error {
	var errorMessage *C.char
	ok := C.pg_powerui_set_charge_limit(C.int(percent), &errorMessage)
	if errorMessage != nil {
		defer C.free(unsafe.Pointer(errorMessage))
	}
	if ok == 0 {
		return fmt.Errorf("%w: failed to set native macOS charge limit to %d%%: %s", ErrNotSupported, percent, cStringOr(errorMessage, "unknown PowerUI error"))
	}
	return nil
}

func cStringOr(s *C.char, fallback string) string {
	if s == nil {
		return fallback
	}
	if value := C.GoString(s); value != "" {
		return value
	}
	return fallback
}
