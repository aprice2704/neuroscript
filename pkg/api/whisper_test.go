// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Provides a focused test for the 'whisper' statement and its corresponding HostContext callback.
// filename: pkg/api/whisper_test.go
// nlines: 83
// risk_rating: MEDIUM

package api_test

import (
	"context"
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

	// A simple tool to create a handle for testing purposes
	handleTool := api.ToolImplementation{
		Spec: api.ToolSpec{
			Name:  "make_handle",
			Group: "test",
			Args:  []api.ArgSpec{{Name: "value", Type: "any"}},
		},
		Func: func(rt api.Runtime, args []any) (any, error) {
			return rt.RegisterHandle(args[0], "test")
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
	handleStr, ok := capturedHandle.(lang.StringValue)
	if !ok {
		t.Fatalf("Expected handle to be a string, got %T", capturedHandle)
	}
	if !strings.HasPrefix(handleStr.Value, "test::") {
		t.Errorf("Expected handle to have 'test::' prefix, got %s", handleStr.Value)
	}

	// Verify data
	dataMap, ok := capturedData.(*lang.MapValue)
	if !ok {
		t.Fatalf("Expected data to be a *lang.MapValue, got %T", capturedData)
	}
	statusVal := dataMap.Value["status"]
	statusStr, _ := statusVal.(lang.StringValue)
	if statusStr.Value != "updated" {
		t.Errorf("Expected data status to be 'updated', got %s", statusStr.Value)
	}
}
