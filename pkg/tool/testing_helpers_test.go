// NeuroScript Version: 0.8.0
// File version: 4
// Purpose: Fixes all compiler errors by defining MockRuntime, correcting types, renaming a shadowing variable, and fixing the tool function call. Implemented HandleRegistry for tool.Runtime interface compliance.
// filename: pkg/tool/testing_helpers_test.go
// nlines: 85
// risk_rating: LOW

package tool_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/testutil"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// MockRuntime is a mock implementation of tool.Runtime for testing purposes.
type MockRuntime struct {
	grantSet *capability.GrantSet
	policy   *policy.ExecPolicy
}

func (m *MockRuntime) GetGrantSet() *capability.GrantSet           { return m.grantSet }
func (m *MockRuntime) GetExecPolicy() *policy.ExecPolicy           { return m.policy }
func (m *MockRuntime) Println(...any)                              {}
func (m *MockRuntime) PromptUser(string) (string, error)           { return "", nil }
func (m *MockRuntime) GetVar(string) (any, bool)                   { return nil, false }
func (m *MockRuntime) SetVar(string, any)                          {}
func (m *MockRuntime) CallTool(types.FullName, []any) (any, error) { return nil, nil }
func (m *MockRuntime) GetLogger() interfaces.Logger                { return nil }
func (m *MockRuntime) SandboxDir() string                          { return "" }
func (m *MockRuntime) ToolRegistry() tool.ToolRegistry             { return nil }
func (m *MockRuntime) LLM() interfaces.LLMClient                   { return nil }

// HandleRegistry is a new required method on the tool.Runtime interface.
func (m *MockRuntime) HandleRegistry() interfaces.HandleRegistry { return nil }

// NOTE: Old handle methods (RegisterHandle, GetHandleValue) have been removed from the interface.

func (m *MockRuntime) AgentModels() interfaces.AgentModelReader     { return nil }
func (m *MockRuntime) AgentModelsAdmin() interfaces.AgentModelAdmin { return nil }

// toolExecTestCase defines a table-driven test case for executing a tool.
type toolExecTestCase struct {
	name      string
	toolName  string
	args      []interface{}
	mock      *MockRuntime
	want      lang.Value
	wantErr   bool
	errEquals error
}

// toolExec is a helper to execute a single tool for testing.
func toolExec(t *testing.T, tc toolExecTestCase) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		interp, err := testutil.NewTestInterpreter(t, nil)
		if err != nil {
			t.Fatalf("failed to create test interpreter: %v", err)
		}

		registry := interp.ToolRegistry()
		toolImpl, found := registry.GetTool(types.FullName(tc.toolName))
		if !found {
			t.Fatalf("tool %q not found in registry", tc.toolName)
		}

		var rt tool.Runtime = interp
		if tc.mock != nil {
			rt = tc.mock
		}

		// Tool functions expect a slice of raw interfaces, not wrapped lang.Value types.
		// The original arguments `tc.args` are already in the correct format.
		_, gotErr := toolImpl.Func(rt, tc.args)

		if (gotErr != nil) != tc.wantErr {
			t.Fatalf("tool.Func() error = %v, wantErr %v", gotErr, tc.wantErr)
		}
		if tc.wantErr && tc.errEquals != nil {
			if gotErr != tc.errEquals {
				t.Fatalf("tool.Func() error = %v, want %v", gotErr, tc.errEquals)
			}
		}
	})
}
