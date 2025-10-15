// filename: pkg/parser/ast_builder_ask_promptuser_test.go
// NeuroScript Version: 0.6.0
// File version: 4
// Purpose: Corrected all test assertions to use the canonical ast.Step.AskStmt field and its sub-fields, resolving compiler errors and panics.
// nlines: 132
// risk_rating: MEDIUM

package parser

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

func TestAskStatementParsing(t *testing.T) {
	t.Run("simple ask statement", func(t *testing.T) {
		script := `
			func main() means
				ask "default-agent", "what is the time?"
			endfunc
		`
		prog := testParseAndBuild(t, script)
		proc := prog.Procedures["main"]
		if len(proc.Steps) != 1 {
			t.Fatalf("Expected 1 step, got %d", len(proc.Steps))
		}
		step := proc.Steps[0]
		if step.Type != "ask" {
			t.Errorf("Expected step type 'ask', got '%s'", step.Type)
		}
		if step.AskStmt == nil {
			t.Fatal("Expected step.AskStmt to be a non-nil AskStmt")
		}
		if step.AskStmt.AgentModelExpr == nil {
			t.Error("Expected AgentModelExpr to be non-nil")
		}
		if step.AskStmt.PromptExpr == nil {
			t.Error("Expected PromptExpr to be non-nil")
		}
		if step.AskStmt.IntoTarget != nil {
			t.Error("Expected IntoTarget to be nil for simple ask")
		}
	})

	t.Run("ask statement with into clause", func(t *testing.T) {
		script := `
			func main() means
				ask "default-agent", "what is the time?" into result
			endfunc
		`
		prog := testParseAndBuild(t, script)
		proc := prog.Procedures["main"]
		step := proc.Steps[0]

		if step.AskStmt == nil {
			t.Fatal("Expected step.AskStmt to be a non-nil AskStmt")
		}
		if step.AskStmt.IntoTarget == nil {
			t.Fatal("Expected IntoTarget to be non-nil")
		}
		if step.AskStmt.IntoTarget.Identifier != "result" {
			t.Errorf("Expected LValue identifier 'result', got '%s'", step.AskStmt.IntoTarget.Identifier)
		}
	})

	t.Run("ask statement with with clause", func(t *testing.T) {
		script := `
			func main() means
				ask "agent", "prompt" with {"temp": 0.8}
			endfunc
		`
		prog := testParseAndBuild(t, script)
		proc := prog.Procedures["main"]
		step := proc.Steps[0]
		if step.AskStmt == nil {
			t.Fatal("Expected step.AskStmt to be a non-nil AskStmt")
		}
		if step.AskStmt.WithOptions == nil {
			t.Fatal("Expected WithOptions expression to be non-nil")
		}
		if _, ok := step.AskStmt.WithOptions.(*ast.MapLiteralNode); !ok {
			t.Errorf("Expected WithOptions to be a MapLiteralNode, got %T", step.AskStmt.WithOptions)
		}
	})

	t.Run("ask statement with with and into clauses", func(t *testing.T) {
		script := `
			func main() means
				ask "agent", "prompt" with {"temp": 0.8} into my_var
			endfunc
		`
		prog := testParseAndBuild(t, script)
		proc := prog.Procedures["main"]
		step := proc.Steps[0]
		if step.AskStmt == nil {
			t.Fatal("Expected step.AskStmt to be a non-nil AskStmt")
		}
		if step.AskStmt.WithOptions == nil {
			t.Error("Expected WithOptions to be non-nil")
		}
		if step.AskStmt.IntoTarget == nil {
			t.Error("Expected IntoTarget to be non-nil")
		}
	})
}

func TestPromptUserStatementParsing(t *testing.T) {
	t.Run("simple promptuser statement", func(t *testing.T) {
		script := `
			func main() means
				promptuser "Enter your name" into user_name
			endfunc
		`
		prog := testParseAndBuild(t, script)
		proc := prog.Procedures["main"]
		if len(proc.Steps) != 1 {
			t.Fatalf("Expected 1 step, got %d", len(proc.Steps))
		}
		step := proc.Steps[0]

		// FIX: Correct the test to check the modern, structured ast.PromptUserStmt
		// instead of the deprecated generic `Values` and `LValues` fields.
		if step.Type != "promptuser" {
			t.Errorf("Expected step type 'promptuser', got '%s'", step.Type)
		}
		if step.PromptUserStmt == nil {
			t.Fatal("Expected step.PromptUserStmt to be a non-nil PromptUserStmt")
		}
		if step.PromptUserStmt.PromptExpr == nil {
			t.Error("Expected PromptExpr to be non-nil")
		}
		if step.PromptUserStmt.IntoTarget == nil {
			t.Fatal("Expected IntoTarget to be non-nil")
		}
		if step.PromptUserStmt.IntoTarget.Identifier != "user_name" {
			t.Errorf("Expected LValue identifier 'user_name', got '%s'", step.PromptUserStmt.IntoTarget.Identifier)
		}
	})
}
