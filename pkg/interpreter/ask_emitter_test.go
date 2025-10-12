// NeuroScript Version: 0.8.0
// File version: 10
// Purpose: Corrected a syntax error in a defer statement and updated ExecPolicy assignment.
// filename: pkg/interpreter/ask_emitter_test.go

package interpreter_test

import (
	"sync"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
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

func TestInterpreter_Ask_EmitterIntegration(t *testing.T) {
	t.Logf("[DEBUG] Turn 1: Starting TestInterpreter_Ask_EmitterIntegration.")
	h := NewTestHarness(t)
	interp := h.Interpreter
	emitter := &mockEmitter{t: t}
	h.HostContext.Emitter = emitter

	provider := test.New()
	providerName := "test-provider"

	configPolicy := &policy.ExecPolicy{
		Context: policy.ContextConfig,
		Grants: capability.NewGrantSet(
			[]capability.Capability{
				{Resource: "model", Verbs: []string{"admin"}, Scopes: []string{"*"}},
			},
			capability.Limits{},
		),
	}
	interp.ExecPolicy = configPolicy
	interp.RegisterProvider(providerName, provider)
	t.Logf("[DEBUG] Turn 2: Harness configured with emitter, policy, and provider.")

	agentModelName := types.AgentModelName("test-agent")
	modelConfig := map[string]lang.Value{
		"provider":            lang.StringValue{Value: providerName},
		"model":               lang.StringValue{Value: "test-model"},
		"tool_loop_permitted": lang.BoolValue{Value: true},
	}
	if err := interp.RegisterAgentModel(agentModelName, modelConfig); err != nil {
		t.Fatalf("Failed to register agent model: %v", err)
	}
	t.Logf("[DEBUG] Turn 3: Agent model registered.")

	script := `command ask "test-agent", "ping" endcommand`
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
