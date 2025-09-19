// NeuroScript Version: 0.7.2
// File version: 2
// Purpose: Defines sentinel errors for the capsule tool package, including one for invalid input data.
// filename: pkg/tool/capsule/errors.go
// nlines: 15
// risk_rating: LOW
package capsule

import "errors"

var (
	// ErrAdminRegistryNotAvailable is returned when a tool requires a writable
	// capsule registry, but the runtime does not provide one.
	ErrAdminRegistryNotAvailable = errors.New("runtime does not provide an admin CapsuleRegistry")

	// ErrInvalidCapsuleData is returned by the Add tool when the provided
	// map is missing required fields or has incorrect types.
	ErrInvalidCapsuleData = errors.New("invalid capsule data: map must contain non-empt 'id' and 'version' string fields")
)
