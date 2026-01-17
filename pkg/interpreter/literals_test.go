// NeuroScript Version: 0.9.6
// File version: 2
// Purpose: Tests evaluation of literals. Fixed NilValue expectation to be a pointer.
// filename: pkg/interpreter/literals_test.go
// serialization: go

package interpreter_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// runLiteralTest is a helper for this file, using the TestHarness to run a single literal expression.
func runLiteralTest(t *testing.T, scriptExpression string) (lang.Value, error) {
	t.Helper()
	h := NewTestHarness(t)
	script := fmt.Sprintf("func main(returns result) means\n\treturn %s\nendfunc", scriptExpression)
	result, runErr := h.Interpreter.ExecuteScriptString("main", script, nil)

	// Explicitly check for a nil concrete error and return a true nil interface.
	if runErr == nil {
		return result, nil
	}
	return result, runErr
}

func TestEvaluateLiterals(t *testing.T) {
	testCases := []struct {
		name   string
		script string
		want   lang.Value
	}{
		// Numbers
		{"Integer", "123", lang.NumberValue{Value: 123}},
		{"Float", "123.456", lang.NumberValue{Value: 123.456}},
		{"Negative", "-5", lang.NumberValue{Value: -5}},

		// Strings - Standard
		{"Double Quoted", `"hello world"`, lang.StringValue{Value: "hello world"}},
		{"Single Quoted", `'hello world'`, lang.StringValue{Value: "hello world"}},
		{"Escaped Double Quote", `"say \"hello\""`, lang.StringValue{Value: `say "hello"`}},
		// Note: Single quote un-escaping logic is in the parser, so runtime just sees the result
		{"Escaped Single Quote", `'it\'s me'`, lang.StringValue{Value: "it's me"}},

		// Strings - Raw (Triple Backtick)
		{"Triple Backtick", "```raw string```", lang.StringValue{Value: "raw string"}},
		{"Triple Backtick Multiline", "```line1\nline2```", lang.StringValue{Value: "line1\nline2"}},

		// Strings - Raw (Triple Single Quote - NEW FEATURE)
		{"Triple Single Quote", "'''raw content'''", lang.StringValue{Value: "raw content"}},
		{"Triple Single Quote Multiline", "'''line A\nline B'''", lang.StringValue{Value: "line A\nline B"}},

		// CRITICAL TEST CASE: Nested Backticks
		// This verifies that we can now safely embed triple backticks inside triple single quotes.
		{"Nested Backticks", "'''contains ```json { \"key\": \"val\" } ``` inside'''", lang.StringValue{Value: "contains ```json { \"key\": \"val\" } ``` inside"}},

		// Booleans
		{"True", "true", lang.BoolValue{Value: true}},
		{"False", "false", lang.BoolValue{Value: false}},

		// Nil
		// FIX: Expect a pointer to NilValue, as that is what the interpreter returns.
		{"Nil", "nil", &lang.NilValue{}},

		// Collections (Basic Evaluation)
		{"List Literal", "[1, 2, 3]", lang.ListValue{Value: []lang.Value{
			lang.NumberValue{Value: 1},
			lang.NumberValue{Value: 2},
			lang.NumberValue{Value: 3},
		}}},
		{"Map Literal", `{"a": 1}`, lang.MapValue{Value: map[string]lang.Value{
			"a": lang.NumberValue{Value: 1},
		}}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := runLiteralTest(t, tc.script)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Result mismatch.\nWant: %#v\nGot:  %#v", tc.want, got)
			}
		})
	}
}
