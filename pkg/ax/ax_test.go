// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Adds compile-time interface satisfaction tests for the ax package.
// filename: pkg/ax/ax_test.go
// nlines: 50
// risk_rating: LOW

package ax

import (
	"context"
	"testing"
)

// mockRunner is a dummy struct simulating a concrete implementation of the Runner.
type mockRunner struct{}

func (m *mockRunner) LoadScript(script []byte) error            { return nil }
func (m *mockRunner) Execute() (any, error)                     { return nil, nil }
func (m *mockRunner) Run(proc string, args ...any) (any, error) { return nil, nil }
func (m *mockRunner) EmitEvent(name, src string, payload any)   {}
func (m *mockRunner) Identity() ID                              { return nil }
func (m *mockRunner) CopyFunctionsFrom(src RunnerCore) error    { return nil }
func (m *mockRunner) Tools() Tools                              { return nil }
func (m *mockRunner) Env() RunEnv                               { return nil }
func (m *mockRunner) Clone() Runner                             { return m }

// mockFactory simulates a concrete implementation of the RunnerFactory.
type mockFactory struct{}

func (m *mockFactory) Env() RunEnv { return nil }
func (m *mockFactory) NewRunner(ctx context.Context, mode RunnerMode, opts RunnerOpts) (Runner, error) {
	return nil, nil
}

// mockRunEnv simulates a concrete implementation of the RunEnv.
type mockRunEnv struct{}

func (m *mockRunEnv) AccountsReader() AccountsReader       { return nil }
func (m *mockRunEnv) AccountsAdmin() AccountsAdmin         { return nil }
func (m *mockRunEnv) AgentModelsReader() AgentModelsReader { return nil }
func (m *mockRunEnv) AgentModelsAdmin() AgentModelsAdmin   { return nil }
func (m *mockRunEnv) CapsulesAdmin() CapsulesAdmin         { return nil }
func (m *mockRunEnv) Tools() Tools                         { return nil }

// TestInterfaceSatisfaction is a compile-time test to ensure that any potential
// concrete types correctly satisfy the interfaces defined in the ax package.
// This test has no runtime assertions; its purpose is to fail compilation if
// an interface contract is broken.
func TestInterfaceSatisfaction(t *testing.T) {
	// These lines will not compile if the mock structs do not
	// correctly implement the interfaces.
	var _ Runner = (*mockRunner)(nil)
	var _ CloneCap = (*mockRunner)(nil)
	var _ RunnerFactory = (*mockFactory)(nil)
	var _ EnvCap = (*mockFactory)(nil)
	var _ RunEnv = (*mockRunEnv)(nil)
}
