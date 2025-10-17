// NeuroScript Version: 0.8.0
// File version: 39
// Purpose: Removes the obsolete 'tool.aeiou.magic' call from the mock response.
// filename: pkg/interpreter/interpreter_ask_integration_test.go
// nlines: 98
// risk_rating: LOW

package interpreter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
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

func loadAskTestScript(t *testing.T, h *TestHarness) {
	t.Helper()
	t.Logf("[DEBUG] Turn X: Loading ask test script.")
	tree, pErr := h.Parser.Parse(askTestScriptV3)
	if pErr != nil {
		t.Fatalf("Failed to parse embedded script: %v", pErr)
	}

	program, _, bErr := h.ASTBuilder.Build(tree)
	if bErr != nil {
		t.Fatalf("Failed to build AST: %v", bErr)
	}
	if err := h.Interpreter.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load program into interpreter: %v", err)
	}
}

func TestAskIntegrationV3(t *testing.T) {
	t.Run("Basic ask statement success", func(t *testing.T) {
		h, mockProv := setupAskTest(t)
		// THE FIX: The AI's job is just to emit the answer.
		// The Go loop handles loop termination.
		actions := `command
		    emit "Victoria"
		endcommand`
		envText, _ := (&aeiou.Envelope{UserData: "{}", Actions: actions}).Compose()
		mockProv.ResponseToReturn = &provider.AIResponse{TextContent: envText}

		loadAskTestScript(t, h)
		t.Logf("[DEBUG] Turn 3: Starting 'TestBasicSuccess' procedure.")

		finalResult, err := h.Interpreter.Run("TestBasicSuccess")
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}

		resultStr, _ := lang.ToString(finalResult)
		if !strings.Contains(resultStr, "Victoria") {
			t.Errorf("Expected result to contain 'Victoria', got '%s'", resultStr)
		}
		t.Logf("[DEBUG] Turn 4: 'TestBasicSuccess' completed.")
	})

	t.Run("Ask statement with provider error", func(t *testing.T) {
		h, mockProv := setupAskTest(t)
		providerErr := errors.New("provider API key invalid")
		mockProv.ErrorToReturn = providerErr
		loadAskTestScript(t, h)
		t.Logf("[DEBUG] Turn 3: Starting 'TestProviderError' procedure.")

		_, err := h.Interpreter.Run("TestProviderError")
		if err == nil {
			t.Fatal("Script execution was expected to fail, but it succeeded.")
		}

		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) || !strings.Contains(rtErr.Error(), "provider API key invalid") {
			t.Errorf("Expected a provider error, but got: %v", err)
		}
		t.Logf("[DEBUG] Turn 4: 'TestProviderError' completed with expected error.")
	})

	t.Run("Ask with non-existent AgentModel", func(t *testing.T) {
		h, _ := setupAskTest(t) // We don't need the provider here
		loadAskTestScript(t, h)
		t.Logf("[DEBUG] Turn 3: Starting 'TestNonExistentAgent' procedure.")
		_, err := h.Interpreter.Run("TestNonExistentAgent")
		if err == nil {
			t.Fatal("Script execution was expected to fail, but it succeeded.")
		}

		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) || rtErr.Code != lang.ErrorCodeKeyNotFound {
			t.Errorf("Expected a KeyNotFound error, but got: %v", err)
		}
		t.Logf("[DEBUG] Turn 4: 'TestNonExistentAgent' completed with expected error.")
	})
}
