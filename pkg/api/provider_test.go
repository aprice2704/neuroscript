// NeuroScript Version: 0.8.0
// File version: 22
// Purpose: Removes incorrect 'Allow("model:admin:*")' from policy builder.
// filename: pkg/api/provider_test.go
// nlines: 109
// risk_rating: LOW

package api_test

import (
	"context"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// mockSimpleProvider is a test provider that emulates the new AI behavior:
// It just emits the answer and does NOT call tool.aeiou.magic.
type mockSimpleProvider struct{}

func (m *mockSimpleProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	// 1. We'd find the prompt in the req.RawPrompt...
	//    For this test, we'll just assume it was "ping".

	// 2. Compose the new, simpler response.
	actions := `
command
    emit "test_provider_ok:pong"
endcommand
`
	env := &aeiou.Envelope{UserData: "{}", Actions: actions}
	respText, _ := env.Compose()
	return &provider.AIResponse{TextContent: respText}, nil
}

func TestAPI_RegisterAndUseProvider(t *testing.T) {
	providerName := "test_provider"

	// The NeuroScript code to be executed.
	scriptContent := `
func main(returns string) means
    ask "test_agent", "ping" into result
    return result
endfunc
`
	// Create an interpreter with a trusted 'config' context to allow registration.
	// FIX: Add grant AND allow the 'tool.agentmodel.register' tool.
	grant := api.MustParse("model:admin:*")
	configPolicy := api.NewPolicyBuilder(api.ContextConfig).
		GrantCap(grant).
		Allow("tool.agentmodel.register").
		// REMOVED: Allow("model:admin:*") <-- This was incorrect.
		Build()

	// FIX: A HostContext is now mandatory for creating an interpreter.
	hc := newTestHostContext(nil)

	// FIX: Create and populate the new ProviderRegistry
	providerRegistry := api.NewProviderRegistry()
	providerAdmin := provider.NewAdmin(providerRegistry, configPolicy)
	if err := providerAdmin.Register(providerName, &mockSimpleProvider{}); err != nil {
		t.Fatalf("Failed to register provider in registry: %v", err)
	}

	interp := api.New(
		api.WithHostContext(hc),
		api.WithExecPolicy(configPolicy),
		api.WithProviderRegistry(providerRegistry), // Inject the registry
	)

	// Register an AgentModel configured to use our test provider.
	agentConfig := map[string]lang.Value{
		"provider": lang.StringValue{Value: providerName},
		"model":    lang.StringValue{Value: "default"},
	}
	// Use the string-based method
	if err := interp.RegisterAgentModel("test_agent", agentConfig); err != nil {
		t.Fatalf("Failed to register agent model: %v", err)
	}

	// Parse and load the script.
	tree, err := api.Parse([]byte(scriptContent), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}
	if _, err := api.ExecWithInterpreter(context.Background(), interp, tree); err != nil {
		t.Fatalf("api.ExecWithInterpreter failed to load definitions: %v", err)
	}

	// Run the procedure.
	// The 'ask' statement will now run the ACTIONS block and return
	// the emitted value. The Go loop manages completion.
	result, err := api.RunProcedure(context.Background(), interp, "main")
	if err != nil {
		t.Fatalf("api.RunProcedure() failed: %v", err)
	}

	// Verify the final result.
	unwrapped, err := api.Unwrap(result)
	if err != nil {
		t.Fatalf("api.Unwrap failed: %v", err)
	}

	val, ok := unwrapped.(string)
	if !ok {
		t.Fatalf("Expected a string return type, but got %T", unwrapped)
	}

	// The mock provider is hard-coded to return "test_provider_ok:pong".
	expectedResponse := "test_provider_ok:pong"
	if !strings.Contains(val, expectedResponse) {
		t.Errorf("Expected response to contain '%s', but got '%s'", expectedResponse, val)
	}
}
