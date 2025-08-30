// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: End-to-end test for the 'whisper' command, loading from a dedicated script file.
// filename: pkg/interpreter/e2e_whisper_from_file_test.go
// nlines: 70
// risk_rating: LOW

package interpreter_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

func TestE2E_WhisperCommandFromFile(t *testing.T) {
	// Load the script from the canonical test file.
	// Note: Adjust the relative path if your test setup is different.
	scriptPath := filepath.Join("..", "antlr", "whisper_feature.ns")
	scriptBytes, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("Failed to read test script '%s': %v", scriptPath, err)
	}
	script := string(scriptBytes)

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
	i.SetWhisperFunc(func(handle, data lang.Value) {
		// Quote the values to match the expected format.
		line := fmt.Sprintf("%q:%q", handle.String(), data.String())
		capturedWhispers = append(capturedWhispers, line)
	})

	// Manually set the required variables for this test script.
	i.SetInitialVariable("stdout", "stdout_handle")
	i.SetInitialVariable("stderr", "stderr_handle")

	_, err = i.Execute(program)
	if err != nil {
		t.Fatalf("Interpreter execution failed: %v", err)
	}

	if len(capturedWhispers) != 2 {
		t.Fatalf("Expected 2 whispers, but got %d", len(capturedWhispers))
	}

	expectedWhisper1 := `"stdout_handle":"This message is whispered to the primary output channel."`
	if capturedWhispers[0] != expectedWhisper1 {
		t.Errorf("Expected whisper 1 to be '%s', got '%s'", expectedWhisper1, capturedWhispers[0])
	}

	expectedWhisper2 := `"stderr_handle":"System is nominal."`
	if capturedWhispers[1] != expectedWhisper2 {
		t.Errorf("Expected whisper 2 to be '%s', got '%s'", expectedWhisper2, capturedWhispers[1])
	}
}
