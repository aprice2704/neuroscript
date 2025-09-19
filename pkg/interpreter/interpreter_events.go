// NeuroScript Version: 0.7.1
// File version: 39
// Purpose: Implements FIXME item to log a warning when an event is emitted with no registered handlers.
// filename: pkg/interpreter/interpreter_events.go
// nlines: 90
// risk_rating: HIGH

package interpreter

import (
	"fmt"
	"math"

	"github.com/aprice2704/neuroscript/pkg/api/shape"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// unwrapForShapeValidation converts a lang.Value to an interface{} suitable for
// the shape validator, crucially preserving integer types where possible.
func unwrapForShapeValidation(v lang.Value) interface{} {
	if v == nil {
		return nil
	}
	switch tv := v.(type) {
	case lang.NumberValue:
		if tv.Value == math.Trunc(tv.Value) {
			return int64(tv.Value)
		}
		return tv.Value
	case *lang.MapValue:
		m := make(map[string]interface{}, len(tv.Value))
		for k, val := range tv.Value {
			m[k] = unwrapForShapeValidation(val)
		}
		return m
	case lang.ListValue:
		l := make([]interface{}, len(tv.Value))
		for i, val := range tv.Value {
			l[i] = unwrapForShapeValidation(val)
		}
		return l
	default:
		return lang.Unwrap(v)
	}
}

func (i *Interpreter) EmitEvent(eventName string, source string, payload lang.Value) {
	i.eventManager.eventHandlersMu.RLock()
	handlers := i.eventManager.eventHandlers[eventName]
	i.eventManager.eventHandlersMu.RUnlock()

	if len(handlers) == 0 {
		i.logger.Warn("Event emitted but no handlers were registered for it", "event_name", eventName, "source", source)
		return
	}

	// FAIL-FAST CONTRACT: If handlers are registered, the host MUST provide the I/O functions
	// for them to communicate. A nil function indicates a host-level configuration error.
	if i.customEmitFunc == nil || i.customWhisperFunc == nil {
		panic(fmt.Sprintf(
			"FATAL: Interpreter (ID: %s) has event handlers for '%s' but is missing custom I/O functions. The host must configure them via SetEmitFunc/SetWhisperFunc to capture handler output.",
			i.id,
			eventName,
		))
	}

	eventObj, err := i.composeCanonicalEvent(eventName, source, payload)
	if err != nil {
		i.logger.Error("Failed to prepare canonical event", "event", eventName, "error", err)
		return
	}

	for _, handler := range handlers {
		handlerInterpreter := i.CloneForEventHandler()
		handlerInterpreter.customEmitFunc = i.customEmitFunc
		handlerInterpreter.customWhisperFunc = i.customWhisperFunc

		if handler.EventVarName != "" {
			handlerInterpreter.SetVariable(handler.EventVarName, eventObj)
		}

		_, _, _, execErr := handlerInterpreter.executeSteps(handler.Body, true, nil)

		if execErr != nil {
			// This is the single, robust error reporting mechanism.
			// The error is sent directly to the host via the out-of-band callback.
			rtErr := ensureRuntimeError(execErr, handler.GetPos(), "ON_EVENT_HANDLER")
			if i.eventHandlerErrorCallback != nil {
				i.eventHandlerErrorCallback(eventName, source, rtErr)
			}
		}
	}
}

// composeCanonicalEvent ensures the payload is wrapped in the standard event shape.
func (i *Interpreter) composeCanonicalEvent(eventName, source string, payload lang.Value) (lang.Value, error) {
	unwrappedPayload := unwrapForShapeValidation(payload)
	if payloadMap, ok := unwrappedPayload.(map[string]interface{}); ok {
		if shape.ValidateNSEvent(payloadMap, nil) == nil {
			return payload, nil
		}
	}

	payloadToCompose, _ := unwrappedPayload.(map[string]interface{})
	composed, err := shape.ComposeNSEvent(eventName, payloadToCompose, &shape.NSEventComposeOptions{AgentID: source})
	if err != nil {
		return nil, fmt.Errorf("failed to compose canonical event: %w", err)
	}

	return lang.Wrap(composed)
}
