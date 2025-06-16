// NeuroScript Version: 0.3.1
// File version: 9
// Purpose: Restored mutex locking for the event handler loop to ensure thread-safe access to the shared variable scope.
// filename: pkg/core/interpreter_events.go
// nlines: 50+
// risk_rating: HIGH

package core

import (
	"time"
)

func (i *Interpreter) EmitEvent(eventName string, source string, payload Value) {
	i.eventHandlersMu.RLock()
	handlers := i.eventHandlers[eventName]
	i.eventHandlersMu.RUnlock()

	if len(handlers) == 0 {
		return
	}

	// Build the event data object once.
	eventDataMap := map[string]Value{
		EventKeyName:    StringValue{Value: eventName},
		EventKeySource:  StringValue{Value: source},
		"timestamp":     TimedateValue{Value: time.Now().UTC()},
		EventKeyPayload: payload,
	}
	if payload == nil {
		eventDataMap[EventKeyPayload] = NilValue{}
	}
	eventObj := EventValue{Value: eventDataMap}

	// Lock the main variables map for the entire duration of the event processing.
	// This prevents data races when tests or other threads try to access variables
	// while the handlers are running.
	i.variablesMu.Lock()
	defer i.variablesMu.Unlock()

	// Execute handlers sequentially, modifying the single global scope.
	for _, handler := range handlers {
		// Temporarily add the event object to the global scope for this handler's execution.
		if handler.EventVarName != "" {
			i.variables[handler.EventVarName] = eventObj
		}

		// Execute the handler's body directly in the global scope.
		// Any variables set will persist for the next handler and after all handlers complete.
		_, _, _, err := i.executeSteps(handler.Body, true, nil)

		// Clean up the temporary event variable from the global scope.
		if handler.EventVarName != "" {
			delete(i.variables, handler.EventVarName)
		}

		if err != nil {
			i.Logger().Error("Error executing 'on event' handler", "event", eventName, "error", err)
			// Decide if an error in one handler should stop others. For now, we continue.
		}
	}
}
