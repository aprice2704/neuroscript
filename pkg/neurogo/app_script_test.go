// NeuroScript Version: 0.3.0
// File version: 0.1.0
// Refactored to use app.ExecuteScriptFile instead of deprecated App.Run.
// filename: pkg/neurogo/app_script_test.go
package neurogo

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/toolsets"
)

func TestApp_RunScriptMode_MultiReturn(t *testing.T) {
	testName := "TestApp_RunScriptMode_MultiReturn"
	scriptName := "multi_return.ns.txt" // Assumes this file is in testdata relative to this test file
	scriptPath := filepath.Join("testdata", scriptName)

	// 1. Create Config
	cfg := Config{}
	cfg.SandboxDir = t.TempDir()
	// cfg.LogFile = "" // Log to stderr for test visibility
	// cfg.LogLevel = "debug"

	// 2. Initialize Logger
	logger, err := adapters.NewSimpleSlogAdapter(os.Stderr, interfaces.LogLevelDebug)
	if err != nil {
		t.Fatalf("%s: Failed to create logger: %v", testName, err)
	}

	// 3. Initialize LLMClient
	llmClient := adapters.NewNoOpLLMClient()

	config := NewConfig()

	// 4. Create App (passing nil for interpreter and aiwm, will set them up next)
	app, _ := NewApp(config, logger, llmClient)

	// 5. Setup Interpreter
	absSandboxDir, err := filepath.Abs(cfg.SandboxDir)
	if err != nil {
		t.Fatalf("%s: Failed to get absolute sandbox path: %v", testName, err)
	}
	var procArgs map[string]interface{}

	interpreter, err := core.NewInterpreter(logger, llmClient, absSandboxDir, procArgs, cfg.LibPaths)
	if err != nil {
		t.Fatalf("%s: Failed to create core.Interpreter: %v", testName, err)
	}
	app.SetInterpreter(interpreter)

	if err := toolsets.RegisterExtendedTools(interpreter); err != nil {
		t.Fatalf("%s: Failed to register extended tools: %v", testName, err)
	}

	// AIWM setup (simplified)
	aiWm, aiWmErr := core.NewAIWorkerManager(logger, app.Config.SandboxDir, llmClient, "", "")
	if aiWmErr != nil {
		t.Logf("%s: Warning - Failed to create AI Worker Manager: %v", testName, aiWmErr)
	} else {
		app.SetAIWorkerManager(aiWm)
	}

	// 6. Execute Script
	ctx := context.Background()
	executionErr := app.ExecuteScriptFile(ctx, scriptPath)

	// 7. Assert Results
	if executionErr != nil {
		t.Errorf("Test '%s': Expected successful execution of script '%s', but got error: %v",
			testName, scriptPath, executionErr)
	}
}

// Add other tests from app_script_test.go here if they also use App.Run
// and need similar refactoring.
