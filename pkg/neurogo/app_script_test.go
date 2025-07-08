// NeuroScript Version: 0.3.1
// File version: 0.3.9
// Purpose: Corrected the NewInterpreter constructor to use the functional options pattern
// and fixed the logger constructor name. Removed non-existent WithLibPaths option.
// filename: pkg/neurogo/app_script_test.go
package neurogo

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/tool"
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all" // load tools
)

// setupTestApp is a helper to reduce boilerplate in tests.
func setupTestApp(t *testing.T) *App {
	t.Helper()
	cfg := Config{SandboxDir: t.TempDir()}
	logger, err := logging.NewSimpleSlogAdapter(os.Stderr, interfaces.LogLevelDebug)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	llmClient := adapters.NewNoOpLLMClient()
	app, _ := NewApp(&cfg, logger, llmClient)

	interp := interpreter.NewInterpreter(
		interpreter.WithLogger(logger),
		interpreter.WithLLMClient(llmClient),
		interpreter.WithSandboxDir(cfg.SandboxDir),
		// interpreter.WithLibPaths(cfg.LibPaths), // This option does not exist yet
	)
	if err != nil {
		t.Fatalf("Failed to create interpreter: %v", err)
	}
	app.SetInterpreter(interp)

	if err := tool.RegisterExtendedTools(interp.ToolRegistry()); err != nil {
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
	scriptPathForTool := filepath.Join("testdata", scriptName)
	filepathArg, err := lang.Wrap(scriptPathForTool)
	if err != nil {
		t.Fatalf("Failed to wrap filepath argument: %v", err)
	}
	toolArgs := map[string]lang.Value{"filepath": filepathArg}
	contentValue, err := app.GetInterpreter().ExecuteTool("FS.Read", toolArgs)
	if err != nil {
		t.Fatalf("Executing FS.Read tool failed: %v", err)
	}
	scriptContent, ok := lang.Unwrap(contentValue).(string)
	if !ok {
		t.Fatalf("FS.Read did not return a string, got %T", lang.Unwrap(contentValue))
	}

	_, err = app.LoadScriptString(ctx, scriptContent)
	if err != nil {
		t.Fatalf("LoadScriptString failed: %v", err)
	}

	_, err = app.RunProcedure(ctx, "main", nil)

	// ASSERT
	if err != nil {
		t.Errorf("RunProcedure for 'main' failed: %v", err)
	}
}

func TestApp_LoadScript_DoesNotExecuteTopLevelCode(t *testing.T) {
	// ARRANGE
	const testScriptWithSideEffect = `
		call TestSetCanary()
		func do_nothing() means
		endfunc
	`

	wasExecuted := false
	canaryTool := tool.ToolImplementation{
		Spec: tool.ToolSpec{Name: "TestSetCanary"},
		Func: func(rt tool.Runtime, args []interface{}) (interface{}, error) {
			wasExecuted = true
			return true, nil
		},
	}
	app := setupTestApp(t)
	if err := app.GetInterpreter().ToolRegistry().RegisterTool(canaryTool); err != nil {
		t.Fatalf("Failed to register canary tool: %v", err)
	}

	// ACT
	_, err := app.LoadScriptString(context.Background(), testScriptWithSideEffect)

	// ASSERT
	if wasExecuted {
		t.Error("Test failed: Top-level statement was executed during LoadScriptString.")
	}
	if err == nil {
		t.Error("Test passed, but LoadScriptString did not return a syntax error as expected.")
	}
}

func TestApp_LoadScript_DoesNotImplicitlyRunMain(t *testing.T) {
	// ARRANGE
	const testScriptWithMainFunc = `
		func main() means
			call TestSetCanary()
		endfunc
	`

	wasExecuted := false
	canaryTool := tool.ToolImplementation{
		Spec: tool.ToolSpec{Name: "TestSetCanary"},
		Func: func(rt tool.Runtime, args []interface{}) (interface{}, error) {
			wasExecuted = true
			return true, nil
		},
	}
	app := setupTestApp(t)
	if err := app.GetInterpreter().ToolRegistry().RegisterTool(canaryTool); err != nil {
		t.Fatalf("Failed to register canary tool: %v", err)
	}

	// ACT
	_, err := app.LoadScriptString(context.Background(), testScriptWithMainFunc)

	// ASSERT
	if wasExecuted {
		t.Error("Test failed: A function named 'main' was executed automatically on load.")
	}
	if err != nil {
		t.Errorf("Test failed: LoadScriptString returned an unexpected error: %v", err)
	}
}
