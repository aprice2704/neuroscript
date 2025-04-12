// filename: neuroscript/pkg/core/evaluation_access_test.go
package core

import (
	"reflect"
	"strings"
	"testing"
)

// TestEvaluateElementAccess - Tests specifically the element access logic
func TestEvaluateElementAccess(t *testing.T) {
	vars := map[string]interface{}{
		// Use a nested map for the map test case
		"listVar": []interface{}{"x", int64(99), "z", []interface{}{"nested"}},
		"mapVar": map[string]interface{}{
			"mKey":  "mVal",
			"mNum":  int64(1),
			"mList": []interface{}{"a"}, // The list is directly under mList
		},
		"idx":     int64(1),
		"key":     "mKey",
		"bad_idx": "one",
		"bad_key": int64(123),
		"nilVar":  nil,
		"name":    "World",
		"numVar":  int64(123),
	}
	// *** FIXED: Use newTestInterpreter from test scope ***
	interp, _ := newTestInterpreter(t, vars, nil) // Get interpreter and ignore sandbox path

	tests := []struct {
		name        string
		inputNode   interface{} // Should be ElementAccessNode
		expected    interface{}
		wantErr     bool
		errContains string
	}{
		// --- List Access ---
		{"List Access Valid Index 0", ElementAccessNode{Collection: VariableNode{Name: "listVar"}, Accessor: NumberLiteralNode{Value: int64(0)}}, "x", false, ""},
		{"List Access Valid Index 1 (Num)", ElementAccessNode{Collection: VariableNode{Name: "listVar"}, Accessor: NumberLiteralNode{Value: int64(1)}}, int64(99), false, ""},
		{"List Access Valid Index Var", ElementAccessNode{Collection: VariableNode{Name: "listVar"}, Accessor: VariableNode{Name: "idx"}}, int64(99), false, ""},
		{"List Access Index Out of Bounds (High)", ElementAccessNode{Collection: VariableNode{Name: "listVar"}, Accessor: NumberLiteralNode{Value: int64(10)}}, nil, true, "index 10 is out of bounds"},
		{"List Access Index Out of Bounds (Neg)", ElementAccessNode{Collection: VariableNode{Name: "listVar"}, Accessor: NumberLiteralNode{Value: int64(-1)}}, nil, true, "index -1 is out of bounds"},
		{"List Access Invalid Index Type (String)", ElementAccessNode{Collection: VariableNode{Name: "listVar"}, Accessor: StringLiteralNode{Value: "one"}}, nil, true, "list index must evaluate to an integer"},
		{"List Access Invalid Index Type (Var)", ElementAccessNode{Collection: VariableNode{Name: "listVar"}, Accessor: VariableNode{Name: "bad_idx"}}, nil, true, "list index must evaluate to an integer"},
		{"List Access Returns List", ElementAccessNode{Collection: VariableNode{Name: "listVar"}, Accessor: NumberLiteralNode{Value: int64(3)}}, []interface{}{"nested"}, false, ""},
		{"List Literal Access", ElementAccessNode{Collection: ListLiteralNode{Elements: []interface{}{StringLiteralNode{Value: "a"}, NumberLiteralNode{Value: int64(5)}}}, Accessor: NumberLiteralNode{Value: int64(0)}}, "a", false, ""},
		{"List Access Error in Collection", ElementAccessNode{Collection: VariableNode{Name: "missing"}, Accessor: NumberLiteralNode{Value: int64(0)}}, nil, true, "variable 'missing' not found"},
		{"List Access Error in Accessor", ElementAccessNode{Collection: VariableNode{Name: "listVar"}, Accessor: VariableNode{Name: "missing"}}, nil, true, "variable 'missing' not found"},
		{"List Access Collection Nil", ElementAccessNode{Collection: VariableNode{Name: "nilVar"}, Accessor: NumberLiteralNode{Value: int64(0)}}, nil, true, "collection evaluated to nil"},
		{"List Access Accessor Nil", ElementAccessNode{Collection: VariableNode{Name: "listVar"}, Accessor: VariableNode{Name: "nilVar"}}, nil, true, "accessor evaluated to nil"},

		// --- Map Access ---
		{"Map Access Valid Key", ElementAccessNode{Collection: VariableNode{Name: "mapVar"}, Accessor: StringLiteralNode{Value: "mKey"}}, "mVal", false, ""},
		{"Map Access Valid Key Num", ElementAccessNode{Collection: VariableNode{Name: "mapVar"}, Accessor: StringLiteralNode{Value: "mNum"}}, int64(1), false, ""},
		{"Map Access Valid Key Var", ElementAccessNode{Collection: VariableNode{Name: "mapVar"}, Accessor: VariableNode{Name: "key"}}, "mVal", false, ""},
		{"Map Access Key Not Found", ElementAccessNode{Collection: VariableNode{Name: "mapVar"}, Accessor: StringLiteralNode{Value: "notFound"}}, nil, true, "key 'notFound' not found"},
		{"Map Access Invalid Key Type (Converted)", ElementAccessNode{Collection: VariableNode{Name: "mapVar"}, Accessor: VariableNode{Name: "bad_key"}}, nil, true, "key '123' not found"},
		{"Map Access Returns List", ElementAccessNode{Collection: VariableNode{Name: "mapVar"}, Accessor: StringLiteralNode{Value: "mList"}}, []interface{}{"a"}, false, ""},
		{"Map Literal Access", ElementAccessNode{Collection: MapLiteralNode{Entries: []MapEntryNode{{Key: StringLiteralNode{Value: "k"}, Value: StringLiteralNode{Value: "v"}}}}, Accessor: StringLiteralNode{Value: "k"}}, "v", false, ""},
		{"Map Access Error in Collection", ElementAccessNode{Collection: VariableNode{Name: "missing"}, Accessor: StringLiteralNode{Value: "k"}}, nil, true, "variable 'missing' not found"},
		{"Map Access Error in Accessor", ElementAccessNode{Collection: VariableNode{Name: "mapVar"}, Accessor: VariableNode{Name: "missing"}}, nil, true, "variable 'missing' not found"},
		{"Map Access Collection Nil", ElementAccessNode{Collection: VariableNode{Name: "nilVar"}, Accessor: StringLiteralNode{Value: "k"}}, nil, true, "collection evaluated to nil"},
		{"Map Access Accessor Nil", ElementAccessNode{Collection: VariableNode{Name: "mapVar"}, Accessor: VariableNode{Name: "nilVar"}}, nil, true, "accessor evaluated to nil"},

		// --- Access on Invalid Types ---
		{"Access on String", ElementAccessNode{Collection: VariableNode{Name: "name"}, Accessor: NumberLiteralNode{Value: int64(0)}}, nil, true, "cannot perform element access using [...] on type string"},
		{"Access on Number", ElementAccessNode{Collection: VariableNode{Name: "numVar"}, Accessor: NumberLiteralNode{Value: int64(0)}}, nil, true, "cannot perform element access using [...] on type int64"},
		{"Access on Nil (Variable)", ElementAccessNode{Collection: VariableNode{Name: "nilVar"}, Accessor: NumberLiteralNode{Value: int64(0)}}, nil, true, "collection evaluated to nil"},

		// --- Nested Access ---
		{"Nested List Access [3][0]", ElementAccessNode{Collection: ElementAccessNode{Collection: VariableNode{Name: "listVar"}, Accessor: NumberLiteralNode{Value: int64(3)}}, Accessor: NumberLiteralNode{Value: int64(0)}}, "nested", false, ""},
		{"Nested Map List Access [\"mList\"][0]", ElementAccessNode{Collection: ElementAccessNode{Collection: VariableNode{Name: "mapVar"}, Accessor: StringLiteralNode{Value: "mList"}}, Accessor: NumberLiteralNode{Value: int64(0)}}, "a", false, ""},
		{"Nested Access Error Outer List", ElementAccessNode{Collection: ElementAccessNode{Collection: VariableNode{Name: "listVar"}, Accessor: NumberLiteralNode{Value: int64(99)}}, Accessor: NumberLiteralNode{Value: int64(0)}}, nil, true, "index 99 is out of bounds"},
		{"Nested Access Error Inner List", ElementAccessNode{Collection: ElementAccessNode{Collection: VariableNode{Name: "listVar"}, Accessor: NumberLiteralNode{Value: int64(3)}}, Accessor: NumberLiteralNode{Value: int64(10)}}, nil, true, "index 10 is out of bounds"},
		{"Nested Access Error Outer Map Key", ElementAccessNode{Collection: ElementAccessNode{Collection: VariableNode{Name: "mapVar"}, Accessor: StringLiteralNode{Value: "badKey"}}, Accessor: NumberLiteralNode{Value: int64(0)}}, nil, true, "key 'badKey' not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// interp is already initialized with vars above
			interp.variables = make(map[string]interface{}, len(vars))
			for k, v := range vars {
				interp.variables[k] = v
			}

			got, err := interp.evaluateExpression(tt.inputNode)

			if (err != nil) != tt.wantErr {
				t.Errorf("TestEvaluateElementAccess(%s) error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && (err == nil || !strings.Contains(err.Error(), tt.errContains)) {
					coreErrorFound := false
					if err != nil {
						if strings.Contains(err.Error(), tt.errContains) {
							coreErrorFound = true
						}
					}
					if !coreErrorFound {
						t.Errorf("TestEvaluateElementAccess(%s) expected error containing %q, got: %v", tt.name, tt.errContains, err)
					}
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
