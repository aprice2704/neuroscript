// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 18:31:17 PDT
// filename: pkg/neurogo/app_script_break_continue_test.go
package neurogo

import (
	"context"
	"os"            // Needed for reading test file
	"path/filepath" // Needed for joining paths

	// "reflect"       // No longer needed for variable comparison
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	// Import the logging interface definition
)

// TestApp_RunScriptMode_BreakContinue tests break/continue execution via the App layer.
// It relies on the internal 'must' statements within the script to cause an error
// if the break/continue logic is incorrect. The Go test only checks if app.Run()
// returns an error.
func TestApp_RunScriptMode_BreakContinue(t *testing.T) {
	// Define path relative to this test file's package directory (neurogo)
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

	// Assert execution success.
	// If any 'must' statement within 'valid_break_continue.ns.txt' fails,
	// app.Run() should return a non-nil error.
	if runErr != nil {
		// Use t.Fatalf as the rest of the test depends on successful execution
		t.Fatalf("Test '%s': Expected successful execution of script '%s' (no 'must' failures), but got error: %v", t.Name(), scriptPath, runErr)
	}

	// --- Assertions Removed ---
	// Variable assertions are removed because interpreter.RunProcedure restores the
	// previous variable scope upon returning, making variables set *inside* the
	// 'main' procedure inaccessible after app.Run() completes.
	// The test now relies on the internal 'must' statements within the script
	// to verify correctness, ensuring app.Run() returns nil on success.

	// Example of removed assertion helper:
	/*
		assertVariableEqualsStd := func(varName string, expectedValue interface{}) {
			t.Helper() // Mark this as a helper function
			if app.interpreter == nil {
				t.Fatalf("Test '%s': App interpreter is nil after execution. Cannot verify variables.", t.Name())
				return // Redundant due to Fatalf
			}
			value, exists := app.interpreter.GetVariable(varName)
			if !exists {
				t.Errorf("Variable '%s' should exist after execution, but was not found", varName)
				return
			}
			if !reflect.DeepEqual(expectedValue, value) {
				t.Errorf("Variable '%s' has unexpected value. Expected: [%v] (%T), Got: [%v] (%T)",
					varName, expectedValue, expectedValue, value, value)
			}
		}
	*/

	// Examples of removed assertions:
	// assertVariableEqualsStd("result", int64(3))
	// assertVariableEqualsStd("counter", int64(5))
	// assertVariableEqualsStd("sum", int64(12))
	// ... etc ...

	// If we reach here without t.Fatalf, the script executed without error.
	t.Logf("Test '%s': Script '%s' executed successfully (no 'must' failures detected via app.Run error).", t.Name(), scriptPath)
}
