// NeuroScript Version: 0.8.0
// File version: 16
// Purpose: Corrected call to provider.NewAdmin to include the ExecPolicy.
// filename: pkg/interpreter/ask_emitter_test.go
// nlines: 115
// risk_rating: LOW

package interpreter_test

import (
	"context"
	"sync"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// mockEmitter is a test double that records calls to the Emitter interface.
type mockEmitter struct {
	mu         sync.Mutex
	started    int
	succeeded  int
	failed     int
	lastCallID string
	t          *testing.T
}

func (m *mockEmitter) EmitLLMCallStarted(info interfaces.LLMCallStartInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.t.Logf("[DEBUG] Turn X: mockEmitter.EmitLLMCallStarted called.")
	m.started++
	m.lastCallID = info.CallID
}

func (m *mockEmitter) EmitLLMCallSucceeded(info interfaces.LLMCallSuccessInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.t.Logf("[DEBUG] Turn X: mockEmitter.EmitLLMCallSucceeded called.")
	if info.CallID == m.lastCallID {
		m.succeeded++
	}
}

func (m *mockEmitter) EmitLLMCallFailed(info interfaces.LLMCallFailureInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.t.Logf("[DEBUG] Turn X: mockEmitter.EmitLLMCallFailed called.")
	if info.CallID == m.lastCallID {
		m.failed++
	}
}

// simplePingProvider is a mock provider that just emits "pong".
type simplePingProvider struct{}

func (m *simplePingProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	// THE FIX: Added required newlines to the 'command' block.
	actions := `
command
    emit "pong"
endcommand`
	env := &aeiou.Envelope{UserData: "{}", Actions: actions}
	respText, _ := env.Compose()
	return &provider.AIResponse{TextContent: respText}, nil
}

func TestInterpreter_Ask_EmitterIntegration(t *testing.T) {
	t.Logf("[DEBUG] Turn 1: Starting TestInterpreter_Ask_EmitterIntegration.")
	h := NewTestHarness(t)
	interp := h.Interpreter
	emitter := &mockEmitter{t: t}
	h.HostContext.Emitter = emitter

	// THE FIX: Use our new simple mock provider instead of the obsolete test.New().
	providerInstance := &simplePingProvider{} // Renamed from 'provider' to avoid shadowing
	providerName := "test-provider"

	// --- FIX: Register provider via the harness's registry ---
	if err := provider.NewAdmin(h.ProviderRegistry, h.Interpreter.GetExecPolicy()).Register(providerName, providerInstance); err != nil {
		t.Fatalf("Failed to register mock provider: %v", err)
	}
	// --- End Fix ---
	t.Logf("[DEBUG] Turn 2: Harness configured with emitter and provider.")

	agentModelName := "test-agent" // FIX: Use string
	modelConfig := map[string]lang.Value{
		"provider":            lang.StringValue{Value: providerName},
		"model":               lang.StringValue{Value: "test-model"},
		"tool_loop_permitted": lang.BoolValue{Value: true},
	}
	// FIX: Pass string 'agentModelName'
	if err := interp.RegisterAgentModel(agentModelName, modelConfig); err != nil {
		t.Fatalf("Failed to register agent model: %v", err)
	}
	t.Logf("[DEBUG] Turn 3: Agent model registered.")

	script := `command
	ask "test-agent", "ping" into reply
	endcommand`
	tree, pErr := h.Parser.Parse(script)
	if pErr != nil {
		t.Fatalf("Failed to parse test script: %v", pErr)
	}
	program, _, bErr := h.ASTBuilder.Build(tree)
	if bErr != nil {
		t.Fatalf("Failed to build AST from parsed script: %v", bErr)
	}
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load script: %v", err)
	}
	t.Logf("[DEBUG] Turn 4: Script loaded.")

	_, err := interp.Execute(program)
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}
	t.Logf("[DEBUG] Turn 5: Script executed.")

	if emitter.started != 1 {
		t.Errorf("Expected emitter.started to be 1, but got %d", emitter.started)
	}
	if emitter.succeeded != 1 {
		t.Errorf("Expected emitter.succeeded to be 1, but got %d", emitter.succeeded)
	}
	if emitter.failed != 0 {
		t.Errorf("Expected emitter.failed to be 0, but got %d", emitter.failed)
	}
	t.Logf("[DEBUG] Turn 6: Assertions passed.")
}
