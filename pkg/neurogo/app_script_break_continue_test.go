// NeuroScript Version: 0.3.0
// File version: 0.1.0
// Refactored to use app.ExecuteScriptFile instead of deprecated App.Run.
// filename: pkg/neurogo/app_script_break_continue_test.go
package neurogo

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/toolsets"
)

func TestApp_RunScriptMode_BreakContinue(t *testing.T) {
	testName := "TestApp_RunScriptMode_BreakContinue"
	scriptName := "valid_break_continue.ns.txt" // Assumes this file is in testdata relative to this test file
	scriptPath := filepath.Join("testdata", scriptName)

	// 1. Create Config
	cfg := Config{}
	cfg.SandboxDir = t.TempDir()
	// cfg.LogFile = "" // Log to stderr for test visibility
	// cfg.LogLevel = "debug"
	// // cfg.APIKey = "test-dummy-key" // Usually not needed with NoOpLLMClient

	// 2. Initialize Logger
	loglev, _ := adapters.LogLevelFromString("debug")
	logger, err := adapters.NewSimpleSlogAdapter(os.Stderr, loglev)
	if err != nil {
		t.Fatalf("%s: Failed to create logger: %v", testName, err)
	}

	// 3. Initialize LLMClient
	llmClient := adapters.NewNoOpLLMClient()

	// 4. Create App (passing nil for interpreter and aiwm, will set them up next)
	app := NewApp(logger)
	if err != nil {
		t.Fatalf("%s: Failed to create App: %v", testName, err)
	}

	// 5. Setup Interpreter
	absSandboxDir, err := filepath.Abs(cfg.SandboxDir)
	if err != nil {
		t.Fatalf("%s: Failed to get absolute sandbox path: %v", testName, err)
	}

	var procArgs map[string]interface{} // Assuming test script doesn't rely on specific procArgs from file
	// If ProcArgsConfig is needed for some tests, it should be handled here.
	// For this specific test, unlikely to be essential.

	interpreter, err := core.NewInterpreter(logger, llmClient, absSandboxDir, procArgs, cfg.LibPaths)
	if err != nil {
		t.Fatalf("%s: Failed to create core.Interpreter: %v", testName, err)
	}
	app.SetInterpreter(interpreter) // Link interpreter to the app

	// Core tools are registered by core.NewInterpreter by default.
	// Register extended toolsets - good practice for app-level tests.
	if err := toolsets.RegisterExtendedTools(interpreter); err != nil {
		t.Fatalf("%s: Failed to register extended tools: %v", testName, err)
	}

	// AIWM setup (simplified - only if strictly needed by NewApp or scripts not using AI tools)
	// For these specific control-flow tests, AIWM is likely not involved.
	// If app.ExecuteScriptFile or other app logic requires a non-nil AIWM:
	aiWm, aiWmErr := core.NewAIWorkerManager(logger, app.Config.SandboxDir, llmClient, "", "")
	if aiWmErr != nil {
		t.Logf("%s: Warning - Failed to create AI Worker Manager: %v (continuing as script might not need it)", testName, aiWmErr)
		// app.SetAIWorkerManager(nil) // Explicitly set to nil if that's acceptable
	} else {
		app.SetAIWorkerManager(aiWm)
		// No need to call app.RegisterAIWorkerTools if the script doesn't use AI tools.
	}

	// 6. Execute Script
	ctx := context.Background()
	executionErr := app.ExecuteScriptFile(ctx, scriptPath)

	// 7. Assert Results
	// The original error message indicates these tests expect no 'must' failures.
	// So, any error from ExecuteScriptFile is a test failure.
	if executionErr != nil {
		t.Errorf("Test '%s': Expected successful execution of script '%s' (no 'must' failures), but got error: %v",
			testName, scriptPath, executionErr)
	}
}
