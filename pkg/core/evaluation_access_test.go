// filename: pkg/core/evaluation_access_test.go
package core

import (
	"reflect"
	"strings"
	"testing"
	// Removed unused import: "github.com/aprice2704/neuroscript/pkg/core/token"
)

// TestEvaluateElementAccess - Tests specifically the element access logic
// Uses AST node definitions from ast.go in this package.

func TestEvaluateElementAccess(t *testing.T) {
	vars := map[string]interface{}{
		"listVar": []interface{}{"x", int64(99), "z", []interface{}{"nested"}},
		"mapVar": map[string]interface{}{
			"mKey":  "mVal",
			"mNum":  int64(1),
			"mList": []interface{}{"a"},
		},
		"idx":     int64(1),
		"key":     "mKey",
		"bad_idx": "one",
		"bad_key": int64(123),
		"nilVar":  nil,
		"name":    "World",
		"numVar":  int64(123),
	}
	interp, _ := NewDefaultTestInterpreter(t) // Use default interpreter from helpers
	for k, v := range vars {
		interp.SetVariable(k, v)
	}

	// Corrected: Use local Position type and make it a pointer
	dummyPos := &Position{Line: 1, Column: 1}

	tests := []struct {
		name        string
		inputNode   Expression // Change inputNode type to Expression for clarity where applicable
		expected    interface{}
		wantErr     bool
		errContains string
	}{
		// --- List Access ---
		{"List Access Valid Index 0", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, "x", false, ""},
		{"List Access Valid Index 1 (Num)", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(1)}}, int64(99), false, ""},
		{"List Access Valid Index Var", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &VariableNode{Pos: dummyPos, Name: "idx"}}, int64(99), false, ""},
		{"List Access Index Out of Bounds (High)", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(10)}}, nil, true, "list index 10 is out of bounds"},
		{"List Access Index Out of Bounds (Neg)", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(-1)}}, nil, true, "list index -1 is out of bounds"},
		{"List Access Invalid Index Type (String)", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &StringLiteralNode{Pos: dummyPos, Value: "one"}}, nil, true, "list index must evaluate to an integer"},
		{"List Access Invalid Index Type (Var)", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &VariableNode{Pos: dummyPos, Name: "bad_idx"}}, nil, true, "list index must evaluate to an integer"},
		{"List Access Returns List", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, []interface{}{"nested"}, false, ""},
		{"List Literal Access", &ElementAccessNode{Pos: dummyPos, Collection: &ListLiteralNode{Pos: dummyPos, Elements: []Expression{&StringLiteralNode{Pos: dummyPos, Value: "a"}, &NumberLiteralNode{Pos: dummyPos, Value: int64(5)}}}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, "a", false, ""},
		{"List Access Error in Collection", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "missing"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, nil, true, "evaluating collection for element access: variable not found: 'missing'"},
		{"List Access Error in Accessor", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &VariableNode{Pos: dummyPos, Name: "missing"}}, nil, true, "evaluating accessor for element access: variable not found: 'missing'"},
		{"List Access Collection Nil", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "nilVar"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, nil, true, "collection evaluated to nil"},
		{"List Access Accessor Nil", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, nil, true, "accessor evaluated to nil"},

		// --- Map Access ---
		{"Map Access Valid Key", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "mapVar"}, Accessor: &StringLiteralNode{Pos: dummyPos, Value: "mKey"}}, "mVal", false, ""},
		{"Map Access Valid Key Num", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "mapVar"}, Accessor: &StringLiteralNode{Pos: dummyPos, Value: "mNum"}}, int64(1), false, ""},
		{"Map Access Valid Key Var", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "mapVar"}, Accessor: &VariableNode{Pos: dummyPos, Name: "key"}}, "mVal", false, ""},
		{"Map Access Key Not Found", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "mapVar"}, Accessor: &StringLiteralNode{Pos: dummyPos, Value: "notFound"}}, nil, true, "key 'notFound' not found"},
		{"Map Access Invalid Key Type (Converted)", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "mapVar"}, Accessor: &VariableNode{Pos: dummyPos, Name: "bad_key"}}, nil, true, "key '123' not found"},
		{"Map Access Returns List", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "mapVar"}, Accessor: &StringLiteralNode{Pos: dummyPos, Value: "mList"}}, []interface{}{"a"}, false, ""},
		// *** CORRECTED: MapEntryNode Key needs VALUE, Value needs POINTER (Expression) ***
		{"Map Literal Access", &ElementAccessNode{Pos: dummyPos, Collection: &MapLiteralNode{Pos: dummyPos, Entries: []MapEntryNode{{Pos: dummyPos, Key: StringLiteralNode{Pos: dummyPos, Value: "k"}, Value: &StringLiteralNode{Pos: dummyPos, Value: "v"}}}}, Accessor: &StringLiteralNode{Pos: dummyPos, Value: "k"}}, "v", false, ""},
		{"Map Access Error in Collection", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "missing"}, Accessor: &StringLiteralNode{Pos: dummyPos, Value: "k"}}, nil, true, "evaluating collection for element access: variable not found: 'missing'"},
		{"Map Access Error in Accessor", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "mapVar"}, Accessor: &VariableNode{Pos: dummyPos, Name: "missing"}}, nil, true, "evaluating accessor for element access: variable not found: 'missing'"},
		{"Map Access Collection Nil", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "nilVar"}, Accessor: &StringLiteralNode{Pos: dummyPos, Value: "k"}}, nil, true, "collection evaluated to nil"},
		{"Map Access Accessor Nil", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "mapVar"}, Accessor: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, nil, true, "accessor evaluated to nil"},

		// --- Access on Invalid Types ---
		{"Access on String", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "name"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, nil, true, "cannot perform element access using [...] on type string"},
		{"Access on Number", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "numVar"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, nil, true, "cannot perform element access using [...] on type int64"},
		{"Access on Nil (Variable)", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "nilVar"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, nil, true, "collection evaluated to nil"},

		// --- Nested Access ---
		{"Nested List Access [3][0]", &ElementAccessNode{Pos: dummyPos, Collection: &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, "nested", false, ""},
		{"Nested Map List Access [\"mList\"][0]", &ElementAccessNode{Pos: dummyPos, Collection: &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "mapVar"}, Accessor: &StringLiteralNode{Pos: dummyPos, Value: "mList"}}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, "a", false, ""},
		{"Nested Access Error Outer List", &ElementAccessNode{Pos: dummyPos, Collection: &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(99)}}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, nil, true, "index 99 is out of bounds"},
		{"Nested Access Error Inner List", &ElementAccessNode{Pos: dummyPos, Collection: &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(10)}}, nil, true, "index 10 is out of bounds"},
		{"Nested Access Error Outer Map Key", &ElementAccessNode{Pos: dummyPos, Collection: &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "mapVar"}, Accessor: &StringLiteralNode{Pos: dummyPos, Value: "badKey"}}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, nil, true, "key 'badKey' not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Re-initialize variables for each subtest
			currentVars := make(map[string]interface{}, len(vars))
			for k, v := range vars {
				currentVars[k] = v
			}
			interp.variables = currentVars

			// Ensure InputNode is an Expression if it's not nil
			var inputExpr Expression
			if tt.inputNode != nil {
				var ok bool
				inputExpr, ok = tt.inputNode.(Expression)
				if !ok {
					t.Fatalf("Test setup error: InputNode (%T) does not implement Expression", tt.inputNode)
				}
			}

			got, err := interp.evaluateExpression(inputExpr) // Pass the asserted Expression

			if (err != nil) != tt.wantErr {
				t.Errorf("TestEvaluateElementAccess(%s) error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && (err == nil || !strings.Contains(err.Error(), tt.errContains)) {
					t.Errorf("TestEvaluateElementAccess(%s) expected error containing %q, got: %v", tt.name, tt.errContains, err)
				} else if err != nil {
					t.Logf("TestEvaluateElementAccess(%s) got expected error: %v", tt.name, err)
				}
			} else {
				if !reflect.DeepEqual(got, tt.expected) {
					t.Errorf("TestEvaluateElementAccess(%s)\nInput Node: %+v\nExpected:   %v (%T)\nGot:        %v (%T)",
						tt.name, tt.inputNode, tt.expected, tt.expected, got, got)
				}
			}
		})
	}
}

// Dummy Struct Definitions (Minimal to satisfy compilation in this file)
// Replace these with actual definitions if they are simple enough or keep assuming they exist.
// This is just to ensure the test file itself compiles based on usage.

// type Position struct { Line int; Column int } // Defined in ast.go
// type Expression interface { GetPos() *Position } // Defined in ast.go

// type VariableNode struct { Pos *Position; Name string }
// func (n *VariableNode) GetPos() *Position { return n.Pos }

// type NumberLiteralNode struct { Pos *Position; Value interface{} }
// func (n *NumberLiteralNode) GetPos() *Position { return n.Pos }

// type StringLiteralNode struct { Pos *Position; Value string; IsRaw bool }
// func (n *StringLiteralNode) GetPos() *Position { return n.Pos }

// type ListLiteralNode struct { Pos *Position; Elements []Expression }
// func (n *ListLiteralNode) GetPos() *Position { return n.Pos }

// type MapLiteralNode struct { Pos *Position; Entries []MapEntryNode }
// func (n *MapLiteralNode) GetPos() *Position { return n.Pos }

// type MapEntryNode struct { Pos *Position; Key StringLiteralNode; Value Expression }
// func (n *MapEntryNode) GetPos() *Position { return n.Key.Pos } // Use Key's Pos

// type ElementAccessNode struct { Pos *Position; Collection Expression; Accessor Expression }
// func (n *ElementAccessNode) GetPos() *Position { return n.Pos }

// Ensure all nodes above also implement Expression marker methods if needed by the interface definition
// func (n *VariableNode) expressionNode() {}
// ... etc for others ...
