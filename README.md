# PowerKit-Go

`powerkit-go` is a Darwin-only Go library for reading and controlling macOS power state through IOKit, SMC, and OS power APIs.

## Scope

Use `powerkit-go` when you need:

- battery, adapter, and charging telemetry
- charging and adapter control
- MagSafe LED control
- Low Power Mode read/write
- sleep assertion control
- event-driven power, sleep, and wake notifications

This library targets Apple Silicon Macs. Intel behavior is not guaranteed.

## Status

- OS: macOS only
- Build tag: `//go:build darwin`
- cgo required
- Mutating control APIs require root

## Install

```bash
go get github.com/peterneutron/powerkit-go
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/peterneutron/powerkit-go/pkg/powerkit"
)

func main() {
	info, err := powerkit.GetSystemInfo()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Charge: %d%%\n", info.IOKit.Battery.CurrentCharge)
	fmt.Printf("Charging: %v\n", info.IOKit.State.IsCharging)
}
```

## Public API

Main package: `github.com/peterneutron/powerkit-go/pkg/powerkit`

Core entrypoints:

- `GetSystemInfo(opts ...FetchOptions) (*SystemInfo, error)`
- `GetSystemInfoContext(ctx context.Context, opts ...FetchOptions) (*SystemInfo, error)`
- `StreamSystemEvents() (<-chan SystemEvent, error)`
- `StreamSystemEventsWithHooks(StreamHooks) (<-chan SystemEvent, error)`
- `GetRawSMCValues(keys []string) (map[string]RawSMCValue, error)`

Control APIs:

- `SetChargingState(ChargingAction) error`
- `SetAdapterState(AdapterAction) error`
- `SetMagsafeLEDState(MagsafeLEDState) error`
- `GetMagsafeLEDState() (state MagsafeLEDState, available bool, err error)`
- `GetLowPowerModeEnabled() (enabled bool, available bool, err error)`
- `SetLowPowerMode(enable bool) error`

Sleep assertions:

- `CreateAssertion(AssertionType, reason string) (AssertionID, error)`
- `ReleaseAssertion(AssertionType)`
- `AllowAllSleep()`

## Event Stream Notes

`StreamSystemEventsWithHooks` supports a synchronous `BeforeSleep` hook.

Use it only for short, bounded pre-sleep work. macOS sleep acknowledgement waits for that hook to return.

## Privileges

- Read telemetry: no root required
- Sleep assertions: no root required
- Charging, adapter, MagSafe, and Low Power Mode writes: root required

## Build

```bash
go build ./...
```

## Verify

```bash
make tests
make vet
make lint
make verify
```

## Docs

Keep the README short. Detailed material lives elsewhere:

- [Contract Details](docs/contracts.md)
- [Release Process](docs/release.md)
- [Agent Instructions](AGENTS.md)

## Release Model

- semver tags on `master`
- additive API changes allowed in patch releases
- breaking API changes require a minor or major bump
- release notes belong in tags and Git hosting releases, not an in-repo changelog

## Safety

This library can change charging and adapter behavior at the hardware-control layer. Prefer read APIs unless you explicitly need mutation.
