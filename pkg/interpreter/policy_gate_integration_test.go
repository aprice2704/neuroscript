// NeuroScript Version: 0.6.0
// File version: 4
// Purpose: Corrected test helper to properly return a nil error interface on success, fixing test assertion failures.
// filename: pkg/interpreter/policy_gate_integration_test.go
// nlines: 120
// risk_rating: HIGH

package interpreter

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
	"github.com/aprice2704/neuroscript/pkg/runtime"
)

func runPolicyIntegrationTest(t *testing.T, policy *runtime.ExecPolicy, script string) (*Interpreter, error) {
	t.Helper()
	interp := NewInterpreter(WithExecPolicy(policy))
	_ = interp.SetInitialVariable("dummy_var", lang.StringValue{Value: "dummy"})
	fullScript := "func main() means\n" + script + "\nendfunc"
	_, rErr := interp.ExecuteScriptString("main", fullScript, nil)

	// If the returned runtime error is nil, we must return a nil error
	// interface to satisfy the `if err != nil` checks in the tests.
	if rErr == nil {
		return interp, nil
	}
	return interp, rErr
}

func TestPolicyGate_Integration(t *testing.T) {
	t.Run("Failure: tool.agentmodel.register is trusted", func(t *testing.T) {
		policy := &runtime.ExecPolicy{
			Context: runtime.ContextNormal,
			Allow:   []string{"*"},
			Grants: capability.NewGrantSet(
				[]capability.Capability{{Resource: "model", Verbs: []string{"admin"}, Scopes: []string{"*"}}},
				capability.Limits{},
			),
		}
		script := `must tool.agentmodel.Register("test", {"provider":"p", "model":"m"})`
		_, err := runPolicyIntegrationTest(t, policy, script)

		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) || !errors.Is(rtErr.Unwrap(), runtime.ErrTrust) {
			t.Errorf("Expected a RuntimeError wrapping ErrTrust, but got: %v", err)
		}
	})

	t.Run("Success: tool.agentmodel.register with correct policy", func(t *testing.T) {
		policy := &runtime.ExecPolicy{
			Context: runtime.ContextConfig,
			Allow:   []string{"tool.agentmodel.*"},
			Grants: capability.NewGrantSet(
				[]capability.Capability{{Resource: "model", Verbs: []string{"admin"}, Scopes: []string{"*"}}},
				capability.Limits{},
			),
		}
		script := `must tool.agentmodel.Register("test", {"provider":"p", "model":"m"})`
		interp, err := runPolicyIntegrationTest(t, policy, script)
		if err != nil {
			t.Fatalf("Expected script to succeed, but it failed: %v", err)
		}
		_, exists := interp.GetAgentModel("test")
		if !exists {
			t.Error("AgentModel was not registered, even though the call should have been permitted.")
		}
	})

	t.Run("Failure: tool.os.getenv without capability", func(t *testing.T) {
		t.Setenv("MY_SECRET", "12345")
		policy := &runtime.ExecPolicy{
			Context: runtime.ContextConfig,
			Allow:   []string{"tool.os.getenv"},
			Grants:  capability.NewGrantSet(nil, capability.Limits{}),
		}
		script := `set secret = tool.os.getenv("MY_SECRET")`
		_, err := runPolicyIntegrationTest(t, policy, script)
		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) || !errors.Is(rtErr.Unwrap(), runtime.ErrCapability) {
			t.Errorf("Expected a RuntimeError wrapping ErrCapability, but got: %v", err)
		}
	})

	t.Run("Success: tool.os.getenv with capability", func(t *testing.T) {
		t.Setenv("MY_SECRET", "12345")
		policy := &runtime.ExecPolicy{
			Context: runtime.ContextConfig,
			Allow:   []string{"tool.os.getenv"},
			Grants: capability.NewGrantSet(
				[]capability.Capability{{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"my_secret"}}},
				capability.Limits{},
			),
		}
		script := `set secret = tool.os.getenv("MY_SECRET")`
		interp, err := runPolicyIntegrationTest(t, policy, script)
		if err != nil {
			t.Fatalf("Expected script to succeed, but it failed: %v", err)
		}
		secret, _ := interp.GetVariable("secret")
		if s, _ := lang.ToString(secret); s != "12345" {
			t.Errorf("Expected secret to be '12345', got '%s'", s)
		}
	})
}
