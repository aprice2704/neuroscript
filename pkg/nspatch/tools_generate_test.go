// NeuroScript Version: 0.3.1
// File version: 0.0.4
// Fix compile errors: Correct NewInterpreter call.
// Test file for GeneratePatch tool.
// filename: pkg/nspatch/tools_generate_test.go

package nspatch_test // Use _test package convention

import (
	"errors" // Needed by helper
	// Needed for error formatting
	"io" // Needed by helper
	"reflect"
	"sort"
	"strings"
	"testing"

	// Import packages necessary for setup and testing
	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	nspatch "github.com/aprice2704/neuroscript/pkg/nspatch" // Import the package under test
	// Need toolsets for registration if using init
	// _ "github.com/aprice2704/neuroscript/pkg/toolsets"
)

// --- Test Helper ---

// sortPatchMaps sorts a slice of patch maps (represented as interface{})
// based on line number, then operation type (delete before insert).
func sortPatchMaps(results []interface{}) {
	sort.SliceStable(results, func(i, j int) bool {
		mapI, okI := results[i].(map[string]interface{})
		mapJ, okJ := results[j].(map[string]interface{})
		if !okI || !okJ {
			return false // Should not happen in valid tests
		}

		lineI, okI1 := mapI["line"].(int64)
		lineJ, okJ1 := mapJ["line"].(int64)
		if !okI1 || !okJ1 {
			return false
		}
		if lineI != lineJ {
			return lineI < lineJ
		}

		// If lines are equal, sort by op (e.g., delete before insert for replace)
		opI, okI2 := mapI["op"].(string)
		opJ, okJ2 := mapJ["op"].(string)
		if !okI2 || !okJ2 {
			return false
		}
		// Define order: delete < insert < replace (though replace shouldn't be generated directly)
		order := map[string]int{"delete": 0, "insert": 1, "replace": 2} // Assuming these are the only ops
		// Handle potential unknown ops gracefully during sort comparison
		orderI, okI3 := order[opI]
		orderJ, okJ3 := order[opJ]
		if !okI3 || !okJ3 {
			return opI < opJ // Fallback to string comparison if op unknown
		}
		return orderI < orderJ
	})
}

// --- Test Cases ---

func TestGeneratePatch(t *testing.T) {
	// --- Test Setup ---
	// Use SimpleSlogAdapter as confirmed available
	logger, err := adapters.NewSimpleSlogAdapter(TestingTBLogWriter(t), interfaces.LogLevelDebug) // Use test helper for output
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	logger.Debug("Test logger initialized")

	llmClient := adapters.NewNoOpLLMClient() // Pass logger if constructor accepts it
	// Sandbox directory is irrelevant for this tool
	// CORRECTED: Added nil for libPaths argument
	interpreter, err := NewInterpreter(logger, llmClient, ".", nil, nil) // Use "." as dummy sandbox, nil for initialVars and libPaths
	if err != nil {
		t.Fatalf("Failed to create interpreter: %v", err)
	}

	// Manually register the tool under test
	registry := interpreter.ToolRegistry()
	err = nspatch.RegisterNsPatchTools(registry) // Call the registration func directly
	if err != nil {
		t.Fatalf("Failed to register nspatch tools: %v", err)
	}

	// --- Test Cases ---
	testCases := []struct {
		name       string
		original   string
		modified   string
		path       interface{} // Use interface{} to test nil explicitly
		wantResult []map[string]interface{}
		wantErr    error // Use error type for specific error checks if needed
	}{
		{
			name:       "No Change",
			original:   "line1\nline2\nline3\n",
			modified:   "line1\nline2\nline3\n",
			path:       "file.txt",
			wantResult: []map[string]interface{}{}, // Expect empty patch list
		},
		{
			name:     "Simple Insert",
			original: "line1\nline3\n",
			modified: "line1\nline2-inserted\nline3\n",
			path:     "file.txt",
			wantResult: []map[string]interface{}{
				{"op": "insert", "file": "file.txt", "line": int64(2), "new": "line2-inserted"},
			},
		},
		{
			name:     "Simple Delete",
			original: "line1\nline2-delete\nline3\n",
			modified: "line1\nline3\n",
			path:     "file.txt",
			wantResult: []map[string]interface{}{
				{"op": "delete", "file": "file.txt", "line": int64(2), "old": "line2-delete"},
			},
		},
		{
			name:     "Simple Replace",
			original: "line1\nline2-replace\nline3\n",
			modified: "line1\nline2-new\nline3\n",
			path:     "file.txt",
			wantResult: []map[string]interface{}{
				// A replace is represented as delete + insert at the same line
				{"op": "delete", "file": "file.txt", "line": int64(2), "old": "line2-replace"},
				{"op": "insert", "file": "file.txt", "line": int64(2), "new": "line2-new"},
			},
		},
		{
			name:     "Insert at Start",
			original: "lineA\nlineB\n",
			modified: "line0-inserted\nlineA\nlineB\n",
			path:     "file.txt",
			wantResult: []map[string]interface{}{
				{"op": "insert", "file": "file.txt", "line": int64(1), "new": "line0-inserted"},
			},
		},
		{
			name:     "Insert at End",
			original: "lineA\nlineB\n",
			modified: "lineA\nlineB\nlineC-inserted\n",
			path:     "file.txt",
			wantResult: []map[string]interface{}{
				{"op": "insert", "file": "file.txt", "line": int64(3), "new": "lineC-inserted"},
			},
		},
		{
			name:     "Delete at Start",
			original: "lineA-delete\nlineB\nlineC\n",
			modified: "lineB\nlineC\n",
			path:     "file.txt",
			wantResult: []map[string]interface{}{
				{"op": "delete", "file": "file.txt", "line": int64(1), "old": "lineA-delete"},
			},
		},
		{
			name:     "Delete at End",
			original: "lineA\nlineB\nlineC-delete\n",
			modified: "lineA\nlineB\n",
			path:     "file.txt",
			wantResult: []map[string]interface{}{
				{"op": "delete", "file": "file.txt", "line": int64(3), "old": "lineC-delete"},
			},
		},
		{
			name:     "Insert into Empty",
			original: "",
			modified: "newline1\nnewline2\n",
			path:     "empty.txt",
			wantResult: []map[string]interface{}{
				{"op": "insert", "file": "empty.txt", "line": int64(1), "new": "newline1"},
				{"op": "insert", "file": "empty.txt", "line": int64(2), "new": "newline2"},
			},
		},
		{
			name:     "Delete to Empty",
			original: "line1\nline2\n",
			modified: "",
			path:     "delete_all.txt",
			wantResult: []map[string]interface{}{
				{"op": "delete", "file": "delete_all.txt", "line": int64(1), "old": "line1"},
				{"op": "delete", "file": "delete_all.txt", "line": int64(2), "old": "line2"},
			},
		},
		{
			name:     "Multiple Changes",
			original: "one\ntwo\nthree\nfour\nfive\n",
			modified: "one\ntwo-changed\nfour\nfive-changed\nsix\n",
			path:     "multi.txt",
			wantResult: []map[string]interface{}{
				{"op": "delete", "file": "multi.txt", "line": int64(2), "old": "two"},
				{"op": "insert", "file": "multi.txt", "line": int64(2), "new": "two-changed"},
				{"op": "delete", "file": "multi.txt", "line": int64(3), "old": "three"},
				{"op": "delete", "file": "multi.txt", "line": int64(5), "old": "five"},
				{"op": "insert", "file": "multi.txt", "line": int64(5), "new": "five-changed"},
				{"op": "insert", "file": "multi.txt", "line": int64(6), "new": "six"},
			},
		},
		{
			name:     "No Trailing Newline Original",
			original: "line1",
			modified: "line1\nline2",
			path:     "no_nl.txt",
			wantResult: []map[string]interface{}{
				{"op": "insert", "file": "no_nl.txt", "line": int64(2), "new": "line2"},
			},
		},
		{
			name:     "No Trailing Newline Modified",
			original: "line1\nline2",
			modified: "line1",
			path:     "no_nl.txt",
			wantResult: []map[string]interface{}{
				{"op": "delete", "file": "no_nl.txt", "line": int64(2), "old": "line2"},
			},
		},
		{
			name:     "Path argument is nil", // Ensure nil path arg works
			original: "a\nb",
			modified: "a\nc",
			path:     nil, // Pass nil for path
			wantResult: []map[string]interface{}{
				{"op": "delete", "file": "", "line": int64(2), "old": "b"},
				{"op": "insert", "file": "", "line": int64(2), "new": "c"},
			},
		},
	}

	// --- Run Tests ---
	toolName := "GeneratePatch" // Tool name used in registration and retrieval
	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// These tests are simple string manipulation, parallel is fine
			t.Parallel()

			// Prepare arguments, handle nil path explicitly
			args := []interface{}{tc.original, tc.modified, tc.path}

			// *** FIXED: Execute tool by retrieving from registry and calling Func ***
			registry := interpreter.ToolRegistry()
			toolImpl, found := registry.GetTool(toolName)
			if !found {
				t.Fatalf("Tool '%s' not found in registry", toolName)
			}

			result, runErr := toolImpl.Func(interpreter, args) // Call the tool's function directly

			// --- Error Checking ---
			if tc.wantErr != nil {
				if runErr == nil {
					t.Errorf("Expected error %v, but got nil", tc.wantErr)
				} else {
					// Check if the specific error is wrapped or is the direct cause
					// For tool execution, often check for  nvalidArgument or specific tool errors
					underlyingErr := tc.wantErr
					if !errors.Is(runErr, underlyingErr) {
						// Allow checking for specific argument errors wrapped by ErrInvalidArgument
						if !(errors.Is(runErr, nvalidArgument) && strings.Contains(runErr.Error(), underlyingErr.Error())) {
							t.Errorf("Expected error wrapping %q, but got %q", underlyingErr, runErr)
						}
					}
				}
				// Don't check result if error was expected
				return
			}
			if runErr != nil {
				// Use %+v for potentially more detailed error stack if available
				t.Fatalf("Did not expect error, but got: %+v", runErr)
			}

			// --- Result Comparison ---
			actualResults, ok := result.([]interface{})
			if !ok {
				t.Fatalf("Expected result type []interface{}, but got %T: %v", result, result)
			}

			// Convert expected result []map[string]interface{} to []interface{} for sorting helper
			wantResultInterfaces := make([]interface{}, len(tc.wantResult))
			for i, v := range tc.wantResult {
				wantResultInterfaces[i] = v
			}

			// Sort both actual and expected results
			sortPatchMaps(actualResults)
			sortPatchMaps(wantResultInterfaces) // Sort the interface slice representation

			// Compare using reflect.DeepEqual
			if !reflect.DeepEqual(actualResults, wantResultInterfaces) {
				t.Errorf("Result mismatch:\nExpected: %#v\nGot:      %#v", wantResultInterfaces, actualResults)
				// Optionally print diffs if useful
				// diff := cmp.Diff(wantResultInterfaces, actualResults)
				// t.Errorf("Diff (-want +got):\n%s", diff)
			}
		})
	}
}

// Helper to get testing.TB compatible writer - useful if adapters are internal
type tbLogWriter struct{ t testing.TB }

func (w tbLogWriter) Write(p []byte) (n int, err error) {
	w.t.Logf("%s", p) // Use t.Logf with %s to ensure output is treated as a single log line
	return len(p), nil
}

// TestingTBLogWriter creates an io.Writer that logs to the testing.TB.
func TestingTBLogWriter(t testing.TB) io.Writer {
	return tbLogWriter{t}
}
