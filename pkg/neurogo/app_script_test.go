// filename: pkg/neurogo/app_script_test.go
package neurogo

import (
	"context"
	// Removed os and path/filepath as createTempScript is removed
	"testing"

	// Import the adapters package for NoOp implementations
	"github.com/aprice2704/neuroscript/pkg/adapters"
	// Keep core import if needed by other tests or helpers in this file
	// "github.com/aprice2704/neuroscript/pkg/core"
)

// Removed createTempScript helper function as it's no longer needed here.
// If other tests use it, keep it; otherwise, it can be deleted.

// Test for executing a script with multi-return functions
func TestApp_RunScriptMode_MultiReturn(t *testing.T) {
	// Setup: Define path relative to the test file's package directory
	scriptPath := "testdata/multi_return.ns.txt" // Assumes testdata is in the same dir or accessible

	// Configure the App
	// Use NoOpLogger and NoOpLLMClient from the adapters package
	logger := adapters.NewNoOpLogger()
	llmClient := adapters.NewNoOpLLMClient()

	cfg := &Config{
		RunScriptMode: true,       // Indicate script mode
		ScriptFile:    scriptPath, // Path to the script file
		TargetArg:     "main",     // Target the main procedure
		EnableLLM:     false,      // Disable LLM for this script test
		// LibPaths and ProcArgs are empty/nil by default
	}

	// Create and run the App
	app := NewApp(logger)     // Create App with the NoOp logger
	app.Config = cfg          // Assign the config
	app.llmClient = llmClient // Assign the NoOp LLM client

	// Execute the script via the App's Run method
	err := app.Run(context.Background())

	// Assert: Check for errors during execution
	// The script itself contains 'must' checks which will cause 'Run' to return an error if they fail.
	if err != nil {
		t.Errorf("Expected successful execution of multi-return script '%s', but got error: %v", scriptPath, err)
	}
}

// Add more tests for app_script.go functionality as needed
// E.g., TestApp_RunScriptMode_FileNotFound, TestApp_RunScriptMode_SyntaxError, etc.
