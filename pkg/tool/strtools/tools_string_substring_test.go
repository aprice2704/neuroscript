// NeuroScript Version: 0.4.0
// File version: 2
// Purpose: Provides additional, focused edge-case tests for the Substring tool.
// filename: pkg/tool/strtools/tools_string_substring_test.go
// nlines: 57
// risk_rating: LOW

package strtools

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestToolSubstringEdgeCases(t *testing.T) {
	interp := newStringTestInterpreter(t)
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Zero_Length", toolName: "Substring", args: MakeArgs("abcdef", int64(2), int64(0)), wantResult: ""},
		{
			name:       "Start_At_End",
			toolName:   "Substring",
			args:       MakeArgs("abcdef", int64(6), int64(1)),
			wantResult: "",
			// Start index is clamped to the end, resulting in an empty string.
		},
		{
			name:       "Start_Past_End",
			toolName:   "Substring",
			args:       MakeArgs("abcdef", int64(10), int64(1)),
			wantResult: "",
			// Start index is clamped to the end, resulting in an empty string.
		},
		{
			name:     "Length_Exceeds_String",
			toolName: "Substring",
			args:     MakeArgs("abcdef", int64(3), int64(10)),
			// End index is clamped to the end of the string.
			wantResult: "def",
		},
		{
			name:       "Start_Plus_Length_Is_End",
			toolName:   "Substring",
			args:       MakeArgs("abcdef", int64(2), int64(4)),
			wantResult: "cdef",
		},
		{
			name:      "Validation_Negative_Start",
			toolName:  "Substring",
			args:      MakeArgs("abcdef", int64(-1), int64(3)),
			wantErrIs: lang.ErrListIndexOutOfBounds,
		},
		{
			name:      "Validation_Negative_Length",
			toolName:  "Substring",
			args:      MakeArgs("abcdef", int64(1), int64(-1)),
			wantErrIs: lang.ErrListIndexOutOfBounds,
		},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, tt)
	}
}
