// NeuroScript Version: 0.7.0
// File version: 4
// Purpose: Fixed the test by abandoning the `ExecInNewInterpreter` helper and instead manually constructing the interpreter with `api.New` and `WithHostContext`. This ensures the stdout buffer is correctly wired.
// filename: pkg/livetest/stdout_test.go
// nlines: 54
// risk_rating: LOW

package livetest_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// TestLive_EmitToStdout verifies that the interpreter's stdout stream, configured
// via the `api.WithHostContext` option, correctly captures the output of the `emit`
// statement. This is a regression test against bugs where this mechanism failed.
func TestLive_EmitToStdout(t *testing.T) {
	// 1. Create a buffer to act as our stdout receiver.
	var stdout bytes.Buffer

	// 2. Define a simple script that emits a known string.
	script := `command
  emit "hello from stdout test"
endcommand`

	// 3. Configure the HostContext using the official builder.
	hostCtx, err := api.NewHostContextBuilder().
		WithLogger(logging.NewTestLogger(t)).
		WithStdout(&stdout).
		WithStdin(os.Stdin).
		WithStderr(os.Stderr).
		Build()
	if err != nil {
		t.Fatalf("Failed to build host context: %v", err)
	}

	// 4. Manually construct the interpreter and execute the script.
	// This avoids using `ExecInNewInterpreter`, which might not pass the context correctly.
	interp := api.New(api.WithHostContext(hostCtx))
	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}

	if _, err := api.ExecWithInterpreter(context.Background(), interp, tree); err != nil {
		t.Fatalf("ExecWithInterpreter failed unexpectedly: %v", err)
	}

	// 5. Assert that the buffer contains the exact expected output.
	// The `emit` statement automatically adds a newline.
	expected := "hello from stdout test\n"
	if got := stdout.String(); got != expected {
		t.Errorf("stdout mismatch:\n  Got: %q\n  Want: %q", got, expected)
	}
}
