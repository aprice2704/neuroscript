// filename: pkg/runtime/evaluation_event_handler_test.go
// NeuroScript Version: 0.5.2
// File version: 9
// Purpose: Updated empty function blocks in tests to be syntactically valid.
// nlines: 135+
// risk_rating: LOW

package runtime

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// setupEventHandlerTest parses `script`, builds its AST, loads it into a fresh
// Interpreter, and returns the ready Interpreter for use in assertions.
func setupEventHandlerTest(t *testing.T, script string) (*Interpreter, error) {
	t.Helper()

	logger := llm.NewTestLogger(t)

	interp, err := NewInterpreter(logger, nil, ".", nil, nil)
	if err != nil {
		t.Fatalf("Failed to create new interpreter: %v", err)
	}

	parser := parser.NewParserAPI(logger)
	parseTree, parseErr := parser.Parse(script)
	if parseErr != nil {
		t.Fatalf("Failed to parse script: %v", parseErr)
	}

	astBuilder := parser.NewASTBuilder(logger)
	prog, _, err := astBuilder.Build(parseTree)
	if err != nil {
		t.Fatalf("Failed to build AST: %v", err)
	}

	if err := interp.LoadProgram(prog); err != nil {
		t.Fatalf("Failed to load program into interpreter: %v", err)
	}

	return interp, nil
}

func TestOnEventHandling(t *testing.T) {
	t.Run("Basic event handler sets variable from payload", func(t *testing.T) {
		script := `
			on event "user_login" as data do
				set payload_map = data["payload"]
				set login_name = payload_map["username"]
			endon

			func main() means
				set _ = nil
			endfunc
			`

		interp, err := setupEventHandlerTest(t, script)
		if err != nil {
			t.Fatal(err)
		}

		payload := lang.NewMapValue(map[string]lang.Value{"username": lang.StringValue{Value: "testuser"}})
		interp.EmitEvent("user_login", "auth_system", payload)

		val, exists := interp.GetVariable("login_name")
		if !exists {
			t.Fatal("Variable 'login_name' was not set by the event handler")
		}
		if strVal, ok := val.(StringValue); !ok || strVal.Value != "testuser" {
			t.Errorf("Expected login_name to be 'testuser', got %v", val)
		}
	})

	t.Run("Multiple handlers for the same event", func(t *testing.T) {
		script := `
			on event "test_event" as e1 do
				set var_a = 1
			endon

			on event "test_event" as e2 do
				set var_b = 2
			endon
			
			func main() means
				set _ = nil
			endfunc
			`

		interp, err := setupEventHandlerTest(t, script)
		if err != nil {
			t.Fatal(err)
		}
		interp.EmitEvent("test_event", "test", nil)

		t.Log("Dumping variables after event emission to check state:")
		testutil.DebugDumpVariables(interp, t)

		valA, _ := interp.GetVariable("var_a")
		if numA, ok := valA.(NumberValue); !ok || numA.Value != 1 {
			t.Errorf("Expected var_a to be 1, got %v", valA)
		}

		valB, _ := interp.GetVariable("var_b")
		if numB, ok := valB.(NumberValue); !ok || numB.Value != 2 {
			t.Errorf("Expected var_b to be 2, got %v", valB)
		}
	})

	t.Run("Event name can be a variable (dynamic)", func(t *testing.T) {
		script := `
			func main() means
				set my_event = "some_event"
			endfunc

			on event my_event as e do
				set x = 1
			endon
			`

		logger := llm.NewTestLogger(t)
		parser := parser.NewParserAPI(logger)
		parseTree, parseErr := parser.Parse(script)
		if parseErr != nil {
			t.Fatalf("Failed to parse script: %v", parseErr)
		}

		astBuilder := parser.NewASTBuilder(logger)
		_, _, err := astBuilder.Build(parseTree)

		if err != nil {
			t.Fatalf("Expected AST build to succeed for dynamic event name, but it failed: %v", err)
		}
	})
}
