// NeuroScript Version: 0.8.0
// File version: 1.0.0
// Purpose: Test to prevent regression on 'return' statements in 'on event' handlers.
// filename: pkg/interpreter/events_return_test.go
// nlines: 66
// risk_rating: LOW

package interpreter_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// TestEventReturnRegression verifies that a 'return' statement inside an
// 'on event' handler does NOT incorrectly trigger an ErrReturnViolation.
// This test prevents regressions on the bug fixed in events.go (File version 45)
// where the 'isInHandler' flag was incorrectly set to true for the entire
// event handler body.
func TestEventReturnRegression(t *testing.T) {
	script := `
on event "test.event.return" as ev do
  emit "Handler started."
  return "I am returning from an event."
  emit "This line should never be reached."
endon
`
	t.Logf("[DEBUG] Turn 1: Starting TestEventReturnRegression.")
	h := NewTestHarness(t)
	interp := h.Interpreter

	// Setup a variable to capture any error from the event handler
	var handlerErr error
	var mu sync.Mutex

	// Configure the error callback. This is where the bug would surface.
	interp.HostContext().EventHandlerErrorCallback = func(eventName, source string, err *lang.RuntimeError) {
		mu.Lock()
		defer mu.Unlock()
		handlerErr = err
		t.Logf("[DEBUG] EventHandlerErrorCallback was triggered: %v", err)
	}

	// Parse and load the script [cite: 3]
	tree, pErr := h.Parser.Parse(script)
	if pErr != nil {
		t.Fatalf("Failed to parse script: %v", pErr)
	}
	program, _, bErr := h.ASTBuilder.Build(tree)
	if bErr != nil {
		t.Fatalf("Failed to build AST: %v", bErr)
	}
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}
	t.Logf("[DEBUG] Turn 2: Script parsed and loaded with 'on event' handler.")

	// --- Execute ---
	// Emit the event that triggers the handler.
	interp.EmitEvent("test.event.return", "test-source", &lang.NilValue{})
	t.Logf("[DEBUG] Turn 3: EmitEvent completed.")

	// --- Assert ---
	mu.Lock()
	defer mu.Unlock()

	if handlerErr != nil {
		// This is the failure condition.
		if errors.Is(handlerErr, lang.ErrReturnViolation) { //
			t.Fatalf("REGRESSION DETECTED: 'return' in event handler incorrectly caused ErrReturnViolation: %v", handlerErr)
		}
		t.Fatalf("Event handler returned an unexpected error: %v", handlerErr)
	}

	t.Logf("[DEBUG] Turn 4: Assertion passed. No error was captured by the callback.")
}
