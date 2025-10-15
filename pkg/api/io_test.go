// NeuroScript Version: 0.8.0
// File version: 10
// Purpose: Provides focused tests for default and custom IO behaviors, including emit and a correctly-behaved custom tool.
// filename: pkg/api/io_test.go
// nlines: 109
// risk_rating: MEDIUM

package api_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// TestInterpreter_DefaultIO verifies that both 'emit' and a well-behaved tool
// call write to the configured Stdout stream by default.
func TestInterpreter_DefaultIO(t *testing.T) {
	script := `command
    emit "message from emit"
    call tool.test.print("message from tool")
endcommand`

	var stdout bytes.Buffer
	hc, err := api.NewHostContextBuilder().
		WithLogger(logging.NewNoOpLogger()).
		WithStdout(&stdout).
		WithStdin(os.Stdin).
		WithStderr(io.Discard).
		Build()
	if err != nil {
		t.Fatalf("Failed to build host context: %v", err)
	}

	// A test-local tool that correctly uses the runtime's Println method.
	testPrintTool := api.ToolImplementation{
		Spec: api.ToolSpec{Name: "print", Group: "test", Args: []api.ArgSpec{{Name: "msg", Type: "string"}}},
		Func: func(rt api.Runtime, args []any) (any, error) {
			rt.Println(args...)
			return nil, nil
		},
	}

	policy := api.NewPolicyBuilder(api.ContextNormal).Allow("tool.test.print").Build()
	interp := api.New(
		api.WithHostContext(hc),
		api.WithExecPolicy(policy),
	)
	interp.ToolRegistry().RegisterTool(testPrintTool)

	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse() failed: %v", err)
	}

	_, err = api.ExecWithInterpreter(context.Background(), interp, tree)
	if err != nil {
		t.Fatalf("ExecWithInterpreter() failed unexpectedly: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "message from emit") {
		t.Errorf("Expected stdout to contain 'message from emit', but it didn't. Got: %q", output)
	}
	if !strings.Contains(output, "message from tool") {
		t.Errorf("Expected stdout to contain 'message from tool', but it didn't. Got: %q", output)
	}
}

// TestInterpreter_CustomEmitFuncWithTool verifies that a custom EmitFunc is
// correctly called while a well-behaved tool's output still goes to stdout.
func TestInterpreter_CustomEmitFuncWithTool(t *testing.T) {
	script := `command
    emit "custom message"
    call tool.test.print("stdout message")
endcommand`

	var stdout bytes.Buffer
	var capturedEmitValue api.Value
	var wg sync.WaitGroup
	wg.Add(1)

	hc, err := api.NewHostContextBuilder().
		WithLogger(logging.NewNoOpLogger()).
		WithStdout(&stdout).
		WithStdin(os.Stdin).
		WithStderr(io.Discard).
		WithEmitFunc(func(v api.Value) {
			capturedEmitValue = v
			wg.Done()
		}).
		Build()
	if err != nil {
		t.Fatalf("Failed to build host context: %v", err)
	}

	testPrintTool := api.ToolImplementation{
		Spec: api.ToolSpec{Name: "print", Group: "test", Args: []api.ArgSpec{{Name: "msg", Type: "string"}}},
		Func: func(rt api.Runtime, args []any) (any, error) {
			rt.Println(args...)
			return nil, nil
		},
	}

	policy := api.NewPolicyBuilder(api.ContextNormal).Allow("tool.test.print").Build()
	interp := api.New(
		api.WithHostContext(hc),
		api.WithExecPolicy(policy),
	)
	interp.ToolRegistry().RegisterTool(testPrintTool)

	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse() failed: %v", err)
	}

	_, err = api.ExecWithInterpreter(context.Background(), interp, tree)
	if err != nil {
		t.Fatalf("ExecWithInterpreter() failed unexpectedly: %v", err)
	}

	wg.Wait()

	if capturedEmitValue == nil {
		t.Fatal("Custom EmitFunc was not called")
	}
	unwrapped, _ := api.Unwrap(capturedEmitValue)
	if unwrapped != "custom message" {
		t.Errorf("Expected captured emit value to be 'custom message', but got %q", unwrapped)
	}
	output := stdout.String()
	if strings.Contains(output, "custom message") {
		t.Errorf("Expected stdout to NOT contain 'custom message', but it did.")
	}

	if !strings.Contains(output, "stdout message") {
		t.Errorf("Expected stdout to contain 'stdout message', but it didn't. Got: %q", output)
	}
}
