// NeuroScript/FDM Major Version: 1
// File version: 6
// Purpose: Tests 'ask' hook. Fixes mockAeiouService to fully implement interfaces.AeiouOrchestrator. Removes fallback test as panic is expected behavior for invalid types.
// filename: pkg/api/exec_hook_test.go
// nlines: 105

package api_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// mockAeiouService implements the AeiouOrchestrator interface for testing.
type mockAeiouService struct {
	t              *testing.T
	expectedAgent  string
	expectedPrompt string
	valueToReturn  lang.Value
	called         bool
	receivedInterp any
	receivedAgent  string
	receivedPrompt string
}

// RunAskLoop matches the signature in interfaces.AeiouOrchestrator.
func (m *mockAeiouService) RunAskLoop(
	callingInterp any,
	agentModelName string,
	initialPrompt string,
) (any, error) {
	m.called = true
	m.receivedInterp = callingInterp
	m.receivedAgent = agentModelName
	m.receivedPrompt = initialPrompt

	if _, ok := callingInterp.(*api.Interpreter); !ok {
		m.t.Errorf("RunAskLoop was called with wrong interpreter type. Got %T, Want *api.Interpreter", callingInterp)
	}

	if agentModelName != m.expectedAgent {
		m.t.Errorf("RunAskLoop agent mismatch. Got '%s', Want '%s'", agentModelName, m.expectedAgent)
	}
	if initialPrompt != m.expectedPrompt {
		m.t.Errorf("RunAskLoop prompt mismatch. Got '%s', Want '%s'", initialPrompt, m.expectedPrompt)
	}

	return m.valueToReturn, nil
}

// ListActiveLoops satisfies the AeiouOrchestrator interface.
func (m *mockAeiouService) ListActiveLoops() []interfaces.ActiveLoopInfo {
	return []interfaces.ActiveLoopInfo{}
}

// CancelLoop satisfies the AeiouOrchestrator interface.
func (m *mockAeiouService) CancelLoop(loopID string) error {
	return nil
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
		WithLogger(&mockLogger{}).
		WithStdout(&stdout).
		WithStdin(os.Stdin).
		WithStderr(os.Stderr).
		WithServiceRegistry(registry).
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

	if !mockService.called {
		t.Fatal("The mock AeiouOrchestrator service was never called.")
	}

	output := strings.TrimSpace(stdout.String())
	if output != expectedReturn {
		t.Errorf("Stdout mismatch:\n  Got: %q\n  Want: %q", output, expectedReturn)
	}

	t.Log("SUCCESS: The 'ask' statement correctly hooked into the mock service.")
}
