// NeuroScript Version: 0.8.0
// File version: 9
// Purpose: Corrected final test assertions to align with the centralized policy logic.
// filename: pkg/interpreter/integration_policy_test.go
// nlines: 120
// risk_rating: HIGH

package interpreter

import (
	"errors"
	"os"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/policy"
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all" // Ensure tools are registered
)

func runPolicyIntegrationTest(t *testing.T, p *policy.ExecPolicy, script string) (*Interpreter, error) {
	t.Helper()
	hostCtx := &HostContext{
		Logger: logging.NewTestLogger(t),
		Stdout: os.Stdout,
		Stdin:  os.Stdin,
		Stderr: os.Stderr,
	}
	interp := NewInterpreter(WithHostContext(hostCtx), WithExecPolicy(p))
	interp.RegisterProvider("p", nil)
	fullScript := "func main() means\n" + script + "\nendfunc"
	_, rErr := interp.ExecuteScriptString("main", fullScript, nil)
	if rErr == nil {
		return interp, nil
	}
	return interp, rErr
}

func TestPolicyGate_Integration(t *testing.T) {
	t.Run("Failure: tool.agentmodel.register is trusted and requires config context", func(t *testing.T) {
		p := &policy.ExecPolicy{Context: policy.ContextNormal, Allow: []string{"*"}}
		script := `must tool.agentmodel.Register("test", {"provider":"p", "model":"m"})`
		_, err := runPolicyIntegrationTest(t, p, script)

		if !errors.Is(err, policy.ErrTrust) {
			t.Errorf("Expected an error wrapping ErrTrust due to context violation, but got: %v", err)
		}
	})

	t.Run("Success: tool.agentmodel.register with correct policy", func(t *testing.T) {
		p := &policy.ExecPolicy{
			Context: policy.ContextConfig,
			Allow:   []string{"*"},
			Grants: capability.NewGrantSet(
				[]capability.Capability{{Resource: "model", Verbs: []string{"admin"}, Scopes: []string{"*"}}},
				capability.Limits{},
			),
		}
		script := `must tool.agentmodel.Register("test", {"provider":"p", "model":"m"})`
		interp, err := runPolicyIntegrationTest(t, p, script)
		if err != nil {
			t.Fatalf("Expected script to succeed, but it failed: %v", err)
		}
		_, exists := interp.AgentModels().Get("test")
		if !exists {
			t.Error("AgentModel was not registered, even though the call should have been permitted.")
		}
	})

	t.Run("Failure: tool.os.getenv without capability", func(t *testing.T) {
		t.Setenv("MY_SECRET", "12345")
		p := &policy.ExecPolicy{
			Context: policy.ContextConfig,
			Allow:   []string{"*"},
			Grants:  capability.NewGrantSet(nil, capability.Limits{}),
		}
		script := `set secret = tool.os.getenv("MY_SECRET")`
		_, err := runPolicyIntegrationTest(t, p, script)

		if !errors.Is(err, policy.ErrCapability) {
			t.Errorf("Expected a RuntimeError wrapping ErrCapability, but got: %v", err)
		}
	})

	t.Run("Success: tool.os.getenv with capability", func(t *testing.T) {
		t.Setenv("MY_SECRET", "12345")
		p := &policy.ExecPolicy{
			Context: policy.ContextConfig,
			Allow:   []string{"*"},
			Grants: capability.NewGrantSet(
				[]capability.Capability{{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"MY_SECRET"}}},
				capability.Limits{},
			),
		}
		script := `set secret = tool.os.getenv("MY_SECRET")`
		interp, err := runPolicyIntegrationTest(t, p, script)
		if err != nil {
			t.Fatalf("Expected script to succeed, but it failed: %v", err)
		}
		secret, _ := interp.GetVariable("secret")
		if s, _ := lang.ToString(secret); s != "12345" {
			t.Errorf("Expected secret to be '12345', got '%s'", s)
		}
	})
}
