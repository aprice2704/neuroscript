// filename: pkg/parser/ast_builder_positions_test.go
// NeuroScript Version: 0.6.0
// File version: 2
// Purpose: Corrected test by splitting the script into valid library and command scripts to respect grammar rules.
// nlines: 130
// risk_rating: LOW

package parser

import (
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// checkNodePositions is a helper that asserts a node has valid start and end positions.
func checkNodePositions(t *testing.T, node ast.Node, name string) {
	t.Helper()
	if node == nil {
		t.Errorf("%s: node is nil", name)
		return
	}

	startPos := node.GetPos()
	endPos := node.End()

	if startPos == nil {
		t.Errorf("%s: StartPos (GetPos()) is nil", name)
	}
	if endPos == nil {
		t.Errorf("%s: StopPos (End()) is nil", name)
	}

	if startPos != nil && endPos != nil {
		if endPos.Line < startPos.Line {
			t.Errorf("%s: StopPos line (%d) is before StartPos line (%d)", name, endPos.Line, startPos.Line)
		}
		if endPos.Line == startPos.Line && endPos.Column < startPos.Column {
			t.Errorf("%s: StopPos column (%d) is before StartPos column (%d) on the same line", name, endPos.Column, startPos.Column)
		}
	}
}

// checkStepPositions recursively checks all steps within a slice.
func checkStepPositions(t *testing.T, steps []ast.Step, prefix string) {
	t.Helper()
	for i, step := range steps {
		stepName := fmt.Sprintf("%s.Step[%d](%s)", prefix, i, step.Type)
		checkNodePositions(t, &step, stepName)

		if len(step.Body) > 0 {
			checkStepPositions(t, step.Body, stepName+".Body")
		}
		if len(step.ElseBody) > 0 {
			checkStepPositions(t, step.ElseBody, stepName+".ElseBody")
		}
	}
}

func TestNodePositions(t *testing.T) {
	logger := logging.NewTestLogger(t)
	parserAPI := NewParserAPI(logger)

	t.Run("Library Script Positions", func(t *testing.T) {
		script := `
:: file: pos_test_lib.ns

on event "test.event" do
	emit "event"
endon

func main() means
	set y = 10
	if y > 5
		emit "greater"
	else
		emit "less or equal"
	endif

	for each i in [1,2,3]
		emit i
	endfor

	while y > 0
		set y = y - 1
	endwhile

	on error do
		emit "error in func"
	endon
endfunc
`
		tree, err := parserAPI.Parse(script)
		if err != nil {
			t.Fatalf("Parse() failed unexpectedly: %v", err)
		}

		builder := NewASTBuilder(logger)
		program, _, err := builder.Build(tree)
		if err != nil {
			t.Fatalf("Build() failed unexpectedly: %v", err)
		}

		// Check OnEventDecl
		if len(program.Events) != 1 {
			t.Fatalf("Expected 1 OnEventDecl, got %d", len(program.Events))
		}
		eventDecl := program.Events[0]
		checkNodePositions(t, eventDecl, "OnEventDecl")
		checkStepPositions(t, eventDecl.Body, "OnEventDecl")

		// Check Procedure and its contents
		proc, ok := program.Procedures["main"]
		if !ok {
			t.Fatal("Procedure 'main' not found")
		}
		checkNodePositions(t, proc, "Procedure:main")
		checkStepPositions(t, proc.Steps, "Procedure:main")
		if len(proc.ErrorHandlers) != 1 {
			t.Fatalf("Expected 1 error handler in procedure, got %d", len(proc.ErrorHandlers))
		}
		checkNodePositions(t, proc.ErrorHandlers[0], "Procedure:main.ErrorHandler")
		checkStepPositions(t, proc.ErrorHandlers[0].Body, "Procedure:main.ErrorHandler.Body")
	})

	t.Run("Command Script Positions", func(t *testing.T) {
		script := `
command
	set x = 1
	on error do
		emit "error in command"
	endon
endcommand
`
		tree, err := parserAPI.Parse(script)
		if err != nil {
			t.Fatalf("Parse() failed unexpectedly: %v", err)
		}

		builder := NewASTBuilder(logger)
		program, _, err := builder.Build(tree)
		if err != nil {
			t.Fatalf("Build() failed unexpectedly: %v", err)
		}

		// Check CommandNode
		if len(program.Commands) != 1 {
			t.Fatalf("Expected 1 CommandNode, got %d", len(program.Commands))
		}
		cmd := program.Commands[0]
		checkNodePositions(t, cmd, "CommandNode")
		checkStepPositions(t, cmd.Body, "CommandNode.Body")
		if len(cmd.ErrorHandlers) != 1 {
			t.Fatalf("Expected 1 error handler in command, got %d", len(cmd.ErrorHandlers))
		}
		checkNodePositions(t, cmd.ErrorHandlers[0], "CommandNode.ErrorHandler")
		checkStepPositions(t, cmd.ErrorHandlers[0].Body, "CommandNode.ErrorHandler.Body")
	})
}
