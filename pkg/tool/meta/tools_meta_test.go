// NeuroScript Version: 0.8.0
// File version: 9
// Purpose: Removes test for deprecated getTool and corrects assertion for listTools.
// filename: pkg/tool/meta/tools_meta_test.go
// nlines: 88
// risk_rating: MEDIUM

package meta_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/tool/meta"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// mockRuntime is a minimal mock implementation of tool.Runtime for this test file.
type mockRuntime struct {
	registry   tool.ToolRegistry
	execPolicy *policy.ExecPolicy
}

// Statically assert that *mockRuntime satisfies the tool.Runtime interface.
var _ tool.Runtime = (*mockRuntime)(nil)

func (m *mockRuntime) ToolRegistry() tool.ToolRegistry   { return m.registry }
func (m *mockRuntime) GetExecPolicy() *policy.ExecPolicy { return m.execPolicy }
func (m *mockRuntime) GetGrantSet() *capability.GrantSet {
	if m.execPolicy != nil {
		return &m.execPolicy.Grants
	}
	return &capability.GrantSet{}
}

// --- Unused tool.Runtime methods ---
func (m *mockRuntime) Println(args ...any)                                   { fmt.Fprintln(os.Stderr, args...) } // DEBUG
func (m *mockRuntime) PromptUser(prompt string) (string, error)              { return "", nil }
func (m *mockRuntime) GetVar(name string) (any, bool)                        { return nil, false }
func (m *mockRuntime) SetVar(name string, val any)                           {}
func (m *mockRuntime) CallTool(name types.FullName, args []any) (any, error) { return nil, nil }
func (m *mockRuntime) GetLogger() interfaces.Logger                          { return nil } // Use Println for test debug
func (m *mockRuntime) SandboxDir() string                                    { return "" }
func (m *mockRuntime) LLM() interfaces.LLMClient                             { return nil }
func (m *mockRuntime) RegisterHandle(obj interface{}, typePrefix string) (string, error) {
	return "", nil
}
func (m *mockRuntime) GetHandleValue(handle, prefix string) (interface{}, error) { return nil, nil }
func (m *mockRuntime) AgentModels() interfaces.AgentModelReader                  { return nil }
func (m *mockRuntime) AgentModelsAdmin() interfaces.AgentModelAdmin              { return nil }

func newTestRuntime(t *testing.T) *mockRuntime {
	t.Helper()
	rt := &mockRuntime{}
	registry := tool.NewToolRegistry(rt)
	rt.registry = registry
	rt.execPolicy = &policy.ExecPolicy{
		Allow: []string{"tool.meta.*"}, // Allow all meta tools for testing
	}
	if err := meta.RegisterTools(registry); err != nil {
		t.Fatalf("Failed to register meta tools: %v", err)
	}
	return rt
}

func TestToolMetaListTools(t *testing.T) {
	rt := newTestRuntime(t)
	dummyTool := tool.ToolImplementation{
		Spec: tool.ToolSpec{Name: "dummy", Group: "test"},
		Func: func(rt tool.Runtime, args []any) (any, error) { return "ok", nil },
	}
	if _, err := rt.ToolRegistry().RegisterTool(dummyTool); err != nil {
		t.Fatalf("Failed to register dummy tool: %v", err)
	}

	result, err := rt.ToolRegistry().ExecuteTool("tool.meta.listTools", map[string]lang.Value{})
	if err != nil {
		t.Fatalf("ExecuteTool failed unexpectedly: %v", err)
	}

	unwrapped := lang.Unwrap(result)
	specs, ok := unwrapped.([]interface{})
	if !ok {
		t.Fatalf("Expected result to be a slice of specs, but got %T", unwrapped)
	}

	// Should find the 2 meta tools + the 1 dummy tool
	if len(specs) != 3 {
		t.Errorf("Expected to find 3 tools, but got %d", len(specs))
	}
}
