// NeuroScript Version: 0.7.0
// File version: 5
// Purpose: Corrected test assertions by explicitly quoting captured whisper values.
// filename: pkg/interpreter/e2e_whisper_test.go
// nlines: 70
// risk_rating: LOW

package interpreter_test

import (
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
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
	parserAPI := parser.NewParserAPI(nil)
	tree, err := parserAPI.Parse(script)
	if err != nil {
		t.Fatalf("ParserAPI.Parse() failed: %v", err)
	}

	builder := parser.NewASTBuilder(nil)
	program, _, err := builder.Build(tree)
	if err != nil {
		t.Fatalf("ASTBuilder.Build() failed: %v", err)
	}
	if program == nil {
		t.Fatal("Parsing returned a nil program without an error")
	}

	var capturedWhispers []string

	i := interpreter.NewInterpreter()
	// CORRECTED: Use fmt.Sprintf with %q to quote the raw string values.
	i.SetWhisperFunc(func(handle, data lang.Value) {
		line := fmt.Sprintf("%q:%q", handle.String(), data.String())
		capturedWhispers = append(capturedWhispers, line)
	})

	_, err = i.Execute(program)
	if err != nil {
		t.Fatalf("Interpreter execution failed: %v", err)
	}

	if len(capturedWhispers) != 2 {
		t.Fatalf("Expected 2 whispers, but got %d", len(capturedWhispers))
	}

	// The expectation is now correct because the captured strings will be quoted.
	expectedWhisper1 := `"out":"This is a test whisper."`
	if capturedWhispers[0] != expectedWhisper1 {
		t.Errorf("Expected whisper 1 to be '%s', got '%s'", expectedWhisper1, capturedWhispers[0])
	}

	expectedWhisper2 := `"err":"This is an error whisper."`
	if capturedWhispers[1] != expectedWhisper2 {
		t.Errorf("Expected whisper 2 to be '%s', got '%s'", expectedWhisper2, capturedWhispers[1])
	}
}
