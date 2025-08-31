// NeuroScript Version: 0.7.0
// File version: 5
// Purpose: Simplified the prompt, relying on the bootstrap capsule for formatting rules.
// filename: pkg/livetest/oneshot_test.go
// nlines: 88
// risk_rating: HIGH
package livetest_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
)

// setupOneShotTest creates an interpreter configured for a live one-shot test.
// It skips the test if the GEMINI_API_KEY environment variable is not set.
func setupOneShotTest(t *testing.T) *api.Interpreter {
	t.Helper()

	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("Skipping live test: GEMINI_API_KEY is not set.")
	}

	// Create an interpreter in a trusted context to allow agent registration.
	configPolicy := &api.ExecPolicy{
		Context: api.ContextConfig,
	}

	interp := api.New(api.WithExecPolicy(configPolicy))

	// Configure the agent to use the API key from the environment.
	agentConfig := map[string]any{
		"provider":            "google",
		"model":               "gemini-1.5-flash",
		"api_key_ref":         "GEMINI_API_KEY",
		"tool_loop_permitted": false, // This is a one-shot agent.
	}

	if err := interp.RegisterAgentModel("live_agent", agentConfig); err != nil {
		t.Fatalf("Failed to register agent model for live test: %v", err)
	}

	return interp
}

// TestLive_OneShotQuery tests a simple, single-turn factual question.
func TestLive_OneShotQuery(t *testing.T) {
	interp := setupOneShotTest(t)

	script := `
func main(returns result) means
    # The prompt is now just the user's question.
    # Formatting rules are provided by the bootstrap capsule automatically.
    set prompt = "What were the names of the three astronauts who flew on the Apollo 13 mission?"

    ask "live_agent", prompt into result
    return result
endfunc
`
	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}
	if err := api.LoadFromUnit(interp, &api.LoadedUnit{Tree: tree}); err != nil {
		t.Fatalf("api.LoadFromUnit failed: %v", err)
	}

	result, err := api.RunProcedure(context.Background(), interp, "main")
	if err != nil {
		// Log the raw error which might contain the full AI response for debugging.
		t.Logf("Full error from RunProcedure: %v", err)
		t.Fatalf("api.RunProcedure failed unexpectedly: %v", err)
	}

	unwrapped, _ := api.Unwrap(result)
	answer, ok := unwrapped.(string)
	if !ok {
		t.Fatalf("Expected a string result, but got %T", unwrapped)
	}

	t.Logf("One-shot query final answer: %s", answer)

	// Verify the answer contains the key names.
	for _, name := range []string{"Lovell", "Swigert", "Haise"} {
		if !strings.Contains(answer, name) {
			t.Errorf("Expected answer to contain '%s', but it did not.", name)
		}
	}
}
