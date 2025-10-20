// NeuroScript Version: 0.8.0
// File version: 19
// Purpose: Corrects test failure by using api.NewPolicyBuilder to construct the policy.
// filename: pkg/api/autoprovider_test.go
// nlines: 86
// risk_rating: LOW

package api_test

import (
	"context" // DEBUG
	// DEBUG
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/provider/test"
)

// TestAPI_AutoProviderRegistration verifies that a provider registered via the
// top-level API function is correctly configured and accessible to scripts via 'ask'.
func TestAPI_AutoProviderRegistration(t *testing.T) {
	// DEBUG
	// fmt.Fprintf(os.Stderr, "[DEBUG] TestAPI_AutoProviderRegistration: START\n")

	// 1. Define a script that uses an AgentModel.
	scriptContent := `
func main(returns string) means
    # The 'ask' statement uses an AgentModel, which in turn uses our registered provider.
    ask "test_agent", "What is a large language model?" into result
    return result
endfunc
`
	// 2. Configure a policy that allows running in a trusted 'config' context.
	// This is required to call RegisterAgentModel.
	// THE FIX: Use the api.NewPolicyBuilder to correctly create the policy
	// with the required 'model:admin:*' grant.
	configPolicy := api.NewPolicyBuilder(policy.ContextConfig).
		Grant("model:admin:*").
		Build()

	interp := api.New(
		api.WithHostContext(newTestHostContext(nil)),
		interpreter.WithExecPolicy(configPolicy),
	)

	// 3. Register the mock provider.
	interp.RegisterProvider("mock", test.New())
	// fmt.Fprintf(os.Stderr, "[DEBUG] TestAPI_AutoProviderRegistration: Mock provider registered.\n")

	// 4. Register an AgentModel using native Go types.
	agentConfig := map[string]any{
		"provider": "mock",
		"model":    "test-model",
	}
	if err := interp.RegisterAgentModel("test_agent", agentConfig); err != nil {
		// fmt.Fprintf(os.Stderr, "[DEBUG] TestAPI_AutoProviderRegistration: RegisterAgentModel FAILED: %v\n", err) // DEBUG
		t.Fatalf("Failed to register agent model: %v", err)
	}
	// fmt.Fprintf(os.Stderr, "[DEBUG] TestAPI_AutoProviderRegistration: Agent model registered.\n")

	// 5. Parse and load the script.
	tree, err := api.Parse([]byte(scriptContent), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}
	if _, err := api.ExecWithInterpreter(context.Background(), interp, tree); err != nil {
		t.Fatalf("api.ExecWithInterpreter failed: %v", err)
	}
	// fmt.Fprintf(os.Stderr, "[DEBUG] TestAPI_AutoProviderRegistration: Script loaded.\n")

	// 6. Run the procedure.
	// The 'ask' statement will internally create a V3 envelope with a USERDATA
	// section like: {"subject":"ask","fields":{"prompt":"What is a large language model?"}}
	result, err := api.RunProcedure(context.Background(), interp, "main")
	if err != nil {
		// fmt.Fprintf(os.Stderr, "[DEBUG] TestAPI_AutoProviderRegistration: RunProcedure FAILED: %v\n", err) // DEBUG
		t.Fatalf("api.RunProcedure failed unexpectedly: %v", err)
	}
	// fmt.Fprintf(os.Stderr, "[DEBUG] TestAPI_AutoProviderRegistration: RunProcedure complete.\n")

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
	// fmt.Fprintf(os.Stderr, "[DEBUG] TestAPI_AutoProviderRegistration: END\n")
}
