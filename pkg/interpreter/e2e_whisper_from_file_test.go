// NeuroScript Version: 0.7.1
// File version: 4
// Purpose: Added a Stdout writer to the HostContext to prevent panic during 'emit'.
// filename: pkg/interpreter/e2e_whisper_from_file_test.go
// nlines: 90
// risk_rating: LOW

package interpreter_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

func TestE2E_WhisperCommandFromFile(t *testing.T) {
	// Load the script from the canonical test file.
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

	t.Logf("[DEBUG] Setting up HostContext and initial globals")

	// 1. Setup the HostContext with all required I/O components.
	hostCtx := &interpreter.HostContext{
		Logger: logging.NewTestLogger(t),
		Stdout: &bytes.Buffer{}, // This is the required fix.
		WhisperFunc: func(handle, data lang.Value) {
			line := fmt.Sprintf("%q:%q", handle.String(), data.String())
			t.Logf("[DEBUG] WhisperFunc captured: %s", line)
			capturedWhispers = append(capturedWhispers, line)
		},
	}

	// 2. Setup the initial global variables.
	globals := map[string]interface{}{
		"stdout": "stdout_handle",
		"stderr": "stderr_handle",
	}

	// 3. Create the interpreter with the new options.
	i := interpreter.NewInterpreter(
		interpreter.WithHostContext(hostCtx),
		interpreter.WithGlobals(globals),
	)

	// 4. Load and execute the program using the post-refactor API.
	if err := i.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Interpreter load failed: %v", err)
	}

	_, err = i.Execute(program)
	if err != nil {
		t.Fatalf("Interpreter execution failed: %v", err)
	}

	t.Logf("[DEBUG] Execution complete. Captured %d whispers.", len(capturedWhispers))

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
