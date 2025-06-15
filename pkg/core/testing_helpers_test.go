// filename: pkg/core/testing_helpers_test.go
package core

// Import errors package for errors.Is

// Keep for deepEqualWithTolerance
// Keep for deepEqual comparison

// Logger/Adapter imports likely not needed if setup moved to helpers.go
// "github.com/aprice2704/neuroscript/pkg/core/token" // Ensure token is imported if Position needed directly - Assuming it's available via core package implicitly or defined in helpers.go context

// --- Placeholders for other helpers potentially defined in original ---
// func runValidationTestCases(...) { ... }

// Ensure core errors are accessible if needed by helpers here
var (
	_ = ErrValidationArgCount
	_ = ErrValidationRequiredArgNil
	_ = ErrValidationTypeMismatch
	// Add other error variables used here if needed
	_ = ErrMustConditionFailed
	_ = ErrTypeAssertionFailed // Example if used internally
)

// --- END FILE ---
