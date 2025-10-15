// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Tests the interpreter's core I/O functionality, like emitting to stdout.
// filename: pkg/interpreter/io_test.go
// nlines: 53
// risk_rating: LOW

package interpreter_test

import (
	"bytes"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

func TestInterpreter_EmitToStdout(t *testing.T) {
	t.Run("emit statement writes to configured stdout", func(t *testing.T) {
		t.Logf("[DEBUG] Turn 1: Starting 'emit statement writes to configured stdout' test.")
		h := NewTestHarness(t)

		// 1. Create a buffer to capture the interpreter's standard output.
		var stdoutBuffer bytes.Buffer

		// 2. The harness's HostContext is mutable, so we replace its Stdout
		// to intercept any writes.
		h.HostContext.Stdout = &stdoutBuffer

		// 3. Set the custom EmitFunc to nil. This forces the interpreter to fall back
		// to its default behavior, which is to write to the Stdout stream.
		h.HostContext.EmitFunc = nil

		t.Logf("[DEBUG] Turn 2: Test harness I/O configured to capture stdout.")

		script := `
			command
				emit "hello, world"
			endcommand
		`
		interp := h.Interpreter

		// 4. Parse, build, load and execute the script.
		tree, pErr := h.Parser.Parse(script)
		if pErr != nil {
			t.Fatalf("Parser failed: %v", pErr)
		}
		program, _, bErr := h.ASTBuilder.Build(tree)
		if bErr != nil {
			t.Fatalf("AST Builder failed: %v", bErr)
		}

		if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
			t.Fatalf("Failed to load program: %v", err)
		}

		_, execErr := interp.Execute(program)
		if execErr != nil {
			t.Fatalf("Script execution failed unexpectedly: %v", execErr)
		}
		t.Logf("[DEBUG] Turn 3: Script executed.")

		// 5. Assert that the captured output matches what the script emitted.
		// The default emit via fmt.Fprintln adds a newline character.
		expectedOutput := "hello, world\n"
		if got := stdoutBuffer.String(); got != expectedOutput {
			t.Errorf("Stdout mismatch.\n  Got:      %q\n  Expected: %q", got, expectedOutput)
		}
		t.Logf("[DEBUG] Turn 4: Assertions passed.")
	})
}
