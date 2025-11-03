// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Tests 'ask' hook. Fixes mockAeiouService signature to match the interface's (any, error) return.
// filename: pkg/api/exec_hook_test.go
// nlines: 120

package api_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// mockAeiouService implements the AeiouOrchestrator interface for testing.
// It uses 'any' for the interpreter arg to match the cycle-breaking signature
// in pkg/interfaces/aeiou.go.
type mockAeiouService struct {
	t              *testing.T
	expectedAgent  string
	expectedPrompt string
	valueToReturn  lang.Value // This can still be lang.Value
	called         bool
	receivedInterp any
	receivedAgent  string
	receivedPrompt string
}

// FIX: The signature now returns (any, error) to match the interface.
func (m *mockAeiouService) RunAskLoop(
	callingInterp any,
	agentModelName string,
	initialPrompt string,
) (any, error) { // <<< WAS (lang.Value, error)
	m.called = true
	m.receivedInterp = callingInterp
	m.receivedAgent = agentModelName
	m.receivedPrompt = initialPrompt

	// Check that the passed interpreter is the *api.Interpreter
	if _, ok := callingInterp.(*api.Interpreter); !ok {
		m.t.Errorf("RunAskLoop was called with wrong interpreter type. Got %T, Want *api.Interpreter", callingInterp)
	}

	if agentModelName != m.expectedAgent {
		m.t.Errorf("RunAskLoop agent mismatch. Got '%s', Want '%s'", agentModelName, m.expectedAgent)
	}
	if initialPrompt != m.expectedPrompt {
		m.t.Errorf("RunAskLoop prompt mismatch. Got '%s', Want '%s'", initialPrompt, m.expectedPrompt)
	}

	// We return the specific lang.Value, which satisfies the 'any' return type.
	return m.valueToReturn, nil
}

// TestAsk_ServiceHook verifies that the 'ask' statement correctly calls
// an AeiouOrchestrator service provided via the HostContext.
func TestAsk_ServiceHook(t *testing.T) {
	// --- ARRANGE ---
	var stdout bytes.Buffer
	expectedAgent := "test_agent_hook"
	expectedPrompt := "hello from test"
	expectedReturn := "hello from mock service"

	// 1. Create the mock service.
	mockService := &mockAeiouService{
		t:              t,
		expectedAgent:  expectedAgent,
		expectedPrompt: expectedPrompt,
		valueToReturn:  lang.StringValue{Value: expectedReturn},
	}

	// 2. Create the service registry containing the mock.
	registry := map[string]any{
		interfaces.AeiouServiceKey: mockService,
	}

	// 3. Create a HostContext and inject the registry.
	hc, err := api.NewHostContextBuilder().
		WithLogger(&mockLogger{}). // Assumes mockLogger is in another _test.go file
		WithStdout(&stdout).
		WithStdin(os.Stdin).
		WithStderr(os.Stderr).
		WithServiceRegistry(registry). // <-- Inject the service registry
		Build()
	if err != nil {
		t.Fatalf("Failed to build host context: %v", err)
	}

	// 4. Create the interpreter.
	interp := api.New(api.WithHostContext(hc))

	// 5. Define the script that triggers the hook.
	script := `
	command
		ask "test_agent_hook", "hello from test" into result
		emit result
	endcommand
	`

	// --- ACT ---
	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	_, execErr := api.ExecWithInterpreter(context.Background(), interp, tree)

	// --- ASSERT ---
	if execErr != nil {
		t.Fatalf("ExecWithInterpreter failed unexpectedly: %v", execErr)
	}

	// Check that the mock service was actually called
	if !mockService.called {
		t.Fatal("The mock AeiouOrchestrator service was never called.")
	}

	// Check that the 'emit' statement output the mock's return value
	output := strings.TrimSpace(stdout.String())
	if output != expectedReturn {
		t.Errorf("Stdout mismatch:\n  Got: %q\n  Want: %q", output, expectedReturn)
	}

	t.Log("SUCCESS: The 'ask' statement correctly hooked into the mock service.")
}

// TestAsk_ServiceHook_WrongType_Fallback verifies that the 'ask' statement
// gracefully falls back to the legacy internal loop if the registered
// service is of the wrong type, and does not panic.
func TestAsk_ServiceHook_WrongType_Fallback(t *testing.T) {
	// --- ARRANGE ---
	var stdout bytes.Buffer

	// 1. Create a misconfigured registry.
	registry := map[string]any{
		interfaces.AeiouServiceKey: "i-am-a-string-not-a-service", // <-- Wrong type
	}

	// 2. Create a HostContext and inject the bad registry.
	hc, err := api.NewHostContextBuilder().
		WithLogger(&mockLogger{}).
		WithStdout(&stdout).
		WithStdin(os.Stdin).
		WithStderr(os.Stderr).
		WithServiceRegistry(registry).
		Build()
	if err != nil {
		t.Fatalf("Failed to build host context: %v", err)
	}

	// 3. Create the interpreter.
	interp := api.New(api.WithHostContext(hc))

	// 4. Define a script that will fail if the *legacy* path is taken.
	script := `
	command
		# This agent does not exist, so the legacy path will fail.
		# If the hook code panicked, this would never be reached.
		ask "no_such_agent_exists", "test prompt" into result
	endcommand
	`

	// --- ACT ---
	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	_, execErr := api.ExecWithInterpreter(context.Background(), interp, tree)

	// --- ASSERT ---
	if execErr == nil {
		t.Fatal("ExecWithInterpreter succeeded, but was expected to fail.")
	}

	// We expect a specific *runtime* error from the legacy path,
	// *not* a Go panic from the hook path.
	var rtErr *api.RuntimeError
	if !errors.As(execErr, &rtErr) {
		t.Fatalf("Expected a *lang.RuntimeError, but got a different error (or panic): %T: %v", execErr, execErr)
	}

	expectedMsg := "AgentModel 'no_such_agent_exists' is not registered"
	if !strings.Contains(rtErr.Error(), expectedMsg) {
		t.Errorf("Error message mismatch:\n  Got: %v\n  Want: %v", rtErr.Error(), expectedMsg)
	}

	t.Logf("SUCCESS: 'ask' hook correctly fell back to legacy path and returned expected error: %v", execErr)
}
