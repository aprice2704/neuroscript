package core

import (
	"reflect"
	"strings"
	"testing"
)

// --- Interpreter Test Specific Helper ---
func newTestInterpreter(vars map[string]interface{}, lastResult interface{}) *Interpreter {
	interp := NewInterpreter()
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

// --- Unit Tests for Interpreter Helpers ---

// TestSplitExpression
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
		{"Parentheses", `"A" + (b + "C")`, []string{`"A"`, `+`, `(b + "C")`}},
		{"Placeholder Inside Parentheses", `"A" + ({{var}} + "C")`, []string{`"A"`, `+`, `({{var}} + "C")`}},
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

// TestResolvePlaceholders
func TestResolvePlaceholders(t *testing.T) {
	vars := map[string]interface{}{"name": "World", "obj": "Test"}
	interp := newTestInterpreter(vars, "LAST")
	interp.variables["full"] = "Hello {{name}} using {{obj}} returning {{__last_call_result}}"
	interp.variables["nested"] = "Value is {{full}}"
	interp.variables["self"] = "{{self}}"
	interp.variables["loop1"] = "{{loop2}}"
	interp.variables["loop2"] = "{{loop1}}"
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"No Placeholders", `"Hello World!"`, `"Hello World!"`}, {"Single Found", `"Hello {{name}}!"`, `"Hello World!"`},
		{"Single Not Found", `"Hello {{user}}!"`, `"Hello {{user}}!"`}, {"Multiple Found", `"{{obj}} for {{name}}"`, `"Test for World"`},
		{"Multiple Mixed", `"{{obj}} for {{user}}"`, `"Test for {{user}}"`}, {"Adjacent", `{{obj}}{{name}}`, `TestWorld`},
		{"Empty Input", "", ""}, {"Placeholder Only Found", `{{name}}`, `World`}, {"Placeholder Only Not Found", `{{user}}`, `{{user}}`},
		{"Last Call Result Placeholder", `{{__last_call_result}}`, `LAST`}, {"Placeholder with Last Call", `Result was: {{__last_call_result}}`, `Result was: LAST`},
		{"Variable containing placeholders", `{{full}}`, `Hello World using Test returning LAST`}, {"Nested Variable", `{{nested}}`, `Value is Hello World using Test returning LAST`},
		{"Literal __last_call_result", `__last_call_result`, `__last_call_result`}, {"Self Referential", `{{self}}`, `{{self}}`},
		{"Recursive Loop", `{{loop1}}`, `{{loop1}}`}, // Expect original on max depth
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

// TestResolveValue
func TestResolveValue(t *testing.T) {
	vars := map[string]interface{}{"name": "World", "obj": "Test", "filename": "actual_file.txt", "greeting": "Hello {{name}}", "num": 123, "items": []string{"a", "b"}}
	interp := newTestInterpreter(vars, "LAST_RESULT")
	tests := []struct {
		name     string
		input    string
		expected interface{}
		found    bool
	}{
		{"Literal String", `"Hello"`, `"Hello"`, false}, {"Plain Variable Found", `name`, `World`, true},
		{"Plain Variable Not Found", `user`, `user`, false}, {"Placeholder String", `{{name}}`, `{{name}}`, false},
		{"Variable Containing Placeholder", `greeting`, `Hello {{name}}`, true}, {"Quoted Placeholder", `"{{name}}"`, `"{{name}}"`, false},
		{"Not a plain var (space)", `my var`, `my var`, false}, {"Not a plain var (+)", `"a"+b`, `"a"+b`, false},
		{"Last Call Result Keyword", `__last_call_result`, `LAST_RESULT`, true}, {"Quoted Last Call", `"{{__last_call_result}}"`, `"{{__last_call_result}}"`, false},
		{"Number Var", `num`, 123, true}, {"Slice Var", `items`, []string{"a", "b"}, true},
		{"Empty String Input", `""`, `""`, false}, {"Whitespace Input", `  `, ``, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found := interp.resolveValue(tt.input)
			if found != tt.found {
				t.Errorf("resolveValue(%q) found mismatch: Expected: %v, Got: %v", tt.input, tt.found, found)
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("resolveValue(%q):\nExpected: %v (%T)\nGot:      %v (%T)", tt.input, tt.expected, tt.expected, got, got)
			}
		})
	}
}

// TestEvaluateExpression
func TestEvaluateExpression(t *testing.T) {
	vars := map[string]interface{}{"name": "World", "base_path": "skills/", "file_ext": ".ns", "sanitized": "my_skill", "greeting": "Hello {{name}}", "num": 123, "items": []string{"x", "y"}}
	lastResult := "LastCallResult"
	interp := newTestInterpreter(vars, lastResult)
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"Literal String", `"Hello World"`, `Hello World`}, {"Simple Variable", `name`, `World`},
		{"Placeholder in Literal", `"Hello {{name}}"`, `Hello World`}, {"Last Call Result", `__last_call_result`, `LastCallResult`},
		{"Variable Not Found", `unknown_var`, `unknown_var`}, {"Placeholder Not Found", `{{unknown_var}}`, `{{unknown_var}}`},
		{"Empty String Literal", `""`, ``}, {"Literal + Literal", `"Hello" + " " + "World"`, `Hello World`},
		{"Literal + Variable", `"Hello " + name`, `Hello World`}, {"Variable + Literal", `name + "!"`, `World!`},
		{"Variable + Variable", `name + name`, `WorldWorld`}, {"Literal + Placeholder", `"Path: " + {{base_path}}`, `Path: skills/`},
		{"Placeholder + Literal", `{{base_path}} + "file.txt"`, `skills/file.txt`}, {"Placeholder + Placeholder", `{{base_path}} + {{name}}`, `skills/World`},
		{"Lit + Var + Lit", `"skills/" + sanitized + ".ns"`, `skills/my_skill.ns`}, {"Lit + Placeholder + Placeholder", `"File: " + {{sanitized}} + {{file_ext}}`, `File: my_skill.ns`},
		{"Var + Lit + Var", `base_path + sanitized + file_ext`, `skills/my_skill.ns`}, {"String with internal '+'", `"Value is a+b"`, `Value is a+b`},
		{"Ends with +", `"Hello" + name +`, `HelloWorld`}, // Now evaluates valid part
		{"Starts with +", `+ "Hello"`, `+ "Hello"`}, {"Missing +", `"Hello" name`, `"Hello" name`}, {"Adjacent Vars", `name sanitized`, `name sanitized`},
		{"Variable with internal placeholder", `greeting`, `Hello {{name}}`}, {"Placeholder resolving to last call", `{{__last_call_result}}`, `LastCallResult`},
		{"Concat involving last call", `"Result: " + __last_call_result`, `Result: LastCallResult`},
		{"Number Var Direct", `num`, 123}, {"Slice Var Direct", `items`, []string{"x", "y"}},
		{"Concat Number Var", `"Count: " + num`, `Count: 123`}, {"Concat Slice Var", `items + " z"`, `[x y] z`},
		{"Parenthesized Expression", `(name + "!")`, `World!`}, {"Literal + Parenthesized", `"Prefix: " + (name + "!")`, `Prefix: World!`},
		{"Placeholder Inside Parentheses", `"Value: " + ({{name}} + "...")`, `Value: World...`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp.lastCallResult = lastResult
			interp.variables = vars
			got := interp.evaluateExpression(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("evaluateExpression(%q):\nExpected: %v (%T)\nGot:      %v (%T)", tt.input, tt.expected, tt.expected, got, got)
			}
		})
	}
}

// --- Test Suite for executeSteps (Blocks, Loops, Tools) ---
type executeStepsTestCase struct {
	name           string
	inputSteps     []Step
	initialVars    map[string]interface{}
	expectedVars   map[string]interface{}
	expectedResult interface{}
	expectError    bool
	errorContains  string
}

func runExecuteStepsTest(t *testing.T, tc executeStepsTestCase) {
	t.Helper()
	interp := newTestInterpreter(tc.initialVars, nil)
	finalResult, _, err := interp.executeSteps(tc.inputSteps)
	if tc.expectError {
		if err == nil {
			t.Errorf("Expected an error, but got nil")
		} else if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
			t.Errorf("Expected error containing %q, but got: %v", tc.errorContains, err)
		}
	} else {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
	if !tc.expectError && !reflect.DeepEqual(finalResult, tc.expectedResult) {
		t.Errorf("Final result mismatch:\nExpected: %v (%T)\nGot:      %v (%T)", tc.expectedResult, tc.expectedResult, finalResult, finalResult)
	}
	if tc.expectedVars != nil && err == nil {
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

// TestExecuteStepsBlocksAndLoops
func TestExecuteStepsBlocksAndLoops(t *testing.T) {
	testCases := []executeStepsTestCase{
		{name: "IF true basic", inputSteps: []Step{{Type: "IF", Cond: `"true"`, Value: []Step{{Type: "SET", Target: "x", Value: `"Inside"`}}}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"x": "Inside"}, expectedResult: nil, expectError: false},
		{name: "IF false basic", inputSteps: []Step{{Type: "IF", Cond: `"false"`, Value: []Step{{Type: "SET", Target: "x", Value: `"Inside"`}}}, {Type: "SET", Target: "y", Value: `"Outside"`}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"y": "Outside"}, expectedResult: nil, expectError: false},
		{name: "IF condition var true", inputSteps: []Step{{Type: "IF", Cond: `{{cond_var}}`, Value: []Step{{Type: "SET", Target: "x", Value: `"Inside"`}}}}, initialVars: map[string]interface{}{"cond_var": "true"}, expectedVars: map[string]interface{}{"cond_var": "true", "x": "Inside"}, expectedResult: nil, expectError: false},
		{name: "IF string eq true", inputSteps: []Step{{Type: "SET", Target: "a", Value: `"hello"`}, {Type: "IF", Cond: `{{a}} == "hello"`, Value: []Step{{Type: "SET", Target: "result", Value: `"eq_true"`}}}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"a": "hello", "result": "eq_true"}, expectedResult: nil, expectError: false},
		{name: "IF string gt true (b > a)", inputSteps: []Step{{Type: "IF", Cond: `"b" > "a"`, Value: []Step{{Type: "SET", Target: "result", Value: `"gt_true"`}}}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"result": "gt_true"}, expectedResult: nil, expectError: false},
		{name: "IF string lt true (a < b)", inputSteps: []Step{{Type: "IF", Cond: `"a" < "b"`, Value: []Step{{Type: "SET", Target: "result", Value: `"lt_true"`}}}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"result": "lt_true"}, expectedResult: nil, expectError: false},
		{name: "IF string gte true (b >= b)", inputSteps: []Step{{Type: "IF", Cond: `"b" >= "b"`, Value: []Step{{Type: "SET", Target: "result", Value: `"gte_true"`}}}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"result": "gte_true"}, expectedResult: nil, expectError: false},
		{name: "IF string lte true (a <= b)", inputSteps: []Step{{Type: "IF", Cond: `"a" <= "b"`, Value: []Step{{Type: "SET", Target: "result", Value: `"lte_true"`}}}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"result": "lte_true"}, expectedResult: nil, expectError: false},
		{name: "IF empty block", inputSteps: []Step{{Type: "IF", Cond: `"true"`, Value: []Step{}}, {Type: "SET", Target: "y", Value: `"After"`}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"y": "After"}, expectedResult: nil, expectError: false},
		{name: "IF block with RETURN", inputSteps: []Step{{Type: "SET", Target: "status", Value: `"Started"`}, {Type: "IF", Cond: `"true"`, Value: []Step{{Type: "SET", Target: "x", Value: `"Inside"`}, {Type: "RETURN", Value: `"ReturnedFromIf"`}, {Type: "SET", Target: "y", Value: `"NotReached"`}}}, {Type: "SET", Target: "status", Value: `"Finished"`}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"status": "Started", "x": "Inside"}, expectedResult: "ReturnedFromIf", expectError: false},
		{name: "WHILE false initially", inputSteps: []Step{{Type: "SET", Target: "count", Value: `"0"`}, {Type: "WHILE", Cond: `{{count}} == "1"`, Value: []Step{{Type: "SET", Target: "x", Value: `"InsideLoop"`}}}, {Type: "SET", Target: "y", Value: `"AfterLoop"`}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"count": "0", "y": "AfterLoop"}, expectedResult: nil, expectError: false},
		{name: "WHILE runs once", inputSteps: []Step{{Type: "SET", Target: "run", Value: `"true"`}, {Type: "SET", Target: "counter", Value: `"0"`}, {Type: "WHILE", Cond: `{{run}} == "true"`, Value: []Step{{Type: "SET", Target: "run", Value: `"false"`}, {Type: "SET", Target: "counter", Value: `"1"`}}}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"run": "false", "counter": "1"}, expectedResult: nil, expectError: false},
		{name: "WHILE block with RETURN", inputSteps: []Step{{Type: "SET", Target: "run", Value: `"true"`}, {Type: "WHILE", Cond: `{{run}} == "true"`, Value: []Step{{Type: "RETURN", Value: `"ReturnedFromWhile"`}}}, {Type: "SET", Target: "status", Value: `"Finished"`}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"run": "true"}, expectedResult: "ReturnedFromWhile", expectError: false},
		{name: "FOR EACH string char iteration", inputSteps: []Step{{Type: "SET", Target: "input", Value: `"Hi!"`}, {Type: "SET", Target: "output", Value: `""`}, {Type: "FOR", Target: "char", Cond: `{{input}}`, Value: []Step{{Type: "SET", Target: "output", Value: `{{output}} + {{char}} + "-"`}}}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"input": "Hi!", "output": "H-i-!-"}, expectedResult: nil, expectError: false},
		{name: "FOR EACH comma split fallback", inputSteps: []Step{{Type: "SET", Target: "input", Value: `"a, b ,c"`}, {Type: "SET", Target: "output", Value: `""`}, {Type: "FOR", Target: "item", Cond: `{{input}}`, Value: []Step{{Type: "SET", Target: "output", Value: `{{output}} + "-" + {{item}}`}}}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"input": "a, b ,c", "output": "-a-b-c"}, expectedResult: nil, expectError: false},
		{name: "FOR EACH string with RETURN", inputSteps: []Step{{Type: "SET", Target: "input", Value: `"stop"`}, {Type: "SET", Target: "output", Value: `""`}, {Type: "FOR", Target: "char", Cond: `{{input}}`, Value: []Step{{Type: "SET", Target: "output", Value: `{{output}} + {{char}}`}, {Type: "IF", Cond: `{{char}} == "o"`, Value: []Step{{Type: "RETURN", Value: `"Stopped at o"`}}}}}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"input": "stop", "output": "sto"}, expectedResult: "Stopped at o", expectError: false},
		{name: "CALL TOOL StringLength", inputSteps: []Step{{Type: "SET", Target: "myStr", Value: `"Hello 你好"`}, {Type: "CALL", Target: "TOOL.StringLength", Args: []string{"{{myStr}}"}}, {Type: "SET", Target: "lenResult", Value: `__last_call_result`}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"myStr": "Hello 你好", "lenResult": "8"}, expectedResult: nil, expectError: false},
		{name: "CALL TOOL Substring", inputSteps: []Step{{Type: "SET", Target: "myStr", Value: `"你好世界"`}, {Type: "CALL", Target: "TOOL.Substring", Args: []string{"{{myStr}}", `"1"`, `"3"`}}, {Type: "SET", Target: "subResult", Value: `__last_call_result`}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"myStr": "你好世界", "subResult": "好世"}, expectedResult: nil, expectError: false},
		{name: "CALL TOOL ToUpper/ToLower", inputSteps: []Step{{Type: "SET", Target: "myStr", Value: `"Test"`}, {Type: "CALL", Target: "TOOL.ToUpper", Args: []string{"{{myStr}}"}}, {Type: "SET", Target: "upper", Value: `__last_call_result`}, {Type: "CALL", Target: "TOOL.ToLower", Args: []string{"{{upper}}"}}, {Type: "SET", Target: "lower", Value: `__last_call_result`}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"myStr": "Test", "upper": "TEST", "lower": "test"}, expectedResult: nil, expectError: false},
		{name: "CALL TOOL TrimSpace", inputSteps: []Step{{Type: "SET", Target: "myStr", Value: `"  spaced \n"`}, {Type: "CALL", Target: "TOOL.TrimSpace", Args: []string{"{{myStr}}"}}, {Type: "SET", Target: "trimmed", Value: `__last_call_result`}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"myStr": "  spaced \n", "trimmed": "spaced"}, expectedResult: nil, expectError: false},
		{name: "CALL TOOL SplitString and JoinStrings", inputSteps: []Step{{Type: "SET", Target: "myStr", Value: `"a-b-c"`}, {Type: "CALL", Target: "TOOL.SplitString", Args: []string{"{{myStr}}", `"-"`}}, {Type: "SET", Target: "splitResult", Value: `__last_call_result`}, {Type: "CALL", Target: "TOOL.JoinStrings", Args: []string{"splitResult", `","`}}, {Type: "SET", Target: "joinedResult", Value: `__last_call_result`}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"myStr": "a-b-c", "splitResult": []string{"a", "b", "c"}, "joinedResult": "a,b,c"}, expectedResult: nil, expectError: false},
		{name: "CALL TOOL SplitWords", inputSteps: []Step{{Type: "SET", Target: "myStr", Value: `"  hello\t world \n"`}, {Type: "CALL", Target: "TOOL.SplitWords", Args: []string{"{{myStr}}"}}, {Type: "SET", Target: "words", Value: `__last_call_result`}, {Type: "CALL", Target: "TOOL.JoinStrings", Args: []string{"words", `" "`}}, {Type: "SET", Target: "joinedWords", Value: `__last_call_result`}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"myStr": "  hello\t world \n", "words": []string{"hello", "world"}, "joinedWords": "hello world"}, expectedResult: nil, expectError: false},
		{name: "CALL TOOL ReplaceAll", inputSteps: []Step{{Type: "SET", Target: "myStr", Value: `"one two one"`}, {Type: "CALL", Target: "TOOL.ReplaceAll", Args: []string{"{{myStr}}", `"one"`, `"three"`}}, {Type: "SET", Target: "replaced", Value: `__last_call_result`}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"myStr": "one two one", "replaced": "three two three"}, expectedResult: nil, expectError: false},
		{name: "CALL TOOL Contains", inputSteps: []Step{{Type: "SET", Target: "myStr", Value: `"hello world"`}, {Type: "CALL", Target: "TOOL.Contains", Args: []string{"{{myStr}}", `"world"`}}, {Type: "SET", Target: "containsResult", Value: `__last_call_result`}, {Type: "CALL", Target: "TOOL.Contains", Args: []string{"{{myStr}}", `"xyz"`}}, {Type: "SET", Target: "notContainsResult", Value: `__last_call_result`}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"myStr": "hello world", "containsResult": "true", "notContainsResult": "false"}, expectedResult: nil, expectError: false},
		{name: "CALL TOOL HasPrefix/HasSuffix", inputSteps: []Step{{Type: "SET", Target: "myStr", Value: `"start middle end"`}, {Type: "CALL", Target: "TOOL.HasPrefix", Args: []string{"{{myStr}}", `"start"`}}, {Type: "SET", Target: "prefixTrue", Value: `__last_call_result`}, {Type: "CALL", Target: "TOOL.HasPrefix", Args: []string{"{{myStr}}", `"middle"`}}, {Type: "SET", Target: "prefixFalse", Value: `__last_call_result`}, {Type: "CALL", Target: "TOOL.HasSuffix", Args: []string{"{{myStr}}", `"end"`}}, {Type: "SET", Target: "suffixTrue", Value: `__last_call_result`}, {Type: "CALL", Target: "TOOL.HasSuffix", Args: []string{"{{myStr}}", `"middle"`}}, {Type: "SET", Target: "suffixFalse", Value: `__last_call_result`}}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"myStr": "start middle end", "prefixTrue": "true", "prefixFalse": "false", "suffixTrue": "true", "suffixFalse": "false"}, expectedResult: nil, expectError: false},
		{
			name:           "CALL TOOL Substring Wrong Arg Type",
			inputSteps:     []Step{{Type: "CALL", Target: "TOOL.Substring", Args: []string{`"hello"`, `"one"`, `"three"`}}},
			initialVars:    map[string]interface{}{},
			expectedResult: nil, expectError: true,
			// ** FIX: Correct expected error string (remove inner single quotes) **
			errorContains: `expected int, but received "one" which cannot be converted to int`,
		},
	} // End testCases slice

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { runExecuteStepsTest(t, tc) })
	}
} // End TestExecuteStepsBlocksAndLoops
