// NeuroScript Version: 0.7.0
// File version: 18
// Purpose: Refactored to use a local mock provider and the correct AgentModelAdmin API.
// filename: pkg/interpreter/interpreter_steps_ask_test.go
// nlines: 115
// risk_rating: MEDIUM

package interpreter

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// mockSimpleProvider returns a single, final response.
type mockSimpleProvider struct{}

func (m *mockSimpleProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	responseActions := `command emit "The capital of Canada is Ottawa." endcommand`
	env := &aeiou.Envelope{Actions: responseActions}
	// The real ask loop doesn't need a LOOP:done signal if it's a one-shot response,
	// so we don't include one here to test that path.
	respText, _ := env.Compose()
	return &provider.AIResponse{TextContent: respText}, nil
}

// mockErrorProvider always returns an error.
type mockErrorProvider struct{}

func (m *mockErrorProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	return nil, errors.New("API limit reached")
}

func TestAskStatementExecution(t *testing.T) {
	t.Run("Simple ask into variable", func(t *testing.T) {
		// 1. Setup the mock provider.
		mockProv := &mockSimpleProvider{}

		// 2. Setup the interpreter.
		interp, _ := NewTestInterpreter(t, nil, nil, true)
		interp.RegisterProvider("mock_provider", mockProv)
		_ = interp.AgentModelsAdmin().Register("test_agent", map[string]any{
			"provider": "mock_provider",
			"model":    "mock_model",
		})

		// 3. Define the AST step to execute.
		step := ast.Step{
			Type: "ask",
			AskStmt: &ast.AskStmt{
				AgentModelExpr: &ast.StringLiteralNode{Value: "test_agent"},
				PromptExpr:     &ast.StringLiteralNode{Value: "What is the capital of Canada?"},
				IntoTarget:     &ast.LValueNode{Identifier: "result"},
			},
		}

		// 4. Execute and assert.
		_, err := interp.executeAsk(step)
		if err != nil {
			t.Fatalf("executeAsk failed: %v", err)
		}

		resultVar, exists := interp.GetVariable("result")
		if !exists {
			t.Fatal("Variable 'result' was not set by the ask statement")
		}
		resultStr, _ := lang.ToString(resultVar)
		expectedResult := "The capital of Canada is Ottawa."
		if resultStr != expectedResult {
			t.Errorf("Expected result variable to be '%s', got '%s'", expectedResult, resultStr)
		}
	})

	t.Run("Ask with provider returning an error", func(t *testing.T) {
		// 1. Setup the mock provider to return an error.
		mockProv := &mockErrorProvider{}

		// 2. Setup the interpreter.
		interp, _ := NewTestInterpreter(t, nil, nil, true)
		interp.RegisterProvider("mock_provider", mockProv)
		_ = interp.AgentModelsAdmin().Register("test_agent_err", map[string]any{
			"provider": "mock_provider",
			"model":    "err_model",
		})

		// 3. Define the AST step.
		step := ast.Step{
			Type: "ask",
			AskStmt: &ast.AskStmt{
				AgentModelExpr: &ast.StringLiteralNode{Value: "test_agent_err"},
				PromptExpr:     &ast.StringLiteralNode{Value: "This will fail"},
			},
		}

		// 4. Execute and assert the error.
		_, err := interp.executeAsk(step)

		if err == nil {
			t.Fatal("Expected an error from the provider, but got nil")
		}
		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) {
			t.Fatalf("Expected a RuntimeError, but got %T", err)
		}
		if !strings.Contains(rtErr.Error(), "AI provider call failed") {
			t.Errorf("Expected error to be about provider failure, got: %v", rtErr)
		}
	})
}
