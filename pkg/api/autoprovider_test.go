// NeuroScript Version: 0.7.0
// File version: 13
// Purpose: Corrected the test to align with the V3 'ask' statement, which generates a JSON object in the USERDATA section, and updated agent registration to use native Go types. Made the result check less brittle.
// filename: pkg/api/autoprovider_test.go
// nlines: 100
// risk_rating: LOW

package api_test

import (
	"context"
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
	configPolicy := &policy.ExecPolicy{
		Context: policy.ContextConfig,
	}
	interp := api.New(interpreter.WithExecPolicy(configPolicy))

	// 3. Register the mock provider.
	interp.RegisterProvider("mock", test.New())

	// 4. Register an AgentModel using native Go types.
	agentConfig := map[string]any{
		"provider": "mock",
		"model":    "test-model",
	}
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
	// The 'ask' statement will internally create a V3 envelope with a USERDATA
	// section like: {"subject":"ask","fields":{"prompt":"What is a large language model?"}}
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
