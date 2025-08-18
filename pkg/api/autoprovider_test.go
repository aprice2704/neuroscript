// NeuroScript Version: 0.6.0
// File version: 2
// Purpose: Verifies that default providers are auto-registered by api.New(). Removes unused variable.
// filename: pkg/api/autoprovider_test.go
// nlines: 64
// risk_rating: MEDIUM

package api_test

import (
	"context"
	"os"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
)

// TestAPI_AutoProviderRegistration confirms that a standard interpreter created
// via api.New() has the default providers (e.g., 'google') registered and ready
// for use without any manual registration steps.
func TestAPI_AutoProviderRegistration(t *testing.T) {
	// This test requires a live API key to be set.
	if os.Getenv("GOOGLE_API_KEY") == "" {
		t.Skip("Skipping live API test: GOOGLE_API_KEY is not set.")
	}

	// 1. Define a script that uses the 'google' provider, which should exist by default.
	// We need a trusted context to register a model. api.NewConfigInterpreter
	// calls api.New() internally, so this correctly tests the new default behavior.
	requiredGrants := []api.Capability{
		{Resource: "model", Verbs: []string{"admin"}, Scopes: []string{"*"}},
		{Resource: "model", Verbs: []string{"use"}, Scopes: []string{"*"}},
		{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"*"}},
		{Resource: "net", Verbs: []string{"read", "write"}, Scopes: []string{"*"}},
	}
	trustedInterp := api.NewConfigInterpreter(nil, requiredGrants)

	script := `
func main(returns result) means
  must tool.agentmodel.Register("default_google", {\
    "provider": "google",\
    "model": "gemini-1.5-flash",\
    "api_key_ref": "GOOGLE_API_KEY"\
  })
  ask "default_google", "briefly explain what a large language model is" into result
  return result
endfunc
`
	// 2. Parse and load the script into the trusted interpreter.
	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse() failed: %v", err)
	}
	if _, err := api.ExecWithInterpreter(context.Background(), trustedInterp, tree); err != nil {
		t.Fatalf("api.ExecWithInterpreter() failed to load script: %v", err)
	}

	// 3. Run the main procedure, which will trigger the 'ask' statement.
	result, err := api.RunProcedure(context.Background(), trustedInterp, "main")
	if err != nil {
		t.Fatalf("api.RunProcedure() failed: %v", err)
	}

	// 4. Verify that we got a non-empty string response.
	unwrapped, _ := api.Unwrap(result)
	val, ok := unwrapped.(string)
	if !ok {
		t.Fatalf("Expected result to be a string, but got %T", unwrapped)
	}
	if val == "" {
		t.Error("Expected a non-empty response from the Google provider, but got an empty string.")
	}

	t.Logf("Received valid response from auto-registered google provider: %s...", val[:50])
}
