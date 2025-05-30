// NeuroScript Version: 0.3.5
// File version: 0.0.3 // Switched to sentinel error checking (errors.Is) and cleared errContains for those cases.
// filename: pkg/core/evaluation_access_test.go
package core

import (
	"errors" // Required for errors.Is in the test helper, and for defining sentinel vars if needed locally (not here)
	"reflect"
	"strings" // Still used by the test runner for errContains if ExpectedErrorIs is nil
	"testing"
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
	// Use default interpreter from helpers; NewDefaultTestInterpreter is defined in testing_helpers.go
	interp, _ := NewDefaultTestInterpreter(t)
	for k, v := range vars {
		interp.SetVariable(k, v)
	}

	dummyPos := &Position{Line: 1, Column: 1, File: "test_access.go"}

	tests := []struct {
		name            string
		inputNode       Expression
		expected        interface{}
		wantErr         bool
		ExpectedErrorIs error  // USE THIS for sentinel checks
		errContains     string // Fallback or for errors without specific sentinels
	}{
		// --- List Access ---
		{"List Access Valid Index 0", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, "x", false, nil, ""},
		{"List Access Valid Index 1 (Num)", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(1)}}, int64(99), false, nil, ""},
		{"List Access Valid Index Var", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &VariableNode{Pos: dummyPos, Name: "idx"}}, int64(99), false, nil, ""},
		{"List Access Index Out of Bounds (High)", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(10)}}, nil, true, ErrListIndexOutOfBounds, ""},
		{"List Access Index Out of Bounds (Neg)", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(-1)}}, nil, true, ErrListIndexOutOfBounds, ""},
		{"List Access Invalid Index Type (String)", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &StringLiteralNode{Pos: dummyPos, Value: "one"}}, nil, true, ErrListInvalidIndexType, ""},
		{"List Access Invalid Index Type (Var)", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &VariableNode{Pos: dummyPos, Name: "bad_idx"}}, nil, true, ErrListInvalidIndexType, ""},
		{"List Access Returns List", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, []interface{}{"nested"}, false, nil, ""},
		{"List Literal Access", &ElementAccessNode{Pos: dummyPos, Collection: &ListLiteralNode{Pos: dummyPos, Elements: []Expression{&StringLiteralNode{Pos: dummyPos, Value: "a"}, &NumberLiteralNode{Pos: dummyPos, Value: int64(5)}}}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, "a", false, nil, ""},
		{"List Access Error in Collection", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "missing"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, nil, true, ErrVariableNotFound, ""}, // MODIFIED
		{"List Access Error in Accessor", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &VariableNode{Pos: dummyPos, Name: "missing"}}, nil, true, ErrVariableNotFound, ""},        // MODIFIED
		{"List Access Collection Nil", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "nilVar"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, nil, true, ErrCollectionIsNil, ""},
		{"List Access Accessor Nil", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, nil, true, ErrAccessorIsNil, ""},

		// --- Map Access ---
		{"Map Access Valid Key", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "mapVar"}, Accessor: &StringLiteralNode{Pos: dummyPos, Value: "mKey"}}, "mVal", false, nil, ""},
		{"Map Access Valid Key Num", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "mapVar"}, Accessor: &StringLiteralNode{Pos: dummyPos, Value: "mNum"}}, int64(1), false, nil, ""},
		{"Map Access Valid Key Var", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "mapVar"}, Accessor: &VariableNode{Pos: dummyPos, Name: "key"}}, "mVal", false, nil, ""},
		{"Map Access Key Not Found", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "mapVar"}, Accessor: &StringLiteralNode{Pos: dummyPos, Value: "notFound"}}, nil, true, ErrMapKeyNotFound, ""},
		{"Map Access Invalid Key Type (Converted to string)", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "mapVar"}, Accessor: &VariableNode{Pos: dummyPos, Name: "bad_key"}}, nil, true, ErrMapKeyNotFound, ""}, // Expecting key "123" not found
		{"Map Access Returns List", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "mapVar"}, Accessor: &StringLiteralNode{Pos: dummyPos, Value: "mList"}}, []interface{}{"a"}, false, nil, ""},
		{"Map Literal Access", &ElementAccessNode{Pos: dummyPos, Collection: &MapLiteralNode{Pos: dummyPos, Entries: []*MapEntryNode{{Pos: dummyPos, Key: &StringLiteralNode{Pos: dummyPos, Value: "k"}, Value: &StringLiteralNode{Pos: dummyPos, Value: "v"}}}}, Accessor: &StringLiteralNode{Pos: dummyPos, Value: "k"}}, "v", false, nil, ""},
		{"Map Access Error in Collection", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "missing"}, Accessor: &StringLiteralNode{Pos: dummyPos, Value: "k"}}, nil, true, ErrVariableNotFound, ""}, // MODIFIED
		{"Map Access Error in Accessor", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "mapVar"}, Accessor: &VariableNode{Pos: dummyPos, Name: "missing"}}, nil, true, ErrVariableNotFound, ""},    // MODIFIED
		{"Map Access Collection Nil", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "nilVar"}, Accessor: &StringLiteralNode{Pos: dummyPos, Value: "k"}}, nil, true, ErrCollectionIsNil, ""},
		{"Map Access Accessor Nil", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "mapVar"}, Accessor: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, nil, true, ErrAccessorIsNil, ""},

		// --- Access on Invalid Types ---
		{"Access on String", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "name"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, nil, true, ErrCannotAccessType, ""},
		{"Access on Number", &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "numVar"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, nil, true, ErrCannotAccessType, ""},
		// Note: "Access on Nil (Variable)" is covered by "List Access Collection Nil" and "Map Access Collection Nil" using ErrCollectionIsNil

		// --- Nested Access ---
		{"Nested List Access [3][0]", &ElementAccessNode{Pos: dummyPos, Collection: &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, "nested", false, nil, ""},
		{"Nested Map List Access [\"mList\"][0]", &ElementAccessNode{Pos: dummyPos, Collection: &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "mapVar"}, Accessor: &StringLiteralNode{Pos: dummyPos, Value: "mList"}}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, "a", false, nil, ""},
		{"Nested Access Error Outer List", &ElementAccessNode{Pos: dummyPos, Collection: &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(99)}}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, nil, true, ErrListIndexOutOfBounds, ""}, // Error from outer access
		{"Nested Access Error Inner List", &ElementAccessNode{Pos: dummyPos, Collection: &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "listVar"}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(10)}}, nil, true, ErrListIndexOutOfBounds, ""}, // Error from inner access
		{"Nested Access Error Outer Map Key", &ElementAccessNode{Pos: dummyPos, Collection: &ElementAccessNode{Pos: dummyPos, Collection: &VariableNode{Pos: dummyPos, Name: "mapVar"}, Accessor: &StringLiteralNode{Pos: dummyPos, Value: "badKey"}}, Accessor: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, nil, true, ErrMapKeyNotFound, ""},      // Error from outer access
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Re-initialize variables for each subtest to ensure test isolation
			// This uses the SetVariable method of the interpreter, which should handle logging if implemented.
			// NewDefaultTestInterpreter already creates a clean variable set.
			// We re-populate it with the 'vars' specific to this test function.
			cleanInterp, _ := NewDefaultTestInterpreter(t) // Get a fresh interpreter for isolation
			for k, v := range vars {
				if err := cleanInterp.SetVariable(k, v); err != nil {
					t.Fatalf("Failed to set initial variable %s for test %s: %v", k, tt.name, err)
				}
			}
			// Use this freshly populated interpreter for the current test case
			currentInterp := cleanInterp

			var inputExpr Expression
			if tt.inputNode != nil {
				var ok bool
				inputExpr, ok = tt.inputNode.(Expression)
				if !ok {
					t.Fatalf("Test setup error: InputNode (%T) does not implement Expression", tt.inputNode)
				}
			}

			got, err := currentInterp.evaluateExpression(inputExpr)

			if (err != nil) != tt.wantErr {
				t.Errorf("TestEvaluateElementAccess(%s) error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.ExpectedErrorIs != nil {
					if !errors.Is(err, tt.ExpectedErrorIs) {
						t.Errorf("TestEvaluateElementAccess(%s) expected error type [%v], but got [%v] (type %T)", tt.name, tt.ExpectedErrorIs, err, err)
					} else {
						t.Logf("TestEvaluateElementAccess(%s) got expected error type [%v]: %v", tt.name, tt.ExpectedErrorIs, err)
					}
				} else if tt.errContains != "" { // Fallback to errContains if no specific sentinel is expected
					if err == nil || !strings.Contains(err.Error(), tt.errContains) {
						t.Errorf("TestEvaluateElementAccess(%s) expected error containing %q, got: %v", tt.name, tt.errContains, err)
					} else {
						t.Logf("TestEvaluateElementAccess(%s) got expected error containing %q: %v", tt.name, tt.errContains, err)
					}
				} else if err != nil { // wantErr is true, but no specific check, just log the error
					t.Logf("TestEvaluateElementAccess(%s) got expected error: %v", tt.name, err)
				}
			} else { // No error wanted
				if !reflect.DeepEqual(got, tt.expected) {
					t.Errorf("TestEvaluateElementAccess(%s)\nInput Node: %+v\nExpected:   %#v (%T)\nGot:        %#v (%T)",
						tt.name, tt.inputNode, tt.expected, tt.expected, got, got)
				}
			}
		})
	}
}

// nlines: 148
// risk_rating: LOW
