// filename: pkg/parser/ast_builder_ask_promptuser_test.go
// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Provides dedicated tests for the new AI 'ask' statement and the renamed 'promptuser' statement.
// nlines: 100
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
		if len(step.Values) != 2 {
			t.Errorf("Expected 2 values (agent, prompt), got %d", len(step.Values))
		}
		if len(step.LValues) != 0 {
			t.Errorf("Expected 0 LValues, got %d", len(step.LValues))
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
		if len(step.LValues) != 1 {
			t.Errorf("Expected 1 LValue, got %d", len(step.LValues))
		}
		if step.LValues[0].Identifier != "result" {
			t.Errorf("Expected LValue identifier 'result', got '%s'", step.LValues[0].Identifier)
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
		if len(step.Values) != 3 {
			t.Errorf("Expected 3 values (agent, prompt, with), got %d", len(step.Values))
		}
		if _, ok := step.Values[2].(*ast.MapLiteralNode); !ok {
			t.Errorf("Expected third value to be a MapLiteralNode, got %T", step.Values[2])
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
		if len(step.Values) != 3 {
			t.Errorf("Expected 3 values, got %d", len(step.Values))
		}
		if len(step.LValues) != 1 {
			t.Errorf("Expected 1 LValue, got %d", len(step.LValues))
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
		if step.Type != "promptuser" {
			t.Errorf("Expected step type 'promptuser', got '%s'", step.Type)
		}
		if len(step.Values) != 1 {
			t.Errorf("Expected 1 value (prompt), got %d", len(step.Values))
		}
		if len(step.LValues) != 1 {
			t.Errorf("Expected 1 LValue, got %d", len(step.LValues))
		}
		if step.LValues[0].Identifier != "user_name" {
			t.Errorf("Expected LValue identifier 'user_name', got '%s'", step.LValues[0].Identifier)
		}
	})
}
