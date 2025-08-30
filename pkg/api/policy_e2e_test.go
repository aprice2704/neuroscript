// NeuroScript Version: 0.6.0
// File version: 4
// Purpose: Added a test to explicitly verify that a default interpreter is unprivileged.
// filename: pkg/api/policy_e2e_test.go
// nlines: 160
// risk_rating: HIGH

package api_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
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
	interp := api.New(
		api.WithTool(counterTool),
		interpreter.WithExecPolicy(policy),
	)
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
		interp := api.NewConfigInterpreter(
			[]string{"tool.sys.setConfig"},
			[]capability.Capability{},
			api.WithTool(privilegedTool),
		)
		if err := api.LoadFromUnit(interp, &api.LoadedUnit{Tree: tree}); err != nil {
			t.Fatalf("api.LoadFromUnit() failed: %v", err)
		}

		_, err := api.RunProcedure(context.Background(), interp, "main")

		if err != nil {
			t.Fatalf("Expected privileged tool to succeed in a config context, but it failed: %v", err)
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

	// Create a standard, default-configured interpreter.
	// It must be unprivileged.
	interp := api.New(api.WithTool(privilegedTool))

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
