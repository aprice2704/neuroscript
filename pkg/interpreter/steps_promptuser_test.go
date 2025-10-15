// NeuroScript Version: 0.8.0
// File version: 7
// Purpose: Corrected test to check the procedure's return value instead of a variable in a discarded scope.
// filename: pkg/interpreter/interpreter_steps_promptuser_test.go
// nlines: 63
// risk_rating: LOW

package interpreter_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestPromptUserStatement(t *testing.T) {
	t.Run("promptuser into variable", func(t *testing.T) {
		t.Logf("[DEBUG] Turn 1: Starting 'promptuser into variable' test.")
		h := NewTestHarness(t)
		interp := h.Interpreter

		userInput := "yes, it is a test\n"
		// Configure the harness's Stdin and Stdout for this test
		h.HostContext.Stdin = strings.NewReader(userInput)
		var stdout bytes.Buffer
		h.HostContext.Stdout = &stdout
		t.Logf("[DEBUG] Turn 2: Test harness I/O configured.")

		// The test script now returns the variable. This allows us to test the result
		// without making assumptions about variable scope after the procedure has finished.
		script := `
			func main() means
				promptuser "Is this a test?" into user_response
				return user_response
			endfunc
		`
		// 1. Parse the script string into a parse tree.
		parseTree, parseErr := h.Parser.Parse(script)
		if parseErr != nil {
			t.Fatalf("Failed to parse script: %v", parseErr)
		}

		// 2. Build the parse tree into an Abstract Syntax Tree (AST).
		programAST, _, buildErr := h.ASTBuilder.Build(parseTree)
		if buildErr != nil {
			t.Fatalf("Failed to build AST from parse tree: %v", buildErr)
		}

		// 3. Load the AST into the interpreter.
		if err := interp.Load(&interfaces.Tree{Root: programAST}); err != nil {
			t.Fatalf("Failed to load AST into interpreter: %v", err)
		}

		// 4. Run the 'main' procedure and capture its return value.
		resultVal, err := interp.Run("main")
		if err != nil {
			t.Fatalf("interp.Run(\"main\") failed: %v", err)
		}
		t.Logf("[DEBUG] Turn 3: Script executed.")

		expectedPrompt := "Is this a test? "
		if stdout.String() != expectedPrompt {
			t.Errorf("Expected prompt '%s', got '%s'", expectedPrompt, stdout.String())
		}

		// 5. Assert against the returned value.
		if resultVal == nil {
			t.Fatal("Expected a return value from the script, but got nil")
		}
		resultStr, _ := lang.ToString(resultVal)
		if resultStr != strings.TrimSpace(userInput) {
			t.Errorf("Expected result variable to be '%s', got '%s'", strings.TrimSpace(userInput), resultStr)
		}
		t.Logf("[DEBUG] Turn 4: Assertions passed.")
	})
}
