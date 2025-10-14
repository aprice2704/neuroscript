// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Added tests for privilege inheritance, function redefinition, and variable scoping between root and forks.
// filename: pkg/interpreter/inheritance_test.go
// nlines: 200
// risk_rating: LOW

package interpreter_test

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
)

func TestInterpreter_Inheritance(t *testing.T) {

	t.Run("Functions loaded in root are available in forks", func(t *testing.T) {
		t.Logf("[DEBUG] Starting test: Functions loaded in root are available in forks")
		h := NewTestHarness(t)
		rootInterpreter := h.Interpreter

		// 1. Load a script with a function into the root interpreter.
		rootScript := `
			func root_function() means
				return "hello from root"
			endfunc
		`
		tree, _ := h.Parser.Parse(rootScript)
		program, _, _ := h.ASTBuilder.Build(tree)
		if err := rootInterpreter.Load(&interfaces.Tree{Root: program}); err != nil {
			t.Fatalf("Failed to load script into root interpreter: %v", err)
		}
		t.Logf("[DEBUG] Root script loaded.")

		// 2. Fork the interpreter.
		forkedInterpreter := rootInterpreter.Clone()
		t.Logf("[DEBUG] Interpreter forked.")

		// 3. Execute a script in the FORK that calls the ROOT function.
		forkScript := `
			func main() means
				return root_function()
			endfunc
		`

		// We must APPEND the new script to the fork, not load, to preserve inherited procedures.
		forkTree, _ := h.Parser.Parse(forkScript)
		forkProgram, _, _ := h.ASTBuilder.Build(forkTree)
		if err := forkedInterpreter.AppendScript(&interfaces.Tree{Root: forkProgram}); err != nil {
			t.Fatalf("Failed to append script to forked interpreter: %v", err)
		}
		t.Logf("[DEBUG] Fork script appended.")

		// 4. Run the 'main' procedure on the fork.
		result, err := forkedInterpreter.Run("main")
		if err != nil {
			t.Fatalf("Forked interpreter failed to run script calling root function: %v", err)
		}
		t.Logf("[DEBUG] Forked interpreter ran 'main'.")

		// 5. Verify the result.
		expected := lang.StringValue{Value: "hello from root"}
		if result != expected {
			t.Errorf("Expected result from root function, got %#v", result)
		}
		t.Logf("[DEBUG] Test passed.")
	})

	t.Run("Functions loaded in a fork are NOT available in the root", func(t *testing.T) {
		t.Logf("[DEBUG] Starting test: Functions loaded in a fork are NOT available in the root")
		h := NewTestHarness(t)
		rootInterpreter := h.Interpreter
		forkedInterpreter := rootInterpreter.Clone()
		t.Logf("[DEBUG] Interpreters created and forked.")

		// 1. Load a function only into the fork.
		forkScript := `
			func fork_only_function() means
				return "this should not be callable from root"
			endfunc
		`
		tree, _ := h.Parser.Parse(forkScript)
		program, _, _ := h.ASTBuilder.Build(tree)
		if err := forkedInterpreter.Load(&interfaces.Tree{Root: program}); err != nil {
			t.Fatalf("Failed to load script into fork: %v", err)
		}
		t.Logf("[DEBUG] Script with 'fork_only_function' loaded into fork.")

		// 2. Attempt to call the fork's function from the root.
		_, err := rootInterpreter.Run("fork_only_function")
		if err == nil {
			t.Fatal("Root interpreter was able to call a function defined in a fork, but it should not have been possible.")
		}

		if _, ok := err.(*lang.RuntimeError); !ok || err.(*lang.RuntimeError).Code != lang.ErrorCodeProcNotFound {
			t.Errorf("Expected a ProcedureNotFound error, but got: %v", err)
		}
		t.Logf("[DEBUG] Correctly received expected error: %v", err)
		t.Logf("[DEBUG] Test passed.")
	})

	t.Run("AgentModels registered in root are available in forks", func(t *testing.T) {
		t.Logf("[DEBUG] Starting test: AgentModels registered in root are available in forks")
		h := NewTestHarness(t)
		rootInterpreter := h.Interpreter

		// 1. Register an AgentModel in the root.
		config := map[string]lang.Value{
			"provider": lang.StringValue{Value: "p"},
			"model":    lang.StringValue{Value: "m"},
		}
		if err := rootInterpreter.RegisterAgentModel("shared_agent", config); err != nil {
			t.Fatalf("Failed to register agent model in root: %v", err)
		}
		t.Logf("[DEBUG] AgentModel 'shared_agent' registered in root.")

		// 2. Fork the interpreter.
		forkedInterpreter := rootInterpreter.Clone()
		t.Logf("[DEBUG] Interpreter forked.")

		// 3. Use a tool in the fork to access the agent model.
		forkScript := `
			func main() means
				return tool.agentmodel.Get("shared_agent")
			endfunc
		`
		tree, _ := h.Parser.Parse(forkScript)
		program, _, _ := h.ASTBuilder.Build(tree)
		if err := forkedInterpreter.AppendScript(&interfaces.Tree{Root: program}); err != nil {
			t.Fatalf("Failed to load script into fork: %v", err)
		}
		t.Logf("[DEBUG] Script loaded into fork.")

		// 4. Run and check the result.
		result, err := forkedInterpreter.Run("main")
		if err != nil {
			t.Fatalf("Fork failed to run script accessing root's agent model: %v", err)
		}
		t.Logf("[DEBUG] Fork 'main' executed.")

		resultMap, ok := result.(*lang.MapValue)
		if !ok {
			t.Fatalf("Expected tool call to return a map, but got %T", result)
		}
		if model, _ := resultMap.Value["model"].(lang.StringValue); model.Value != "m" {
			t.Errorf("Fork retrieved incorrect model data. Expected model 'm', got '%s'", model.Value)
		}
		t.Logf("[DEBUG] Test passed.")
	})

	t.Run("Fork inherits parent policy and cannot be less privileged", func(t *testing.T) {
		h := NewTestHarness(t)
		// The harness provides a fully privileged root interpreter by default.
		rootInterpreter := h.Interpreter

		forkedInterpreter := rootInterpreter.Clone()

		// Attempt a privileged operation from the forked interpreter.
		// Because the parent was privileged, the fork should be too.
		err := forkedInterpreter.RegisterAgentModel("test_in_fork", map[string]lang.Value{
			"provider": lang.StringValue{Value: "p"},
			"model":    lang.StringValue{Value: "m"},
		})

		if err != nil {
			t.Fatalf("Fork of a privileged interpreter failed a privileged operation: %v", err)
		}
	})

	t.Run("Fork of unprivileged parent is also unprivileged", func(t *testing.T) {
		h := NewTestHarness(t)
		// Create a new, unprivileged interpreter for this test.
		unprivilegedPolicy := policy.NewBuilder(policy.ContextNormal).Build()
		rootInterpreter := interpreter.NewInterpreter(
			interpreter.WithHostContext(h.HostContext),
			interpreter.WithExecPolicy(unprivilegedPolicy),
		)

		forkedInterpreter := rootInterpreter.Clone()

		err := forkedInterpreter.RegisterAgentModel("test_in_fork", map[string]lang.Value{
			"provider": lang.StringValue{Value: "p"},
			"model":    lang.StringValue{Value: "m"},
		})

		if err == nil {
			t.Fatal("Fork of an unprivileged interpreter was able to perform a privileged action.")
		}
	})

	t.Run("Fork cannot redefine a function from root", func(t *testing.T) {
		h := NewTestHarness(t)
		rootInterpreter := h.Interpreter

		rootScript := `func my_func() means return "root" endfunc`
		tree, _ := h.Parser.Parse(rootScript)
		program, _, _ := h.ASTBuilder.Build(tree)
		if err := rootInterpreter.Load(&interfaces.Tree{Root: program}); err != nil {
			t.Fatalf("Failed to load root script: %v", err)
		}

		forkedInterpreter := rootInterpreter.Clone()

		conflictScript := `func my_func() means return "fork" endfunc`
		conflictTree, _ := h.Parser.Parse(conflictScript)
		conflictProgram, _, _ := h.ASTBuilder.Build(conflictTree)
		err := forkedInterpreter.AppendScript(&interfaces.Tree{Root: conflictProgram})

		if err == nil {
			t.Fatal("Expected an error when redefining a function in a fork, but got nil.")
		}
		if !errors.Is(err, lang.ErrProcedureExists) {
			t.Errorf("Expected ErrProcedureExists, but got: %v", err)
		}
	})

	t.Run("Variables set in a fork do not leak to parent", func(t *testing.T) {
		h := NewTestHarness(t)
		rootInterpreter := h.Interpreter
		forkedInterpreter := rootInterpreter.Clone()

		// Set a variable directly on the forked interpreter's state.
		err := forkedInterpreter.SetVariable("leaky_var", lang.StringValue{Value: "i_am_in_a_fork"})
		if err != nil {
			t.Fatalf("SetVariable on fork failed: %v", err)
		}

		// Attempt to retrieve the same variable from the root.
		_, exists := rootInterpreter.GetVariable("leaky_var")
		if exists {
			t.Fatal("Variable 'leaky_var' from fork leaked into the root interpreter's scope.")
		}
	})
}
