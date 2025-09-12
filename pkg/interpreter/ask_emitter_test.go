// NeuroScript Version: 0.7.2
// File version: 8
// Purpose: Corrects the final test failure by using a simple script string and a non-empty prompt, which ensures a valid AEIOU envelope is created.
// filename: pkg/interpreter/ask_emitter_test.go

package interpreter

import (
	"sync"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
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

// WithEmitter is a local test helper that mirrors the api.WithEmitter option
// to avoid a circular dependency in tests.
func WithEmitter(emitter interfaces.Emitter) InterpreterOption {
	return func(i *Interpreter) {
		i.SetEmitter(emitter)
	}
}

func TestInterpreter_Ask_EmitterIntegration(t *testing.T) {
	// --- ARRANGE ---
	emitter := &mockEmitter{}
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
	interp := NewInterpreter(WithEmitter(emitter), WithExecPolicy(configPolicy))
	interp.RegisterProvider(providerName, provider)

	agentModelName := types.AgentModelName("test-agent")
	modelConfig := map[string]lang.Value{
		"provider":            lang.StringValue{Value: providerName},
		"model":               lang.StringValue{Value: "test-model"},
		"tool_loop_permitted": lang.BoolValue{Value: true}, // Allow looping for the mock provider
	}
	if err := interp.RegisterAgentModel(agentModelName, modelConfig); err != nil {
		t.Fatalf("Failed to register agent model: %v", err)
	}

	// FIX: Use a simple script string instead of manual AST construction.
	// FIX: Use a non-empty prompt to ensure UserData is valid.
	script := `command 
   ask "test-agent", "ping" 
endcommand`
	p := parser.NewParserAPI(nil)
	tree, pErr := p.Parse(script)
	if pErr != nil {
		t.Fatalf("Failed to parse test script: %v", pErr)
	}
	program, _, bErr := parser.NewASTBuilder(nil).Build(tree)
	if bErr != nil {
		t.Fatalf("Failed to build AST from parsed script: %v", bErr)
	}
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load script: %v", err)
	}

	// --- ACT ---
	_, err := interp.ExecuteCommands()
	if err != nil {
		t.Fatalf("ExecuteCommands() failed: %v", err)
	}

	// --- ASSERT ---
	if emitter.started != 1 {
		t.Errorf("Expected emitter.started to be 1, but got %d", emitter.started)
	}
	if emitter.succeeded != 1 {
		t.Errorf("Expected emitter.succeeded to be 1, but got %d", emitter.succeeded)
	}
	if emitter.failed != 0 {
		t.Errorf("Expected emitter.failed to be 0, but got %d", emitter.failed)
	}
}
