// NeuroScript Version: 0.6.0
// File version: 3
// Purpose: Refactored to use the centralized TestHarness for robust and consistent interpreter initialization.
// filename: pkg/interpreter/interpreter_steps_promptuser_test.go
// nlines: 65
// risk_rating: LOW

package interpreter_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
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

		step := ast.Step{
			Type: "promptuser",
			PromptUserStmt: &ast.PromptUserStmt{
				PromptExpr: &ast.StringLiteralNode{Value: "Is this a test?"},
				IntoTarget: &ast.LValueNode{Identifier: "user_response"},
			},
		}

		if step.PromptUserStmt == nil {
			t.Skip("Skipping test: ast.Step does not yet have PromptUserStmt field.")
		}

		// This test calls an unexported method, so we need to use a build tag to make it accessible
		// or refactor the test to call a public method. For now, we will assume it is accessible.
		// _, err := interp.executePromptUser(step)
		// For now, we will simulate the execution via a script
		script := `
			func main() means
				promptuser "Is this a test?" into user_response
			endfunc
		`
		_, err := interp.ExecuteScriptString("main", script, nil)

		if err != nil {
			t.Fatalf("executePromptUser failed: %v", err)
		}
		t.Logf("[DEBUG] Turn 3: Script executed.")

		expectedPrompt := "Is this a test? "
		if stdout.String() != expectedPrompt {
			t.Errorf("Expected prompt '%s', got '%s'", expectedPrompt, stdout.String())
		}

		resultVar, exists := interp.GetVariable("user_response")
		if !exists {
			t.Fatal("Variable 'user_response' was not set by the promptuser statement")
		}
		resultStr, _ := lang.ToString(resultVar)
		if resultStr != strings.TrimSpace(userInput) {
			t.Errorf("Expected result variable to be '%s', got '%s'", strings.TrimSpace(userInput), resultStr)
		}
		t.Logf("[DEBUG] Turn 4: Assertions passed.")
	})
}
