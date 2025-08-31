// NeuroScript Version: 0.7.0
// File version: 4
// Purpose: Updated prompts to explicitly instruct the AI to use the AEIOU envelope format.
// filename: pkg/livetest/ask_loop_livetest.go
// nlines: 213
// risk_rating: HIGH

package livetest_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
)

// setupLiveTest creates a new interpreter, registers a live Google provider,
// and configures a general-purpose agent model for the tests.
// It skips the test if the GEMINI_API_KEY environment variable is not set.
func setupLiveTest(t *testing.T) *api.Interpreter {
	t.Helper()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: GEMINI_API_KEY is not set.")
	}

	// Create an interpreter in a trusted context to allow agent registration.
	configPolicy := &api.ExecPolicy{
		Context: api.ContextConfig,
		Allow:   []string{"tool.fs.*", "tool.str.*"},
	}

	// Note: api.New() automatically registers the "google" provider.
	interp := api.New(api.WithExecPolicy(configPolicy))

	// Configure the agent to use the API key from the environment.
	agentConfig := map[string]any{
		"provider":            "google",
		"model":               "gemini-1.5-flash",
		"secret_ref":          "GEMINI_API_KEY", // This tells the provider where to find the key.
		"tool_loop_permitted": true,
		"max_turns":           5,
	}

	if err := interp.RegisterAgentModel("live_agent", agentConfig); err != nil {
		t.Fatalf("Failed to register agent model for live test: %v", err)
	}

	return interp
}

// TestLive_SingleToolCall tests the AI's ability to follow instructions
// to call a single tool to answer a question.
func TestLive_SingleToolCall(t *testing.T) {
	interp := setupLiveTest(t)
	tempDir := t.TempDir()
	interp.SetSandboxDir(tempDir) // Set the sandbox for the fs tool

	goalFile := filepath.Join(tempDir, "goal.txt")
	goalContent := "The project goal is to create a robust and secure scripting language."
	if err := os.WriteFile(goalFile, []byte(goalContent), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	prompt := `You MUST format your entire response as a valid AEIOU envelope. Your goal is to answer the user's question. To do so, you MUST call the tool 'tool.fs.Read' with the file path "goal.txt" and place the content of that file inside an 'emit' statement in your ACTIONS block. User's question: What is the project's goal?`
	script := fmt.Sprintf(`
func main(returns result) means
    ask "live_agent", %q into result
    return result
endfunc
`, prompt)

	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}
	if err := api.LoadFromUnit(interp, &api.LoadedUnit{Tree: tree}); err != nil {
		t.Fatalf("api.LoadFromUnit failed: %v", err)
	}

	result, err := api.RunProcedure(context.Background(), interp, "main")
	if err != nil {
		t.Fatalf("api.RunProcedure failed unexpectedly: %v", err)
	}

	unwrapped, _ := api.Unwrap(result)
	answer, ok := unwrapped.(string)
	if !ok {
		t.Fatalf("Expected a string result, but got %T", unwrapped)
	}

	t.Logf("Single-tool call answer: %s", answer)

	if !strings.Contains(answer, goalContent) {
		t.Errorf("Expected the AI's answer to contain the file content, but it did not.")
	}
}

// TestLive_ComplexToolComposition tests a multi-step task requiring the AI
// to chain multiple tool calls together in a specific sequence.
func TestLive_ComplexToolComposition(t *testing.T) {
	interp := setupLiveTest(t)
	tempDir := t.TempDir()
	interp.SetSandboxDir(tempDir) // Set the sandbox for the fs tool

	// Setup the initial files
	part1File := filepath.Join(tempDir, "part1.txt")
	part2File := filepath.Join(tempDir, "part2.txt")
	summaryFile := filepath.Join(tempDir, "summary.txt")

	part1Content := "NeuroScript is a language."
	part2Content := "It is designed for safety."
	expectedSummary := "NeuroScript is a language. It is designed for safety."

	os.WriteFile(part1File, []byte(part1Content), 0600)
	os.WriteFile(part2File, []byte(part2Content), 0600)

	prompt := `You MUST format your entire response as a valid AEIOU envelope. Follow these steps exactly and do not add any extra commentary in your final output. 1. Call 'tool.fs.Read' to get the content of "part1.txt". 2. Call 'tool.fs.Read' to get the content of "part2.txt". 3. Call 'tool.str.Join' with the results from steps 1 and 2 and a space " " as the separator. 4. Call 'tool.fs.Write' to save the result from step 3 into a new file named "summary.txt". 5. Finally, emit the content you wrote to "summary.txt" in your ACTIONS block.`
	script := fmt.Sprintf(`
func main(returns result) means
    ask "live_agent", %q into result
    return result
endfunc
`, prompt)

	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}
	if err := api.LoadFromUnit(interp, &api.LoadedUnit{Tree: tree}); err != nil {
		t.Fatalf("api.LoadFromUnit failed: %v", err)
	}

	result, err := api.RunProcedure(context.Background(), interp, "main")
	if err != nil {
		t.Fatalf("api.RunProcedure failed unexpectedly: %v", err)
	}

	// Verify the final emitted result from the 'ask' statement
	unwrapped, _ := api.Unwrap(result)
	answer, ok := unwrapped.(string)
	if !ok {
		t.Fatalf("Expected a string result, but got %T", unwrapped)
	}

	t.Logf("Multi-tool composition answer: %s", answer)
	if !strings.Contains(answer, expectedSummary) {
		t.Errorf("Expected final emitted result to be '%s', but got '%s'", expectedSummary, answer)
	}

	// Also verify the side-effect: the summary file was written correctly.
	summaryData, err := os.ReadFile(summaryFile)
	if err != nil {
		t.Fatalf("The summary.txt file was not created by the agent: %v", err)
	}
	if string(summaryData) != expectedSummary {
		t.Errorf("Expected summary.txt to contain '%s', but got '%s'", expectedSummary, string(summaryData))
	}
}
