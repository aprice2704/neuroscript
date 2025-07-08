// NeuroScript Version: 0.4.1
// File version: 1
// Purpose: Provides comprehensive unit tests for the time tool package.
// filename: pkg/tool/time/tools_time_test.go
// nlines: 121
// risk_rating: LOW

package time

import (
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/tool"
)

// TestAdapterFunctions verifies the top-level adapter functions that bridge the interpreter
// and the tool's implementation.
func TestAdapterFunctions(t *testing.T) {
	// The adapter functions require a tool.Runtime, but it's not used in the current
	// implementation. We can pass nil.
	var mockInterpreter tool.Runtime = nil

	t.Run("adaptToolTimeNow", func(t *testing.T) {
		t.Run("happy path - no args", func(t *testing.T) {
			val, err := adaptToolTimeNow(mockInterpreter, []interface{}{})
			if err != nil {
				t.Fatalf("adaptToolTimeNow() with valid args returned an unexpected error: %v", err)
			}
			if _, ok := val.(time.Time); !ok {
				t.Errorf("adaptToolTimeNow() did not return a time.Time value, got %T", val)
			}
		})

		t.Run("unhappy path - wrong arg count", func(t *testing.T) {
			_, err := adaptToolTimeNow(mockInterpreter, []interface{}{"invalid arg"})
			if err == nil {
				t.Error("adaptToolTimeNow() with invalid args did not return an error")
			}
		})
	})

	t.Run("adaptToolTimeSleep", func(t *testing.T) {
		testCases := []struct {
			name    string
			args    []interface{}
			wantErr bool
		}{
			{"happy path", []interface{}{0.01}, false},
			{"unhappy path - wrong arg count", []interface{}{}, true},
			{"unhappy path - wrong arg type", []interface{}{"string"}, true},
			{"unhappy path - negative duration", []interface{}{-1.0}, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := adaptToolTimeSleep(mockInterpreter, tc.args)
				if (err != nil) != tc.wantErr {
					t.Errorf("adaptToolTimeSleep() error = %v, wantErr %v", err, tc.wantErr)
				}
			})
		}
	})
}

// TestValidationFunctions ensures the argument validation logic is correct for each tool.
func TestValidationFunctions(t *testing.T) {
	t.Run("validateTimeNow", func(t *testing.T) {
		testCases := []struct {
			name    string
			args    []interface{}
			wantErr bool
		}{
			{"correct number of args (0)", []interface{}{}, false},
			{"incorrect number of args (1)", []interface{}{"foo"}, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := validateTimeNow(tc.args)
				if (err != nil) != tc.wantErr {
					t.Errorf("validateTimeNow() error = %v, wantErr %v", err, tc.wantErr)
				}
			})
		}
	})

	t.Run("validateTimeSleep", func(t *testing.T) {
		testCases := []struct {
			name    string
			args    []interface{}
			wantErr bool
		}{
			{"correct args (float64)", []interface{}{1.5}, false},
			{"incorrect arg type (int)", []interface{}{1}, true},
			{"incorrect arg type (string)", []interface{}{"1.5"}, true},
			{"incorrect number of args (0)", []interface{}{}, true},
			{"incorrect number of args (2)", []interface{}{1.0, 2.0}, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := validateTimeSleep(tc.args)
				if (err != nil) != tc.wantErr {
					t.Errorf("validateTimeSleep() error = %v, wantErr %v", err, tc.wantErr)
				}
			})
		}
	})
}

// TestImplementationFunctions tests the raw tool logic.
func TestImplementationFunctions(t *testing.T) {
	t.Run("implTimeNow", func(t *testing.T) {
		now, err := implTimeNow()
		if err != nil {
			t.Fatalf("implTimeNow() returned an unexpected error: %v", err)
		}
		if time.Since(now) > time.Second {
			t.Errorf("implTimeNow() returned a time that is too far in the past: %v", now)
		}
	})

	t.Run("implTimeSleep", func(t *testing.T) {
		testCases := []struct {
			name       string
			duration   float64
			wantErr    bool
			minSleepMS int64 // Minimum expected sleep time in milliseconds
		}{
			{"positive duration", 0.02, false, 15}, // Sleep 20ms, expect >= 15ms
			{"zero duration", 0, false, 0},
			{"negative duration", -1.0, true, 0},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				start := time.Now()
				_, err := implTimeSleep(tc.duration)
				elapsed := time.Since(start)

				if (err != nil) != tc.wantErr {
					t.Errorf("implTimeSleep() error = %v, wantErr %v", err, tc.wantErr)
					return
				}

				if !tc.wantErr && elapsed.Milliseconds() < tc.minSleepMS {
					t.Errorf("implTimeSleep() slept for %v, less than min expected %dms", elapsed, tc.minSleepMS)
				}
			})
		}
	})
}
