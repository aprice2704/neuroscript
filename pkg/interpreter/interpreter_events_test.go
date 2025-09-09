// NeuroScript Version: 0.7.1
// File version: 5
// Purpose: Uses the canonical type-preserving unwrap helper from pkg/lang for validation.
// filename: pkg/interpreter/interpreter_events_test.go
// nlines: 100
// risk_rating: LOW

package interpreter_test

import (
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api/shape"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// setupEventTest creates an interpreter, loads a script with an event handler
// that emits its data, and captures the emitted value.
func setupEventTest(t *testing.T) (*interpreter.Interpreter, *lang.Value) {
	t.Helper()

	// The handler will emit the 'data' variable it receives.
	script := `
	on event "test_event" as data do
		emit data
	endon
	`
	interp, err := interpreter.NewTestInterpreter(t, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to create test interpreter: %v", err)
	}

	var capturedEmit lang.Value

	// Capture the first thing that gets emitted.
	interp.SetEmitFunc(func(v lang.Value) {
		if capturedEmit == nil {
			capturedEmit = v
		}
	})

	// Load the script to register the handler.
	parserAPI := parser.NewParserAPI(interp.GetLogger())
	p, pErr := parserAPI.Parse(script)
	if pErr != nil {
		t.Fatalf("Failed to parse script: %v", pErr)
	}
	program, _, bErr := parser.NewASTBuilder(interp.GetLogger()).Build(p)
	if bErr != nil {
		t.Fatalf("Failed to build AST: %v", bErr)
	}
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	return interp, &capturedEmit
}

func TestEmitEvent_CanonicalShape(t *testing.T) {
	t.Run("Pre-formatted canonical event is passed through", func(t *testing.T) {
		interp, capturedEmit := setupEventTest(t)

		// 1. Create a canonical event using the shape API.
		originalPayload, err := shape.ComposeNSEvent("user.login", map[string]interface{}{"user": "alice"}, nil)
		if err != nil {
			t.Fatalf("Failed to compose canonical event: %v", err)
		}
		wrappedPayload, _ := lang.Wrap(originalPayload)

		// 2. Emit it.
		interp.EmitEvent("test_event", "source_system", wrappedPayload)

		// 3. Assert the captured (emitted) event has the same content as the one we sent.
		if !reflect.DeepEqual(lang.UnwrapForShapeValidation(*capturedEmit), lang.UnwrapForShapeValidation(wrappedPayload)) {
			t.Errorf("Event was changed. Expected pass-through.\nGot:  %#v\nWant: %#v", lang.UnwrapForShapeValidation(*capturedEmit), lang.UnwrapForShapeValidation(wrappedPayload))
		}
	})

	t.Run("Raw payload is wrapped into a canonical event", func(t *testing.T) {
		interp, capturedEmit := setupEventTest(t)

		// 1. Create a raw payload.
		rawPayload := lang.NewMapValue(map[string]lang.Value{"user": lang.StringValue{Value: "bob"}})

		// 2. Emit it.
		interp.EmitEvent("test_event", "source_system", rawPayload)

		// 3. Assert the captured (emitted) event is now a valid canonical event.
		// CRITICAL: Use the type-preserving unwrapper before validating.
		unwrappedCaptured := lang.UnwrapForShapeValidation(*capturedEmit)
		capturedMap, ok := unwrappedCaptured.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected captured event to be a map, but got %T", unwrappedCaptured)
		}

		if err := shape.ValidateNSEvent(capturedMap, nil); err != nil {
			t.Errorf("The emitted event was not wrapped in a valid canonical shape: %v", err)
		}
	})
}
