// NeuroScript Version: 0.6.0
// File version: 7.0.0
// Purpose: Aligned test to use the canonical types.AgentModel, resolving type assertion failures.
// filename: pkg/interpreter/interpreter_steps_ask_test.go
// nlines: 120
// risk_rating: MEDIUM

package interpreter

import (
	"context"
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// --- Mock AI Provider for Testing ---

type mockProvider struct {
	LastRequest      provider.AIRequest
	ResponseToReturn *provider.AIResponse
	ErrorToReturn    error
}

func (m *mockProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	m.LastRequest = req
	if m.ErrorToReturn != nil {
		return nil, m.ErrorToReturn
	}
	return m.ResponseToReturn, nil
}

// --- Tests ---

func TestAskStatementExecution(t *testing.T) {
	t.Run("Simple ask into variable", func(t *testing.T) {
		mockProv := &mockProvider{
			ResponseToReturn: &provider.AIResponse{TextContent: "The capital of Canada is Ottawa."},
		}
		interp, _ := NewTestInterpreter(t, nil, nil, true) // Run with privileges
		interp.RegisterProvider("mock_provider", mockProv)
		_ = interp.RegisterAgentModel("test_agent", map[string]lang.Value{
			"provider": lang.StringValue{Value: "mock_provider"},
			"model":    lang.StringValue{Value: "mock_model"},
		})

		step := ast.Step{
			Type: "ask",
			AskStmt: &ast.AskStmt{
				AgentModelExpr: &ast.StringLiteralNode{Value: "test_agent"},
				PromptExpr:     &ast.StringLiteralNode{Value: "What is the capital of Canada?"},
				IntoTarget:     &ast.LValueNode{Identifier: "result"},
			},
		}

		_, err := interp.executeAsk(step)
		if err != nil {
			t.Fatalf("executeAsk failed: %v", err)
		}

		expectedPrompt := "What is the capital of Canada?"
		if mockProv.LastRequest.Prompt != expectedPrompt {
			t.Errorf("Expected prompt '%s', got '%s'", expectedPrompt, mockProv.LastRequest.Prompt)
		}

		resultVar, exists := interp.GetVariable("result")
		if !exists {
			t.Fatal("Variable 'result' was not set by the ask statement")
		}
		resultStr, _ := lang.ToString(resultVar)
		if resultStr != mockProv.ResponseToReturn.TextContent {
			t.Errorf("Expected result variable to be '%s', got '%s'", mockProv.ResponseToReturn.TextContent, resultStr)
		}
	})

	t.Run("Ask with provider returning an error", func(t *testing.T) {
		mockProv := &mockProvider{
			ErrorToReturn: errors.New("API limit reached"),
		}
		interp, _ := NewTestInterpreter(t, nil, nil, true) // Run with privileges
		interp.RegisterProvider("mock_provider", mockProv)
		_ = interp.RegisterAgentModel("test_agent", map[string]lang.Value{
			"provider": lang.StringValue{Value: "mock_provider"},
			"model":    lang.StringValue{Value: "mock_model"},
		})

		step := ast.Step{
			Type: "ask",
			AskStmt: &ast.AskStmt{
				AgentModelExpr: &ast.StringLiteralNode{Value: "test_agent"},
				PromptExpr:     &ast.StringLiteralNode{Value: "This will fail"},
			},
		}

		_, err := interp.executeAsk(step)

		if err == nil {
			t.Fatal("Expected an error from the provider, but got nil")
		}
		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) {
			t.Fatalf("Expected a RuntimeError, but got %T", err)
		}
		// The error from the provider is wrapped as an external error.
		if rtErr.Code != lang.ErrorCodeExternal {
			t.Errorf("Expected error code %v, got %v", lang.ErrorCodeExternal, rtErr.Code)
		}
	})
}
