// NeuroScript Version: 0.3.1
// File version: 0.2.7
// Purpose: Added a permissive execution policy to the interpreter to grant the fs:read capability, resolving the test failure.
// filename: pkg/neurogo/app_script_break_continue_test.go
package neurogo

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
	"github.com/aprice2704/neuroscript/pkg/runtime"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestApp_RunScriptMode_BreakContinue(t *testing.T) {
	testName := "TestApp_RunScriptMode_BreakContinue"
	scriptName := "valid_break_continue.ns.txt"
	sourceScriptPath := filepath.Join("testdata", scriptName)

	// 1. Create Config & App
	cfg := Config{}
	cfg.SandboxDir = t.TempDir()
	loglev, err := logging.LogLevelFromString("debug")
	if err != nil {
		t.Fatalf("%s: Failed to create log level: %v", testName, err)
	}
	logger, err := logging.NewSimpleSlogAdapter(os.Stderr, loglev)
	if err != nil {
		t.Fatalf("%s: Failed to create logger: %v", testName, err)
	}
	llmClient := adapters.NewNoOpLLMClient()
	config := NewConfig()
	app, err := NewApp(config, logger, llmClient)
	if err != nil {
		t.Fatalf("%s: Failed to create App: %v", testName, err)
	}

	// 2. Setup Interpreter and Sandbox
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

	// FIX: Add a permissive policy for testing to grant necessary capabilities.
	policy := &runtime.ExecPolicy{
		Context: runtime.ContextNormal,
		Allow:   []string{"*"}, // Allow all tools by name for testing.
		Grants: capability.GrantSet{
			Grants: []capability.Capability{
				capability.New(capability.ResFS, capability.VerbRead, "*"),
			},
		},
	}

	// Corrected Interpreter constructor with all options
	interp := interpreter.NewInterpreter(
		interpreter.WithLogger(logger),
		interpreter.WithLLMClient(llmClient),
		interpreter.WithSandboxDir(cfg.SandboxDir),
		interpreter.WithExecPolicy(policy), // Apply the policy
	)
	app.SetInterpreter(interp)

	if err := tool.RegisterExtendedTools(interp.ToolRegistry()); err != nil {
		t.Fatalf("%s: Failed to register extended tools: %v", testName, err)
	}

	// 3. Read, Load, and Execute Script
	ctx := context.Background()

	scriptPathForTool := filepath.Join("testdata", scriptName)
	filepathArg, err := lang.Wrap(scriptPathForTool)
	if err != nil {
		t.Fatalf("Failed to wrap script path argument: %v", err)
	}
	toolArgs := map[string]lang.Value{"filepath": filepathArg}
	contentValue, err := app.GetInterpreter().ExecuteTool("fs.read", toolArgs)
	if err != nil {
		t.Fatalf("Executing tool 'fs.read' failed: %v", err)
	}
	scriptContent, ok := lang.Unwrap(contentValue).(string)
	if !ok {
		t.Fatalf("'fs.read' did not return a string, got %T", lang.Unwrap(contentValue))
	}

	_, err = app.LoadScriptString(ctx, scriptContent)
	if err != nil {
		t.Fatalf("LoadScriptString failed: %v", err)
	}

	_, executionErr := app.RunProcedure(ctx, "main", nil)

	// 4. Assert Results
	if executionErr != nil {
		t.Errorf("Test '%s': Expected successful execution of script '%s', but got error: %v",
			testName, sourceScriptPath, executionErr)
	}
}
