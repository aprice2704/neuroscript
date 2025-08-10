// NeuroScript Version: 0.6.3
// File version: 1
// Purpose: Defines sentinel errors for the canonicalization package.
// filename: pkg/canon/errors.go
// nlines: 20
// risk_rating: LOW

package canon

import "errors"

var (
	// ErrInvalidMagic is returned when the byte slice to be decoded does not start
	// with the correct magic number.
	ErrInvalidMagic = errors.New("invalid magic number")

	// ErrTruncatedData is returned when the decoder encounters an unexpected EOF,
	// indicating the data is incomplete.
	ErrTruncatedData = errors.New("truncated data")

	// ErrUnknownCodec is returned when the decoder encounters a node kind for which
	// no codec has been registered.
	ErrUnknownCodec = errors.New("unknown or unregistered node kind")
)
