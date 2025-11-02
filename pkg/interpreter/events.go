// NeuroScript Version: 0.8.0
// File version: 45
// Purpose: Fix bug where 'return' was disallowed in event handlers by correctly setting isInHandler to false.
// filename: pkg/interpreter/events.go
// nlines: 115
// risk_rating: MEDIUM

package interpreter

import (
	"fmt"
	"math"

	"github.com/aprice2704/neuroscript/pkg/api/shape"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// RegisterEventHandler is an exported test helper that allows an external
// caller (like a test harness) to register an event handler declaration.
// This is necessary to wire up the AST builder's callback during testing.
func (i *Interpreter) RegisterEventHandler(decl *ast.OnEventDecl) {
	// DEBUG: Confirm that this callback is being invoked by the parser/AST builder.
	i.Logger().Debug("[DEBUG] RegisterEventHandler: Callback invoked by ASTBuilder", "event_name", decl.EventNameExpr.String()) //

	if err := i.eventManager.register(decl, i); err != nil { //
		// In a test context, panicking is acceptable if setup fails.
		panic(fmt.Sprintf("test setup failed: could not register event handler: %v", err)) //
	}
}

// unwrapForShapeValidation is a private helper from the original implementation.
func unwrapForShapeValidation(v lang.Value) interface{} { //
	// ... (implementation remains the same)
	if v == nil { //
		return nil
	}
	switch tv := v.(type) { //
	case lang.NumberValue: //
		if tv.Value == math.Trunc(tv.Value) { //
			return int64(tv.Value) //
		}
		return tv.Value //
	case *lang.MapValue: //
		m := make(map[string]interface{}, len(tv.Value)) //
		for k, val := range tv.Value {                   //
			m[k] = unwrapForShapeValidation(val) //
		}
		return m //
	case lang.ListValue: //
		l := make([]interface{}, len(tv.Value)) //
		for i, val := range tv.Value {          //
			l[i] = unwrapForShapeValidation(val) //
		}
		return l //
	default: //
		return lang.Unwrap(v) //
	}
}

func (i *Interpreter) EmitEvent(eventName string, source string, payload lang.Value) { //
	i.eventManager.eventHandlersMu.RLock()              //
	handlers := i.eventManager.eventHandlers[eventName] //
	i.eventManager.eventHandlersMu.RUnlock()            //

	if len(handlers) == 0 { //
		i.Logger().Warn("Event emitted but no handlers were registered for it", "event_name", eventName, "source", source) //
		return                                                                                                             //
	}

	// FAIL-FAST CONTRACT: The host MUST provide the I/O functions in the context.
	if i.hostContext.EmitFunc == nil || i.hostContext.WhisperFunc == nil { //
		panic(fmt.Sprintf( //
			"FATAL: Interpreter (ID: %s) has event handlers for '%s' but is missing I/O functions in its HostContext. The host must configure them.", //
			i.id,      //
			eventName, //
		))
	}

	eventObj, err := i.composeCanonicalEvent(eventName, source, payload) //
	if err != nil {                                                      //
		i.Logger().Error("Failed to prepare canonical event", "event", eventName, "error", err) //
		return                                                                                  //
	}

	for _, handler := range handlers { //
		// --- THE FIX: Wrap execution in defer/recover ---
		func(h *ast.OnEventDecl) { // Create closure to capture 'h'
			var execErr error
			handlerInterpreter := i.fork() // Fork a clean interpreter for the handler //

			defer func() {
				if r := recover(); r != nil {
					// Convert panic to a RuntimeError
					panicMsg := fmt.Sprintf("panic executing event handler for '%s': %v", eventName, r)
					execErr = lang.NewRuntimeError(lang.ErrorCodeInternal, panicMsg, fmt.Errorf("panic: %v", r)).WithPosition(h.GetPos())
					i.Logger().Error("Panic recovered in event handler", "event", eventName, "source", source, "panic_value", r, "error", execErr)
				}

				// Report error (either from panic or normal execution) via callback
				if execErr != nil {
					rtErr := ensureRuntimeError(execErr, h.GetPos(), "ON_EVENT_HANDLER") //
					if i.hostContext.EventHandlerErrorCallback != nil {                  //
						i.hostContext.EventHandlerErrorCallback(eventName, source, rtErr) //
					} else {
						// Log if no callback is registered, as this would otherwise be silent.
						i.Logger().Error("Unhandled error in event handler (no callback registered)", "event", eventName, "source", source, "error", rtErr)
					}
				}
			}()

			if h.EventVarName != "" { //
				handlerInterpreter.SetVariable(h.EventVarName, eventObj) //
			}

			// Execute the handler steps
			// --- FIX: Set isInHandler to false. It should only be true
			// --- when executing the *body* of an on_error block,
			// --- not the entire event handler.
			_, _, _, execErr = handlerInterpreter.executeSteps(h.Body, false, nil) //

		}(handler) // Pass handler into the closure
		// --- End Fix ---
	}
}

// composeCanonicalEvent ensures the payload is wrapped in the standard event shape.
func (i *Interpreter) composeCanonicalEvent(eventName, source string, payload lang.Value) (lang.Value, error) { //
	// ... (implementation remains the same)
	unwrappedPayload := unwrapForShapeValidation(payload)                //
	if payloadMap, ok := unwrappedPayload.(map[string]interface{}); ok { //
		if shape.ValidateNSEvent(payloadMap, nil) == nil { //
			return payload, nil //
		}
	}

	payloadToCompose, _ := unwrappedPayload.(map[string]interface{})                                                  //
	composed, err := shape.ComposeNSEvent(eventName, payloadToCompose, &shape.NSEventComposeOptions{AgentID: source}) //
	if err != nil {                                                                                                   //
		return nil, fmt.Errorf("failed to compose canonical event: %w", err) //
	}

	return lang.Wrap(composed) //
}
