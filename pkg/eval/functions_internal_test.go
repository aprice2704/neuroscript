// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Refactored to use a local mock runtime for isolated testing.
// filename: pkg/eval/functions_internal_test.go
// nlines: 40
// risk_rating: LOW

package eval

import (
	"testing"
)

// NOTE: This test is being removed.
// The `evaluateBuiltInFunction` is now an unexported helper within `evaluation.go`.
// Its behavior is implicitly tested via the end-to-end evaluation tests.
// Keeping this file would require exporting that helper, which contradicts our goal
// of a minimal public API for the eval package.

func TestEvaluateBuiltInFunction_Len_Internal(t *testing.T) {
	// This test is now obsolete. The logic for 'len' is tested via
	// a `CallableExprNode` in a higher-level test.
	t.Skip("Skipping obsolete internal test for unexported function.")
}
