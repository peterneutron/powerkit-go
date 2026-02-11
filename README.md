# PowerKit-Go

`powerkit-go` is a Darwin-only Go library for reading and controlling macOS power-related state via IOKit/SMC and OS power interfaces.

## Status
Pre-release. Public APIs can still change.

## Platform
- macOS only (`//go:build darwin`)
- cgo required (links CoreFoundation/IOKit/Foundation)

## Install
```bash
go get github.com/peterneutron/powerkit-go
```

## Core API
Public package: `github.com/peterneutron/powerkit-go/pkg/powerkit`

### Telemetry
- `GetSystemInfo(opts ...FetchOptions) (*SystemInfo, error)`
- `GetSystemInfoContext(ctx, opts ...FetchOptions) (*SystemInfo, error)`
- `GetRawSMCValues(keys []string) (map[string]RawSMCValue, error)`
- `StreamSystemEvents() (<-chan SystemEvent, error)`

### Control APIs (privileged)
These return `ErrPermissionRequired` when not root.

- Charging/adapter:
  - `SetChargingState(ChargingAction)`
  - `SetAdapterState(AdapterAction)`
- MagSafe LED:
  - `SetMagsafeLEDState(MagsafeLEDState)`
  - `GetMagsafeLEDState() (state MagsafeLEDState, available bool, err error)`
  - `GetMagsafeStatus() (MagsafeStatus, error)`
- Low Power Mode:
  - `GetLowPowerModeEnabled() (enabled bool, available bool, err error)`
  - `SetLowPowerMode(enable bool)`
  - `ToggleLowPowerMode()`

Context variants are available for mutating calls (`*Context` methods).

### Sleep assertions (no root required)
- `CreateAssertion(AssertionType, reason string) (AssertionID, error)`
- `ReleaseAssertion(AssertionType)`
- `AllowAllSleep()`
- `IsAssertionActive(AssertionType) bool`
- `GetAssertionID(AssertionType) (AssertionID, bool)`

## Typed Errors
- `ErrPermissionRequired`
- `ErrNotSupported`
- `ErrTransientIO`

## Minimal Example
```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/peterneutron/powerkit-go/pkg/powerkit"
)

func main() {
	info, err := powerkit.GetSystemInfoContext(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(info.IOKit.Battery.CurrentCharge)
}
```

## CLI (`powerkit-cli`)
Build:
```bash
go build -o powerkit-cli ./cmd/powerkit-cli
```

Examples:
```bash
./powerkit-cli all
./powerkit-cli watch
./powerkit-cli raw FNum ACLC
sudo ./powerkit-cli charging off
sudo ./powerkit-cli magsafe set-color amber
./powerkit-cli lowpower get
```

## Development
From `powerkit-go/`:

- `make tests`
- `make vet`
- `make lint`
- `make verify`

## Safety Notes
- SMC writes can affect charging/adapter behavior. Use write APIs carefully.
- Prefer read APIs and assertions unless you specifically need SMC control.
