// NeuroScript Version: 0.7.0
// File version: 6
// Purpose: Corrected the account registration in the test script to use 'api_key' instead of 'apiKey', aligning with the new JSON key standard.
// filename: pkg/interpreter/interpreter_state_persistence_test.go
// nlines: 120
// risk_rating: LOW

package interpreter_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

const statePersistenceAndSandboxingTestScript = `
# This script is for testing state persistence and variable sandboxing.

func _SetupState() means
    # Register an account and a model. These must persist.
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
    # Try to access the model registered in the setup phase.
    # This will fail if the state was not persisted.
    set model = tool.agentmodel.Get("test_agent_for_persistence")
    must not is_error(model)
    return "check_ok"
endfunc

func _SetLocalVariable() means
    # This variable should NOT persist in the caller's scope.
    set this_should_not_leak = "i am a local variable"
    return "local_set_ok"
endfunc
`

// TestStatePersistence_StoresPersistAcrossRuns verifies that state modifications made in one
// Run() call (like registering an agent) are still present in a subsequent
// Run() call on the same interpreter instance.
func TestStatePersistence_StoresPersistAcrossRuns(t *testing.T) {
	interp, err := interpreter.NewTestInterpreter(t, nil, nil, true)
	if err != nil {
		t.Fatalf("Failed to create privileged test interpreter: %v", err)
	}

	parserAPI := parser.NewParserAPI(interp.GetLogger())
	p, pErr := parserAPI.Parse(statePersistenceAndSandboxingTestScript)
	if pErr != nil {
		t.Fatalf("Failed to parse script: %v", pErr)
	}
	program, _, bErr := parser.NewASTBuilder(interp.GetLogger()).Build(p)
	if bErr != nil {
		t.Fatalf("Failed to build AST: %v", bErr)
	}
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Run the setup procedure.
	_, setupErr := interp.Run("_SetupState")
	if setupErr != nil {
		t.Fatalf("Setup procedure failed unexpectedly: %v", setupErr)
	}

	// Run the check procedure ON THE SAME INTERPRETER. It should find the agent.
	_, checkErr := interp.Run("_CheckState")
	if checkErr != nil {
		t.Fatalf("Check procedure failed, indicating state was not persisted: %v", checkErr)
	}

	t.Log("Successfully verified that store state persists across multiple Run() calls.")
}

// TestStatePersistence_VariablesAreSandboxed verifies that variables set inside a
// procedure call are properly sandboxed and do not "leak" into the parent
// interpreter's scope.
func TestStatePersistence_VariablesAreSandboxed(t *testing.T) {
	interp, err := interpreter.NewTestInterpreter(t, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to create test interpreter: %v", err)
	}

	parserAPI := parser.NewParserAPI(interp.GetLogger())
	p, pErr := parserAPI.Parse(statePersistenceAndSandboxingTestScript)
	if pErr != nil {
		t.Fatalf("Failed to parse script: %v", pErr)
	}
	program, _, bErr := parser.NewASTBuilder(interp.GetLogger()).Build(p)
	if bErr != nil {
		t.Fatalf("Failed to build AST: %v", bErr)
	}
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Run the procedure that sets a local variable.
	_, runErr := interp.Run("_SetLocalVariable")
	if runErr != nil {
		t.Fatalf("Procedure failed unexpectedly: %v", runErr)
	}

	// Check the interpreter's state. The variable should NOT exist.
	_, exists := interp.GetVariable("this_should_not_leak")
	if exists {
		t.Fatal("Variable 'this_should_not_leak' leaked from sandboxed procedure into the parent interpreter's scope.")
	}

	t.Log("Successfully verified that local procedure variables do not persist.")
}
