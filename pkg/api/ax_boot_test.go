// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: FIX: Removed invalid type assertion on the concrete factory type, resolving the build error.
// filename: pkg/api/ax_boot_test.go
// nlines: 45
// risk_rating: MEDIUM

package api

import (
	"context"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ax"
)

func TestAX_BootProcess(t *testing.T) {
	ctx := context.Background()
	bootID := &mockID{did: "did:test:boot"}
	baseRT := &mockRuntime{id: bootID}

	fac, err := NewAXFactory(ctx, ax.RunnerOpts{SandboxDir: "/tmp/ax-test"}, baseRT, bootID)
	if err != nil {
		t.Fatalf("NewAXFactory() failed: %v", err)
	}

	bootLibScript := `
        func get_lib_msg(returns string) means
            return "from the library"
        endfunc
    `
	bootCmdScript := `
        command
            set boot_var = "secret"
        endcommand
    `
	// Load library functions directly into the factory's root interpreter
	libTree, err := Parse([]byte(bootLibScript), ParseSkipComments)
	if err != nil {
		t.Fatalf("boot: failed to parse library script: %v", err)
	}

	// FIX: Removed invalid type assertion. 'fac' is already the concrete type.
	if err := fac.root.AppendScript(libTree); err != nil {
		t.Fatalf("boot: failed to append library script to root: %v", err)
	}

	// Run the command script in a separate config runner
	configRunner, err := fac.NewRunner(ctx, ax.RunnerConfig, ax.RunnerOpts{})
	if err != nil {
		t.Fatalf("NewRunner(Config) for boot failed: %v", err)
	}
	if err := configRunner.LoadScript([]byte(bootCmdScript)); err != nil {
		t.Fatalf("boot: LoadScript(cmd) failed: %v", err)
	}
	if _, err := configRunner.Execute(); err != nil {
		t.Fatalf("boot: Execute() failed: %v", err)
	}
}
