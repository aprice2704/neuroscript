// NeuroScript Version: 0.7.0
// File version: 10
// Purpose: Corrected the agentic capsule detection marker to be more reliable.
// filename: pkg/llmconn/llmconn_test.go
// nlines: 225
// risk_rating: LOW

package llmconn

import (
	"context"
	"errors"
	"strings"
	"testing"

	// This blank import is critical. It forces the Go compiler to include
	// the bootstrap capsule package in the test binary, which in turn
	// triggers its init() function, populating the global capsule registry.
	// Without this, capsule.GetLatest() would fail to find the required prompts.
	_ "github.com/aprice2704/neuroscript/pkg/capsule"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/provider/test"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// capturingMockProvider is a mock used to capture the request sent to it.
type capturingMockProvider struct {
	lastRequest      *provider.AIRequest
	responseToReturn *provider.AIResponse
	errorToReturn    error
}

func (p *capturingMockProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	p.lastRequest = &req
	if p.errorToReturn != nil {
		return nil, p.errorToReturn
	}
	if p.responseToReturn != nil {
		return p.responseToReturn, nil
	}
	// Return a default response
	respText, _ := test.WrapResponseInAEIOU("captured")
	return &provider.AIResponse{TextContent: respText}, nil
}

func TestNewLLMConn(t *testing.T) {
	mockProvider := test.New()
	model := &types.AgentModel{
		Name: "test-model",
	}

	t.Run("Successful creation", func(t *testing.T) {
		conn, err := New(model, mockProvider)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if conn == nil {
			t.Fatal("Expected a valid LLMConn instance, got nil")
		}
	})

	t.Run("Fails with nil model", func(t *testing.T) {
		_, err := New(nil, mockProvider)
		if !errors.Is(err, ErrModelNotSet) {
			t.Errorf("Expected error %v, got %v", ErrModelNotSet, err)
		}
	})

	t.Run("Fails with nil provider", func(t *testing.T) {
		_, err := New(model, nil)
		if !errors.Is(err, ErrProviderNotSet) {
			t.Errorf("Expected error %v, got %v", ErrProviderNotSet, err)
		}
	})
}

func TestLLMConn_Converse_Lifecycle(t *testing.T) {
	mockProvider := test.New()
	model := &types.AgentModel{
		Name:     "test-model",
		MaxTurns: 2,
		Tools: types.ToolConfig{
			ToolLoopPermitted: true,
		},
	}
	ctx := context.Background()

	t.Run("Successful conversation turn", func(t *testing.T) {
		conn, err := New(model, mockProvider)
		if err != nil {
			t.Fatal(err)
		}

		inputEnv := &aeiou.Envelope{UserData: "ping", Actions: "command endcommand"}
		resp, err := conn.Converse(ctx, inputEnv)
		if err != nil {
			t.Fatalf("Converse() failed: %v", err)
		}
		if resp == nil {
			t.Fatal("Expected a response, got nil")
		}
		if conn.TurnCount() != 1 {
			t.Errorf("Expected turn count to be 1, got %d", conn.TurnCount())
		}
	})

	t.Run("Exceeds max turns", func(t *testing.T) {
		conn, err := New(model, mockProvider)
		if err != nil {
			t.Fatal(err)
		}
		inputEnv := &aeiou.Envelope{UserData: "ping", Actions: "command endcommand"}

		// First two turns should succeed
		_, _ = conn.Converse(ctx, inputEnv)
		_, _ = conn.Converse(ctx, inputEnv)
		if conn.TurnCount() != 2 {
			t.Fatalf("Expected turn count to be 2, got %d", conn.TurnCount())
		}

		// Third turn should fail
		_, err = conn.Converse(ctx, inputEnv)
		if !errors.Is(err, ErrMaxTurnsExceeded) {
			t.Errorf("Expected error %v, got %v", ErrMaxTurnsExceeded, err)
		}
	})
}

func TestLLMConn_Converse_RequestPopulation(t *testing.T) {
	const testAPIKey = "test-secret-key"
	mockProvider := &capturingMockProvider{}
	ctx := context.Background()
	inputEnv := &aeiou.Envelope{UserData: "test prompt", Actions: "command endcommand"}

	// Define robust markers to look for in the prompts.
	agenticMarker := `make observable progress`
	oneshotMarker := `Do not alter any part of the envelope except ACTIONS.`

	t.Run("Agentic model gets agentic bootstrap", func(t *testing.T) {
		model := &types.AgentModel{
			APIKey: testAPIKey,
			Tools:  types.ToolConfig{ToolLoopPermitted: true},
		}
		conn, _ := New(model, mockProvider)
		_, err := conn.Converse(ctx, inputEnv)
		if err != nil {
			t.Fatalf("Converse() failed: %v", err)
		}
		if mockProvider.lastRequest == nil {
			t.Fatal("Provider was not called")
		}
		if mockProvider.lastRequest.APIKey != testAPIKey {
			t.Errorf("APIKey mismatch: got '%s'", mockProvider.lastRequest.APIKey)
		}
		if !strings.Contains(mockProvider.lastRequest.Prompt, agenticMarker) {
			t.Error("Prompt is missing agentic-specific bootstrap text")
		}
	})

	t.Run("One-shot model gets one-shot bootstrap", func(t *testing.T) {
		model := &types.AgentModel{
			APIKey: testAPIKey,
			Tools:  types.ToolConfig{ToolLoopPermitted: false},
		}
		conn, _ := New(model, mockProvider)
		_, err := conn.Converse(ctx, inputEnv)
		if err != nil {
			t.Fatalf("Converse() failed: %v", err)
		}
		if mockProvider.lastRequest == nil {
			t.Fatal("Provider was not called")
		}
		// Positive check: ensure it still gets a bootstrap prompt.
		if !strings.Contains(mockProvider.lastRequest.Prompt, oneshotMarker) {
			t.Error("Prompt appears to be missing one-shot bootstrap text")
		}
		// Negative check: ensure it does NOT get the agentic/looping instructions.
		if strings.Contains(mockProvider.lastRequest.Prompt, agenticMarker) {
			t.Error("One-shot prompt should not contain agentic-specific bootstrap text")
		}
	})

	t.Run("Provider error is wrapped correctly", func(t *testing.T) {
		providerErr := errors.New("API rate limit exceeded")
		mockProvider.errorToReturn = providerErr
		conn, _ := New(&types.AgentModel{}, mockProvider)

		_, err := conn.Converse(ctx, inputEnv)

		if err == nil {
			t.Fatal("Expected an error but got nil")
		}
		if !errors.Is(err, providerErr) {
			t.Errorf("Returned error does not wrap the original provider error. Got: %v", err)
		}
		if !strings.Contains(err.Error(), "provider chat failed on turn 1") {
			t.Error("Error message is missing the correct turn count wrapper")
		}
	})

	t.Run("State accumulates correctly", func(t *testing.T) {
		mockProvider.errorToReturn = nil
		mockProvider.responseToReturn = &provider.AIResponse{
			InputTokens:  10,
			OutputTokens: 20,
			Cost:         0.005,
		}
		conn, _ := New(&types.AgentModel{}, mockProvider)

		_, err := conn.Converse(ctx, inputEnv)
		if err != nil {
			t.Fatalf("Converse failed on first call: %v", err)
		}

		// First turn check
		if conn.TurnCount() != 1 {
			t.Errorf("Expected turn count 1, got %d", conn.TurnCount())
		}
		if conn.TotalTokens() != 30 {
			t.Errorf("Expected total tokens 30, got %d", conn.TotalTokens())
		}
		if conn.TotalCost() != 0.005 {
			t.Errorf("Expected total cost 0.005, got %f", conn.TotalCost())
		}

		// Second turn
		_, err = conn.Converse(ctx, inputEnv)
		if err != nil {
			t.Fatalf("Converse failed on second call: %v", err)
		}

		// Second turn check (accumulation)
		if conn.TurnCount() != 2 {
			t.Errorf("Expected turn count 2, got %d", conn.TurnCount())
		}
		if conn.TotalTokens() != 60 {
			t.Errorf("Expected total tokens 60, got %d", conn.TotalTokens())
		}
		if conn.TotalCost() != 0.010 {
			t.Errorf("Expected total cost 0.010, got %f", conn.TotalCost())
		}
	})
}
