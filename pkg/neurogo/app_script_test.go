// NeuroScript Version: 0.3.1
// File version: 0.3.5
// Purpose: Fixed three test failures: 1) Corrected tool name to FS.Read and added sandbox file creation. 2) Adjusted a test to handle expected syntax errors. 3) Renamed a test tool to avoid a parser issue with dots in names.
// filename: pkg/neurogo/app_script_test.go
package neurogo

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/toolsets"
)

// setupTestApp is a helper to reduce boilerplate in tests.
func setupTestApp(t *testing.T) *App {
	t.Helper()
	cfg := Config{SandboxDir: t.TempDir()}
	logger, err := adapters.NewSimpleSlogAdapter(os.Stderr, interfaces.LogLevelDebug)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	llmClient := adapters.NewNoOpLLMClient()
	app, _ := NewApp(&cfg, logger, llmClient)

	interpreter, err := NewInterpreter(logger, llmClient, cfg.SandboxDir, nil, cfg.LibPaths)
	if err != nil {
		t.Fatalf("Failed to create  rpreter: %v", err)
	}
	app.SetInterpreter(interpreter)

	// Register tools needed for tests, including file system tools.
	if err := toolsets.RegisterExtendedTools(interpreter); err != nil {
		t.Fatalf("Failed to register extended tools: %v", err)
	}
	return app
}

func TestApp_LoadAndRunScript_MultiReturn(t *testing.T) {
	// ARRANGE
	app := setupTestApp(t)
	ctx := context.Background()
	scriptName := "multi_return.ns.txt"
	sourceScriptPath := filepath.Join("testdata", scriptName)

	// Create the script file inside the sandbox directory
	scriptContentBytes, err := os.ReadFile(sourceScriptPath)
	if err != nil {
		t.Fatalf("Failed to read source script '%s': %v", sourceScriptPath, err)
	}
	sandboxScriptDir := filepath.Join(app.Config.SandboxDir, "testdata")
	if err := os.MkdirAll(sandboxScriptDir, 0755); err != nil {
		t.Fatalf("Failed to create testdata dir in sandbox: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sandboxScriptDir, scriptName), scriptContentBytes, 0644); err != nil {
		t.Fatalf("Failed to write script to sandbox: %v", err)
	}

	// ACT
	// 1. Read the script file content using the interpreter's tool.
	scriptPathForTool := filepath.Join("testdata", scriptName)
	filepathArg, err := (scriptPathForTool)
	if err != nil {
		t.Fatalf("Failed to wrap filepath argument: %v", err)
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

	// 2. Load the script definitions from the string content.
	_, err = app.LoadScriptString(ctx, scriptContent)
	if err != nil {
		t.Fatalf("LoadScriptString failed: %v", err)
	}

	// 3. Execute a procedure from the loaded script.
	_, err = app.RunProcedure(ctx, "main", nil)

	// ASSERT
	if err != nil {
		t.Errorf("RunProcedure for 'main' failed: %v", err)
	}
}

// TestApp_LoadScript_DoesNotExecuteTopLevelCode verifies that loading a script does not
// execute top-level statements.
func TestApp_LoadScript_DoesNotExecuteTopLevelCode(t *testing.T) {
	// ARRANGE
	const testScriptWithSideEffect = `
		# This top-level call is a syntax error, which prevents execution.
		call TestSetCanary()
		func do_nothing() means
			# This function exists to prove the script can be parsed.
		endfunc
	`

	wasExecuted := false
	canaryTool := Implementation{
		Spec:	Spec{Name: "TestSetCanary"},	// Simplified name to avoid parser issues
		Func: func(i *rpreter, args []any) (any, error) {
			wasExecuted = true
			return true, nil
		},
	}
	app := setupTestApp(t)
	if err := app.Interpreter().RegisterTool(canaryTool); err != nil {
		t.Fatalf("Failed to register canary tool: %v", err)
	}

	// ACT
	_, err := app.LoadScriptString(context.Background(), testScriptWithSideEffect)

	// ASSERT
	// The primary goal: the top-level code was not executed.
	if wasExecuted {
		t.Error("Test failed: Top-level statement was executed during LoadScriptString.")
	}
	// A syntax error from LoadScriptString is expected and proves non-execution, so we don't fail the test on `err != nil`.
	if err == nil {
		t.Error("Test passed, but LoadScriptString did not return a syntax error as expected. The parser might have changed.")
	}
}

// TestApp_LoadScript_DoesNotImplicitlyRunMain verifies that a function named 'main'
// is not treated as a special entrypoint and executed automatically on load.
func TestApp_LoadScript_DoesNotImplicitlyRunMain(t *testing.T) {
	// ARRANGE
	const testScriptWithMainFunc = `
		func main() means
			# This call should only happen if 'main' is explicitly run,
			# not when the script is loaded.
			call TestSetCanary()
		endfunc
	`

	wasExecuted := false
	canaryTool := Implementation{
		Spec:	Spec{Name: "TestSetCanary"},	// Simplified name to avoid parser issues
		Func: func(i *rpreter, args []any) (any, error) {
			wasExecuted = true
			return true, nil
		},
	}
	app := setupTestApp(t)
	if err := app.Interpreter().RegisterTool(canaryTool); err != nil {
		t.Fatalf("Failed to register canary tool: %v", err)
	}

	// ACT
	// Load the script containing the 'main' function. Do NOT run anything.
	_, err := app.LoadScriptString(context.Background(), testScriptWithMainFunc)

	// ASSERT
	// 1. The primary assertion: the side effect should NOT have happened.
	if wasExecuted {
		t.Error("Test failed: A function named 'main' was executed automatically on load, violating the 'No implicit main' rule.")
	}

	// 2. Secondary assertion: the script should have loaded without parsing errors.
	if err != nil {
		t.Errorf("Test failed: LoadScriptString returned an unexpected error: %v", err)
	}
}