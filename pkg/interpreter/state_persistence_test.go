// NeuroScript Version: 0.8.0
// File version: 8
// Purpose: Corrected ExecPolicy assignment to align with the post-refactor API.
// filename: pkg/interpreter/interpreter_state_persistence_test.go
// nlines: 125
// risk_rating: LOW

package interpreter_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/policy"
)

const statePersistenceAndSandboxingTestScript = `
# This script is for testing state persistence and variable sandboxing.

func _SetupState() means
    must tool.account.Register("test_account", {\
        "kind": "test",\
        "provider": "test_provider",\
        "api_key": "123"\
    })
    must tool.agentmodel.Register("test_agent_for_persistence", {\
        "provider": "p",\
        "model": "m",\
        "AccountName": "test_account"\
    })
    return "setup_ok"
endfunc

func _CheckState() means
    set model = tool.agentmodel.Get("test_agent_for_persistence")
    must not is_error(model)
    return "check_ok"
endfunc

func _SetLocalVariable() means
    set this_should_not_leak = "i am a local variable"
    return "local_set_ok"
endfunc
`

func TestStatePersistence_StoresPersistAcrossRuns(t *testing.T) {
	t.Logf("[DEBUG] Turn 1: Starting TestStatePersistence_StoresPersistAcrossRuns.")
	h := NewTestHarness(t)
	// Provide permissive policy to allow admin tools to run
	h.Interpreter.ExecPolicy = &policy.ExecPolicy{Context: policy.ContextConfig, Allow: []string{"*"}}
	interp := h.Interpreter

	tree, pErr := h.Parser.Parse(statePersistenceAndSandboxingTestScript)
	if pErr != nil {
		t.Fatalf("Failed to parse script: %v", pErr)
	}
	program, _, bErr := h.ASTBuilder.Build(tree)
	if bErr != nil {
		t.Fatalf("Failed to build AST: %v", bErr)
	}
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}
	t.Logf("[DEBUG] Turn 2: Script parsed and loaded.")

	_, setupErr := interp.Run("_SetupState")
	if setupErr != nil {
		t.Fatalf("Setup procedure failed unexpectedly: %v", setupErr)
	}
	t.Logf("[DEBUG] Turn 3: '_SetupState' procedure executed.")

	_, checkErr := interp.Run("_CheckState")
	if checkErr != nil {
		t.Fatalf("Check procedure failed, indicating state was not persisted: %v", checkErr)
	}
	t.Logf("[DEBUG] Turn 4: '_CheckState' procedure executed successfully.")
	t.Log("Successfully verified that store state persists across multiple Run() calls.")
}

func TestStatePersistence_VariablesAreSandboxed(t *testing.T) {
	t.Logf("[DEBUG] Turn 1: Starting TestStatePersistence_VariablesAreSandboxed.")
	h := NewTestHarness(t)
	interp := h.Interpreter

	tree, pErr := h.Parser.Parse(statePersistenceAndSandboxingTestScript)
	if pErr != nil {
		t.Fatalf("Failed to parse script: %v", pErr)
	}
	program, _, bErr := h.ASTBuilder.Build(tree)
	if bErr != nil {
		t.Fatalf("Failed to build AST: %v", bErr)
	}
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}
	t.Logf("[DEBUG] Turn 2: Script parsed and loaded.")

	_, runErr := interp.Run("_SetLocalVariable")
	if runErr != nil {
		t.Fatalf("Procedure failed unexpectedly: %v", runErr)
	}
	t.Logf("[DEBUG] Turn 3: '_SetLocalVariable' procedure executed.")

	_, exists := interp.GetVariable("this_should_not_leak")
	if exists {
		t.Fatal("Variable 'this_should_not_leak' leaked from sandboxed procedure into the parent interpreter's scope.")
	}
	t.Logf("[DEBUG] Turn 4: Assertion passed, variable did not leak.")
	t.Log("Successfully verified that local procedure variables do not persist.")
}
