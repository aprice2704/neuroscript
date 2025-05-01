// filename: pkg/neurogo/app_script_test.go
package neurogo

import (
	"context"
	"log/slog" // Import slog
	"os"       // Import os for stderr
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/logging" // Import the logging interface definition
	// "github.com/aprice2704/neuroscript/pkg/core"
)

// Test for executing a script with multi-return functions
func TestApp_RunScriptMode_MultiReturn(t *testing.T) {
	// Setup: Define path relative to the test file's package directory
	scriptPath := "testdata/multi_return.ns.txt"

	// Configure the App
	// --- CORRECTED: Use SlogAdapter properly ---
	// 1. Create the slog handler
	slogHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Ensure Debug level to see stack traces
	})
	// 2. Create the slog Logger instance using the handler
	slogLogger := slog.New(slogHandler)
	// 3. Create the adapter using the slog Logger instance
	var logger logging.Logger // Declare logger with the interface type
	loggerAdapter, err := adapters.NewSlogAdapter(slogLogger)
	if err != nil {
		// Handle error during logger creation - fail the test
		t.Fatalf("Failed to create SlogAdapter for testing: %v", err)
	}
	logger = loggerAdapter // Assign the created adapter to the interface variable
	// --- END CORRECTION ---

	llmClient := adapters.NewNoOpLLMClient() // Keep LLM as NoOp

	cfg := &Config{
		RunScriptMode: true,       // Indicate script mode
		ScriptFile:    scriptPath, // Path to the script file
		TargetArg:     "main",     // Target the main procedure
		EnableLLM:     false,      // Disable LLM for this script test
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
