package core

import (
	"reflect"
	"testing"
)

// --- Interpreter Test Specific Helper ---
func newTestInterpreter(vars map[string]interface{}, lastResult interface{}) *Interpreter {
	interp := NewInterpreter()
	if vars != nil {
		interp.variables = vars
	}
	interp.lastCallResult = lastResult
	return interp
}

// --- Unit Tests for Interpreter Helpers ---

// TestSplitExpression seems stable, no changes needed based on failures
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
		{"Placeholder With Space", `{{ file name }}`, []string{`{{ file name }}`}}, // Assuming spaces allowed in placeholders
		{"Concat Placeholders With Space", `{{ greeting }} + {{ user }}`, []string{`{{ greeting }}`, `+`, `{{ user }}`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitExpression(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("splitExpression(%q):\nExpected: %v\nGot:      %v", tt.input, tt.expected, got)
			}
		})
	}
}

// TestResolvePlaceholders seems stable
func TestResolvePlaceholders(t *testing.T) {
	vars := map[string]interface{}{"name": "World", "obj": "Test"}
	interp := newTestInterpreter(vars, nil)
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"No Placeholders", `"Hello World!"`, `"Hello World!"`},
		{"Single Found", `"Hello {{name}}!"`, `"Hello World!"`},
		{"Single Not Found", `"Hello {{user}}!"`, `"Hello {{user}}!"`}, // Expect placeholder kept
		{"Multiple Found", `"{{obj}} for {{name}}"`, `"Test for World"`},
		{"Multiple Mixed", `"{{obj}} for {{user}}"`, `"Test for {{user}}"`}, // Expect user kept
		{"Adjacent", `{{obj}}{{name}}`, `TestWorld`},
		{"Empty Input", "", ""},
		{"Placeholder Only Found", `{{name}}`, `World`},
		{"Placeholder Only Not Found", `{{user}}`, `{{user}}`}, // Expect placeholder kept
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

// TestResolveValue - Adjusted expectation for Variable_Containing_Placeholder
func TestResolveValue(t *testing.T) {
	vars := map[string]interface{}{"name": "World", "obj": "Test", "filename": "actual_file.txt", "greeting": "Hello {{name}}"}
	interp := newTestInterpreter(vars, nil)
	tests := []struct {
		name     string
		input    string
		expected interface{} // Now comparing interface{}
	}{
		{"Literal String", `"Hello"`, `"Hello"`},          // resolveValue returns string literal
		{"Plain Variable Found", `name`, `World`},         // resolveValue returns string value
		{"Plain Variable Not Found", `user`, `user`},      // resolveValue returns input string
		{"Placeholder Found", `{{name}}`, `World`},        // resolveValue resolves placeholder -> string
		{"Placeholder Not Found", `{{user}}`, `{{user}}`}, // resolveValue returns placeholder string
		// ** Updated Expectation: resolveValue returns RAW variable content **
		{"Variable Containing Placeholder", `greeting`, `Hello {{name}}`},
		{"Quoted Placeholder", `"{{name}}"`, `"World"`},     // resolveValue resolves placeholder inside literal -> string
		{"Not a plain var (has space)", `my var`, `my var`}, // resolveValue returns input string
		{"Not a plain var (has +)", `"a"+b`, `"a"+b`},       // resolveValue returns input string
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := interp.resolveValue(tt.input)
			// Use DeepEqual for interface{} comparison
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("resolveValue(%q):\nExpected: %v (%T)\nGot:      %v (%T)", tt.input, tt.expected, tt.expected, got, got)
			}
		})
	}
}

// TestEvaluateExpression - Corrected input for Lit_+_Placeholder_+_Lit
// TestEvaluateExpression - Corrected input for Lit_+_Placeholder_+_Lit
func TestEvaluateExpression(t *testing.T) {
	// ** FIX: Added "greeting" variable to the map **
	vars := map[string]interface{}{"name": "World", "base_path": "skills/", "file_ext": ".ns", "sanitized": "my_skill", "commit_msg": "Initial commit", "greeting": "Hello {{name}}"}
	lastResult := "LastCallResult"
	interp := newTestInterpreter(vars, lastResult)
	tests := []struct {
		name     string
		input    string
		expected interface{} // Comparing interface{} as result could be non-string later
	}{
		// ... existing test cases ...
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
		// Tests for invalid concatenation patterns (should fallback to evaluating whole string)
		{"Ends with +", `"Hello" + name +`, `"Hello" + name +`},
		{"Starts with +", `+ "Hello"`, `+ "Hello"`},
		{"Missing +", `"Hello" name`, `"Hello" name`},
		{"Adjacent Vars", `name sanitized`, `name sanitized`},
		// ** This test should now pass with the updated vars map **
		{"Variable with internal placeholder", `greeting`, `Hello World`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset lastCallResult specifically for the relevant test
			currentLastResult := interp.lastCallResult // Backup
			if tt.input == "__last_call_result" {
				interp.lastCallResult = lastResult
			} else {
				// ** Ensure greeting var is available for its test **
				if tt.input != "greeting" {
					interp.lastCallResult = nil // Clear for other tests
				}
			}

			got := interp.evaluateExpression(tt.input)

			// Restore lastCallResult
			interp.lastCallResult = currentLastResult

			// Use DeepEqual for interface{} comparison
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("evaluateExpression(%q):\nExpected: %v (%T)\nGot:      %v (%T)", tt.input, tt.expected, tt.expected, got, got)
			}
		})
	}
}
