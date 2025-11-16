// NeuroScript Version: 0.7.2
// File version: 3
// Purpose: Defines sentinel errors for the capsule tool package, including one for invalid input data.
// Latest change: Removed ErrAdminRegistryNotAvailable as it's no longer used.
// filename: pkg/tool/capsule/errors.go
// nlines: 11
// risk_rating: LOW
package capsule

import "errors"

var (
	// ErrAdminRegistryNotAvailable -- REMOVED. The admin registry is now the store.

	// ErrInvalidCapsuleData is returned by the Add tool when the provided
	// map is missing required fields or has incorrect types.
	ErrInvalidCapsuleData = errors.New("invalid capsule data: map must contain non-empt 'id' and 'version' string fields")
)
