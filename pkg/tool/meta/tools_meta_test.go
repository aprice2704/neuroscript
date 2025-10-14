// NeuroScript Version: 0.8.0
// File version: 5
// Purpose: Corrected package name and function calls to resolve build errors.
// filename: pkg/tool/meta/tools_meta_test.go
// nlines: 104
// risk_rating: LOW

package meta_test

import (
	"errors"
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
func (m *mockRuntime) Println(...any)                                        {}
func (m *mockRuntime) PromptUser(prompt string) (string, error)              { return "", nil }
func (m *mockRuntime) GetVar(name string) (any, bool)                        { return nil, false }
func (m *mockRuntime) SetVar(name string, val any)                           {}
func (m *mockRuntime) CallTool(name types.FullName, args []any) (any, error) { return nil, nil }
func (m *mockRuntime) GetLogger() interfaces.Logger                          { return nil }
func (m *mockRuntime) SandboxDir() string                                    { return "" }
func (m *mockRuntime) LLM() interfaces.LLMClient                             { return nil }
func (m *mockRuntime) RegisterHandle(obj interface{}, typePrefix string) (string, error) {
	return "", nil
}
func (m *mockRuntime) GetHandleValue(handle, prefix string) (interface{}, error) { return nil, nil }
func (m *mockRuntime) AgentModels() interfaces.AgentModelReader                  { return nil }
func (m *mockRuntime) AgentModelsAdmin() interfaces.AgentModelAdmin              { return nil }

func newTestRuntime(t *testing.T) tool.Runtime {
	t.Helper()
	rt := &mockRuntime{}
	registry := tool.NewToolRegistry(rt)
	rt.registry = registry
	if err := meta.RegisterTools(registry); err != nil {
		t.Fatalf("Failed to register meta tools: %v", err)
	}
	return rt
}

func TestToolMetaGetTool(t *testing.T) {
	rt := newTestRuntime(t)

	dummyTool := tool.ToolImplementation{
		Spec: tool.ToolSpec{Name: "dummy", Group: "test"},
		Func: func(rt tool.Runtime, args []any) (any, error) { return "ok", nil },
	}
	if _, err := rt.ToolRegistry().RegisterTool(dummyTool); err != nil {
		t.Fatalf("Failed to register dummy tool: %v", err)
	}

	testCases := []struct {
		name          string
		args          []any
		wantFound     bool
		wantFullName  string
		wantToolErrIs error
	}{
		{"Success - Find existing tool", []any{"tool.test.dummy"}, true, "tool.test.dummy", nil},
		{"Failure - Tool not found", []any{"tool.nonexistent.tool"}, false, "", nil},
		{"Failure - Invalid argument type", []any{123}, false, "", lang.ErrInvalidArgument},
		{"Failure - Too few arguments", []any{}, false, "", lang.ErrArgumentMismatch},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := meta.GetTool(rt, tc.args)

			if tc.wantToolErrIs != nil {
				if !errors.Is(err, tc.wantToolErrIs) {
					t.Errorf("Expected error wrapping [%v], but got: %v", tc.wantToolErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				t.Fatalf("Expected result to be a map, got %T", result)
			}

			if found, _ := resultMap["found"].(bool); found != tc.wantFound {
				t.Errorf("Mismatched 'found' status: got %v, want %v", found, tc.wantFound)
			}

			if tc.wantFound {
				spec, _ := resultMap["spec"].(tool.ToolSpec)
				if types.FullName(spec.FullName) != types.FullName(tc.wantFullName) {
					t.Errorf("Mismatched tool name: got %s, want %s", spec.FullName, tc.wantFullName)
				}
			}
		})
	}
}
