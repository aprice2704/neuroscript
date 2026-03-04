// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 8
// :: description: Tests evaluation of literals. Fixed newline interpolation in TestInterpolatedStrings.
// :: latestChange: Updated TestInterpolatedStrings to use {{@nl}} and {{@tbt}} for reliability and UI safety.
// :: filename: pkg/interpreter/literals_test.go
// :: serialization: go

package interpreter_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// runLiteralTest returns result, nil to satisfy Go's error interface vs typed-nil concrete pointers.
func runLiteralTest(t *testing.T, scriptExpression string) (lang.Value, error) {
	t.Helper()
	h := NewTestHarness(t)
	script := fmt.Sprintf("func main(returns result) means\n\treturn %s\nendfunc", scriptExpression)
	result, runErr := h.Interpreter.ExecuteScriptString("main", script, nil)

	if runErr == nil {
		return result, nil
	}
	return result, runErr
}

func TestEvaluateLiterals(t *testing.T) {
	// Using hex escapes for backticks to keep the Go source syntactically clear
	// and avoid UI rendering issues.
	bt3 := "\x60\x60\x60"

	testCases := []struct {
		name   string
		script string
		want   lang.Value
	}{
		{"Integer", "123", lang.NumberValue{Value: 123}},
		{"Float", "123.456", lang.NumberValue{Value: 123.456}},
		{"Char function", "char(65)", lang.StringValue{Value: "A"}},
		{"Ord function", "ord('A')", lang.NumberValue{Value: 65}},
		{"Triple Backtick", bt3 + "raw string" + bt3, lang.StringValue{Value: "raw string"}},
		{"Triple Single Quote", "'''raw content'''", lang.StringValue{Value: "raw content"}},
		{"Double Bracket", "[[raw content]]", lang.StringValue{Value: "raw content"}},
		{"Interpolation Constants", "[[a {{@nl}} b {{@tbt}}]]", lang.StringValue{Value: "a \n b ```"}},
		{"Nil", "nil", &lang.NilValue{}},
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

func TestInterpolatedStrings(t *testing.T) {
	h := NewTestHarness(t)

	// We use the new interpolation symbols {{@tbt}} and {{@nl}} inside the script.
	// This ensures the backticks and newlines are generated correctly by the parser
	// regardless of Go string escaping rules.
	script := "func main(returns result) means\n" +
		"\tset name = \"Alice\"\n" +
		"\tset numVar = 42\n" +
		"\tset result = [[Hello {{name}}! {{@tbt}}python{{@nl}}print(\"{{numVar}}\"){{@nl}}{{@tbt}}]]\n" +
		"\treturn result\n" +
		"endfunc"

	result, err := h.Interpreter.ExecuteScriptString("main", script, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// The expected value should contain actual newline characters (0x0A) and backticks.
	expected := "Hello Alice! ```python\nprint(\"42\")\n```"

	if got, ok := result.(lang.StringValue); !ok || got.Value != expected {
		t.Errorf("Result mismatch.\nWant: %q\nGot:  %#v", expected, result)
	}
}
