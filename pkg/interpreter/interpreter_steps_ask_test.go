// NeuroScript Version: 0.5.2
// File version: 1.4.0
// Purpose: Corrected the mock tool implementation to handle primitive Go types, fixing the final test failure.
// filename: pkg/interpreter/interpreter_steps_ask_test.go
// nlines: 160
// risk_rating: MEDIUM

package interpreter

import (
	"context"
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/google/generative-ai-go/genai"
)

// --- Mock LLM Client for Testing ---

type mockLLMClient struct {
	LastPrompt         string
	ResponseToReturn   string
	ToolCallToReturn   *interfaces.ToolCall
	ErrorToReturn      error
	ToolsPresentedWith []interfaces.ToolDefinition
}

func (m *mockLLMClient) Ask(ctx context.Context, turns []*interfaces.ConversationTurn) (*interfaces.ConversationTurn, error) {
	if m.ErrorToReturn != nil {
		return nil, m.ErrorToReturn
	}
	if len(turns) > 0 {
		m.LastPrompt = turns[len(turns)-1].Content
	}
	return &interfaces.ConversationTurn{Role: "model", Content: m.ResponseToReturn}, nil
}

func (m *mockLLMClient) AskWithTools(ctx context.Context, turns []*interfaces.ConversationTurn, tools []interfaces.ToolDefinition) (*interfaces.ConversationTurn, []*interfaces.ToolCall, error) {
	if m.ErrorToReturn != nil {
		return nil, nil, m.ErrorToReturn
	}
	if len(turns) > 0 {
		m.LastPrompt = turns[len(turns)-1].Content
	}
	m.ToolsPresentedWith = tools

	var toolCalls []*interfaces.ToolCall
	if m.ToolCallToReturn != nil {
		toolCalls = append(toolCalls, m.ToolCallToReturn)
	}
	return &interfaces.ConversationTurn{Role: "model", Content: m.ResponseToReturn}, toolCalls, nil
}

func (m *mockLLMClient) Client() *genai.Client { return nil }

func (m *mockLLMClient) Embed(ctx context.Context, text string) ([]float32, error) {
	if m.ErrorToReturn != nil {
		return nil, m.ErrorToReturn
	}
	return make([]float32, 16), nil
}

// --- Tests ---

func TestAskStatement(t *testing.T) {
	t.Run("Simple ask into variable", func(t *testing.T) {
		mockLLM := &mockLLMClient{
			ResponseToReturn: "The capital of Canada is Ottawa.",
		}
		interp, _ := newLocalTestInterpreter(t, nil, nil)
		interp.aiWorker = mockLLM

		script := `ask "What is the capital of Canada?" into result`
		_, err := interp.ExecuteScriptString("ask_test", script, nil)
		if err != nil {
			t.Fatalf("ExecuteScriptString failed: %v", err)
		}

		expectedPrompt := "What is the capital of Canada?"
		if mockLLM.LastPrompt != expectedPrompt {
			t.Errorf("Expected prompt '%s', got '%s'", expectedPrompt, mockLLM.LastPrompt)
		}

		resultVar, exists := interp.GetVariable("result")
		if !exists {
			t.Fatal("Variable 'result' was not set by the ask statement")
		}
		resultStr, _ := lang.ToString(resultVar)
		if resultStr != mockLLM.ResponseToReturn {
			t.Errorf("Expected result variable to be '%s', got '%s'", mockLLM.ResponseToReturn, resultStr)
		}
	})

	t.Run("Ask AI with tool calling", func(t *testing.T) {
		mockLLM := &mockLLMClient{
			ToolCallToReturn: &interfaces.ToolCall{
				ID:   "call_123",
				Name: "GetWeather",
				Arguments: map[string]interface{}{
					"location": "Ottawa, ON",
				},
			},
		}

		interp, _ := newLocalTestInterpreter(t, nil, nil)
		interp.aiWorker = mockLLM

		var toolWasCalledWith string
		weatherTool := tool.ToolImplementation{
			Spec: tool.ToolSpec{Name: "GetWeather", Args: []tool.ArgSpec{{Name: "location", Type: "string"}}},
			Func: func(rt tool.Runtime, args []interface{}) (interface{}, error) {
				// FIX: The argument from the AI comes in as a primitive Go type.
				// We need to type-assert it, not use lang.ToString.
				if len(args) > 0 {
					if val, ok := args[0].(string); ok {
						toolWasCalledWith = val
					}
				}
				return lang.StringValue{Value: "Sunny, 19C"}, nil
			},
		}
		interp.ToolRegistry().RegisterTool(weatherTool)

		step := ast.Step{
			Type: "ask ai",
			Values: []ast.Expression{
				&ast.StringLiteralNode{Value: "What is the weather in Ottawa?"},
			},
		}

		err := interp.executeAskAI(step)
		if err != nil {
			t.Fatalf("executeAskAI failed: %v", err)
		}

		if toolWasCalledWith != "Ottawa, ON" {
			t.Errorf("Expected mock tool to be called with 'Ottawa, ON', got '%s'", toolWasCalledWith)
		}
	})

	t.Run("Ask with LLM returning an error", func(t *testing.T) {
		mockLLM := &mockLLMClient{
			ErrorToReturn: errors.New("API limit reached"),
		}
		interp, _ := newLocalTestInterpreter(t, nil, nil)
		interp.aiWorker = mockLLM

		script := `ask "This will fail" into result`
		_, err := interp.ExecuteScriptString("ask_fail_test", script, nil)

		if err == nil {
			t.Fatal("Expected an error from the LLM, but got nil")
		}
		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) {
			t.Fatalf("Expected a RuntimeError, but got %T", err)
		}
		if rtErr.Code != lang.ErrorCodeLLMError {
			t.Errorf("Expected error code %v, got %v", lang.ErrorCodeLLMError, rtErr.Code)
		}
	})
}
