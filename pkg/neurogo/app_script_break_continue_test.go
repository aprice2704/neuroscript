// NeuroScript Version: 0.3.1
// File version: 0.2.2
// Purpose: Fixed test failure by ensuring the test script file is created within the sandbox directory before being accessed by the FS.Read tool.
// filename: pkg/neurogo/app_script_break_continue_test.go
package neurogo

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/toolsets"
)

func TestApp_RunScriptMode_BreakContinue(t *testing.T) {
	testName := "TestApp_RunScriptMode_BreakContinue"
	scriptName := "valid_break_continue.ns.txt"
	sourceScriptPath := filepath.Join("testdata", scriptName)	// Path to the source file in the project

	// 1. Create Config & App
	cfg := Config{}
	cfg.SandboxDir = t.TempDir()
	loglev, _ := adapters.LogLevelFromString("debug")
	logger, err := adapters.NewSimpleSlogAdapter(os.Stderr, loglev)
	if err != nil {
		t.Fatalf("%s: Failed to create logger: %v", testName, err)
	}
	llmClient := adapters.NewNoOpLLMClient()
	config := NewConfig()
	app, _ := NewApp(config, logger, llmClient)
	if err != nil {
		t.Fatalf("%s: Failed to create App: %v", testName, err)
	}

	// 2. Setup Interpreter and Sandbox
	// This setup ensures the test script is actually present within the sandbox
	// before the interpreter tries to read it.
	scriptContentBytes, err := os.ReadFile(sourceScriptPath)
	if err != nil {
		t.Fatalf("Failed to read source test script '%s': %v", sourceScriptPath, err)
	}
	sandboxScriptDir := filepath.Join(cfg.SandboxDir, "testdata")
	if err := os.MkdirAll(sandboxScriptDir, 0755); err != nil {
		t.Fatalf("Failed to create testdata dir in sandbox: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sandboxScriptDir, scriptName), scriptContentBytes, 0644); err != nil {
		t.Fatalf("Failed to write script to sandbox: %v", err)
	}

	interpreter, err := NewInterpreter(logger, llmClient, cfg.SandboxDir, nil, cfg.LibPaths)
	if err != nil {
		t.Fatalf("%s: Failed to create  rpreter: %v", testName, err)
	}
	app.SetInterpreter(interpreter)
	if err := toolsets.RegisterExtendedTools(interpreter); err != nil {
		t.Fatalf("%s: Failed to register extended tools: %v", testName, err)
	}
	aiWm, aiWmErr := IWorkerManager(logger, app.Config.SandboxDir, llmClient, "", "")
	if aiWmErr != nil {
		t.Logf("%s: Warning - Failed to create AI Worker Manager: %v", testName, aiWmErr)
	} else {
		app.SetAIWorkerManager(aiWm)
	}

	// 3. Read, Load, and Execute Script using the new protocol
	ctx := context.Background()

	// 3a. Read the script file content using the interpreter's tool.
	// The path must be relative to the sandbox root.
	scriptPathForTool := filepath.Join("testdata", scriptName)
	filepathArg, err := (scriptPathForTool)
	if err != nil {
		t.Fatalf("Failed to wrap script path argument: %v", err)
	}
	toolArgs := map[string]e{"filepath": filepathArg}
	contentValue, err := app.Interpreter().ExecuteTool("FS.Read", toolArgs)
	if err != nil {
		t.Fatalf("Executing FS.Read tool failed: %v", err)
	}
	scriptContent, ok := ap(contentValue).(string)
	if !ok {
		t.Fatalf("FS.Read did not return a string, got %T", ap(contentValue))
	}

	// 3b. Load the script definitions from the content string.
	_, err = app.LoadScriptString(ctx, scriptContent)
	if err != nil {
		t.Fatalf("LoadScriptString failed: %v", err)
	}

	// 3c. Run the main procedure from the script.
	_, executionErr := app.RunProcedure(ctx, "main", nil)

	// 4. Assert Results
	if executionErr != nil {
		t.Errorf("Test '%s': Expected successful execution of script '%s', but got error: %v",
			testName, sourceScriptPath, executionErr)
	}
}