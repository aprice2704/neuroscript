// NeuroScript Version: 0.7.2
// File version: 3
// Purpose: Corrects the test failures by providing a valid AEIOU envelope and using a provider mock that correctly simulates a failure within the conversation loop.
// filename: pkg/llmconn/emitter_test.go

package llmconn

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/provider/test"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// mockEmitter is a test double that records calls to the Emitter interface.
type mockEmitter struct {
	mu         sync.Mutex
	started    int
	succeeded  int
	failed     int
	lastCallID string
}

func (m *mockEmitter) EmitLLMCallStarted(info interfaces.LLMCallStartInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.started++
	m.lastCallID = info.CallID
}

func (m *mockEmitter) EmitLLMCallSucceeded(info interfaces.LLMCallSuccessInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if info.CallID == m.lastCallID {
		m.succeeded++
	}
}

func (m *mockEmitter) EmitLLMCallFailed(info interfaces.LLMCallFailureInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if info.CallID == m.lastCallID {
		m.failed++
	}
}

func TestLLMConn_Emitter(t *testing.T) {
	ctx := context.Background()
	model := &types.AgentModel{Name: "test-model"}
	// FIX: A valid envelope must have an Actions section to pass validation.
	env := &aeiou.Envelope{UserData: "test", Actions: "command endcommand"}

	t.Run("Success case", func(t *testing.T) {
		emitter := &mockEmitter{}
		provider := test.New()
		conn, err := New(model, provider, emitter)
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		_, err = conn.Converse(ctx, env)
		if err != nil {
			t.Fatalf("Converse() failed: %v", err)
		}

		if emitter.started != 1 {
			t.Errorf("Expected EmitLLMCallStarted to be called 1 time, but got %d", emitter.started)
		}
		if emitter.succeeded != 1 {
			t.Errorf("Expected EmitLLMCallSucceeded to be called 1 time, but got %d", emitter.succeeded)
		}
		if emitter.failed != 0 {
			t.Errorf("Expected EmitLLMCallFailed to be called 0 times, but got %d", emitter.failed)
		}
	})

	t.Run("Failure case", func(t *testing.T) {
		emitter := &mockEmitter{}
		// FIX: Use a mock provider that returns an error from its Chat method.
		// This ensures the Converse method runs and calls the emitter.
		failingProvider := &mockFailingProvider{err: errors.New("provider failed")}
		conn, err := New(model, failingProvider, emitter)
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		_, err = conn.Converse(ctx, env)
		if err == nil {
			t.Fatal("Converse() should have returned an error, but did not")
		}

		if emitter.started != 1 {
			t.Errorf("Expected EmitLLMCallStarted to be called 1 time, but got %d", emitter.started)
		}
		if emitter.succeeded != 0 {
			t.Errorf("Expected EmitLLMCallSucceeded to be called 0 times, but got %d", emitter.succeeded)
		}
		if emitter.failed != 1 {
			t.Errorf("Expected EmitLLMCallFailed to be called 1 time, but got %d", emitter.failed)
		}
	})
}

// mockFailingProvider is a simple provider that always returns an error.
type mockFailingProvider struct {
	err error
}

func (p *mockFailingProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	return nil, p.err
}
