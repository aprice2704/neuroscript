// NeuroScript Version: 0.3.0
// File version: 0.1.1
// Updated Config struct field names
// filename: pkg/neurogo/app_script_test.go
// nlines: 45
// risk_rating: LOW
package neurogo

import (
	"context" // Import slog
	"os"      // Import os for stderr
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/logging" // Import the logging interface definition
	// "github.com/aprice2704/neuroscript/pkg/core"
)

// Test for executing a script with multi-return functions
func TestApp_RunScriptMode_MultiReturn(t *testing.T) {
	// Setup: Define path relative to the test file's package directory
	scriptPath := "testdata/multi_return.ns.txt"

	var logger logging.Logger // Declare logger with the interface type
	loggerAdapter, err := adapters.NewSimpleSlogAdapter(os.Stderr, logging.LogLevelDebug)
	if err != nil {
		// Handle error during logger creation - fail the test
		t.Fatalf("Failed to create SlogAdapter for testing: %v", err)
	}
	logger = loggerAdapter // Assign the created adapter to the interface variable
	// --- END CORRECTION ---

	llmClient := adapters.NewNoOpLLMClient() // Keep LLM as NoOp

	cfg := &Config{
		StartupScript: scriptPath, // CORRECTED: Was ScriptFile
		TargetArg:     "main",     // Target the main procedure
		// REMOVED: RunScriptMode: true,
		// REMOVED: EnableLLM:     false,
	}

	// Create and run the App
	app := NewApp(logger)     // Create App with the correctly initialized logger
	app.Config = cfg          // Assign the config
	app.llmClient = llmClient // Assign the NoOp LLM client

	// Execute the script via the App's Run method
	runErr := app.Run(context.Background()) // Renamed err to runErr for clarity

	// Assert: Check for errors during execution
	if runErr != nil {
		// Include the test name in the error message for clarity
		t.Errorf("Test '%s': Expected successful execution of script '%s', but got error: %v", t.Name(), scriptPath, runErr)
	}
}

// Add more tests for app_script.go functionality as needed
