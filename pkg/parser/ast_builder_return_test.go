// pkg/parser/ast_builder_return_test.go

package parser

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

func TestReturnStatement(t *testing.T) {
	t.Run("multiple return values are in correct order", func(t *testing.T) {
		script := `
            func main() means
                return "a", 1, true
            endfunc
        `
		logger := logging.NewTestLogger(t)
		parserAPI := NewParserAPI(logger)
		tree, err := parserAPI.Parse(script)
		if err != nil {
			t.Fatalf("Parse() failed: %v", err)
		}

		builder := NewASTBuilder(logger)
		program, _, err := builder.Build(tree)
		if err != nil {
			t.Fatalf("Build() failed: %v", err)
		}

		mainProc, ok := program.Procedures["main"]
		if !ok {
			t.Fatal("main procedure not found")
		}

		if len(mainProc.Steps) != 1 {
			t.Fatalf("expected 1 step, got %d", len(mainProc.Steps))
		}

		returnStep := mainProc.Steps[0]
		if returnStep.Type != "return" {
			t.Fatalf("expected return step, got %s", returnStep.Type)
		}

		if len(returnStep.Values) != 3 {
			t.Fatalf("expected 3 return values, got %d", len(returnStep.Values))
		}

		expectedValues := []ast.Expression{
			&ast.StringLiteralNode{Value: "a"},
			&ast.NumberLiteralNode{Value: int64(1)},
			&ast.BooleanLiteralNode{Value: true},
		}

		for i, val := range returnStep.Values {
			// Can't do a deep equal on the position, so we check the type and value
			switch expected := expectedValues[i].(type) {
			case *ast.StringLiteralNode:
				if actual, ok := val.(*ast.StringLiteralNode); ok {
					if expected.Value != actual.Value {
						t.Errorf("Expected return value %d to be %s, got %s", i, expected.Value, actual.Value)
					}
				} else {
					t.Errorf("Expected return value %d to be a string, got %T", i, val)
				}
			case *ast.NumberLiteralNode:
				if actual, ok := val.(*ast.NumberLiteralNode); ok {
					// FIX: Make comparison robust to underlying numeric types (e.g. int vs int64 vs float64).
					// The parser produces a number type that doesn't DeepEqual int64, so we cast to compare the value.
					var actualAsInt64 int64
					switch v := actual.Value.(type) {
					case int64:
						actualAsInt64 = v
					case int:
						actualAsInt64 = int64(v)
					case float64:
						actualAsInt64 = int64(v)
					default:
						t.Fatalf("unhandled actual numeric type in test: %T", v)
					}

					if expectedNode, ok := expected.Value.(int64); ok {
						if expectedNode != actualAsInt64 {
							t.Errorf("Expected return value %d to be %v, got %v", i, expectedNode, actual.Value)
						}
					} else {
						t.Fatalf("unhandled expected numeric type in test: %T", expected.Value)
					}
				} else {
					t.Errorf("Expected return value %d to be a number, got %T", i, val)
				}
			case *ast.BooleanLiteralNode:
				if actual, ok := val.(*ast.BooleanLiteralNode); ok {
					if expected.Value != actual.Value {
						t.Errorf("Expected return value %d to be %t, got %t", i, expected.Value, actual.Value)
					}
				} else {
					t.Errorf("Expected return value %d to be a boolean, got %T", i, val)
				}
			default:
				t.Errorf("unhandled expected type: %T", expected)
			}
		}
	})
}
