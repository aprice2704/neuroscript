// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Provides a failing test case to demonstrate that string tools requiring integer arguments do not correctly coerce floating-point inputs.
// filename: pkg/tool/strtools/tools_string_coercion_test.go
// nlines: 60
// risk_rating: LOW

package strtools

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
)

func TestToolStringArgumentCoercion(t *testing.T) {
	interp := interpreter.NewInterpreter()
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		// --- Substring Coercion Tests ---
		{
			name:       "Substring with float64 start index",
			toolName:   "Substring",
			args:       MakeArgs("hello world", float64(6.0), float64(5.0)),
			wantResult: "world",
			// This test is designed to FAIL. Currently, it will error out with an argument mismatch.
			// Once the fix is applied, this test should PASS.
			wantErrIs: nil,
		},
		{
			name:       "Substring with float64 length",
			toolName:   "Substring",
			args:       MakeArgs("hello", float64(0.0), float64(4.0)),
			wantResult: "hell",
			wantErrIs:  nil,
		},
		{
			name:       "Substring with non-integer float",
			toolName:   "Substring",
			args:       MakeArgs("hello", float64(1.5), float64(3.5)),
			wantResult: "ell", // Should truncate to start=1, length=3
			wantErrIs:  nil,
		},

		// --- Replace Coercion Tests ---
		{
			name:       "Replace with float64 count",
			toolName:   "Replace",
			args:       MakeArgs("ababab", "ab", "cd", float64(2.0)),
			wantResult: "cdcdab",
			wantErrIs:  nil,
		},
		{
			name:       "Replace with non-integer float count",
			toolName:   "Replace",
			args:       MakeArgs("ababab", "ab", "cd", float64(2.7)),
			wantResult: "cdcdab", // Should truncate to count=2
			wantErrIs:  nil,
		},
	}
	for _, tt := range tests {
		// We use the existing helper, which will show the failure.
		testStringToolHelper(t, interp, tt)
	}
}
