// NeuroScript Version: 0.8.0
// File version: 9
// Purpose: FIX: Removed invalid type assertion and corrected policy modification to use the interpreter's parcel.
// filename: pkg/api/ax_user_runner_test.go
// nlines: 125
// risk_rating: HIGH

package api

import (
	"context"
	"sync"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ax"
	"github.com/aprice2704/neuroscript/pkg/ax/contract"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
)

func TestAX_UserRunner_FunctionInheritanceAndStateIsolation(t *testing.T) {
	ctx := context.Background()
	fac, _ := NewAXFactory(ctx, ax.RunnerOpts{}, &mockRuntime{}, &mockID{did: "did:test:boot"})

	// Load a library function into the root interpreter.
	libScript := `
		func get_lib_msg(returns string) means
			return "from the library"
		endfunc
	`
	libTree, _ := Parse([]byte(libScript), ParseSkipComments)
	// FIX: Remove invalid type assertion. 'fac' is already the concrete type.
	fac.root.AppendScript(libTree)

	// Run a command in a config runner to set a variable.
	cmdScript := `
		command
			set boot_var = "secret"
		endcommand
	`
	configRunner, _ := fac.NewRunner(ctx, ax.RunnerConfig, ax.RunnerOpts{})
	configRunner.LoadScript([]byte(cmdScript))
	configRunner.Execute()

	// Now, create a user runner and test its behavior.
	userRunner, err := fac.NewRunner(ctx, ax.RunnerUser, ax.RunnerOpts{})
	if err != nil {
		t.Fatalf("NewRunner(User) failed: %v", err)
	}

	// Test function inheritance.
	userFuncScript := `
		func main(returns string) means
			return get_lib_msg()
		endfunc
	`
	res, err := AXRunScript(ctx, userRunner, []byte(userFuncScript), "main")
	if err != nil {
		t.Fatalf("AXRunScript() failed: %v", err)
	}
	if s, ok := res.(string); !ok || s != "from the library" {
		t.Errorf("Inherited function returned wrong value: got %q, want 'from the library'", s)
	}

	// Test state isolation.
	interp, _ := AXInterpreter(userRunner)
	_, found := interp.GetVariable("boot_var")
	if found {
		t.Error("State (boot_var) leaked from boot runner to user runner")
	}
}

func TestAX_UserRunner_CommandExecution(t *testing.T) {
	ctx := context.Background()
	fac, _ := NewAXFactory(ctx, ax.RunnerOpts{}, &mockRuntime{}, &mockID{})
	userRunner, _ := fac.NewRunner(ctx, ax.RunnerUser, ax.RunnerOpts{})

	var emittedValue any
	var wg sync.WaitGroup
	wg.Add(1)

	interp, _ := AXInterpreter(userRunner)
	interp.SetEmitFunc(func(v lang.Value) {
		emittedValue = lang.Unwrap(v)
		wg.Done()
	})

	cmdScript := `
		command
			emit "command block was executed"
		endcommand
	`
	if err := userRunner.LoadScript([]byte(cmdScript)); err != nil {
		t.Fatalf("LoadScript() failed: %v", err)
	}
	if _, err := userRunner.Execute(); err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	wg.Wait()
	expected := "command block was executed"
	if s, ok := emittedValue.(string); !ok || s != expected {
		t.Errorf("Execute() emitted %q, want %q", s, expected)
	}
}

func TestAX_UserRunner_IdentityAndTools(t *testing.T) {
	ctx := context.Background()
	userID := &mockID{did: "did:test:user123"}
	userRT := &mockRuntime{id: userID}

	// Use the factory to create the runner, which is the correct pattern.
	fac, err := NewAXFactory(ctx, ax.RunnerOpts{}, userRT, userID)
	if err != nil {
		t.Fatalf("NewAXFactory() failed: %v", err)
	}
	userRunner, err := fac.NewRunner(ctx, ax.RunnerUser, ax.RunnerOpts{})
	if err != nil {
		t.Fatalf("NewRunner(User) failed: %v", err)
	}

	// User runners have a deny-by-default policy, so we must allow the tool for the test.
	interp, _ := AXInterpreter(userRunner)
	// FIX: Access the policy via the interpreter's parcel.
	allowPolicy := policy.NewBuilder(policy.ContextUser).Allow("tool.host.whoami").Build()
	interp.internal.SetParcel(interp.internal.GetParcel().Fork(func(m *contract.ParcelMut) {
		m.Policy = allowPolicy
	}))

	toolImpl := ToolImplementation{
		Spec: ToolSpec{Name: "whoami", Group: "host"},
		Func: decoupledTool,
	}
	if err := userRunner.Tools().Register("tool.host.whoami", toolImpl); err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	toolScript := `
		func main(returns string) means
			return tool.host.whoami()
		endfunc
	`
	res, err := AXRunScript(ctx, userRunner, []byte(toolScript), "main")
	if err != nil {
		t.Fatalf("AXRunScript() with tool failed: %v", err)
	}

	expected := "called by: did:test:user123"
	if s, ok := res.(string); !ok || s != expected {
		t.Errorf("Identity tool returned %q, want %q", s, expected)
	}
}
