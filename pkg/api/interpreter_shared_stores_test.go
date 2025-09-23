// NeuroScript Version: 0.7.3
// File version: 2
// Purpose: Corrects the test script to use lowercase tool names and proper line continuations.
// filename: pkg/api/interpreter_shared_stores_test.go
// nlines: 127
// risk_rating: HIGH

package api_test

import (
	"context"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
)

// TestInterpreter_SharedStores confirms that two separate interpreters can share
// state by being configured with the same host-managed AccountStore and AgentModelStore.
func TestInterpreter_SharedStores(t *testing.T) {
	// --- Phase 1: Host-Level Store Creation ---

	// The host application (e.g., FDM) creates the stores once.
	sharedAccountStore := api.NewAccountStore()
	sharedAgentModelStore := api.NewAgentModelStore()

	// --- Phase 2: Trusted "Writer" Interpreter ---

	// A script to register a new account and a new agent model.
	writerScript := `
command
    # Register an account that the second interpreter will need.
    must tool.account.register("shared-acct", {\
        "kind": "llm",\
        "provider": "test",\
        "api_key": "key-123"\
    })

    # Register an agent model that uses the shared account.
    must tool.agentmodel.register("shared-model", {\
        "provider": "test",\
        "model": "model-1",\
        "account_name": "shared-acct"\
    })
endcommand
`
	// Configure a trusted interpreter to run the writer script.
	// It is given write access to the shared stores.
	writerOpts := []api.Option{
		api.WithAccountStore(sharedAccountStore),
		api.WithAgentModelStore(sharedAgentModelStore),
	}
	writerAllowedTools := []string{"tool.account.register", "tool.agentmodel.register"}
	writerGrants := []api.Capability{
		api.NewWithVerbs(api.ResAccount, []string{api.VerbAdmin}, []string{"*"}),
		api.NewWithVerbs(api.ResModel, []string{api.VerbAdmin}, []string{"*"}),
	}
	writerInterp := api.NewConfigInterpreter(writerAllowedTools, writerGrants, writerOpts...)

	// Run the writer script to populate the shared stores.
	tree, err := api.Parse([]byte(writerScript), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Writer: api.Parse() failed: %v", err)
	}
	_, err = api.ExecWithInterpreter(context.Background(), writerInterp, tree)
	if err != nil {
		t.Fatalf("Writer: api.ExecWithInterpreter() failed: %v", err)
	}

	// At this point, writerInterp can be discarded, but the shared stores now contain data.

	// --- Phase 3: Unprivileged "Reader" Interpreter ---

	// A script to verify that the data from the writer is accessible.
	readerScript := `
func main(returns bool) means
    set model_exists = tool.agentmodel.exists("shared-model")
    set acct_exists = tool.account.exists("shared-acct")
    return model_exists and acct_exists
endfunc
`
	// Configure a new, unprivileged interpreter.
	// It is given access to the SAME shared stores.
	readerOpts := []api.Option{
		api.WithAccountStore(sharedAccountStore),
		api.WithAgentModelStore(sharedAgentModelStore),
	}
	readerAllowedTools := []string{"tool.agentmodel.exists", "tool.account.exists"}
	readerGrants := []api.Capability{
		api.NewWithVerbs(api.ResAccount, []string{api.VerbRead}, []string{"*"}),
		api.NewWithVerbs(api.ResModel, []string{api.VerbRead}, []string{"*"}),
	}
	// Note: Using NewConfigInterpreter just for convenience of setting tools/grants.
	// A standard api.New() with a manually built policy would also work.
	readerInterp := api.NewConfigInterpreter(readerAllowedTools, readerGrants, readerOpts...)

	// Load the reader script's definitions.
	tree, err = api.Parse([]byte(readerScript), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Reader: api.Parse() failed: %v", err)
	}
	// We need to load the function definition.
	if err := api.LoadFromUnit(readerInterp, &api.LoadedUnit{Tree: tree}); err != nil {
		t.Fatalf("Reader: LoadFromUnit() failed: %v", err)
	}

	// Run the reader's main procedure.
	result, err := api.RunProcedure(context.Background(), readerInterp, "main")
	if err != nil {
		t.Fatalf("Reader: api.RunProcedure() failed: %v", err)
	}

	// --- Verification ---
	unwrapped, err := api.Unwrap(result)
	if err != nil {
		t.Fatalf("api.Unwrap() failed: %v", err)
	}

	finalResult, ok := unwrapped.(bool)
	if !ok {
		t.Fatalf("Expected a boolean result, but got %T", unwrapped)
	}

	if !finalResult {
		t.Error("Test failed: The reader interpreter could not see the state created by the writer interpreter.")
	}
}
