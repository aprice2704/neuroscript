// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Contains functional tests for the OS.Getenv tool.
// filename: pkg/tool/os/tools_os_env_test.go
// nlines: 65
// risk_rating: MEDIUM

package os_test

import (
	"os"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestToolGetenvFunctional(t *testing.T) {
	const testVarName = "NEUROSCRIPT_TEST_VAR"
	const testVarValue = "test_value_123"

	tests := []osTestCase{
		{
			name:     "Success: Get Existing Env Var",
			toolName: "Getenv",
			args:     []interface{}{testVarName},
			setupFunc: func(t *testing.T) error {
				return os.Setenv(testVarName, testVarValue)
			},
			wantResult: testVarValue,
		},
		{
			name:     "Success: Get Non-Existent Env Var",
			toolName: "Getenv",
			args:     []interface{}{"NON_EXISTENT_VAR_AJP"},
			setupFunc: func(t *testing.T) error {
				// Ensure it's not set
				return os.Unsetenv("NON_EXISTENT_VAR_AJP")
			},
			wantResult: "", // Expect empty string, not an error
		},
		{
			name:          "Fail: Missing Argument",
			toolName:      "Getenv",
			args:          []interface{}{},
			wantToolErrIs: lang.ErrArgumentMismatch,
		},
		{
			name:          "Fail: Empty VarName",
			toolName:      "Getenv",
			args:          []interface{}{""},
			wantToolErrIs: lang.ErrInvalidArgument,
		},
		{
			name:          "Fail: Wrong Argument Type",
			toolName:      "Getenv",
			args:          []interface{}{12345},
			wantToolErrIs: lang.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment variable after test
			t.Cleanup(func() {
				os.Unsetenv(testVarName)
			})
			testOsToolHelper(t, tt)
		})
	}
}
