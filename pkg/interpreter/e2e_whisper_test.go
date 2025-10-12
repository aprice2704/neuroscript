// NeuroScript Version: 0.8.0
// File version: 8
// Purpose: Corrected test setup to provide a valid HostContext with a logger and I/O streams.
// filename: pkg/interpreter/e2e_whisper_test.go
// nlines: 85
// risk_rating: LOW

package interpreter_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

func TestE2E_WhisperCommand(t *testing.T) {
	script := `
// filename: whisper_feature.ns
command
    set stdout = "out"
    set stderr = "err"
    whisper stdout, "This is a test whisper."
    whisper stderr, "This is an error whisper."
endcommand
`
	// DEBUG: Add printf to show test start
	t.Logf("[DEBUG] Starting TestE2E_WhisperCommand")

	parserAPI := parser.NewParserAPI(logging.NewTestLogger(t))
	tree, err := parserAPI.Parse(script)
	if err != nil {
		t.Fatalf("ParserAPI.Parse() failed: %v", err)
	}

	builder := parser.NewASTBuilder(logging.NewTestLogger(t))
	program, _, err := builder.Build(tree)
	if err != nil {
		t.Fatalf("ASTBuilder.Build() failed: %v", err)
	}
	if program == nil {
		t.Fatal("Parsing returned a nil program without an error")
	}

	var capturedWhispers []string

	t.Logf("[DEBUG] Creating HostContext and setting WhisperFunc")

	// 1. Create a HostContext with the whisper callback and all mandatory fields.
	hostCtx := &interpreter.HostContext{
		Logger: logging.NewTestLogger(t),
		Stdout: os.Stdout,
		Stdin:  os.Stdin,
		Stderr: os.Stderr,
		WhisperFunc: func(handle, data lang.Value) {
			line := fmt.Sprintf("%q:%q", handle.String(), data.String())
			// DEBUG: Log each time the whisper function is called
			t.Logf("[DEBUG] WhisperFunc called. Captured line: %s", line)
			capturedWhispers = append(capturedWhispers, line)
		},
	}

	// 2. Instantiate the interpreter with the HostContext.
	i := interpreter.NewInterpreter(interpreter.WithHostContext(hostCtx))

	// 3. Load the program, wrapping it in the required interfaces.Tree struct.
	if err := i.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Interpreter load failed: %v", err)
	}

	// 4. Execute the loaded program using the unified Execute method.
	_, err = i.Execute(program)
	if err != nil {
		t.Fatalf("Interpreter execution failed: %v", err)
	}

	// DEBUG: Log the final state of the captured whispers before assertion
	t.Logf("[DEBUG] Execution finished. Total whispers captured: %d", len(capturedWhispers))
	for idx, w := range capturedWhispers {
		t.Logf("[DEBUG] Captured[%d]: %s", idx, w)
	}

	if len(capturedWhispers) != 2 {
		t.Fatalf("Expected 2 whispers, but got %d", len(capturedWhispers))
	}

	expectedWhisper1 := `"out":"This is a test whisper."`
	if capturedWhispers[0] != expectedWhisper1 {
		t.Errorf("Expected whisper 1 to be '%s', got '%s'", expectedWhisper1, capturedWhispers[0])
	}

	expectedWhisper2 := `"err":"This is an error whisper."`
	if capturedWhispers[1] != expectedWhisper2 {
		t.Errorf("Expected whisper 2 to be '%s', got '%s'", expectedWhisper2, capturedWhispers[1])
	}
}
