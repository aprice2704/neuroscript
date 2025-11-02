// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Defines sentinel errors for the provider package.
// filename: pkg/provider/errors.go
// nlines: 12
// risk_rating: LOW

package provider

import "errors"

var (
	// ErrProviderNotFound is returned when a requested provider is not registered.
	ErrProviderNotFound = errors.New("provider not found")
)
