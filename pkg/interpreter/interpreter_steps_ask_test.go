// NeuroScript Version: 0.7.0
// File version: 20
// Purpose: Corrected mock provider to generate syntactically valid ACTIONS blocks with newlines.
// filename: pkg/interpreter/interpreter_steps_ask_test.go
// nlines: 118
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
	responseActions := "command\n  emit \"The capital of Canada is Ottawa.\"\nendcommand"
	env := &aeiou.Envelope{Actions: responseActions}
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
		mockProv := &mockSimpleProvider{}
		interp, _ := NewTestInterpreter(t, nil, nil, true)
		interp.RegisterProvider("mock_provider", mockProv)
		err := interp.AgentModelsAdmin().Register("test_agent", map[string]any{
			"provider":          "mock_provider",
			"model":             "mock_model",
			"toolLoopPermitted": true,
		})
		if err != nil {
			t.Fatalf("Failed to register agent model: %v", err)
		}

		step := ast.Step{
			Type: "ask",
			AskStmt: &ast.AskStmt{
				AgentModelExpr: &ast.StringLiteralNode{Value: "test_agent"},
				PromptExpr:     &ast.StringLiteralNode{Value: "What is the capital of Canada?"},
				IntoTarget:     &ast.LValueNode{Identifier: "result"},
			},
		}

		_, err = interp.executeAsk(step)
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
		mockProv := &mockErrorProvider{}
		interp, _ := NewTestInterpreter(t, nil, nil, true)
		interp.RegisterProvider("mock_provider", mockProv)
		err := interp.AgentModelsAdmin().Register("test_agent_err", map[string]any{
			"provider":          "mock_provider",
			"model":             "err_model",
			"toolLoopPermitted": true,
		})
		if err != nil {
			t.Fatalf("Failed to register agent model: %v", err)
		}

		step := ast.Step{
			Type: "ask",
			AskStmt: &ast.AskStmt{
				AgentModelExpr: &ast.StringLiteralNode{Value: "test_agent_err"},
				PromptExpr:     &ast.StringLiteralNode{Value: "This will fail"},
			},
		}

		_, err = interp.executeAsk(step)

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
