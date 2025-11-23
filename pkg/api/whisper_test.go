// NeuroScript Version: 0.8.0
// File version: 4
// Purpose: Corrects type assertion in TestInterpreter_WhisperFunc from *lang.MapValue to lang.MapValue. Updates handle creation to use the new HandleRegistry API.
// filename: pkg/api/whisper_test.go
// nlines: 92
// risk_rating: MEDIUM

package api_test

import (
	"context"
	"fmt" // DEBUG
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// TestInterpreter_WhisperFunc verifies that the 'whisper' statement correctly
// invokes a custom WhisperFunc with the right handle and payload.
func TestInterpreter_WhisperFunc(t *testing.T) {
	script := `command
    set h = tool.test.make_handle("my_object")
    whisper h, {"status": "updated"}
endcommand`
	var capturedHandle, capturedData api.Value
	var wg sync.WaitGroup
	wg.Add(1)

	hc, err := api.NewHostContextBuilder().
		WithLogger(logging.NewNoOpLogger()).
		WithStdout(io.Discard).
		WithStdin(os.Stdin).
		WithStderr(io.Discard).
		WithWhisperFunc(func(handle, data api.Value) {
			capturedHandle = handle
			capturedData = data
			wg.Done()
		}).
		Build()
	if err != nil {
		t.Fatalf("Failed to build host context: %v", err)
	}

	const testHandleKind = "test"

	// A simple tool to create a handle for testing purposes
	handleTool := api.ToolImplementation{
		Spec: api.ToolSpec{
			Name:  "make_handle",
			Group: "test",
			Args:  []api.ArgSpec{{Name: "value", Type: "any"}},
		},
		Func: func(rt api.Runtime, args []any) (any, error) {
			// FIX: Use HandleRegistry().NewHandle(...)
			handleValue, err := rt.HandleRegistry().NewHandle(args[0], testHandleKind)
			if err != nil {
				return nil, err
			}
			// FIX: Return the HandleValue, not just the ID string.
			// The original code returned the ID string, which was wrapped as lang.StringValue
			// and checked later. We must return the HandleValue itself to be wrapped as the
			// new lang.HandleValue, but the test's next step uses it as a string.
			// Reverting to the string return to preserve the original test logic check.
			return handleValue.HandleID(), nil
		},
	}

	policy := api.NewPolicyBuilder(api.ContextNormal).Allow("tool.test.make_handle").Build()
	interp := api.New(
		api.WithHostContext(hc),
		api.WithExecPolicy(policy),
	)
	interp.ToolRegistry().RegisterTool(handleTool)

	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse() failed: %v", err)
	}
	_, err = api.ExecWithInterpreter(context.Background(), interp, tree)
	if err != nil {
		t.Fatalf("ExecWithInterpreter() failed unexpectedly: %v", err)
	}

	wg.Wait()

	if capturedHandle == nil {
		t.Fatal("WhisperFunc was not called")
	}

	// Verify handle
	// The NS `set h = ...` wraps the string ID as a StringValue because `make_handle` returns a string.
	handleStr, ok := capturedHandle.(lang.StringValue)
	if !ok {
		t.Fatalf("Expected handle to be a string, got %T", capturedHandle)
	}
	if !strings.HasPrefix(handleStr.Value, testHandleKind+"-") {
		// Note: The previous handle implementation used "test::", the new mock uses "test-".
		// We update the check to match the new mock helper's internal format "kind-id".
		t.Errorf("Expected handle to have '%s-' prefix, got %s", testHandleKind, handleStr.Value)
	}

	// DEBUG: Log the type of the captured data
	fmt.Fprintf(os.Stderr, "--- DEBUG: TestInterpreter_WhisperFunc: capturedData type is %T ---\n", capturedData)

	// Verify data
	// FIX: The interpreter now wraps maps as lang.MapValue (value), not *lang.MapValue (pointer).
	dataMap, ok := capturedData.(lang.MapValue)
	if !ok {
		t.Fatalf("Expected data to be a lang.MapValue, got %T", capturedData)
	}
	statusVal := dataMap.Value["status"]
	statusStr, _ := statusVal.(lang.StringValue)
	if statusStr.Value != "updated" {
		t.Errorf("Expected data status to be 'updated', got %s", statusStr.Value)
	}
}
