// filename: pkg/parser/ast_builder_nested_map_test.go
// NeuroScript Version: 0.6.3
// File version: 3
// Purpose: Updated assertions to type-cast MapEntryNode.Key (which is now Expression) to StringLiteralNode.
// nlines: 55
// risk_rating: LOW

package parser

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

// TestNestedMapInCallStatement reproduces a bug where a nested map literal, when
// used directly as an argument in a call statement, is not parsed correctly.
func TestNestedMapInCallStatement(t *testing.T) {
	script := `
        func main() means
            call tool.fdm.events.Emit({"type": "test.topic", "payload": {"message":"hello from lotfi"}})
        endfunc
    `
	prog := testParseAndBuild(t, script)
	proc := prog.Procedures["main"]
	if len(proc.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(proc.Steps))
	}
	step := proc.Steps[0]
	if step.Type != "call" {
		t.Fatalf("Expected a 'call' statement, but got '%s'", step.Type)
	}

	callExpr := step.Call
	if callExpr == nil {
		t.Fatal("Call expression is nil")
	}

	if len(callExpr.Arguments) != 1 {
		t.Fatalf("Expected 1 argument for the call, but got %d", len(callExpr.Arguments))
	}

	arg, ok := callExpr.Arguments[0].(*ast.MapLiteralNode)
	if !ok {
		t.Fatalf("Expected argument to be a MapLiteralNode, but got %T", callExpr.Arguments[0])
	}

	if len(arg.Entries) != 2 {
		t.Fatalf("Expected 2 entries in the outer map, but got %d", len(arg.Entries))
	}

	// Find the 'payload' entry
	var payloadValue ast.Expression
	for _, entry := range arg.Entries {
		// FIX: Cast Key to *ast.StringLiteralNode
		if keyLit, ok := entry.Key.(*ast.StringLiteralNode); ok && keyLit.Value == "payload" {
			payloadValue = entry.Value
			break
		}
	}

	if payloadValue == nil {
		t.Fatal("Could not find 'payload' entry in the map")
	}

	nestedMap, ok := payloadValue.(*ast.MapLiteralNode)
	if !ok {
		t.Fatalf("Expected 'payload' value to be a nested MapLiteralNode, but got %T", payloadValue)
	}

	if len(nestedMap.Entries) != 1 {
		t.Fatalf("Expected 1 entry in the nested map, but got %d", len(nestedMap.Entries))
	}

	// FIX: Cast Key to *ast.StringLiteralNode
	nestedKeyLit, ok := nestedMap.Entries[0].Key.(*ast.StringLiteralNode)
	if !ok {
		t.Fatalf("Expected nested map key to be *ast.StringLiteralNode, got %T", nestedMap.Entries[0].Key)
	}

	if nestedKeyLit.Value != "message" {
		t.Errorf("Expected nested map key to be 'message', but got '%s'", nestedKeyLit.Value)
	}
}
