// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Provides unit tests for the LLMConn stateful connection manager.
// filename: pkg/llmconn/llmconn_test.go
// nlines: 105
// risk_rating: LOW

package llmconn

import (
	"context"
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/provider/test"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestNewLLMConn(t *testing.T) {
	mockProvider := test.New()
	loopableModel := &types.AgentModel{
		Name: "test-model",
		Tools: types.ToolConfig{
			ToolLoopPermitted: true,
		},
	}
	nonLoopableModel := &types.AgentModel{
		Name: "test-model-no-loop",
		Tools: types.ToolConfig{
			ToolLoopPermitted: false,
		},
	}

	t.Run("Successful creation", func(t *testing.T) {
		conn, err := New(loopableModel, mockProvider)
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
		_, err := New(loopableModel, nil)
		if !errors.Is(err, ErrProviderNotSet) {
			t.Errorf("Expected error %v, got %v", ErrProviderNotSet, err)
		}
	})

	t.Run("Fails if loops are not permitted", func(t *testing.T) {
		_, err := New(nonLoopableModel, mockProvider)
		if !errors.Is(err, ErrLoopNotPermitted) {
			t.Errorf("Expected error %v, got %v", ErrLoopNotPermitted, err)
		}
	})
}

func TestLLMConn_Converse(t *testing.T) {
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

		inputEnv := &aeiou.Envelope{Orchestration: "ping"}
		resp, err := conn.Converse(ctx, inputEnv)
		if err != nil {
			t.Fatalf("Converse() failed: %v", err)
		}
		if resp == nil {
			t.Fatal("Expected a response, got nil")
		}
		if conn.turnCount != 1 {
			t.Errorf("Expected turn count to be 1, got %d", conn.turnCount)
		}
	})

	t.Run("Exceeds max turns", func(t *testing.T) {
		conn, err := New(model, mockProvider)
		if err != nil {
			t.Fatal(err)
		}
		inputEnv := &aeiou.Envelope{Orchestration: "ping"}

		// First two turns should succeed
		_, _ = conn.Converse(ctx, inputEnv)
		_, _ = conn.Converse(ctx, inputEnv)
		if conn.turnCount != 2 {
			t.Fatalf("Expected turn count to be 2, got %d", conn.turnCount)
		}

		// Third turn should fail
		_, err = conn.Converse(ctx, inputEnv)
		if !errors.Is(err, ErrMaxTurnsExceeded) {
			t.Errorf("Expected error %v, got %v", ErrMaxTurnsExceeded, err)
		}
	})
}
