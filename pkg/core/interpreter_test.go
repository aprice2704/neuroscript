package core

import (
	"reflect"
	"testing"
)

// --- Interpreter Test Specific Helper ---
func newTestInterpreter(vars map[string]interface{}, lastResult interface{}) *Interpreter {
	interp := NewInterpreter()
	if vars != nil {
		// Deep copy initial vars to avoid modification pollution across tests if needed
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

// --- Unit Tests for Interpreter Helpers (Keep existing ones) ---

// TestSplitExpression (Keep if desired, tests independent logic)
func TestSplitExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"Empty", "", []string{}},
		{"Simple Literal", `"hello world"`, []string{`"hello world"`}},
		{"Simple Variable", `myVar`, []string{`myVar`}},
		{"Placeholder", `{{myVar}}`, []string{`{{myVar}}`}},
		{"Literal + Placeholder", `"Hello " + {{name}}`, []string{`"Hello "`, `+`, `{{name}}`}},
		{"Placeholder + Literal", `{{greeting}} + "!"`, []string{`{{greeting}}`, `+`, `"!"`}},
		{"Literal + Var + Literal", `"skills/" + sanitized_base + ".ns"`, []string{`"skills/"`, `+`, `sanitized_base`, `+`, `".ns"`}},
		{"No Spaces", `"a"+b+"c"`, []string{`"a"`, `+`, `b`, `+`, `"c"`}},
		{"Extra Spaces", ` "a"  +  b +   "c" `, []string{`"a"`, `+`, `b`, `+`, `"c"`}},
		{"Plus Inside Quotes", `"a + b"`, []string{`"a + b"`}},
		{"Placeholder With Space", `{{ file name }}`, []string{`{{ file name }}`}},
		{"Concat Placeholders With Space", `{{ greeting }} + {{ user }}`, []string{`{{ greeting }}`, `+`, `{{ user }}`}},
		{"List Literal String", `["a", "b"]`, []string{`["a", "b"]`}}, // Assuming parser returns string for now
		{"Map Literal String", `{"k": "v"}`, []string{`{"k": "v"}`}},  // Assuming parser returns string for now
		{"Concat with List", `Prefix + ["a"]`, []string{`Prefix`, `+`, `["a"]`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Assuming splitExpression is globally accessible or part of interpreter for testing
			// If splitExpression is not public, this test might need adjustment or removal
			// For now, assuming access for demonstration
			got := splitExpression(tt.input) // splitExpression is actually in interpreter_b.go
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("splitExpression(%q):\nExpected: %v\nGot:      %v", tt.input, tt.expected, got)
			}
		})
	}
}

// TestResolvePlaceholders (Keep if desired)
func TestResolvePlaceholders(t *testing.T) {
	vars := map[string]interface{}{"name": "World", "obj": "Test"}
	interp := newTestInterpreter(vars, "LAST") // Add last result for testing
	interp.variables["full"] = "Hello {{name}} using {{obj}} returning {{__last_call_result}}"

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
		{"Last Call Result", `__last_call_result`, `LAST`},
		{"Placeholder with Last Call", `Result was: {{__last_call_result}}`, `Result was: LAST`},
		{"Nested/Combined", `{{full}}`, `Hello World using Test returning LAST`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := interp.resolvePlaceholders(tt.input)
			if got != tt.expected {
				t.Errorf("resolvePlaceholders(%q):\nExpected: %q\nGot:      %q", tt.input, tt.expected, got)
			}
		})
	}
}

// TestResolveValue (Keep if desired)
func TestResolveValue(t *testing.T) {
	vars := map[string]interface{}{"name": "World", "obj": "Test", "filename": "actual_file.txt", "greeting": "Hello {{name}}"}
	interp := newTestInterpreter(vars, "LAST_RESULT")

	tests := []struct {
		name     string
		input    string
		expected interface{} // Comparing interface{}
	}{
		{"Literal String", `"Hello"`, `"Hello"`},
		{"Plain Variable Found", `name`, `World`},
		{"Plain Variable Not Found", `user`, `user`},                      // Returns input string
		{"Placeholder Found", `{{name}}`, `World`},                        // Resolves placeholder -> string
		{"Placeholder Not Found", `{{user}}`, `{{user}}`},                 // Returns placeholder string
		{"Variable Containing Placeholder", `greeting`, `Hello {{name}}`}, // Returns RAW variable value
		{"Quoted Placeholder", `"{{name}}"`, `"World"`},                   // Resolves placeholder inside literal -> string
		{"Not a plain var (space)", `my var`, `my var`},
		{"Not a plain var (+)", `"a"+b`, `"a"+b`},
		{"Last Call Result Keyword", `__last_call_result`, `LAST_RESULT`}, // Returns actual last result
		{"Quoted Last Call", `"{{__last_call_result}}"`, `"LAST_RESULT"`}, // Resolves inside literal
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := interp.resolveValue(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("resolveValue(%q):\nExpected: %v (%T)\nGot:      %v (%T)", tt.input, tt.expected, tt.expected, got, got)
			}
		})
	}
}

// TestEvaluateExpression (Keep if desired)
func TestEvaluateExpression(t *testing.T) {
	vars := map[string]interface{}{"name": "World", "base_path": "skills/", "file_ext": ".ns", "sanitized": "my_skill", "greeting": "Hello {{name}}"}
	lastResult := "LastCallResult"
	interp := newTestInterpreter(vars, lastResult)

	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"Literal String", `"Hello World"`, `Hello World`},
		{"Simple Variable", `name`, `World`},
		{"Placeholder in Literal", `"Hello {{name}}"`, `Hello World`},
		{"Last Call Result", `__last_call_result`, `LastCallResult`},
		{"Variable Not Found", `unknown_var`, `unknown_var`},
		{"Placeholder Not Found", `{{unknown_var}}`, `{{unknown_var}}`},
		{"Empty String", `""`, ``},
		{"Literal + Literal", `"Hello" + " " + "World"`, `Hello World`},
		{"Literal + Variable", `"Hello " + name`, `Hello World`},
		{"Variable + Literal", `name + "!"`, `World!`},
		{"Variable + Variable", `name + name`, `WorldWorld`},
		{"Literal + Placeholder", `"Path: " + {{base_path}}`, `Path: skills/`},
		{"Placeholder + Literal", `{{base_path}} + "file.txt"`, `skills/file.txt`},
		{"Placeholder + Placeholder", `{{base_path}} + {{name}}`, `skills/World`},
		{"Lit + Var + Lit", `"skills/" + sanitized + ".ns"`, `skills/my_skill.ns`},
		{"Lit + Placeholder + Lit", `"File: " + {{sanitized}} + {{file_ext}}`, `File: my_skill.ns`},
		{"Var + Lit + Var", `base_path + sanitized + file_ext`, `skills/my_skill.ns`},
		{"String with internal '+'", `"Value is a+b"`, `Value is a+b`},
		{"Ends with +", `"Hello" + name +`, `"Hello" + name +`},           // Invalid concat fallback
		{"Starts with +", `+ "Hello"`, `+ "Hello"`},                       // Invalid concat fallback
		{"Missing +", `"Hello" name`, `"Hello" name`},                     // Invalid pattern fallback
		{"Adjacent Vars", `name sanitized`, `name sanitized`},             // Invalid pattern fallback
		{"Variable with internal placeholder", `greeting`, `Hello World`}, // evaluateExpression resolves internal placeholders
		{"Placeholder resolving to last call", `{{__last_call_result}}`, `LastCallResult`},
		{"Concat involving last call", `"Result: " + __last_call_result`, `Result: LastCallResult`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Restore last call result before each test if needed
			interp.lastCallResult = lastResult
			got := interp.evaluateExpression(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("evaluateExpression(%q):\nExpected: %v (%T)\nGot:      %v (%T)", tt.input, tt.expected, tt.expected, got, got)
			}
		})
	}
}

// --- NEW: Test Suite for executeSteps ---

// Helper struct for executeSteps tests
type executeStepsTestCase struct {
	name           string
	inputSteps     []Step
	initialVars    map[string]interface{}
	expectedVars   map[string]interface{} // Check specific vars *after* execution
	expectedResult interface{}            // Expected return value from executeSteps
	expectError    bool                   // True if an error is expected
}

// Helper function to run executeSteps tests
func runExecuteStepsTest(t *testing.T, tc executeStepsTestCase) {
	t.Helper()
	interp := newTestInterpreter(tc.initialVars, nil) // Start fresh interpreter for each case

	// It's often easier to debug by running directly rather than through RunProcedure
	finalResult, _, err := interp.executeSteps(tc.inputSteps)

	// Check error expectation
	if tc.expectError {
		if err == nil {
			t.Errorf("Expected an error, but got nil")
		}
		// Optionally check for specific error string contains(err.Error(), tc.expectedErrorString)
	} else {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}

	// Check final result expectation
	if !reflect.DeepEqual(finalResult, tc.expectedResult) {
		t.Errorf("Final result mismatch:\nExpected: %v (%T)\nGot:      %v (%T)", tc.expectedResult, tc.expectedResult, finalResult, finalResult)
	}

	// Check expected variables state (only check keys present in expectedVars)
	if tc.expectedVars != nil {
		for key, expectedValue := range tc.expectedVars {
			actualValue, exists := interp.variables[key]
			if !exists {
				t.Errorf("Expected variable '%s' not found in final state", key)
			} else if !reflect.DeepEqual(actualValue, expectedValue) {
				t.Errorf("Variable '%s' mismatch:\nExpected: %v (%T)\nGot:      %v (%T)", key, expectedValue, expectedValue, actualValue, actualValue)
			}
		}
	}
}

func TestExecuteStepsBlocksAndLoops(t *testing.T) {
	testCases := []executeStepsTestCase{
		// --- IF Block Tests ---
		{
			name: "IF true basic",
			inputSteps: []Step{
				{Type: "IF", Cond: `"true"`, Value: []Step{ // Block body
					{Type: "SET", Target: "x", Value: `"Inside"`},
				}},
			},
			initialVars:    map[string]interface{}{},
			expectedVars:   map[string]interface{}{"x": "Inside"},
			expectedResult: nil, // No explicit return
			expectError:    false,
		},
		{
			name: "IF false basic",
			inputSteps: []Step{
				{Type: "IF", Cond: `"false"`, Value: []Step{
					{Type: "SET", Target: "x", Value: `"Inside"`},
				}},
				{Type: "SET", Target: "y", Value: `"Outside"`},
			},
			initialVars:    map[string]interface{}{},
			expectedVars:   map[string]interface{}{"y": "Outside"}, // x should not be set
			expectedResult: nil,
			expectError:    false,
		},
		{
			name: "IF condition var true",
			inputSteps: []Step{
				{Type: "IF", Cond: `{{cond_var}}`, Value: []Step{
					{Type: "SET", Target: "x", Value: `"Inside"`},
				}},
			},
			initialVars:    map[string]interface{}{"cond_var": "true"},
			expectedVars:   map[string]interface{}{"cond_var": "true", "x": "Inside"},
			expectedResult: nil,
			expectError:    false,
		},
		{
			name: "IF empty block",
			inputSteps: []Step{
				{Type: "IF", Cond: `"true"`, Value: []Step{}}, // Empty block
				{Type: "SET", Target: "y", Value: `"After"`},
			},
			initialVars:    map[string]interface{}{},
			expectedVars:   map[string]interface{}{"y": "After"},
			expectedResult: nil,
			expectError:    false,
		},
		{
			name: "IF block with RETURN",
			inputSteps: []Step{
				{Type: "SET", Target: "status", Value: `"Started"`},
				{Type: "IF", Cond: `"true"`, Value: []Step{
					{Type: "SET", Target: "x", Value: `"Inside"`},
					{Type: "RETURN", Value: `"ReturnedFromIf"`}, // Explicit RETURN
					{Type: "SET", Target: "y", Value: `"NotReached"`},
				}},
				{Type: "SET", Target: "status", Value: `"Finished"`}, // Should not be reached
			},
			initialVars:    map[string]interface{}{},
			expectedVars:   map[string]interface{}{"status": "Started", "x": "Inside"}, // Check state before RETURN
			expectedResult: "ReturnedFromIf",                                           // Expect return value
			expectError:    false,
		},
		// --- WHILE Block Tests ---
		{
			name: "WHILE false initially",
			inputSteps: []Step{
				{Type: "SET", Target: "count", Value: `"0"`},
				{Type: "WHILE", Cond: `{{count}} == "1"`, Value: []Step{ // Condition initially false
					{Type: "SET", Target: "x", Value: `"InsideLoop"`},
				}},
				{Type: "SET", Target: "y", Value: `"AfterLoop"`},
			},
			initialVars:    map[string]interface{}{},
			expectedVars:   map[string]interface{}{"count": "0", "y": "AfterLoop"}, // x not set
			expectedResult: nil,
			expectError:    false,
		},
		{
			name: "WHILE runs once",
			inputSteps: []Step{
				{Type: "SET", Target: "run", Value: `"true"`},
				{Type: "SET", Target: "counter", Value: `"0"`},
				{Type: "WHILE", Cond: `{{run}} == "true"`, Value: []Step{
					{Type: "SET", Target: "run", Value: `"false"`}, // Stop after one iteration
					{Type: "SET", Target: "counter", Value: `"1"`},
				}},
			},
			initialVars:    map[string]interface{}{},
			expectedVars:   map[string]interface{}{"run": "false", "counter": "1"},
			expectedResult: nil,
			expectError:    false,
		},
		// TODO: Add WHILE loop N times test (needs arithmetic support)
		{
			name: "WHILE block with RETURN",
			inputSteps: []Step{
				{Type: "SET", Target: "run", Value: `"true"`},
				{Type: "WHILE", Cond: `{{run}} == "true"`, Value: []Step{
					{Type: "RETURN", Value: `"ReturnedFromWhile"`},
					{Type: "SET", Target: "run", Value: `"false"`}, // Not reached
				}},
				{Type: "SET", Target: "status", Value: `"Finished"`}, // Should not be reached
			},
			initialVars:    map[string]interface{}{},
			expectedVars:   map[string]interface{}{"run": "true"}, // State before RETURN
			expectedResult: "ReturnedFromWhile",
			expectError:    false,
		},
		// --- FOR EACH Tests ---
		{
			name: "FOR EACH string (abc)",
			inputSteps: []Step{
				{Type: "SET", Target: "input", Value: `"abc"`},
				{Type: "SET", Target: "output", Value: `""`},
				{Type: "FOR", Target: "char", Cond: `{{input}}`, Value: []Step{ // Iterate over string
					{Type: "SET", Target: "output", Value: `{{output}} + {{char}}`}, // Concatenate chars
				}},
			},
			initialVars:    map[string]interface{}{},
			expectedVars:   map[string]interface{}{"input": "abc", "output": "abc"}, // output should be "abc"
			expectedResult: nil,
			expectError:    false,
		},
		{
			name: "FOR EACH string empty",
			inputSteps: []Step{
				{Type: "SET", Target: "input", Value: `""`},
				{Type: "SET", Target: "output", Value: `"start"`},
				{Type: "FOR", Target: "char", Cond: `{{input}}`, Value: []Step{
					{Type: "SET", Target: "output", Value: `"in_loop"`},
				}},
			},
			initialVars:    map[string]interface{}{},
			expectedVars:   map[string]interface{}{"input": "", "output": "start"}, // output unchanged
			expectedResult: nil,
			expectError:    false,
		},
		{
			name: "FOR EACH string with RETURN",
			inputSteps: []Step{
				{Type: "SET", Target: "input", Value: `"stop"`},
				{Type: "SET", Target: "output", Value: `""`},
				{Type: "FOR", Target: "char", Cond: `{{input}}`, Value: []Step{
					{Type: "SET", Target: "output", Value: `{{output}} + {{char}}`},
					{Type: "IF", Cond: `{{char}} == "o"`, Value: []Step{ // If char is 'o'
						{Type: "RETURN", Value: `"Stopped at o"`},
					}},
				}},
			},
			initialVars:    map[string]interface{}{},
			expectedVars:   map[string]interface{}{"input": "stop", "output": "sto"}, // Output before RETURN
			expectedResult: "Stopped at o",
			expectError:    false,
		},
		{
			name: "FOR EACH comma split fallback",
			inputSteps: []Step{
				{Type: "SET", Target: "input", Value: `"a, b ,c"`}, // Comma separated
				{Type: "SET", Target: "output", Value: `""`},
				{Type: "FOR", Target: "item", Cond: `{{input}}`, Value: []Step{ // Iterate fallback
					// Prepend item with dash
					{Type: "SET", Target: "output", Value: `{{output}} + "-" + {{item}}`},
				}},
			},
			initialVars: map[string]interface{}{},
			// Expects iteration over "a", "b", "c" (after trim)
			expectedVars:   map[string]interface{}{"input": "a, b ,c", "output": "-a-b-c"},
			expectedResult: nil,
			expectError:    false,
		},
		// TODO: Add tests for FOR EACH on native lists/maps when implemented
		// TODO: Add tests for nested blocks
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runExecuteStepsTest(t, tc)
		})
	}
}
