// NeuroScript Version: 0.5.2
// File version: 10
// Purpose: Refactored script to use an intermediate variable for map access to avoid parser limitations with chained access.
// filename: pkg/core/evaluation_event_handler_test.go
// nlines: 137
// risk_rating: LOW

package core

import (
	"testing"
)

// setupEventHandlerTest now uses the new LoadProgram method.
func setupEventHandlerTest(t *testing.T, script string) *Interpreter {
	t.Helper()
	logger := NewTestLogger(t)
	interp, err := NewInterpreter(logger, nil, ".", nil, nil)
	if err != nil {
		t.Fatalf("Failed to create new interpreter: %v", err)
	}

	parser := NewParserAPI(logger)
	parseTree, parseErr := parser.Parse(script)
	if parseErr != nil {
		t.Fatalf("Failed to parse script: %v", parseErr)
	}

	astBuilder := NewASTBuilder(logger)
	prog, syntaxErr, validationErr := astBuilder.Build(parseTree)
	if syntaxErr != nil {
		t.Fatalf("Syntax error found during AST build: %v", syntaxErr)
	}
	if validationErr != nil {
		t.Fatalf("Validation error found during AST build: %v", validationErr)
	}

	// Load the entire program (procs and events) at once.
	if err := interp.LoadProgram(prog); err != nil {
		t.Fatalf("Failed to load program into interpreter: %v", err)
	}

	return interp
}

func TestOnEventHandling(t *testing.T) {
	t.Run("Basic event handler sets variable from payload", func(t *testing.T) {
		// Corrected Script: Using an intermediate variable to work around chained access limitations.
		script := `
		on event "user_login" as data
			set payload_map = data["Payload"]
			set login_name = payload_map["username"]
		endevent

		func main() means
		endfunc
		`
		interp := setupEventHandlerTest(t, script)
		payload := NewMapValue(map[string]Value{"username": StringValue{Value: "testuser"}})
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
		// This script is already syntactically correct.
		script := `
		on event "test_event" as e1
			set var_a = 1
		endevent

		on event "test_event" as e2
			set var_b = 2
		endevent
		
		func main() means
		endfunc
		`
		interp := setupEventHandlerTest(t, script)
		interp.EmitEvent("test_event", "test", nil)

		valA, _ := interp.GetVariable("var_a")
		if numA, ok := valA.(NumberValue); !ok || numA.Value != 1 {
			t.Errorf("Expected var_a to be 1, got %v", valA)
		}

		valB, _ := interp.GetVariable("var_b")
		if numB, ok := valB.(NumberValue); !ok || numB.Value != 2 {
			t.Errorf("Expected var_b to be 2, got %v", valB)
		}
	})

	t.Run("Event name must be a static string", func(t *testing.T) {
		// This script is correct for testing the validation logic.
		script := `
		func main() means
			set my_event = "some_event"
		endfunc

		on event my_event as e
			set x = 1
		endevent
		`
		logger := NewTestLogger(t)
		parser := NewParserAPI(logger)
		parseTree, parseErr := parser.Parse(script)
		if parseErr != nil {
			t.Fatalf("Failed to parse script: %v", parseErr)
		}

		// The error is expected from the AST builder's validation phase.
		astBuilder := NewASTBuilder(logger)
		_, syntaxErr, validationErr := astBuilder.Build(parseTree)
		if syntaxErr != nil {
			t.Fatalf("Unexpected syntax error during AST build: %v", syntaxErr)
		}
		if validationErr == nil {
			t.Fatal("Expected a validation error for non-static event name, but got nil")
		}
	})
}
