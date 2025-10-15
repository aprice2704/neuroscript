// NeuroScript Version: 0.7.0
// File version: 55
// Purpose: No code changes needed; test passes with the corrected model name in the associated 'oneshot.txt' script file.
// filename: pkg/livetest/oneshot_test.go
// nlines: 194
// risk_rating: HIGH
package livetest_test

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

//go:embed test_scripts/oneshot.txt
var oneShotScriptTemplate string

// abbreviate returns a shortened version of a multi-line string for cleaner test logs.
func abbreviate(s string, maxLines int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	var sb strings.Builder
	for i := 0; i < maxLines/2; i++ {
		sb.WriteString(lines[i])
		sb.WriteString("\n")
	}
	sb.WriteString(fmt.Sprintf("... (%d lines omitted) ...\n", len(lines)-maxLines))
	for i := len(lines) - maxLines/2; i < len(lines); i++ {
		sb.WriteString(lines[i])
		sb.WriteString("\n")
	}
	return sb.String()
}

// setupLiveInterpreter creates the privileged interpreter instance.
func setupLiveInterpreter(t *testing.T, stdout *bytes.Buffer) *api.Interpreter {
	t.Helper()

	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("Skipping live test: GEMINI_API_KEY is not set.")
	}

	allowedTools := []string{
		"tool.os.Getenv",
		"tool.account.*",
		"tool.agentmodel.*",
		"tool.aeiou.ComposeEnvelope",
		"tool.aeiou.magic", // Allow the magic tool to be called by the LLM's code.
	}
	requiredGrants := []api.Capability{
		api.NewCapability(api.ResEnv, api.VerbRead, "GEMINI_API_KEY"),
		api.NewWithVerbs(api.ResModel, []string{api.VerbAdmin, api.VerbRead}, []string{"*"}),
		api.NewWithVerbs("account", []string{api.VerbAdmin, api.VerbRead}, []string{"*"}),
	}

	transcriptWriter := &aiTranscriptLogger{t: t}

	hostCtx, err := api.NewHostContextBuilder().
		WithLogger(logging.NewTestLogger(t)).
		WithStdout(stdout).
		WithStdin(os.Stdin).
		WithStderr(os.Stderr).
		Build()
	if err != nil {
		t.Fatalf("Failed to build host context: %v", err)
	}

	extraOpts := []api.Option{
		api.WithAITranscriptWriter(transcriptWriter),
		api.WithHostContext(hostCtx),
	}

	return api.NewConfigInterpreter(allowedTools, requiredGrants, extraOpts...)
}

// TestLive_OneShotQuery tests a simple, single-turn factual question.
func TestLive_OneShotQuery(t *testing.T) {
	var testOutput bytes.Buffer
	interp := setupLiveInterpreter(t, &testOutput)

	finalScript := oneShotScriptTemplate

	t.Logf("--- Assembled Script (Abbreviated) ---\n%s", abbreviate(finalScript, 20))

	// --- Execute the final, composed script ---
	tree, err := api.Parse([]byte(finalScript), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}
	if err := api.LoadFromUnit(interp, &api.LoadedUnit{Tree: tree}); err != nil {
		t.Fatalf("api.LoadFromUnit failed: %v", err)
	}

	// === STAGE 1: Run Setup ===
	if _, err := api.RunProcedure(context.Background(), interp, "setup"); err != nil {
		t.Logf("Setup script output on failure:\n%s", testOutput.String())
		t.Fatalf("Failed to run setup procedure: %v", err)
	}
	t.Logf("Setup procedure finished successfully.")
	t.Logf("Setup output:\n%s", testOutput.String())

	// === STAGE 2: Immediately Verify State ===
	testOutput.Reset()
	// NOTE: The function is named 'verify' in the script.
	if _, err := api.RunProcedure(context.Background(), interp, "verify"); err != nil {
		t.Logf("Verification script output on failure:\n%s", testOutput.String())
		t.Fatalf("Failed to run verify procedure: %v", err)
	}
	t.Logf("Verification output after setup:\n%s", testOutput.String())

	// === STAGE 3: Run Main Test Logic ===
	testOutput.Reset()
	result, err := api.RunProcedure(context.Background(), interp, "main")
	t.Logf("Main procedure output:\n%s", testOutput.String())
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
