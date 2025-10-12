// NeuroScript Version: 0.7.1
// File version: 6
// Purpose: Refactored to use the centralized TestHarness for robust and consistent interpreter initialization.
// filename: pkg/interpreter/interpreter_events_test.go
// nlines: 105
// risk_rating: LOW

package interpreter_test

import (
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api/shape"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// setupEventTest creates an interpreter, loads a script with an event handler
// that emits its data, and captures the emitted value.
func setupEventTest(t *testing.T) (*interpreter.Interpreter, *lang.Value) {
	t.Helper()
	h := NewTestHarness(t)
	t.Logf("[DEBUG] Turn 1 (setupEventTest): Test harness created.")

	script := `
	on event "test_event" as data do
		emit data
	endon
	`
	var capturedEmit lang.Value

	h.HostContext.EmitFunc = func(v lang.Value) {
		t.Logf("[DEBUG] Turn X (setupEventTest): EmitFunc captured: %#v", v)
		if capturedEmit == nil {
			capturedEmit = v
		}
	}
	t.Logf("[DEBUG] Turn 2 (setupEventTest): HostContext configured.")

	tree, pErr := h.Parser.Parse(script)
	if pErr != nil {
		t.Fatalf("Failed to parse script: %v", pErr)
	}
	program, _, bErr := h.ASTBuilder.Build(tree)
	if bErr != nil {
		t.Fatalf("Failed to build AST: %v", bErr)
	}
	if err := h.Interpreter.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}
	t.Logf("[DEBUG] Turn 3 (setupEventTest): Script loaded.")

	return h.Interpreter, &capturedEmit
}

func TestEmitEvent_CanonicalShape(t *testing.T) {
	t.Run("Pre-formatted canonical event is passed through", func(t *testing.T) {
		t.Logf("[DEBUG] Turn 1: Starting 'Pre-formatted canonical event' test.")
		interp, capturedEmit := setupEventTest(t)

		originalPayload, err := shape.ComposeNSEvent("user.login", map[string]interface{}{"user": "alice"}, nil)
		if err != nil {
			t.Fatalf("Failed to compose canonical event: %v", err)
		}
		wrappedPayload, _ := lang.Wrap(originalPayload)
		t.Logf("[DEBUG] Turn 2: Original payload created.")

		interp.EmitEvent("test_event", "source_system", wrappedPayload)
		t.Logf("[DEBUG] Turn 3: Event emitted.")

		if !reflect.DeepEqual(lang.UnwrapForShapeValidation(*capturedEmit), lang.UnwrapForShapeValidation(wrappedPayload)) {
			t.Errorf("Event was changed. Expected pass-through.\nGot:  %#v\nWant: %#v", lang.UnwrapForShapeValidation(*capturedEmit), lang.UnwrapForShapeValidation(wrappedPayload))
		}
		t.Logf("[DEBUG] Turn 4: Assertion passed.")
	})

	t.Run("Raw payload is wrapped into a canonical event", func(t *testing.T) {
		t.Logf("[DEBUG] Turn 1: Starting 'Raw payload is wrapped' test.")
		interp, capturedEmit := setupEventTest(t)

		rawPayload := lang.NewMapValue(map[string]lang.Value{"user": lang.StringValue{Value: "bob"}})
		t.Logf("[DEBUG] Turn 2: Raw payload created.")

		interp.EmitEvent("test_event", "source_system", rawPayload)
		t.Logf("[DEBUG] Turn 3: Event emitted.")

		unwrappedCaptured := lang.UnwrapForShapeValidation(*capturedEmit)
		capturedMap, ok := unwrappedCaptured.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected captured event to be a map, but got %T", unwrappedCaptured)
		}

		if err := shape.ValidateNSEvent(capturedMap, nil); err != nil {
			t.Errorf("The emitted event was not wrapped in a valid canonical shape: %v", err)
		}
		t.Logf("[DEBUG] Turn 4: Assertion passed.")
	})
}
