//go:build darwin

package powerkit

import (
	"errors"
	"fmt"
	"os"
)

var (
	// ErrPermissionRequired indicates the operation needs elevated privileges.
	ErrPermissionRequired = errors.New("permission required")
	// ErrNotSupported indicates a feature is unavailable on the current hardware/system.
	ErrNotSupported = errors.New("not supported")
	// ErrTransientIO indicates a temporary operating-system I/O failure.
	ErrTransientIO = errors.New("transient io failure")
)

func requireRoot(op string) error {
	if os.Geteuid() == 0 {
		return nil
	}
	return fmt.Errorf("%w: %s requires root", ErrPermissionRequired, op)
}
