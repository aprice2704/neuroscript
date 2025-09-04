// NeuroScript Version: 0.7.0
// File version: 36
// Purpose: Reverted test scripts to use simple string prompts, aligning with the corrected 'ask' statement logic that now handles envelope creation internally.
// filename: pkg/interpreter/interpreter_ask_integration_test.go
// nlines: 135
// risk_rating: LOW

package interpreter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/llmconn"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// The test script now uses simple, readable string prompts.
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
		interp, _ := setupAskTestV3(t)
		// Use a mock connector for this test instead of a raw provider
		mockConn := llmconn.NewMock(t,
			llmconn.Done("Victoria"),
		)
		interp.RegisterProvider("mock_ask_provider", mockConn)
		loadAskTestScript(t, interp)

		finalResult, err := interp.Run("TestBasicSuccess")
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}

		resultStr, _ := lang.ToString(finalResult)
		if !strings.Contains(resultStr, "Victoria") {
			t.Errorf("Expected result to contain 'Victoria', got '%s'", resultStr)
		}
	})

	t.Run("Ask statement with provider error", func(t *testing.T) {
		interp, _ := setupAskTestV3(t)
		providerErr := errors.New("provider API key invalid")
		mockConn := llmconn.NewMock(t, llmconn.Error(providerErr))
		interp.RegisterProvider("mock_ask_provider", mockConn)
		loadAskTestScript(t, interp)

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
