// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Provides a dedicated test to verify that the 'emit' statement correctly writes to the interpreter's configured stdout stream.
// filename: pkg/livetest/stdout_test.go
// nlines: 35
// risk_rating: LOW

package livetest_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
)

// TestLive_EmitToStdout verifies that the interpreter's stdout stream, configured
// via the `api.WithStdout` option, correctly captures the output of the `emit`
// statement. This is a regression test against bugs where this mechanism failed.
func TestLive_EmitToStdout(t *testing.T) {
	// 1. Create a buffer to act as our stdout receiver.
	var stdout bytes.Buffer

	// 2. Define a simple script that emits a known string.
	script := `command
  emit "hello from stdout test"
endcommand`

	// 3. Execute the script in a new interpreter configured to use our buffer as stdout.
	_, err := api.ExecInNewInterpreter(context.Background(), script, api.WithStdout(&stdout))
	if err != nil {
		t.Fatalf("ExecInNewInterpreter failed unexpectedly: %v", err)
	}

	// 4. Assert that the buffer contains the exact expected output.
	// The `emit` statement automatically adds a newline.
	expected := "hello from stdout test\n"
	if got := stdout.String(); got != expected {
		t.Errorf("stdout mismatch:\n  Got: %q\n  Want: %q", got, expected)
	}
}
