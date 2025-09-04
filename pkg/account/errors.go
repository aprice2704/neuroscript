// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Defines sentinel errors for the account package.
// filename: pkg/account/errors.go
// nlines: 12
// risk_rating: LOW

package account

import "errors"

var (
	// ErrInvalidConfiguration is returned when an account configuration is missing
	// required fields or contains invalid values.
	ErrInvalidConfiguration = errors.New("invalid account configuration")
)
