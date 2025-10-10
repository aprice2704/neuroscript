// NeuroScript Version: 0.8.0
// File version: 7
// Purpose: REFACTOR: Changed package to `api` to access unexported test helpers and fixed mock field name.
// filename: pkg/api/interpreter_shared_stores_test.go
// nlines: 98
// risk_rating: HIGH

package api

import (
	"context"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ax"
)

// TestAX_SharedCatalogs confirms that two separate runners created from the same
// factory can share state via the underlying, host-managed AccountStore and AgentModelStore.
func TestAX_SharedCatalogs(t *testing.T) {
	ctx := context.Background()
	// --- Phase 1: Factory Creation ---
	// The host application creates the factory once. The factory now implicitly
	// owns and manages the shared stores (catalogs).
	factory, err := NewAXFactory(ctx, ax.RunnerOpts{}, &mockRuntime{}, &mockID{did: "did:test:host"})
	if err != nil {
		t.Fatalf("NewAXFactory() failed: %v", err)
	}

	// --- Phase 2: Trusted "Writer" Runner ---

	writerScript := `
command
    must tool.account.register("shared-acct", {\
        "kind": "llm",\
        "provider": "test",\
        "api_key": "key-123"\
    })
    must tool.agentmodel.register("shared-model", {\
        "provider": "test",\
        "model": "model-1",\
        "account_name": "shared-acct"\
    })
endcommand
`
	// Create a privileged CONFIG runner. It gets access to the factory's RunEnv.
	writerRunner, err := factory.NewRunner(ctx, ax.RunnerConfig, ax.RunnerOpts{})
	if err != nil {
		t.Fatalf("Writer: NewRunner(Config) failed: %v", err)
	}

	// Run the writer script to populate the shared catalogs.
	if err := writerRunner.LoadScript([]byte(writerScript)); err != nil {
		t.Fatalf("Writer: LoadScript() failed: %v", err)
	}
	if _, err := writerRunner.Execute(); err != nil {
		t.Fatalf("Writer: Execute() failed: %v", err)
	}

	// At this point, writerRunner can be discarded. The factory's catalogs now contain data.

	// --- Phase 3: Unprivileged "Reader" Runner ---

	readerScript := `
func main(returns bool) means
    set model_exists = tool.agentmodel.exists("shared-model")
    set acct_exists = tool.account.exists("shared-acct")
    return model_exists and acct_exists
endfunc
`
	// Create a new, unprivileged USER runner from the SAME factory.
	readerRunner, err := factory.NewRunner(ctx, ax.RunnerUser, ax.RunnerOpts{})
	if err != nil {
		t.Fatalf("Reader: NewRunner(User) failed: %v", err)
	}

	// Run the reader's main procedure.
	result, err := AXRunScript(ctx, readerRunner, []byte(readerScript), "main")
	if err != nil {
		t.Fatalf("Reader: AXRunScript() failed: %v", err)
	}

	// --- Verification ---
	finalResult, ok := result.(bool)
	if !ok {
		t.Fatalf("Expected a boolean result, but got %T", result)
	}

	if !finalResult {
		t.Error("Test failed: The reader runner could not see the state created by the writer runner.")
	}
}
