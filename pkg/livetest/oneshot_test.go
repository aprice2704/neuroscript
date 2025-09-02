// NeuroScript Version: 0.7.0
// File version: 19
// Purpose: Corrected the debug script to iterate over the lists returned by tool calls, as 'emit' does not support printing lists directly.
// filename: pkg/livetest/oneshot_test.go
// nlines: 145
// risk_rating: HIGH
package livetest_test

import (
	"bytes"
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

	// 1. Define the setup script to register the account and agent model.
	setupScript := `
	command
		set key = tool.os.Getenv("GEMINI_API_KEY")
		if key == nil or key == ""
			fail "GEMINI_API_KEY not found by setup script"
		endif
		must tool.account.Register("google-ci", {\
			"kind": "llm", "provider": "google", "apiKey": key\
		})
		must tool.agentmodel.Register("live_agent", {\
			"provider": "google",\
			"model": "gemini-1.5-flash",\
			"AccountName": "google-ci",\
			"tool_loop_permitted": false\
		})
	endcommand
	`

	// 2. Define permissions, including for the debug/listing tools.
	allowedTools := []string{
		"tool.os.Getenv",
		"tool.account.*",
		"tool.agentmodel.*",
		"tool.provider.*",
		"tool.str.*",
	}
	requiredGrants := []api.Capability{
		api.NewCapability(api.ResEnv, api.VerbRead, "GEMINI_API_KEY"),
		api.NewWithVerbs(api.ResModel, []string{api.VerbAdmin, api.VerbRead}, []string{"*"}),
		api.NewWithVerbs("account", []string{api.VerbAdmin, api.VerbRead}, []string{"*"}),
		api.NewWithVerbs("provider", []string{api.VerbRead}, []string{"*"}),
	}

	// 3. Create a trusted interpreter.
	interp := api.NewConfigInterpreter(allowedTools, requiredGrants)

	// 4. Parse and execute the setup script.
	tree, err := api.Parse([]byte(setupScript), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Failed to parse setup script: %v", err)
	}
	if _, err := api.ExecWithInterpreter(context.Background(), interp, tree); err != nil {
		t.Fatalf("Failed to execute setup script: %v", err)
	}

	return interp
}

// TestLive_OneShotQuery tests a simple, single-turn factual question.
func TestLive_OneShotQuery(t *testing.T) {
	interp := setupOneShotTest(t)

	// This script now iterates over the lists to print them.
	script := `
func main(returns result) means
    emit "--- DEBUG: STATE BEFORE ASK ---"
    emit "ACCOUNTS:"
    set account_list = tool.account.List()
    for each name in account_list
        emit "  - " + name
    endfor

    emit "AGENT MODELS:"
    set agent_list = tool.agentmodel.List()
    for each name in agent_list
        emit "  - " + name
    endfor
    emit "-----------------------------"

    set prompt = "What were the names of the three astronauts who flew on the Apollo 13 mission?"
    ask "live_agent", prompt into result
    return result
endfunc
`
	var testOutput bytes.Buffer
	interp.SetStdout(&testOutput)

	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}
	if err := api.LoadFromUnit(interp, &api.LoadedUnit{Tree: tree}); err != nil {
		t.Fatalf("api.LoadFromUnit failed: %v", err)
	}

	result, err := api.RunProcedure(context.Background(), interp, "main")
	t.Logf("Test script debug output:\n%s", testOutput.String())
	if err != nil {
		t.Logf("Full error from RunProcedure: %v", err)
		t.Fatalf("api.RunProcedure failed unexpectedly: %v", err)
	}

	unwrapped, _ := api.Unwrap(result)
	answer, ok := unwrapped.(string)
	if !ok {
		t.Fatalf("Expected a string result, but got %T", unwrapped)
	}

	t.Logf("One-shot query final answer: %s", answer)

	for _, name := range []string{"Lovell", "Swigert", "Haise"} {
		if !strings.Contains(answer, name) {
			t.Errorf("Expected answer to contain '%s', but it did not.", name)
		}
	}
}
