// NeuroScript Version: 0.8.0
// File version: 26
// Purpose: Corrects test failure by adding 'tool.agentmodel.register' to the policy's allow list.
// filename: pkg/api/autoprovider_test.go
// nlines: 93
// risk_rating: LOW

package api_test

import (
	"context" // DEBUG
	// DEBUG
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/provider" // FIX: Import provider package
	"github.com/aprice2704/neuroscript/pkg/provider/test"
)

// TestAPI_AutoProviderRegistration verifies that a provider registered via the
// top-level API function is correctly configured and accessible to scripts via 'ask'.
func TestAPI_AutoProviderRegistration(t *testing.T) {
	// 1. Define a script that uses an AgentModel.
	scriptContent := `
func main(returns string) means
    # The 'ask' statement uses an AgentModel, which in turn uses our registered provider.
    ask "test_agent", "What is a large language model?" into result
    return result
endfunc
`
	// 2. Configure a policy that allows running in a trusted 'config' context.
	grant := api.MustParse("model:admin:*")
	configPolicy := api.NewPolicyBuilder(policy.ContextConfig).
		GrantCap(grant).
		// FIX: The RegisterAgentModel Go method is a wrapper for the tool.
		// The tool must be explicitly allowed.
		Allow("tool.agentmodel.register").
		// REMOVED: Allow("model:admin:*") <-- This was incorrect.
		Build()

	// 3. Create and populate the new ProviderRegistry
	providerRegistry := api.NewProviderRegistry()
	// FIX: Use provider.NewAdmin directly
	providerAdmin := provider.NewAdmin(providerRegistry, configPolicy)
	if err := providerAdmin.Register("mock", test.New()); err != nil {
		t.Fatalf("Failed to register mock provider: %v", err)
	}

	interp := api.New(
		api.WithHostContext(newTestHostContext(nil)),
		api.WithExecPolicy(configPolicy),
		api.WithProviderRegistry(providerRegistry), // Inject the new registry
	)

	// 4. Register an AgentModel using native Go types.
	agentConfig := map[string]lang.Value{
		"provider": lang.StringValue{Value: "mock"},
		"model":    lang.StringValue{Value: "test-model"},
	}
	// Use the string-based method
	if err := interp.RegisterAgentModel("test_agent", agentConfig); err != nil {
		t.Fatalf("Failed to register agent model: %v", err)
	}

	// 5. Parse and load the script.
	tree, err := api.Parse([]byte(scriptContent), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}
	if _, err := api.ExecWithInterpreter(context.Background(), interp, tree); err != nil {
		t.Fatalf("api.ExecWithInterpreter failed: %v", err)
	}

	// 6. Run the procedure.
	result, err := api.RunProcedure(context.Background(), interp, "main")
	if err != nil {
		t.Fatalf("api.RunProcedure failed unexpectedly: %v", err)
	}

	// 7. Verify the result from the mock provider.
	unwrapped, err := api.Unwrap(result)
	if err != nil {
		t.Fatalf("api.Unwrap failed: %v", err)
	}
	val, ok := unwrapped.(string)
	if !ok {
		t.Fatalf("Expected a string return type, but got %T", unwrapped)
	}

	expectedResponse := "large language model"
	if !strings.Contains(val, expectedResponse) {
		t.Errorf("Expected response to contain '%s', but got: '%s'", expectedResponse, val)
	}
}
