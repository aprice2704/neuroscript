// filename: pkg/core/testing_helpers_test.go
package core

import (
	// Keep for errors.Is potentially used elsewhere
	"fmt" // Keep io
	"log" // Keep os
	"path/filepath"
	"reflect" // Keep sort
	"strings"
	"testing"
)

// --- Interpreter Test Specific Helpers ---

// testWriter is a helper to redirect log output to t.Logf
// Make sure this struct is defined, e.g., copied from tools_go_ast_symbol_map_test.go or defined globally here.
// type testWriter struct {
// 	t *testing.T
// }

func (tw testWriter) Write(p []byte) (n int, err error) {
	tw.t.Logf("%s", p) // Use t.Logf to print the log message
	return len(p), nil
}

// newTestInterpreter creates an interpreter instance for testing.
// *** MODIFIED: Uses testWriter for logging to t.Logf ***
func newTestInterpreter(t *testing.T, vars map[string]interface{}, lastResult interface{}) (*Interpreter, string) {
	t.Helper()

	// --- FIX: Use testWriter to redirect logs to t.Logf ---
	// Use Ltime|Lmicroseconds for timing info if helpful, Lshortfile for source line
	testLogger := log.New(testWriter{t}, "[TEST-INTERP] ", log.Ltime|log.Lmicroseconds|log.Lshortfile)
	// --- END FIX ---

	// Create a minimal LLMClient (can keep using testLogger now)
	minimalLLMClient := NewLLMClient("", "", testLogger, false)
	if minimalLLMClient == nil {
		t.Fatal("Failed to create even a minimal LLMClient for testing")
	}

	interp := NewInterpreter(testLogger, minimalLLMClient) // Pass the working logger

	testLogger.Println("Attempting to register core tools...")
	if err := RegisterCoreTools(interp.ToolRegistry()); err != nil {
		testLogger.Printf("FATAL: Failed to register core tools during test setup: %v", err)
		t.Fatalf("FATAL: Failed to register core tools during test setup: %v", err)
	}
	testLogger.Println("Successfully registered core tools.")

	sandboxDirRel := t.TempDir()
	absSandboxDir, err := filepath.Abs(sandboxDirRel)
	if err != nil {
		t.Fatalf("Failed to get absolute path for sandbox %s: %v", sandboxDirRel, err)
	}

	interp.sandboxDir = absSandboxDir
	testLogger.Printf("Sandbox root set in interpreter: %s", absSandboxDir)

	if vars != nil {
		for k, v := range vars {
			interp.variables[k] = v
		}
	}

	interp.lastCallResult = lastResult

	return interp, absSandboxDir
}

// newDefaultTestInterpreter creates a new interpreter with default settings.
// *** MODIFIED: Uses newTestInterpreter which now uses t.Logf ***
func newDefaultTestInterpreter(t *testing.T) (*Interpreter, string) {
	t.Helper()
	return newTestInterpreter(t, nil, nil)
}

// --- Step Creation Helpers (Moved from interpreter_test.go) ---
// (These helpers remain unchanged)
func createTestStep(typ string, target string, valueNode interface{}, argNodes []interface{}) Step { /* ... */
	return newStep(typ, target, nil, valueNode, nil, argNodes)
}
func createIfStep(condNode interface{}, thenSteps []Step, elseSteps []Step) Step { /* ... */
	var thenVal, elseVal interface{}
	if thenSteps != nil {
		thenVal = thenSteps
	}
	if elseSteps != nil {
		elseVal = elseSteps
	}
	return Step{Type: "IF", Cond: condNode, Value: thenVal, ElseValue: elseVal}
}
func createWhileStep(condNode interface{}, bodySteps []Step) Step { /* ... */
	var bodyVal interface{}
	if bodySteps != nil {
		bodyVal = bodySteps
	}
	return Step{Type: "WHILE", Cond: condNode, Value: bodyVal}
}
func createForStep(loopVar string, collectionNode interface{}, bodySteps []Step) Step { /* ... */
	var bodyVal interface{}
	if bodySteps != nil {
		bodyVal = bodySteps
	}
	return Step{Type: "FOR", Target: loopVar, Cond: collectionNode, Value: bodyVal}
}

// runEvalExpressionTest executes a single expression evaluation test case.
// *** MODIFIED: Uses corrected newDefaultTestInterpreter ***
func runEvalExpressionTest(t *testing.T, tc EvalTestCase) {
	t.Helper()
	interp, _ := newDefaultTestInterpreter(t)
	if tc.InitialVars != nil {
		for k, v := range tc.InitialVars {
			interp.variables[k] = v
		}
	}
	interp.lastCallResult = tc.LastResult
	got, err := interp.evaluateExpression(tc.InputNode)
	if tc.WantErr {
		if err == nil {
			t.Errorf("%s: Expected an error, but got nil", tc.Name)
			return
		}
		if tc.ErrContains != "" && !strings.Contains(err.Error(), tc.ErrContains) {
			t.Errorf("%s: Expected error containing %q, got: %v", tc.Name, tc.ErrContains, err)
		}
		return
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", tc.Name)
		return
	}
	if !reflect.DeepEqual(got, tc.Expected) {
		t.Errorf("%s: Result mismatch.\nInput Node: %+v\nExpected:   %v (%T)\nGot:        %v (%T)", tc.Name, tc.InputNode, tc.Expected, tc.Expected, got, got)
	}
}

// --- General Test Helpers ---
// (makeArgs, AssertNoError remain unchanged)
// func makeArgs(vals ...interface{}) []interface{} {
// 	if vals == nil {
// 		return []interface{}{}
// 	}
// 	return vals
// }

func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		t.Fatalf("Expected no error, but got: %v\nContext: %s", err, fmt.Sprint(msgAndArgs...))
	}
}

// --- Struct Definitions (Remain unchanged) ---
type EvalTestCase struct {
	Name        string
	InputNode   interface{}
	InitialVars map[string]interface{}
	LastResult  interface{}
	Expected    interface{}
	WantErr     bool
	ErrContains string
}
type ValidationTestCase struct {
	Name          string
	ToolName      string
	InputArgs     []interface{}
	ArgSpecs      []ArgSpec
	ExpectedArgs  []interface{}
	ExpectedError error
	CheckErrorIs  bool
}

// --- Result Normalization Helpers (From tools_go_ast_package_test.go, ensure they are defined once) ---
// (Include normalizeResultMapPaths, setDefaultResultMapValues, compareErrorString if not already present/imported)
// Assuming these helpers might already be in tools_go_ast_package_test.go, defining them here might cause duplication.
// Ensure they are defined *once* accessible to all tests needing them.
// Example (if needed here):
/*
func normalizeResultMapPaths(t *testing.T, dataMap map[string]interface{}, basePath string) { ... }
func setDefaultResultMapValues(resultMap map[string]interface{}) { ... }
func compareErrorString(t *testing.T, actualMap, expectedMap map[string]interface{}) { ... }
*/

// --- Other potential helpers like runValidationTestCases ---
// Ensure runValidationTestCases is also defined once, potentially here or in a specific _test file.
/*
func runValidationTestCases(t *testing.T, toolName string, testCases []ValidationTestCase) {
	t.Helper()
	interp, _ := newDefaultTestInterpreter(t) // Now uses logging interpreter
	toolImpl, found := interp.ToolRegistry().GetTool(toolName)
	if !found { t.Fatalf("Tool %s not found in registry", toolName) }
	spec := toolImpl.Spec
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			_, err := ValidateAndConvertArgs(spec, tc.InputArgs)
			if tc.ExpectedError != nil {
				if err == nil { t.Errorf("Expected error [%v], got nil", tc.ExpectedError) } else if !errors.Is(err, tc.ExpectedError) { t.Errorf("Expected error wrapping [%v], but errors.Is is false. Got error: [%T] %v", tc.ExpectedError, err, err) } else { t.Logf("Got expected error type via errors.Is: %v", err) }
			} else if err != nil { t.Errorf("Unexpected validation error: %v", err) }
		})
	}
}
*/

// Ensure core errors are accessible if needed by helpers here
var (
	_ = ErrValidationArgCount
	_ = ErrValidationRequiredArgNil
	_ = ErrValidationTypeMismatch
)
