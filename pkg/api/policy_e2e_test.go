// NeuroScript Version: 0.8.0
// File version: 9
// Purpose: Corrects all tests to provide a mandatory HostContext during interpreter creation, resolving a panic.
// filename: pkg/api/policy_e2e_test.go
// nlines: 135
// risk_rating: HIGH

package api_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
)

// TestE2E_LimitEnforcement_ToolCalls verifies that the interpreter stops execution
// when a per-tool call limit is exceeded.
func TestE2E_LimitEnforcement_ToolCalls(t *testing.T) {
	counterTool := api.ToolImplementation{
		Spec: api.ToolSpec{Name: "count", Group: "test"},
		Func: func(rt api.Runtime, args []any) (any, error) { return "counted", nil },
	}
	limitScript := `
func main() means
    call tool.test.count()
    call tool.test.count()
    # This third call should fail.
    call tool.test.count()
endfunc
`
	// Manually construct the policy to include limits, as the high-level
	// helpers do not support them. This is the correct pattern for this test.
	policy := &policy.ExecPolicy{
		Context: policy.ContextNormal,
		Allow:   []string{"tool.test.count"},
		Deny:    []string{},
		Grants: capability.GrantSet{
			Limits: capability.Limits{
				ToolMaxCalls: map[string]int{"tool.test.count": 2},
			},
			Counters: capability.NewCounters(),
		},
	}

	// FIX: Provide a mandatory HostContext
	hc := newTestHostContext(nil)
	interp := api.New(
		api.WithHostContext(hc),
		interpreter.WithExecPolicy(policy),
	)
	if _, err := interp.ToolRegistry().RegisterTool(counterTool); err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	tree, err := api.Parse([]byte(limitScript), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse() failed: %v", err)
	}
	if err := api.LoadFromUnit(interp, &api.LoadedUnit{Tree: tree}); err != nil {
		t.Fatalf("api.LoadFromUnit() failed: %v", err)
	}

	_, err = api.RunProcedure(context.Background(), interp, "main")

	if err == nil {
		t.Fatal("Expected the script to fail due to exceeding tool call limits, but it succeeded.")
	}
	var rtErr *lang.RuntimeError
	if errors.As(err, &rtErr) {
		if rtErr.Code != lang.ErrorCodePolicy {
			t.Errorf("Expected error code %v (ErrorCodePolicy), but got %v", lang.ErrorCodePolicy, rtErr.Code)
		}
	} else {
		t.Errorf("Expected a *lang.RuntimeError, but got a different error type: %v", err)
	}
}

// TestE2E_RequiresTrust_ContextEnforcement verifies that a tool marked with
// 'RequiresTrust = true' can only run in a 'config' context.
func TestE2E_RequiresTrust_ContextEnforcement(t *testing.T) {
	privilegedTool := api.ToolImplementation{
		Spec:          api.ToolSpec{Name: "setConfig", Group: "sys"},
		Func:          func(rt api.Runtime, args []any) (any, error) { return "config set", nil },
		RequiresTrust: true,
	}
	script := `
func main() means
    call tool.sys.setConfig()
endfunc
`
	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse() failed: %v", err)
	}

	t.Run("Success in config context", func(t *testing.T) {
		// FIX: Tool names in policies must be lowercase.
		interp := api.NewConfigInterpreter(
			[]string{"tool.sys.setconfig"},
			[]api.Capability{},
		)
		if _, err := interp.ToolRegistry().RegisterTool(privilegedTool); err != nil {
			t.Fatalf("Failed to register tool: %v", err)
		}
		if err := api.LoadFromUnit(interp, &api.LoadedUnit{Tree: tree}); err != nil {
			t.Fatalf("api.LoadFromUnit() failed: %v", err)
		}

		_, err := api.RunProcedure(context.Background(), interp, "main")
		if err != nil {
			t.Fatalf("Expected privileged tool to succeed in a config context, but it failed: %v", err)
		}
	})

	t.Run("Failure in normal context", func(t *testing.T) {
		// FIX: Provide a mandatory HostContext
		hc := newTestHostContext(nil)
		// A default interpreter is unprivileged.
		interp := api.New(api.WithHostContext(hc))
		if _, err := interp.ToolRegistry().RegisterTool(privilegedTool); err != nil {
			t.Fatalf("Failed to register tool: %v", err)
		}
		if err := api.LoadFromUnit(interp, &api.LoadedUnit{Tree: tree}); err != nil {
			t.Fatalf("api.LoadFromUnit() failed: %v", err)
		}

		_, err := api.RunProcedure(context.Background(), interp, "main")
		if err == nil {
			t.Fatal("Expected privileged tool to fail in a normal context, but it succeeded.")
		}
		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) || rtErr.Code != lang.ErrorCodePolicy {
			t.Errorf("Expected a policy error, but got: %v", err)
		}
	})
}

// TestE2E_DefaultInterpreterIsUnprivileged confirms that a standard interpreter
// created via api.New() is in a 'normal' context and will reject trusted tools.
func TestE2E_DefaultInterpreterIsUnprivileged(t *testing.T) {
	privilegedTool := api.ToolImplementation{
		Spec:          api.ToolSpec{Name: "setConfig", Group: "sys"},
		Func:          func(rt api.Runtime, args []any) (any, error) { return "config set", nil },
		RequiresTrust: true,
	}
	script := `
func main() means
    # This tool requires a trusted context, which the default interpreter should not have.
    call tool.sys.setConfig()
endfunc
`
	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse() failed: %v", err)
	}

	// FIX: Provide a mandatory HostContext
	hc := newTestHostContext(nil)
	// Create a standard, default-configured interpreter. It must be unprivileged.
	interp := api.New(api.WithHostContext(hc))
	if _, err := interp.ToolRegistry().RegisterTool(privilegedTool); err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	if err := api.LoadFromUnit(interp, &api.LoadedUnit{Tree: tree}); err != nil {
		t.Fatalf("api.LoadFromUnit() failed: %v", err)
	}

	_, err = api.RunProcedure(context.Background(), interp, "main")

	// Assert that the call failed with a policy error.
	if err == nil {
		t.Fatal("Expected privileged tool to fail in a default interpreter, but it succeeded.")
	}
	var rtErr *lang.RuntimeError
	if errors.As(err, &rtErr) {
		if rtErr.Code != lang.ErrorCodePolicy {
			t.Errorf("Expected error code %v (ErrorCodePolicy), but got %v", lang.ErrorCodePolicy, rtErr.Code)
		}
	} else {
		t.Errorf("Expected a *lang.RuntimeError, but got a different error type: %v", err)
	}
}
