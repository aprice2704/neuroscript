// filename: pkg/testutil/testing_helpers_test.go
package testutil

import "github.com/aprice2704/neuroscript/pkg/lang"

// Import errors package for errors.Is

// Keep for deepEqualWithTolerance
// Keep for deepEqual comparison

// Logger/Adapter imports likely not needed if setup moved to helpers.go
// "github.com/aprice2704/neuroscript/pkg/core/token" // Ensure token is imported if lang.Position needed directly - Assuming it's available via core package implicitly or defined in helpers.go context

// --- Placeholders for other helpers potentially defined in original ---
// func runValidationTestCases(...) { ... }

// Ensure core errors are accessible if needed by helpers here
var (
	_	= lang.ErrValidationArgCount
	_	= lang.ErrValidationRequiredArgNil
	_	= lang.ErrValidationTypeMismatch
	// Add other error variables used here if needed
	_	= lang.ErrMustConditionFailed
	_	= lang.ErrTypeAssertionFailed	// Example if used internally
)

// --- END FILE ---