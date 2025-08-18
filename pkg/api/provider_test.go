// NeuroScript Version: 0.6.0
// File version: 2
// Purpose: Adds missing env/net grants to the trusted interpreter policy to allow the 'ask' statement to execute.
// filename: pkg/api/provider_test.go
// nlines: 81
// risk_rating: MEDIUM

package api_test

import (
	"context"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	// Import the dummy provider for use in this test.
	"github.com/aprice2704/neuroscript/pkg/api/providers/test"
)

// TestAPI_RegisterAndUseProvider verifies the full workflow of registering a
// custom AI provider and using it via an 'ask' statement in a trusted context.
func TestAPI_RegisterAndUseProvider(t *testing.T) {
	// 1. Define the grants and tools needed for the script to run.
	// The `ask` statement requires env and net grants to satisfy the agent model's runtime envelope.
	requiredGrants := []api.Capability{
		{Resource: "model", Verbs: []string{"admin"}, Scopes: []string{"*"}},
		{Resource: "model", Verbs: []string{"use"}, Scopes: []string{"*"}},
		{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"*"}},
		{Resource: "net", Verbs: []string{"read", "write"}, Scopes: []string{"*"}},
	}
	allowedTools := []string{
		"tool.agentmodel.Register",
		"tool.agentmodel.Ask",
	}

	// 2. Create a trusted interpreter with the necessary permissions.
	interp := api.NewConfigInterpreter(allowedTools, requiredGrants)

	// 3. Instantiate and register the custom test provider.
	testProvider := test.New()
	interp.RegisterProvider("test_provider", testProvider)

	// 4. Define a script that uses the test provider.
	script := `
func main(returns result) means
  must tool.agentmodel.Register("test_agent", {\
    "provider": "test_provider",\
    "model": "test-model-1"\
  })
  ask "test_agent", "ping" into result
  return result
endfunc
`
	// 5. Parse and load the script into the interpreter.
	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse() failed: %v", err)
	}
	if _, err := api.ExecWithInterpreter(context.Background(), interp, tree); err != nil {
		t.Fatalf("api.ExecWithInterpreter() failed to load script: %v", err)
	}

	// 6. Run the main procedure, which will trigger the 'ask' statement.
	result, err := api.RunProcedure(context.Background(), interp, "main")
	if err != nil {
		t.Fatalf("api.RunProcedure() failed: %v", err)
	}

	// 7. Unwrap the result and verify it matches the test provider's expected response.
	unwrapped, _ := api.Unwrap(result)
	val, ok := unwrapped.(string)
	if !ok {
		t.Fatalf("Expected result to be a string, but got %T", unwrapped)
	}

	expectedResponse := "test_provider_ok:ping"
	if val != expectedResponse {
		t.Errorf("Expected response '%s', but got '%s'", expectedResponse, val)
	}
}
