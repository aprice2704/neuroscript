// NeuroScript Version: 0.7.0
// File version: 31
// Purpose: Corrected the call to interp.Load to pass the correct AST structure.
// filename: pkg/interpreter/interpreter_ask_integration_test.go
// nlines: 125
// risk_rating: MEDIUM

package interpreter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

const askTestScriptV3 = `
func TestBasicSuccess(returns result) means
    ask "test_agent", "What is the capital of BC?" into result
    return result
endfunc

func TestProviderError() means
    ask "test_agent", "This will cause a provider error."
endfunc

func TestNonExistentAgent() means
    ask "no_such_agent", "This will fail."
endfunc
`

// Note: The mockAskProviderV3 struct and setupAskTestV3 function have been
// moved to interpreter_test_helpers.go to avoid redeclaration errors.

func loadAskTestScript(t *testing.T, interp *interpreter.Interpreter) {
	parserAPI := parser.NewParserAPI(interp.GetLogger())
	p, pErr := parserAPI.Parse(askTestScriptV3)
	if pErr != nil {
		t.Fatalf("Failed to parse embedded script: %v", pErr)
	}

	program, _, bErr := parser.NewASTBuilder(interp.GetLogger()).Build(p)
	if bErr != nil {
		t.Fatalf("Failed to build AST: %v", bErr)
	}
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load program into interpreter: %v", err)
	}
}

func TestAskIntegrationV3(t *testing.T) {
	t.Run("Basic ask statement success", func(t *testing.T) {
		interp, mockProv := setupAskTestV3(t)
		loadAskTestScript(t, interp)

		actions := `
		command
			emit "Victoria"
			set p = {"action": "done"}
			emit tool.aeiou.magic("LOOP", p)
		endcommand
		`
		env := &aeiou.Envelope{UserData: "{}", Actions: actions}
		respText, _ := env.Compose()
		mockProv.ResponseToReturn = &provider.AIResponse{TextContent: respText}

		finalResult, err := interp.Run("TestBasicSuccess")
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}

		resultStr, _ := lang.ToString(finalResult)
		if resultStr != "Victoria" {
			t.Errorf("Expected result 'Victoria', got '%s'", resultStr)
		}
	})

	t.Run("Ask statement with provider error", func(t *testing.T) {
		interp, mockProv := setupAskTestV3(t)
		loadAskTestScript(t, interp)
		mockProv.ErrorToReturn = errors.New("provider API key invalid")

		_, err := interp.Run("TestProviderError")
		if err == nil {
			t.Fatal("Script execution was expected to fail, but it succeeded.")
		}

		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) || !strings.Contains(rtErr.Error(), "provider API key invalid") {
			t.Errorf("Expected a provider error, but got: %v", err)
		}
	})

	t.Run("Ask with non-existent AgentModel", func(t *testing.T) {
		interp, _ := setupAskTestV3(t)
		loadAskTestScript(t, interp)
		_, err := interp.Run("TestNonExistentAgent")
		if err == nil {
			t.Fatal("Script execution was expected to fail, but it succeeded.")
		}

		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) || rtErr.Code != lang.ErrorCodeKeyNotFound {
			t.Errorf("Expected a KeyNotFound error, but got: %v", err)
		}
	})
}
