// NeuroScript Version: 0.3.1
// File version: 10
// Purpose: Corrects a deadlock by making mutex locking in EmitEvent more fine-grained.
// filename: pkg/core/interpreter_events.go

package runtime

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

	for _, handler := range handlers {
		// Add the temporary event object to the scope in a thread-safe manner.
		if handler.EventVarName != "" {
			// Lock, modify, and immediately unlock.
			i.variablesMu.Lock()
			i.variables[handler.EventVarName] = eventObj
			i.variablesMu.Unlock()
		}

		// Execute the handler's body. This function and its children will
		// acquire locks as needed, but since the parent lock is released,
		// there will be no deadlock.
		_, _, _, err := i.executeSteps(handler.Body, true, nil)

		// Clean up the temporary event variable in a thread-safe manner.
		if handler.EventVarName != "" {
			// Lock, modify, and immediately unlock.
			i.variablesMu.Lock()
			delete(i.variables, handler.EventVarName)
			i.variablesMu.Unlock()
		}

		if err != nil {
			i.Logger().Error("Error executing 'on event' handler", "event", eventName, "error", err)
		}
	}
}
