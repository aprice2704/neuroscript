// filename: pkg/parser/ast_builder_loops_test.go
// NeuroScript Version: 0.6.1
// File version: 4
// Purpose: Added a specific test to ensure the 'Collection' field in a for-each loop is never nil in the AST, preventing a runtime panic.

package parser

import (
	"testing"
)

func TestWhileLoopParsing(t *testing.T) {
	t.Run("Valid while loop", func(t *testing.T) {
		script := `
			func MyFunc() means
				while x < 10
					set x = x + 1
				endwhile
			endfunc
		`
		prog := testParseAndBuild(t, script)
		proc := prog.Procedures["MyFunc"]
		if len(proc.Steps) != 1 {
			t.Fatalf("Expected 1 statement in procedure body, got %d", len(proc.Steps))
		}
		loop := proc.Steps[0]
		if loop.Type != "while" {
			t.Fatalf("Expected a 'while' loop step, but got type %s", loop.Type)
		}
		if loop.Cond == nil {
			t.Error("Expected while loop to have a condition")
		}
		if len(loop.Body) != 1 {
			t.Errorf("Expected 1 statement in while loop body, got %d", len(loop.Body))
		}
	})

	t.Run("While loop with empty body is a parser error", func(t *testing.T) {
		script := `
			func MyFunc() means
				while x < 10
				endwhile
			endfunc
		`
		testForParserError(t, script)
	})
}

func TestForEachLoopParsing(t *testing.T) {
	t.Run("Valid for-each loop", func(t *testing.T) {
		script := `
			func MyFunc() means
				for each item in myList
					emit item
				endfor
			endfunc
		`
		prog := testParseAndBuild(t, script)
		proc := prog.Procedures["MyFunc"]
		if len(proc.Steps) != 1 {
			t.Fatalf("Expected 1 statement in procedure body, got %d", len(proc.Steps))
		}
		loop := proc.Steps[0]
		if loop.Type != "for" {
			t.Fatalf("Expected a 'for' loop step, but got type %s", loop.Type)
		}
		if loop.LoopVarName != "item" {
			t.Errorf("Expected loop variable to be 'item', got '%s'", loop.LoopVarName)
		}
		if loop.Collection == nil {
			t.Error("BUG CHECK: Expected for-each loop to have a non-nil collection expression")
		}
		if len(loop.Body) != 1 {
			t.Errorf("Expected 1 statement in for-each loop body, got %d", len(loop.Body))
		}
	})

	t.Run("For-each loop with empty body is a parser error", func(t *testing.T) {
		script := `
			func MyFunc() means
				for each item in myList
				endfor
			endfunc
		`
		testForParserError(t, script)
	})
}
