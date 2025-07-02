// pkg/neurodata/checklist/defined_errors.go
package checklist

import "errors"

// --- Defined Errors for Checklist Parsing ---

var (
	// ErrScannerFailed indicates an error occurred during the text scanning process.
	ErrScannerFailed = errors.New("checklist scanner failed")

	// ErrMetadataExtraction indicates an error occurred during metadata extraction.
	ErrMetadataExtraction = errors.New("metadata extraction failed")

	// ErrInternalParser represents unexpected internal states during parsing.
	ErrInternalParser = errors.New("internal checklist parser error")

	// ErrMalformedItem indicates an item line was recognized but had invalid syntax within the delimiters.
	ErrMalformedItem = errors.New("malformed checklist item")

	// ErrNoContent indicates the input contained no valid checklist items or metadata. // <<< ADDED
	ErrNoContent = errors.New("checklist contains no valid items or metadata") // <<< ADDED
)
