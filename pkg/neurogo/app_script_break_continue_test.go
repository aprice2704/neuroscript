// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 14:29:55 PDT
// filename: pkg/neurogo/app_script_break_continue_test.go
package neurogo

import (
	"context"
	"os"            // Needed for reading test file
	"path/filepath" // Needed for joining paths
	"reflect"       // Needed for variable comparison
	"testing"

	// Import core package to access interpreter state after run
	// Be aware: This might cause import cycles depending on package dependencies.
	// If pkg/core imports pkg/neurogo, this test might need to be moved back
	// or restructured (e.g., using emit instead of checking variables).

	"github.com/aprice2704/neuroscript/pkg/adapters"
	// Import the logging interface definition
)

// TestApp_RunScriptMode_BreakContinue tests break/continue execution via the App layer.
func TestApp_RunScriptMode_BreakContinue(t *testing.T) {
	// Define path relative to this test file's package directory (neurogo)
	// Assumes neurogo and core are sibling directories.
	scriptPath := filepath.Join("testdata", "valid_break_continue.ns.txt")
	// Verify the test script exists before running the test
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Fatalf("Test script not found at expected path: %s. Check relative path from pkg/neurogo/", scriptPath)
	}

	// Configure logger (using NoOp for this test)
	logger := adapters.NewNoOpLogger()       // From adapters package
	llmClient := adapters.NewNoOpLLMClient() // From adapters package

	cfg := &Config{ // Assuming Config struct is defined in neurogo/config.go
		RunScriptMode: true,
		ScriptFile:    scriptPath,
		TargetArg:     "main", // Target the main procedure in the script
		EnableLLM:     false,
	}

	// Create and run the App (assuming NewApp is defined in neurogo/app.go)
	app := NewApp(logger)
	if app == nil {
		t.Fatal("Failed to create App")
	}
	app.Config = cfg
	app.llmClient = llmClient // Ensure LLM client is set even if disabled in cfg

	// Execute the script via the App's Run method
	runErr := app.Run(context.Background())

	// Assert execution success
	if runErr != nil {
		// Use t.Fatalf as the rest of the test depends on successful execution
		t.Fatalf("Test '%s': Expected successful execution of script '%s', but got error: %v", t.Name(), scriptPath, runErr)
	}

	// --- Assert Final Variable States ---
	// Access the interpreter state *after* app.Run completes.
	// This assumes app.Run successfully initializes and holds onto the interpreter
	// and that the 'interpreter' field is accessible (e.g., exported or via a getter).
	// If 'interpreter' is not exported, this check needs modification.
	if app.interpreter == nil {
		t.Fatalf("Test '%s': App interpreter is nil after execution. Cannot verify variables.", t.Name())
	}

	// Helper function defined locally for this test file using standard testing
	assertVariableEqualsStd := func(varName string, expectedValue interface{}) {
		t.Helper() // Mark this as a helper function

		// Use GetVariable from the core interpreter instance
		// Assumes GetVariable is an exported method on core.Interpreter
		value, exists := app.interpreter.GetVariable(varName)
		if !exists {
			t.Errorf("Variable '%s' should exist after execution, but was not found", varName)
			return // Stop checking this variable if it doesn't exist
		}

		// Compare values using reflect.DeepEqual
		// Be mindful of potential numeric type differences (e.g., int64 vs float64).
		if !reflect.DeepEqual(expectedValue, value) {
			t.Errorf("Variable '%s' has unexpected value. Expected: [%v] (%T), Got: [%v] (%T)",
				varName, expectedValue, expectedValue, value, value)
		}
	}

	// Assertions based on the final state after 'main' calls all test functions in the script
	// Note: These assume variables set within the called functions affect the interpreter's main scope.
	// Using int64 for expected values.
	assertVariableEqualsStd("result", int64(3))              // From test_while_break (overwritten by test_while_continue's result, then saved to sum)
	assertVariableEqualsStd("counter", int64(5))             // Final value from test_while_continue
	assertVariableEqualsStd("sum", int64(12))                // Saved result from test_while_continue
	assertVariableEqualsStd("processed", int64(2))           // From test_for_break
	assertVariableEqualsStd("total", int64(12))              // From test_for_continue
	assertVariableEqualsStd("outer_loops", int64(3))         // From test_nested_break
	assertVariableEqualsStd("inner_loops_total", int64(6))   // From test_nested_break
	assertVariableEqualsStd("outer_loops_c", int64(2))       // From test_nested_continue
	assertVariableEqualsStd("inner_loops_total_c", int64(6)) // From test_nested_continue
	assertVariableEqualsStd("i", int64(5))                   // Final value from test_last_statement_continue
	assertVariableEqualsStd("count", int64(4))               // Final value from test_last_statement_continue
}
