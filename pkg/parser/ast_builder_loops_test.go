// filename: pkg/parser/ast_builder_loops_test.go
package parser

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

// Helper to get the first step, assuming it's a loop
func getLoopStep(t *testing.T, script string) *ast.Step {
	t.Helper()
	// Use the existing test helper to parse the script and get the procedure body
	bodyNodes := parseStringToProcedureBodyNodes(t, script, "TestProc")
	if len(bodyNodes) == 0 {
		t.Fatalf("TestProc body is empty, expected a loop step")
	}
	loopStep := &bodyNodes[0]
	if loopStep.Type != "for" && loopStep.Type != "while" {
		t.Fatalf("Expected first step to be of type 'for' or 'while', got type %s", loopStep.Type)
	}
	return loopStep
}

func TestForLoop(t *testing.T) {
	script := `
func TestProc() means
    for each item in my_list
        emit item
    endfor
endfunc
`
	forStep := getLoopStep(t, script)

	if forStep.Type != "for" {
		t.Fatalf("Expected step type 'for', got '%s'", forStep.Type)
	}
	if forStep.LoopVarName != "item" {
		t.Errorf("Expected LoopVarName to be 'item', got '%s'", forStep.LoopVarName)
	}
	if _, ok := forStep.Collection.(*ast.VariableNode); !ok {
		t.Errorf("Expected loop collection to be a VariableNode, got %T", forStep.Collection)
	}
	if len(forStep.Body) != 1 {
		t.Errorf("Expected 1 statement in loop body, got %d", len(forStep.Body))
	}
}