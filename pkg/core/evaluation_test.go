package core

import (
	"fmt" // Import fmt for string conversions in tests if needed
	"reflect"
	"testing"
	// "sort" // Not needed here unless comparing map string representations
)

// --- Test Helper (Copied from interpreter_test.go) ---
// Keep helper local to this file as well
func newTestInterpreterEval(vars map[string]interface{}, lastResult interface{}) *Interpreter {
	interp := NewInterpreter(nil) // Pass nil for logger in tests
	// Deep copy initial vars to avoid test interference
	if vars != nil {
		interp.variables = make(map[string]interface{}, len(vars))
		for k, v := range vars {
			interp.variables[k] = v
		}
	} else {
		interp.variables = make(map[string]interface{})
	}
	interp.lastCallResult = lastResult
	return interp
}

// TestResolvePlaceholders - Tests placeholder substitution within strings
func TestResolvePlaceholders(t *testing.T) {
	// Note: If variables hold complex types, Sprintf %v is used for interpolation
	vars := map[string]interface{}{
		"name":  "World",
		"obj":   "Test",
		"count": int64(5),
		"items": []interface{}{1, "two"},
		"map":   map[string]interface{}{"a": 1},
		"greet": "Hello {{name}}", // Var holding string with placeholder
	}
	interp := newTestInterpreterEval(vars, "LAST") // Use local helper
	// Define formats outside loop
	fullFormat := "Hello {{name}} using {{obj}} count {{count}} items %v map %v returning {{__last_call_result}}"
	nestedFormat := "Value is %s"

	// Pre-calculate expected fully resolved strings
	// Use fmt.Sprintf with the *actual values* from vars to get the expected string
	expectedFullResolved := fmt.Sprintf("Hello World using Test count 5 items %v map %v returning LAST", vars["items"], vars["map"])
	expectedNestedResolved := fmt.Sprintf("Value is %s", expectedFullResolved)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"No Placeholders", `"Hello World!"`, `"Hello World!"`},
		{"Single Found", `"Hello {{name}}!"`, `"Hello World!"`},
		{"Single Not Found", `"Hello {{user}}!"`, `"Hello {{user}}!"`},
		{"Multiple Found", `"{{obj}} for {{name}}"`, `"Test for World"`},
		{"Multiple Mixed", `"{{obj}} for {{user}}"`, `"Test for {{user}}"`},
		{"Adjacent", `{{obj}}{{name}}`, `TestWorld`},
		{"Empty Input", "", ""},
		{"Placeholder Only Found", `{{name}}`, `World`},
		{"Placeholder Only Not Found", `{{user}}`, `{{user}}`},
		{"Last Call Result Placeholder", `{{__last_call_result}}`, `LAST`},
		{"Placeholder with Last Call", `Result was: {{__last_call_result}}`, `Result was: LAST`},
		// Check string formatting of complex types via Sprintf
		{"Variable containing placeholders", `{{full}}`, expectedFullResolved}, // <-- Use pre-calculated expected
		{"Nested Variable", `{{nested}}`, expectedNestedResolved},              // <-- Use pre-calculated expected
		{"Number Placeholder", `"Count: {{count}}"`, `"Count: 5"`},
		{"Slice Placeholder", `"Items: {{items}}"`, fmt.Sprintf(`"Items: %v"`, vars["items"])},
		{"Map Placeholder", `"Map: {{map}}"`, fmt.Sprintf(`"Map: %v"`, vars["map"])},
		{"Placeholder to var with placeholder", `"{{greet}}"`, `"Hello World"`}, // Test resolving var that needs resolving
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset vars and construct complex vars for each test run
			interp.variables = make(map[string]interface{}, len(vars))
			for k, v := range vars {
				interp.variables[k] = v
			}
			// Correctly use Sprintf with format specifiers here
			interp.variables["full"] = fmt.Sprintf(fullFormat, vars["items"], vars["map"])
			interp.variables["nested"] = fmt.Sprintf(nestedFormat, interp.variables["full"])
			interp.lastCallResult = "LAST"

			got := interp.resolvePlaceholders(tt.input)

			if got != tt.expected {
				t.Errorf("resolvePlaceholders(%q):\nExpected: %q\nGot:      %q", tt.input, tt.expected, got)
			}
		})
	}
}

// TestEvaluateExpressionAST - Tests evaluation of different AST node types
func TestEvaluateExpressionAST(t *testing.T) {
	vars := map[string]interface{}{
		"name":      "World",
		"base_path": "skills/",
		"file_ext":  ".ns",
		"sanitized": "my_skill",
		"greeting":  "Hello {{name}}", // String containing a placeholder
		"numVar":    int64(123),
		"floatVar":  float64(1.5),
		"boolProp":  true,
		"listVar":   []interface{}{"x", int64(99)},
		"mapVar":    map[string]interface{}{"mKey": "mVal", "mNum": int64(1)},
	}
	lastResult := "LastCallResult"
	interp := newTestInterpreterEval(vars, lastResult) // Use local helper

	tests := []struct {
		name      string
		inputNode interface{} // Input is now an AST node
		expected  interface{}
	}{
		// Literals
		{"String Literal", StringLiteralNode{Value: "Hello World"}, `Hello World`},
		{"Empty String Literal", StringLiteralNode{Value: ""}, ``},
		{"String Lit with Placeholder", StringLiteralNode{Value: "Val={{name}}"}, `Val=World`}, // Placeholders resolved
		{"Number Literal Int", NumberLiteralNode{Value: int64(42)}, int64(42)},
		{"Number Literal Float", NumberLiteralNode{Value: float64(3.14)}, float64(3.14)},
		{"Boolean Literal True", BooleanLiteralNode{Value: true}, true},
		{"Boolean Literal False", BooleanLiteralNode{Value: false}, false},

		// Variables & Special
		{"Simple Variable String", VariableNode{Name: "name"}, `World`},
		{"Variable Not Found", VariableNode{Name: "unknown_var"}, nil}, // Returns nil if not found
		{"Variable Num", VariableNode{Name: "numVar"}, int64(123)},
		{"Variable Float", VariableNode{Name: "floatVar"}, float64(1.5)},
		{"Variable Bool", VariableNode{Name: "boolProp"}, true},
		{"Variable List", VariableNode{Name: "listVar"}, []interface{}{"x", int64(99)}},                          // Returns raw slice
		{"Variable Map", VariableNode{Name: "mapVar"}, map[string]interface{}{"mKey": "mVal", "mNum": int64(1)}}, // Returns raw map
		{"Variable Holding Placeholder String", VariableNode{Name: "greeting"}, "Hello World"},                   // Placeholder resolved now by VariableNode case
		{"Last Call Result", LastCallResultNode{}, `LastCallResult`},

		// Placeholders (evaluate what they refer to, return actual value, resolve placeholders if string)
		{"Placeholder to String", PlaceholderNode{Name: "name"}, `World`},
		{"Placeholder to Number", PlaceholderNode{Name: "numVar"}, int64(123)},                                            // Returns actual number
		{"Placeholder to Bool", PlaceholderNode{Name: "boolProp"}, true},                                                  // Returns actual bool
		{"Placeholder to List", PlaceholderNode{Name: "listVar"}, []interface{}{"x", int64(99)}},                          // Returns actual slice
		{"Placeholder to Map", PlaceholderNode{Name: "mapVar"}, map[string]interface{}{"mKey": "mVal", "mNum": int64(1)}}, // Returns actual map
		{"Placeholder to LastCall", PlaceholderNode{Name: "__last_call_result"}, `LastCallResult`},                        // Returns value
		{"Placeholder Not Found", PlaceholderNode{Name: "unknown"}, nil},                                                  // Returns nil if var not found
		{"Placeholder to var with placeholder", PlaceholderNode{Name: "greeting"}, "Hello World"},                         // Corrected: Placeholder case now resolves inner placeholders

		// Concatenation (Operands are other nodes, results stringified)
		{"Concat Lit + Lit", ConcatenationNode{Operands: []interface{}{StringLiteralNode{Value: "Hello "}, StringLiteralNode{Value: "There"}}}, "Hello There"},
		{"Concat Lit + Var", ConcatenationNode{Operands: []interface{}{StringLiteralNode{Value: "Hello "}, VariableNode{Name: "name"}}}, "Hello World"},
		{"Concat Var + Lit", ConcatenationNode{Operands: []interface{}{VariableNode{Name: "name"}, StringLiteralNode{Value: "!"}}}, "World!"},
		{"Concat Var + Var", ConcatenationNode{Operands: []interface{}{VariableNode{Name: "name"}, VariableNode{Name: "name"}}}, "WorldWorld"},
		{"Concat Lit + Placeholder", ConcatenationNode{Operands: []interface{}{StringLiteralNode{Value: "Path: "}, PlaceholderNode{Name: "base_path"}}}, "Path: skills/"},
		{"Concat Placeholder + Lit", ConcatenationNode{Operands: []interface{}{PlaceholderNode{Name: "base_path"}, StringLiteralNode{Value: "f.txt"}}}, "skills/f.txt"},
		{"Concat Placeholder + Placeholder", ConcatenationNode{Operands: []interface{}{PlaceholderNode{Name: "base_path"}, PlaceholderNode{Name: "name"}}}, "skills/World"},
		{"Concat 3 parts", ConcatenationNode{Operands: []interface{}{StringLiteralNode{Value: "A"}, VariableNode{Name: "name"}, StringLiteralNode{Value: "C"}}}, "AWorldC"},
		{"Concat with Number", ConcatenationNode{Operands: []interface{}{StringLiteralNode{Value: "Count: "}, VariableNode{Name: "numVar"}}}, "Count: 123"},
		{"Concat with Bool", ConcatenationNode{Operands: []interface{}{VariableNode{Name: "boolProp"}, StringLiteralNode{Value: " Value"}}}, "true Value"},
		{"Concat with List", ConcatenationNode{Operands: []interface{}{StringLiteralNode{Value: "List="}, VariableNode{Name: "listVar"}}}, fmt.Sprintf("List=%v", vars["listVar"])}, // Use Sprintf for consistent slice format
		{"Concat with Map", ConcatenationNode{Operands: []interface{}{VariableNode{Name: "mapVar"}, StringLiteralNode{Value: " Data"}}}, fmt.Sprintf("%v Data", vars["mapVar"])},    // Use Sprintf for consistent map format

		// List Literals
		{"Empty List", ListLiteralNode{Elements: []interface{}{}}, []interface{}{}},
		{"Simple List", ListLiteralNode{Elements: []interface{}{NumberLiteralNode{Value: int64(1)}, StringLiteralNode{Value: "two"}, BooleanLiteralNode{Value: true}}}, []interface{}{int64(1), "two", true}},
		{"List with Var", ListLiteralNode{Elements: []interface{}{VariableNode{Name: "numVar"}, VariableNode{Name: "name"}}}, []interface{}{int64(123), "World"}},
		{"List with Placeholder", ListLiteralNode{Elements: []interface{}{PlaceholderNode{Name: "name"}}}, []interface{}{"World"}}, // Placeholder evaluates to its value
		{"Nested List", ListLiteralNode{Elements: []interface{}{NumberLiteralNode{Value: int64(1)}, ListLiteralNode{Elements: []interface{}{StringLiteralNode{Value: "a"}, StringLiteralNode{Value: "b"}}}}}, []interface{}{int64(1), []interface{}{"a", "b"}}},
		{"List with Concatenation", ListLiteralNode{Elements: []interface{}{ConcatenationNode{Operands: []interface{}{VariableNode{Name: "name"}, StringLiteralNode{Value: "!"}}}}}, []interface{}{"World!"}},

		// Map Literals
		{"Empty Map", MapLiteralNode{Entries: []MapEntryNode{}}, map[string]interface{}{}},
		{"Simple Map", MapLiteralNode{Entries: []MapEntryNode{
			{Key: StringLiteralNode{Value: "k1"}, Value: StringLiteralNode{Value: "v1"}},
			{Key: StringLiteralNode{Value: "k2"}, Value: NumberLiteralNode{Value: int64(10)}},
		}}, map[string]interface{}{"k1": "v1", "k2": int64(10)}},
		{"Map with Var Value", MapLiteralNode{Entries: []MapEntryNode{
			{Key: StringLiteralNode{Value: "name"}, Value: VariableNode{Name: "name"}},
			{Key: StringLiteralNode{Value: "count"}, Value: VariableNode{Name: "numVar"}},
		}}, map[string]interface{}{"name": "World", "count": int64(123)}},
		{"Map with Placeholder Value", MapLiteralNode{Entries: []MapEntryNode{
			{Key: StringLiteralNode{Value: "greet"}, Value: PlaceholderNode{Name: "greeting"}}, // Placeholder evaluates to var value -> "Hello {{name}}" -> resolves to "Hello World"
		}}, map[string]interface{}{"greet": "Hello World"}}, // <-- CORRECTED Expected value
		{"Nested Map", MapLiteralNode{Entries: []MapEntryNode{
			{Key: StringLiteralNode{Value: "outer"}, Value: MapLiteralNode{Entries: []MapEntryNode{
				{Key: StringLiteralNode{Value: "inner"}, Value: NumberLiteralNode{Value: int64(5)}}},
			}},
		}}, map[string]interface{}{"outer": map[string]interface{}{"inner": int64(5)}}},
		{"Map with List Value", MapLiteralNode{Entries: []MapEntryNode{
			{Key: StringLiteralNode{Value: "items"}, Value: ListLiteralNode{Elements: []interface{}{NumberLiteralNode{Value: int64(1)}}}},
		}}, map[string]interface{}{"items": []interface{}{int64(1)}}},

		// Direct Go types (passed to evaluateExpression)
		{"Direct String", "PassThrough", "PassThrough"},                           // String gets placeholders resolved
		{"Direct String with Placeholder", "Value is {{name}}", "Value is World"}, // Placeholders resolved
		{"Direct Int", int64(100), int64(100)},
		{"Direct Float", float64(1.23), float64(1.23)},
		{"Direct Bool", true, true},
		{"Direct Nil", nil, nil},
		{"Direct Slice", []interface{}{"a", 1}, []interface{}{"a", 1}},
		{"Direct Map", map[string]interface{}{"x": 1}, map[string]interface{}{"x": 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset interpreter state for each test
			interp.variables = make(map[string]interface{}, len(vars))
			for k, v := range vars {
				interp.variables[k] = v
			}
			interp.lastCallResult = lastResult

			got := interp.evaluateExpression(tt.inputNode)

			// Use reflect.DeepEqual for comprehensive comparison
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("evaluateExpressionAST failed for %s:\nInput Node: %+v\nExpected:   %v (%T)\nGot:        %v (%T)",
					tt.name, tt.inputNode, tt.expected, tt.expected, got, got)
			}
		})
	}
}

// TestEvaluateCondition - Tests condition evaluation logic
func TestEvaluateCondition(t *testing.T) {
	vars := map[string]interface{}{
		"trueVar":   true,
		"falseVar":  false,
		"numOne":    int64(1),
		"numZero":   int64(0),
		"floatOne":  float64(1.0),
		"floatZero": float64(0.0),
		"strTrue":   "true",
		"strFalse":  "false",
		"strOther":  "hello",
	}
	interp := newTestInterpreterEval(vars, nil) // Use local helper

	tests := []struct {
		name     string
		node     interface{} // Input is AST node
		expected bool
		wantErr  bool
	}{
		{"Bool Literal True", BooleanLiteralNode{Value: true}, true, false},
		{"Bool Literal False", BooleanLiteralNode{Value: false}, false, false},
		{"Var Bool True", VariableNode{Name: "trueVar"}, true, false},
		{"Var Bool False", VariableNode{Name: "falseVar"}, false, false},
		{"Var Num NonZero", VariableNode{Name: "numOne"}, true, false},
		{"Var Num Zero", VariableNode{Name: "numZero"}, false, false},
		{"Var Float NonZero", VariableNode{Name: "floatOne"}, true, false},
		{"Var Float Zero", VariableNode{Name: "floatZero"}, false, false},
		{"Var String True", VariableNode{Name: "strTrue"}, true, false},
		{"Var String False", VariableNode{Name: "strFalse"}, false, false},
		{"Var String Other", VariableNode{Name: "strOther"}, false, true}, // Expect error for non-"true"/"false" string
		{"String Literal True", StringLiteralNode{Value: "true"}, true, false},
		{"String Literal False", StringLiteralNode{Value: "false"}, false, false},
		{"String Literal Other", StringLiteralNode{Value: "yes"}, false, true}, // Expect error
		{"Number Literal NonZero", NumberLiteralNode{Value: int64(1)}, true, false},
		{"Number Literal Zero", NumberLiteralNode{Value: int64(0)}, false, false},
		{"Number Literal Float NonZero", NumberLiteralNode{Value: float64(0.1)}, true, false},
		{"Number Literal Float Zero", NumberLiteralNode{Value: float64(0.0)}, false, false},
		{"Variable Not Found Condition", VariableNode{Name: "z"}, false, true},              // Error evaluating unknown var (evaluateExpression returns nil)
		{"List Literal Condition", ListLiteralNode{Elements: []interface{}{}}, false, true}, // Error: list is not bool/num
		{"Map Literal Condition", MapLiteralNode{Entries: []MapEntryNode{}}, false, true},   // Error: map is not bool/num
		// TODO: Add tests for actual comparisons (e.g. EqualityNode) when implemented
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset vars for safety
			interp.variables = make(map[string]interface{}, len(vars))
			for k, v := range vars {
				interp.variables[k] = v
			}

			got, err := interp.evaluateCondition(tt.node)

			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateCondition(%+v) error = %v, wantErr %v", tt.node, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("evaluateCondition(%+v) = %v, want %v", tt.node, got, tt.expected)
			}
		})
	}
}
