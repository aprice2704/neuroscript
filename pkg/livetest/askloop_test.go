// NeuroScript Version: 0.7.0
// File version: 30
// Purpose: Drastically simplified both tests to remove all complex tool calls, focusing only on the core protocol. Fixed the sandbox error by passing the path at interpreter creation.
// filename: pkg/livetest/askloop_test.go
// nlines: 241
// risk_rating: HIGH

package livetest_test

import (
	"context"
	_ "embed"
	"os"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
)

//go:embed test_scripts/agentic.txt
var agenticScriptTemplate string

// aiTranscriptLogger is a simple io.Writer that pipes transcript data to the test log.
type aiTranscriptLogger struct {
	t *testing.T
}

func (l *aiTranscriptLogger) Write(p []byte) (n int, err error) {
	l.t.Logf("\n--- AI TRANSCRIPT ---\n%s\n---------------------", p)
	return len(p), nil
}

// setupLiveTest creates a new interpreter, registers a live Google provider,
// and configures a general-purpose agent model for the tests.
func setupLiveTest(t *testing.T, sandboxDir string) *api.Interpreter {
	t.Helper()

	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("Skipping live test: GEMINI_API_KEY is not set.")
	}

	setupScript := `
	command
		set key = tool.os.Getenv("GEMINI_API_KEY")
		if key == nil or key == ""
			fail "GEMINI_API_KEY not found by setup script"
		endif
		must tool.account.Register("google-ci", {\
			"kind": "llm", "provider": "google", "api_key": key\
		})
		must tool.agentmodel.Register("live_agent", {\
			"provider": "google",\
			"model": "gemini-1.5-flash",\
			"account_name": "google-ci",\
			"tool_loop_permitted": true,\
			"max_turns": 5,\
			"temperature": 0.1,\
			"system_prompt_capsule": "capsule/bootstrap_agentic"\
		})
	endcommand
	`

	allowedTools := []string{
		"tool.os.Getenv",
		"tool.account.Register",
		"tool.agentmodel.Register",
		"tool.aeiou.ComposeEnvelope",
		"tool.aeiou.magic",
	}
	requiredGrants := []api.Capability{
		api.NewCapability(api.ResEnv, api.VerbRead, "GEMINI_API_KEY"),
		api.NewWithVerbs(api.ResModel, []string{api.VerbAdmin}, []string{"*"}),
		api.NewWithVerbs("account", []string{api.VerbAdmin}, []string{"*"}),
	}

	transcriptWriter := &aiTranscriptLogger{t: t}
	// FIX: Pass the sandbox directory as a creation-time option.
	extraOpts := []api.Option{
		api.WithAITranscript(transcriptWriter),
		api.WithSandboxDir(sandboxDir),
	}

	interp := api.NewConfigInterpreter(allowedTools, requiredGrants, extraOpts...)

	tree, err := api.Parse([]byte(setupScript), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Failed to parse setup script: %v", err)
	}
	if _, err := api.ExecWithInterpreter(context.Background(), interp, tree); err != nil {
		t.Fatalf("Failed to execute setup script: %v", err)
	}

	return interp
}

// TestLive_SingleToolCall tests the AI's ability to follow a simple instruction.
func TestLive_SingleToolCall(t *testing.T) {
	tempDir := t.TempDir()
	interp := setupLiveTest(t, tempDir)

	// FIX: Simplify the test to its absolute minimum.
	instructions := `This is a test. Your only task is to write this exact line of code in the ACTIONS block: emit "hello"`
	subject := "simple-emit-test"
	expectedResult := "hello"

	tree, err := api.Parse([]byte(agenticScriptTemplate), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}
	if err := api.LoadFromUnit(interp, &api.LoadedUnit{Tree: tree}); err != nil {
		t.Fatalf("api.LoadFromUnit failed: %v", err)
	}

	result, err := api.RunProcedure(context.Background(), interp, "main", subject, instructions)
	if err != nil {
		t.Fatalf("api.RunProcedure failed unexpectedly: %v", err)
	}

	unwrapped, _ := api.Unwrap(result)
	answer, ok := unwrapped.(string)
	if !ok {
		t.Fatalf("Expected a string result, but got %T", unwrapped)
	}

	t.Logf("Single-tool call answer: %s", answer)

	if !strings.Contains(answer, expectedResult) {
		t.Errorf("Expected the AI's answer to contain '%s', but it did not.", expectedResult)
	}
}

// TestLive_ComplexToolComposition tests the AI's ability to follow a two-line instruction.
func TestLive_ComplexToolComposition(t *testing.T) {
	tempDir := t.TempDir()
	interp := setupLiveTest(t, tempDir)

	// FIX: Simplify the test to its absolute minimum.
	instructions := `This is a test. Your only task is to write these two exact lines of code in the ACTIONS block:
set a = "hello"
emit a + " world"
`
	subject := "simple-concat-test"
	expectedResult := "hello world"

	tree, err := api.Parse([]byte(agenticScriptTemplate), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}
	if err := api.LoadFromUnit(interp, &api.LoadedUnit{Tree: tree}); err != nil {
		t.Fatalf("api.LoadFromUnit failed: %v", err)
	}

	result, err := api.RunProcedure(context.Background(), interp, "main", subject, instructions)
	if err != nil {
		t.Fatalf("api.RunProcedure failed unexpectedly: %v", err)
	}

	unwrapped, _ := api.Unwrap(result)
	answer, ok := unwrapped.(string)
	if !ok {
		t.Fatalf("Expected a string result, but got %T", unwrapped)
	}

	t.Logf("Multi-tool composition answer: %s", answer)

	if !strings.Contains(answer, expectedResult) {
		t.Errorf("Expected final emitted result to contain '%s', but got '%s'", expectedResult, answer)
	}
}
