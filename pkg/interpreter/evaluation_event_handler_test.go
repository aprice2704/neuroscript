// filename: pkg/interpreter/evaluation_event_handler_test.go
// NeuroScript Version: 0.5.2
// File version: 13
// Purpose: Corrected calls to interp.Load to pass the correct AST structure.
// nlines: 135+
// risk_rating: LOW

package interpreter

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// setupEventHandlerTest now returns the output buffer as well.
func setupEventHandlerTest(t *testing.T, script string) (*Interpreter, *bytes.Buffer, error) {
	t.Helper()

	logger := logging.NewTestLogger(t)
	var outputBuffer bytes.Buffer

	// FIX: Use SetEmitFunc to capture the output of 'emit' statements.
	// The 'emit' statement does not write to the interpreter's stdout by default.
	interp := NewInterpreter(WithLogger(logger))
	interp.SetEmitFunc(func(v lang.Value) {
		fmt.Fprintln(&outputBuffer, v.String())
	})

	parserAPI := parser.NewParserAPI(logger)
	parseTree, parseErr := parserAPI.Parse(script)
	if parseErr != nil {
		t.Fatalf("Failed to parse script: %v", parseErr)
	}

	astBuilder := parser.NewASTBuilder(logger)
	prog, _, err := astBuilder.Build(parseTree)
	if err != nil {
		t.Fatalf("Failed to build AST: %v", err)
	}

	if err := interp.Load(&interfaces.Tree{Root: prog}); err != nil {
		t.Fatalf("Failed to load program into interpreter: %v", err)
	}

	return interp, &outputBuffer, nil
}

func TestOnEventHandling(t *testing.T) {
	t.Run("Basic event handler emits variable from payload", func(t *testing.T) {
		// FIX: The handler now emits the result to prove it worked internally.
		script := `
			on event "user_login" as data do
				set payload_map = data["payload"]
				set login_name = payload_map["username"]
				emit login_name
			endon

			func main() means
				set _ = nil
			endfunc
			`

		interp, stdout, err := setupEventHandlerTest(t, script)
		if err != nil {
			t.Fatal(err)
		}

		payload := lang.NewMapValue(map[string]lang.Value{"username": lang.StringValue{Value: "testuser"}})
		interp.EmitEvent("user_login", "auth_system", payload)

		// FIX: Check the output buffer, not the interpreter's variables.
		output := strings.TrimSpace(stdout.String())
		if output != "testuser" {
			t.Errorf("Expected event handler to emit 'testuser', got '%s'", output)
		}

		// Also assert that the variable did NOT leak into the parent scope.
		_, exists := interp.GetVariable("login_name")
		if exists {
			t.Fatal("Variable 'login_name' leaked from sandboxed event handler into the parent interpreter")
		}
	})

	t.Run("Multiple handlers for the same event", func(t *testing.T) {
		// FIX: Handlers now emit their results.
		script := `
			on event "test_event" as e1 do
				set var_a = 1
				emit "handler_a_ran"
			endon

			on event "test_event" as e2 do
				set var_b = 2
				emit "handler_b_ran"
			endon
			
			func main() means
				set _ = nil
			endfunc
			`

		interp, stdout, err := setupEventHandlerTest(t, script)
		if err != nil {
			t.Fatal(err)
		}
		interp.EmitEvent("test_event", "test", nil)

		// FIX: Check the output buffer for proof of execution.
		output := stdout.String()
		if !strings.Contains(output, "handler_a_ran") {
			t.Error("Did not find expected output from first event handler")
		}
		if !strings.Contains(output, "handler_b_ran") {
			t.Error("Did not find expected output from second event handler")
		}

		// Assert that variables did not leak.
		_, existsA := interp.GetVariable("var_a")
		_, existsB := interp.GetVariable("var_b")
		if existsA || existsB {
			t.Error("Variable(s) leaked from sandboxed event handlers")
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

		logger := logging.NewTestLogger(t)
		parserAPI := parser.NewParserAPI(logger)
		parseTree, parseErr := parserAPI.Parse(script)
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
