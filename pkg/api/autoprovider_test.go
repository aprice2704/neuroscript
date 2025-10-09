// NeuroScript Version: 0.7.4
// File version: 16
// Purpose: FIX: Passes the required turn context directly to RunProcedure, which now correctly propagates it to the interpreter and its clones.
// filename: pkg/api/autoprovider_test.go
// nlines: 104
// risk_rating: HIGH

package api_test

import (
	"context"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/provider/test"
	"github.com/google/uuid"
)

// TestAPI_AutoProviderRegistration verifies that a provider registered via the
// top-level API function is correctly configured and accessible to scripts via 'ask'.
func TestAPI_AutoProviderRegistration(t *testing.T) {
	// 1. Define a script that uses an AgentModel.
	scriptContent := `
func main(returns string) means
    ask "test_agent", "What is a large language model?" into result
    return result
endfunc
`
	// 2. Configure a policy that allows running in a trusted 'config' context.
	configPolicy := policy.NewBuilder(api.ContextConfig).
		Allow("tool.aeiou.magic"). // The bootstrap capsule run by the mock provider needs this.
		Build()

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

	// 6. Run the procedure, passing the required AEIOU turn context.
	// The fix in RunProcedure will ensure this context is set on the interpreter
	// and the fix in clone() ensures it propagates to all child interpreters.
	turnCtx := api.ContextWithSessionID(context.Background(), uuid.NewString())
	result, err := api.RunProcedure(turnCtx, interp, "main")
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
