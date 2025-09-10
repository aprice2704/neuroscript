// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Contains functional tests for the OS process and time tools.
// filename: pkg/tool/os/tools_os_proc_test.go
// nlines: 70
// risk_rating: MEDIUM

package os_test

import (
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestToolOsProcFunctional(t *testing.T) {
	tests := []osTestCase{
		{
			name:       "Success: Now returns current time",
			toolName:   "Now",
			args:       []interface{}{},
			wantResult: float64(time.Now().Unix()),
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
				if err != nil {
					t.Fatalf("Now() returned unexpected error: %v", err)
				}
				res, ok := result.(float64)
				if !ok {
					t.Fatalf("Now() did not return a float64, got %T", result)
				}
				// Check if the timestamp is within a reasonable delta (e.g., 2 seconds)
				if time.Now().Unix()-int64(res) > 2 {
					t.Errorf("Now() result %v is too far in the past", res)
				}
			},
		},
		{
			name:     "Success: Sleep with valid duration",
			toolName: "Sleep",
			args:     []interface{}{0.01},
			// No result to check, just absence of error
		},
		{
			name:          "Fail: Sleep duration exceeds policy limit",
			toolName:      "Sleep",
			args:          []interface{}{10.0}, // The default test policy in helpers limits sleep to 5s
			wantToolErrIs: capability.ErrTimeExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testOsToolHelper(t, tt)
		})
	}
}

func TestToolHostnameFunctional(t *testing.T) {
	// Hostname is hard to test against a fixed value, so we just check it runs
	// and returns a non-empty string.
	tt := osTestCase{
		name:     "Success: Hostname returns non-empty string",
		toolName: "Hostname",
		args:     []interface{}{},
		checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
			if err != nil {
				t.Fatalf("Hostname() returned unexpected error: %v", err)
			}
			res, ok := result.(string)
			if !ok {
				t.Fatalf("Hostname() did not return a string, got %T", result)
			}
			if res == "" {
				t.Error("Hostname() returned an empty string")
			}
		},
	}
	t.Run(tt.name, func(t *testing.T) {
		testOsToolHelper(t, tt)
	})
}
