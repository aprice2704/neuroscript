// NeuroScript Version: 0.8.0
// File version: 43
// Purpose: Rewrote test to correctly validate 'ask...into' logic within a sandbox by checking emitted output.
// filename: pkg/interpreter/steps_ask_test.go
// nlines: 70
// risk_rating: LOW

package interpreter_test

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

func TestAskStatementExecutionV3(t *testing.T) {
	t.Run("Simple ask into variable", func(t *testing.T) {
		h, mockProv := setupAskTest(t)
		interp := h.Interpreter
		t.Logf("[DEBUG] Turn 1: Harness created for 'Simple ask into variable' test.")

		actions := `
		command
			emit "The capital of Canada is Ottawa."
			set p = {"action": "done"}
			emit tool.aeiou.magic("LOOP", p)
		endcommand
		`
		env := &aeiou.Envelope{UserData: "{}", Actions: actions}
		respText, _ := env.Compose()
		mockProv.ResponseToReturn = &provider.AIResponse{TextContent: respText}
		t.Logf("[DEBUG] Turn 2: Mock provider configured.")

		// The script now emits the result so we can check it from outside the sandbox.
		script := `command
			ask "test_agent", "What is the capital of Canada?" into result
			emit result
		endcommand`

		// Capture the emitted output.
		var capturedEmits []string
		h.HostContext.EmitFunc = func(v lang.Value) {
			s, _ := lang.ToString(v)
			capturedEmits = append(capturedEmits, s)
		}

		tree, _ := h.Parser.Parse(script)
		program, _, _ := h.ASTBuilder.Build(tree)
		interp.Load(&interfaces.Tree{Root: program})
		t.Logf("[DEBUG] Turn 3: Script parsed and loaded.")

		_, err := interp.Execute(program)
		if err != nil {
			t.Fatalf("execute failed: %v", err)
		}
		t.Logf("[DEBUG] Turn 4: Script executed.")

		if len(capturedEmits) != 1 {
			t.Fatalf("Expected script to emit 1 value, but it emitted %d", len(capturedEmits))
		}
		resultStr := capturedEmits[0]
		expected := "The capital of Canada is Ottawa."
		if resultStr != expected {
			t.Errorf("Expected result variable to be '%s', got '%s'", expected, resultStr)
		}
		t.Logf("[DEBUG] Turn 5: Assertion passed.")
	})

	t.Run("Ask with non-existent agent", func(t *testing.T) {
		h, _ := setupAskTest(t)
		interp := h.Interpreter
		t.Logf("[DEBUG] Turn 1: Harness created for 'Ask with non-existent agent' test.")

		script := `command ask "no_such_agent", "prompt" endcommand`
		tree, _ := h.Parser.Parse(script)
		program, _, _ := h.ASTBuilder.Build(tree)
		interp.Load(&interfaces.Tree{Root: program})
		t.Logf("[DEBUG] Turn 2: Script parsed and loaded.")

		_, err := interp.Execute(program)
		t.Logf("[DEBUG] Turn 3: Script executed, expecting error.")
		if err == nil {
			t.Fatal("Expected an error for non-existent agent, but got nil")
		}
		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) || rtErr.Code != lang.ErrorCodeKeyNotFound {
			t.Errorf("Expected a KeyNotFound error, got %T: %v", err, err)
		}
		t.Logf("[DEBUG] Turn 4: Correctly received expected error: %v", err)
	})
}
