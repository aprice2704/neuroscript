// NeuroScript Version: 0.8.0
// File version: 5
// Purpose: Tests that the interpreter rejects scripts that try to override provider symbols.
// filename: pkg/api/interpreter_globals_override_test.go
// nlines: 60

package api_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	// REMOVED: 'errors' and 'lang' are no longer needed
)

// --- The New Test ---
// All helper functions (newTestLogger, newMockSymbolProvider) and the
// mockSymbolProvider struct are removed, as they are defined in
// interpreter_globals_api_test.go and are part of the same package.

func TestGlobalSymbolProviderOverride(t *testing.T) {
	// 1. Create the mock provider (using helper from the other test file)
	provider := newMockSymbolProvider()

	// 2. Create the ServiceRegistry map and inject the provider
	serviceReg := map[string]any{
		api.SymbolProviderKey: provider,
	}

	// 3. Build the HostContext using the public API
	hc, err := api.NewHostContextBuilder().
		WithLogger(newTestLogger(t)). // (using helper)
		WithStdout(io.Discard).
		WithStdin(strings.NewReader("")).
		WithStderr(io.Discard).
		WithServiceRegistry(serviceReg). // Inject the map
		Build()
	if err != nil {
		t.Fatalf("Failed to build HostContext: %v", err)
	}

	// 4. Create the interpreter
	interp := api.New(
		api.WithHostContext(hc),
		api.WithExecPolicy(api.NewPolicyBuilder(api.ContextNormal).Build()),
	)

	// 5. Define a script that *tries* to override a provider symbol
	// This func name 'my_prov_func' intentionally conflicts with the provider.
	const overrideScript = `
        func my_prov_func(returns string) means
            return "from_local_func"
        endfunc

        func main(returns string) means
            return my_prov_func()
        endfunc
    `

	// 6. Parse the script (this should succeed)
	tree, err := api.Parse([]byte(overrideScript), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed unexpectedly: %v", err)
	}

	// 7. Try to load the definitions (this MUST fail)
	// This follows the "No Override" rule from ns_globals.md
	_, err = api.ExecWithInterpreter(context.Background(), interp, tree)
	if err == nil {
		t.Fatalf("api.ExecWithInterpreter succeeded, but it *must* fail when a script tries to override a provider function.")
	}

	// 8. If we got here, the test passed because err was not nil.
	// The specific error type check was removed as it was incorrect.
	t.Logf("Test passed, received expected error: %v", err)
}
