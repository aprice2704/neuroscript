// NeuroScript Version: 0.7.1
// File version: 1
// Purpose: Defines sentinel errors for the capsule tool package.
// filename: pkg/tool/capsule/errors.go
// nlines: 12
// risk_rating: LOW
package capsule

import "errors"

var (
	// ErrAdminRegistryNotAvailable is returned when a tool requires a writable
	// capsule registry, but the runtime does not provide one.
	ErrAdminRegistryNotAvailable = errors.New("runtime does not provide an admin CapsuleRegistry")
)
