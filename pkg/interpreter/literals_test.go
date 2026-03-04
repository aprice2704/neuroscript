// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 10
// :: description: Tests evaluation of literals. Added severe interpolation test case.
// :: latestChange: Fixed test expectation to match correct runtime string coercion for map spacing and nil values.
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
		{"Interpolation Constants", "[[a {{@nl}} b {{@tbt}}]]", lang.StringValue{Value: "a \n b " + bt3}},
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
	expected := "Hello Alice! \x60\x60\x60python\nprint(\"42\")\n\x60\x60\x60"

	if got, ok := result.(lang.StringValue); !ok || got.Value != expected {
		t.Errorf("Result mismatch.\nWant: %q\nGot:  %#v", expected, result)
	}
}

func TestSevereInterpolations(t *testing.T) {
	h := NewTestHarness(t)

	script := `
func main(returns result) means
	set num = 3.14
	set flag = true
	set obj = {"k": "v"}
	set result = [[Num:{{num}}|Flag:{{flag}}|Obj:{{obj}}|Miss:{{missing}}|Symbols:{{@nl}}{{@cr}}{{@tab}}{{@bt}}{{@tbt}}{{@sq}}{{@dq}}{{@tsq}}{{@tdq}}]]
	return result
endfunc`

	result, err := h.Interpreter.ExecuteScriptString("main", script, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// lang.ToString handles formatting of complex objects and primitives.
	// 'missing' resolves to a nil value, which prints as "" when coerced to string.
	// Maps use a space after the colon in default stringification.
	expected := "Num:3.14|Flag:true|Obj:{\"k\": \"v\"}|Miss:|Symbols:\n\r\t\x60\x60\x60\x60'\x22'''\x22\x22\x22"

	if got, ok := result.(lang.StringValue); !ok || got.Value != expected {
		t.Errorf("Result mismatch.\nWant: %q\nGot:  %q", expected, result)
	}
}
