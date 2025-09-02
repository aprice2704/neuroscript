// NeuroScript Version: 0.7.0
// File version: 12
// Purpose: Updated the inline setup script with the corrected NeuroScript syntax and re-verified the Go setup code.
// filename: pkg/livetest/ask_loop_livetest.go
// nlines: 240
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
			"tool_loop_permitted": true,\
			"max_turns": 5\
		})
	endcommand
	`

	// 2. Define permissions for setup AND for the test logic itself.
	allowedTools := []string{
		"tool.os.Getenv",
		"tool.account.Register",
		"tool.agentmodel.Register",
		"tool.fs.*",
		"tool.str.*",
	}
	requiredGrants := []api.Capability{
		api.NewCapability(api.ResEnv, api.VerbRead, "GEMINI_API_KEY"),
		api.NewWithVerbs(api.ResModel, []string{api.VerbAdmin}, []string{"*"}),
	}

	// 3. Create a trusted interpreter with the necessary permissions.
	interp := api.NewConfigInterpreter(allowedTools, requiredGrants)

	// 4. Execute the setup script to configure the interpreter instance.
	if _, err := api.ExecInNewInterpreter(context.Background(), setupScript, api.WithInterpreter(interp)); err != nil {
		t.Fatalf("Failed to execute setup script: %v", err)
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
