// NeuroScript Version: 0.7.0
// File version: 40
// Purpose: Removed duplicated test helper functions, now using the centralized helpers from 'interpreter_test_helpers.go'.
// filename: pkg/interpreter/interpreter_steps_ask_test.go
// nlines: 60
// risk_rating: MEDIUM

package interpreter_test

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// Note: The mockAskProviderV3 struct and setupAskTestV3 function have been
// moved to interpreter_test_helpers.go to avoid redeclaration errors.

func TestAskStatementExecutionV3(t *testing.T) {
	t.Run("Simple ask into variable", func(t *testing.T) {
		interp, mockProv := setupAskTestV3(t)

		// Set up the mock response for this specific test case
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

		script := `command ask "test_agent", "What is the capital of Canada?" into result endcommand`
		p := parser.NewParserAPI(nil)
		tree, _ := p.Parse(script)
		program, _, _ := parser.NewASTBuilder(nil).Build(tree)

		_, err := interp.Execute(program)
		if err != nil {
			t.Fatalf("executeAsk failed: %v", err)
		}

		resultVar, found := interp.GetVariable("result")
		if !found {
			t.Fatal("result variable not found")
		}
		resultStr, _ := lang.ToString(resultVar)
		expected := "The capital of Canada is Ottawa."
		if resultStr != expected {
			t.Errorf("Expected result variable to be '%s', got '%s'", expected, resultStr)
		}
	})

	t.Run("Ask with non-existent agent", func(t *testing.T) {
		interp, _ := setupAskTestV3(t)
		script := `command ask "no_such_agent", "prompt" endcommand`
		p := parser.NewParserAPI(nil)
		tree, _ := p.Parse(script)
		program, _, _ := parser.NewASTBuilder(nil).Build(tree)

		_, err := interp.Execute(program)
		if err == nil {
			t.Fatal("Expected an error for non-existent agent, but got nil")
		}
		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) || rtErr.Code != lang.ErrorCodeKeyNotFound {
			t.Errorf("Expected a KeyNotFound error, got %T: %v", err, err)
		}
	})
}
