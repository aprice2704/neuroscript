// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Fixes test failure by adding Allow("*") to the config interpreter policy.
// filename: pkg/interpreter/interpreter_test_helpers.go
// nlines: 135

package interpreter

import (
	"io"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/policy"
)

// --- Mock Symbol Provider ---

// mockSymbolProvider implements the interfaces.SymbolProvider to mimic a host (FDM).
// It holds the symbols extracted from the "Config" interpreter.
type mockSymbolProvider struct {
	procs    map[string]*ast.Procedure
	handlers map[string][]*ast.OnEventDecl
	consts   map[string]lang.Value
}

func newMockSymbolProvider(p map[string]*ast.Procedure, h map[string][]*ast.OnEventDecl, c map[string]lang.Value) *mockSymbolProvider {
	return &mockSymbolProvider{procs: p, handlers: h, consts: c}
}

func (m *mockSymbolProvider) GetProcedure(name string) (any, bool) {
	p, ok := m.procs[name]
	return p, ok
}

func (m *mockSymbolProvider) ListProcedures() map[string]any {
	converted := make(map[string]any)
	for k, v := range m.procs {
		converted[k] = v
	}
	return converted
}

func (m *mockSymbolProvider) GetEventHandlers(eventName string) ([]any, bool) {
	h, ok := m.handlers[eventName]
	if !ok {
		return nil, false
	}
	anys := make([]any, len(h))
	for i, v := range h {
		anys[i] = v
	}
	return anys, true
}

func (m *mockSymbolProvider) ListEventHandlers() map[string][]any {
	converted := make(map[string][]any)
	for name, handlers := range m.handlers {
		anys := make([]any, len(handlers))
		for i, h := range handlers {
			anys[i] = h
		}
		converted[name] = anys
	}
	return converted
}

func (m *mockSymbolProvider) GetGlobalConstant(name string) (any, bool) {
	c, ok := m.consts[name]
	return c, ok
}

func (m *mockSymbolProvider) ListGlobalConstants() map[string]any {
	converted := make(map[string]any)
	for k, v := range m.consts {
		converted[k] = v
	}
	return converted
}

// --- Test Helpers ---

func newTestHostContext(t *testing.T, provider interfaces.SymbolProvider) *HostContext {
	reg := make(map[string]any)
	if provider != nil {
		reg[interfaces.SymbolProviderKey] = provider
	}

	hc, err := NewHostContextBuilder().
		WithLogger(logging.NewTestLogger(t)). //
		WithStdout(io.Discard).
		WithStdin(strings.NewReader("")).
		WithStderr(io.Discard).
		WithServiceRegistry(reg).
		Build()
	if err != nil {
		t.Fatalf("Failed to build HostContext: %v", err)
	}
	return hc
}

// newConfigInterpreter creates a privileged interpreter for the "Config Context".
func newConfigInterpreter(t *testing.T) *Interpreter {
	hc := newTestHostContext(t, nil) // No provider
	// FIX: Explicitly allow all tools for the config context.
	// This is required by the new deny-by-default policy builder.
	cfgPolicy := policy.NewBuilder(policy.ContextConfig).Allow("*").Build()
	i := NewInterpreter(
		WithHostContext(hc),
		WithExecPolicy(cfgPolicy),
	)
	return i
}

// newRuntimeInterpreter creates a sandboxed interpreter for the "Runtime Context".
func newRuntimeInterpreter(t *testing.T, provider interfaces.SymbolProvider) *Interpreter {
	hc := newTestHostContext(t, provider) // Inject the provider
	// This policy remains "deny-by-default"
	rtPolicy := policy.NewBuilder(policy.ContextNormal).Build()
	i := NewInterpreter(
		WithHostContext(hc),
		WithExecPolicy(rtPolicy),
	)
	return i
}

// mustLoadString parses and loads a script, failing the test on error.
func mustLoadString(t *testing.T, i *Interpreter, script string) {
	t.Helper()
	tree, pErr := i.Parser().Parse(script)
	if pErr != nil {
		t.Fatalf("Failed to parse script: %v", pErr)
	}
	program, _, bErr := i.ASTBuilder().Build(tree)
	if bErr != nil {
		t.Fatalf("Failed to build AST: %v", bErr)
	}
	if err := i.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load AST: %v", err)
	}
}
