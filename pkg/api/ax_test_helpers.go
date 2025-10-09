// NeuroScript Version: 0.7.4
// File version: 1
// Purpose: Provides test helpers (mocks, tool funcs) for the ax wiring tests.
// filename: pkg/api/ax_test_helpers.go
// nlines: 48
// risk_rating: LOW

package api

import (
	"errors"

	"github.com/aprice2704/neuroscript/pkg/ax"
)

// mockID implements the ax.ID interface for testing.
type mockID struct{ did ax.DID }

func (m *mockID) DID() ax.DID { return m.did }

// mockRuntime implements the api.Runtime and ax.IdentityCap interfaces for testing.
type mockRuntime struct {
	id ax.ID
}

func (m *mockRuntime) Identity() ax.ID { return m.id }

// Stubs to satisfy the api.Runtime interface
func (m *mockRuntime) Println(...any)                                 {}
func (m *mockRuntime) PromptUser(string) (string, error)              { return "", nil }
func (m *mockRuntime) GetVar(string) (any, bool)                      { return nil, false }
func (m *mockRuntime) SetVar(string, any)                             {}
func (m *mockRuntime) CallTool(FullName, []any) (any, error)          { return nil, nil }
func (m *mockRuntime) GetLogger() Logger                              { return nil }
func (m *mockRuntime) SandboxDir() string                             { return "" }
func (m *mockRuntime) ToolRegistry() ToolRegistry                     { return nil }
func (m *mockRuntime) LLM() LLMClient                                 { return nil }
func (m *mockRuntime) RegisterHandle(any, string) (string, error)     { return "", nil }
func (m *mockRuntime) GetHandleValue(string, string) (any, error)     { return nil, nil }
func (m *mockRuntime) AgentModels() AgentModelReader                  { return nil }
func (m *mockRuntime) AgentModelsAdmin() AgentModelAdmin              { return nil }
func (m *mockRuntime) GetGrantSet() *GrantSet                         { return nil }
func (m *mockRuntime) Accounts() AccountReader                        { return nil }
func (m *mockRuntime) AccountsAdmin() AccountAdmin                    { return nil }
func (m *mockRuntime) CapsuleStore() *CapsuleRegistry                 { return nil }
func (m *mockRuntime) CapsuleRegistryForAdmin() *AdminCapsuleRegistry { return nil }

// decoupledTool asserts the ax.IdentityCap, not a concrete runtime type.
func decoupledTool(rt Runtime, _ []any) (any, error) {
	if ri, ok := rt.(ax.IdentityCap); ok && ri.Identity() != nil {
		return "called by: " + string(ri.Identity().DID()), nil
	}
	return nil, errors.New("missing identity capability in runtime")
}
