// NeuroScript Version: 0.8.0
// File version: 5
// Purpose: Tests that the SymbolProvider wiring works from the public API.
// filename: pkg/api/interpreter_globals_api_test.go
// nlines: 114

package api_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/ast" // ADDED: Fixes 'undefined: ast.StringLiteralNode'
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// --- Mock Logger ---
// We need a logger that implements api.Logger.
// Using the re-exported NoOpLogger is easiest.
func newTestLogger(t *testing.T) api.Logger {
	return logging.NewNoOpLogger() // Use the silent logger
}

// --- Mock Symbol Provider ---
// RENAMED: from mockAPIProvider to avoid redeclaration error
type mockSymbolProvider struct {
	procs  map[string]any
	consts map[string]any
}

// RENAMED: from newMockAPIProvider
func newMockSymbolProvider() *mockSymbolProvider {
	// Create the provider's symbols
	providerProc := &ast.Procedure{
		// A simplified AST node for testing
		RequiredParams: []string{},
		Steps: []ast.Step{
			{
				Type: "return",
				Values: []ast.Expression{
					// FIXED: Was ast.StringLiteral, now ast.StringLiteralNode
					&ast.StringLiteralNode{Value: "from_provider_func"},
				},
			},
		},
	}

	providerConst := lang.NumberValue{Value: 777}

	return &mockSymbolProvider{
		procs:  map[string]any{"my_prov_func": providerProc},
		consts: map[string]any{"MY_PROV_CONST": providerConst},
	}
}

func (m *mockSymbolProvider) GetProcedure(name string) (any, bool) {
	p, ok := m.procs[name]
	return p, ok
}
func (m *mockSymbolProvider) ListProcedures() map[string]any                  { return m.procs }
func (m *mockSymbolProvider) GetEventHandlers(eventName string) ([]any, bool) { return nil, false }
func (m *mockSymbolProvider) ListEventHandlers() map[string][]any             { return nil }
func (m *mockSymbolProvider) GetGlobalConstant(name string) (any, bool) {
	c, ok := m.consts[name]
	return c, ok
}
func (m *mockSymbolProvider) ListGlobalConstants() map[string]any { return m.consts }

// --- The API Test ---

func TestGlobalSymbolProviderAPI(t *testing.T) {
	// 1. Create the mock provider
	// RENAMED: to use new constructor
	provider := newMockSymbolProvider()

	// 2. Create the ServiceRegistry map and inject the provider
	//    using the now-public api.SymbolProviderKey.
	serviceReg := map[string]any{
		api.SymbolProviderKey: provider,
	}

	// 3. Build the HostContext using the public API
	hc, err := api.NewHostContextBuilder().
		WithLogger(newTestLogger(t)).
		WithStdout(io.Discard).
		WithStdin(strings.NewReader("")).
		WithStderr(io.Discard).
		WithServiceRegistry(serviceReg). // Inject the map
		Build()
	if err != nil {
		t.Fatalf("Failed to build HostContext: %v", err)
	}

	// 4. Create the interpreter using the public API
	interp := api.New(
		api.WithHostContext(hc),
		api.WithExecPolicy(api.NewPolicyBuilder(api.ContextNormal).Build()),
	)

	// 5. Load a script that *uses* the provider's symbols
	// FIXED: Corrected NeuroScript syntax from 'call...into' to 'set...='
	const runtimeScript = `
        func main(returns string) means
            set res = my_prov_func()
            return res + " and " + MY_PROV_CONST
        endfunc
    `
	mustLoadStringAPI(t, interp, runtimeScript)

	// 6. Run the script and verify the result
	// FIXED: Was interp.RunProcedure, now api.RunProcedure
	result, err := api.RunProcedure(context.Background(), interp, "main")
	if err != nil {
		t.Fatalf("Runtime failed to execute 'main': %v", err)
	}

	resultStr, _ := lang.ToString(result)
	expected := "from_provider_func and 777"
	if resultStr != expected {
		t.Errorf("Test failed:\nExpected: %s\nGot:      %s", expected, resultStr)
	}
}

// mustLoadStringAPI is a test helper for the api_test package
// FIXED: This is now a real function, not a comment.
func mustLoadStringAPI(t *testing.T, i *api.Interpreter, script string) {
	t.Helper()
	// 1. Parse using the public API (from pkg/api/parse.go)
	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}

	// 2. Load definitions using the public API
	// As per api_guide.md, ExecWithInterpreter loads definitions.
	if _, err := api.ExecWithInterpreter(context.Background(), i, tree); err != nil {
		t.Fatalf("api.ExecWithInterpreter (for loading) failed: %v", err)
	}
}
