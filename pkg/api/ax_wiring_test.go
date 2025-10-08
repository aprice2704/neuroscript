// NeuroScript Version: 0.7.4
// File version: 11
// Purpose: Updates the ax wiring test to use the new public helper functions (AXBootLoad, AXRunScript), removing internal type casts.
// filename: pkg/api/ax_wiring_test.go
// nlines: 155
// risk_rating: HIGH

package api

import (
	"context"
	"errors"
	"testing"

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

// TestAXWiring_FullLifecycle covers the checklist from impl_wiring.md using the public helpers.
func TestAXWiring_FullLifecycle(t *testing.T) {
	ctx := context.Background()
	bootID := &mockID{did: "did:test:boot"}
	baseRT := &mockRuntime{id: bootID}

	fac, err := NewAXFactory(ctx, ax.RunnerOpts{SandboxDir: "/tmp/ax-test"}, baseRT, bootID)
	if err != nil {
		t.Fatalf("NewAXFactory() failed: %v", err)
	}

	// 1. Bootloading and Environment Configuration
	bootScript := `
        func get_lib_msg() returns string means
            return "from the library"
        endfunc
        command
            # This variable should not leak to user runners
            set boot_var = "secret"
        endcommand
    `
	if err := AXBootLoad(ctx, fac, []byte(bootScript)); err != nil {
		t.Fatalf("AXBootLoad() failed: %v", err)
	}

	// 2. User Runner Execution
	t.Run("FunctionInheritanceAndStateIsolation", func(t *testing.T) {
		userRunner, err := fac.NewRunner(ctx, ax.RunnerUser, ax.RunnerOpts{})
		if err != nil {
			t.Fatalf("NewRunner(User) failed: %v", err)
		}

		userScript := `
            func main() returns string means
                return get_lib_msg()
            endfunc
        `
		res, err := AXRunScript(ctx, userRunner, []byte(userScript), "main")
		if err != nil {
			t.Fatalf("AXRunScript() failed: %v", err)
		}

		if s, ok := res.(string); !ok || s != "from the library" {
			t.Errorf("Inherited function returned wrong value: got %q, want 'from the library'", s)
		}

		// Verify state isolation by checking for the boot variable.
		interp, ok := AXInterpreter(userRunner)
		if !ok {
			t.Fatal("Could not get internal interpreter for test verification")
		}
		_, found := interp.GetVariable("boot_var")
		if found {
			t.Error("State (boot_var) leaked from boot runner to user runner")
		}
	})

	// 3. Identity and Tool Decoupling
	t.Run("IdentityCapabilityInTools", func(t *testing.T) {
		userID := &mockID{did: "did:test:user123"}
		userRT := &mockRuntime{id: userID}

		// In a real app, the factory might be recreated or reconfigured for a user.
		// For this test, we create a new one to inject the user's runtime.
		userFac, _ := NewAXFactory(ctx, ax.RunnerOpts{}, userRT, userID)

		userRunner, _ := userFac.NewRunner(ctx, ax.RunnerUser, ax.RunnerOpts{})
		decoupledToolImpl := ToolImplementation{
			Spec: ToolSpec{Name: "whoami", Group: "host"},
			Func: decoupledTool,
		}
		if err := userRunner.Tools().Register("tool.host.whoami", decoupledToolImpl); err != nil {
			t.Fatalf("Failed to register tool: %v", err)
		}

		toolScript := `
            func main() returns string means
                return tool.host.whoami()
            endfunc
        `
		res, err := AXRunScript(ctx, userRunner, []byte(toolScript), "main")
		if err != nil {
			t.Fatalf("AXRunScript() with identity-aware tool failed: %v", err)
		}

		expected := "called by: did:test:user123"
		if s, ok := res.(string); !ok || s != expected {
			t.Errorf("Identity tool returned wrong value: got %q, want %q", s, expected)
		}
	})
}
