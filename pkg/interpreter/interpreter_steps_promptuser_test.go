// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Contains unit tests for the 'promptuser' statement.
// filename: pkg/interpreter/interpreter_steps_promptuser_test.go
// nlines: 60
// risk_rating: LOW

package interpreter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestPromptUserStatement(t *testing.T) {
	t.Run("promptuser into variable", func(t *testing.T) {
		// Mock user input
		userInput := "yes, it is a test\n"
		stdin := strings.NewReader(userInput)
		var stdout bytes.Buffer

		interp, _ := newLocalTestInterpreter(t, nil, nil)
		interp.SetStdin(stdin)
		interp.SetStdout(&stdout)

		step := ast.Step{
			Type: "promptuser",
			PromptUserStmt: &ast.PromptUserStmt{
				PromptExpr: &ast.StringLiteralNode{Value: "Is this a test?"},
				IntoTarget: &ast.LValueNode{Identifier: "user_response"},
			},
		}

		// This assumes the AST team has added PromptUserStmt to ast.Step
		// If not, this test will fail to compile until that change is made.
		if step.PromptUserStmt == nil {
			t.Skip("Skipping test: ast.Step does not yet have PromptUserStmt field.")
		}

		_, err := interp.executePromptUser(step)
		if err != nil {
			t.Fatalf("executePromptUser failed: %v", err)
		}

		// Check prompt was written to stdout
		expectedPrompt := "Is this a test? "
		if stdout.String() != expectedPrompt {
			t.Errorf("Expected prompt '%s', got '%s'", expectedPrompt, stdout.String())
		}

		// Check variable was set
		resultVar, exists := interp.GetVariable("user_response")
		if !exists {
			t.Fatal("Variable 'user_response' was not set by the promptuser statement")
		}
		resultStr, _ := lang.ToString(resultVar)
		if resultStr != strings.TrimSpace(userInput) {
			t.Errorf("Expected result variable to be '%s', got '%s'", strings.TrimSpace(userInput), resultStr)
		}
	})
}
