// NeuroScript Version: 0.7.4
// File version: 7
// Purpose: FIX: Rewrote IdentityAndTools test to use the factory for runner creation and correctly set the tool policy.
// filename: pkg/api/ax_user_runner_test.go
// nlines: 125
// risk_rating: HIGH

package api

import (
	"context"
	"sync"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ax"
	"github.com/aprice2704/neuroscript/pkg/lang"
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

	// FIX: Use the factory to create the runner, which is the correct pattern.
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
	interp.internal.ExecPolicy.Allow = append(interp.internal.ExecPolicy.Allow, "tool.host.whoami")

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
