// NeuroScript Version: 0.3.1
// File version: 6
// Purpose: Corrected initialization of EventValue to match its map-wrapper definition.
// filename: pkg/core/interpreter_events.go
// nlines: 50
// risk_rating: MEDIUM

package core

import "time"

func (i *Interpreter) EmitEvent(eventName string, source string, payload Value) {
	i.eventHandlersMu.RLock()
	handlers := i.eventHandlers[eventName]
	i.eventHandlersMu.RUnlock()

	if len(handlers) == 0 {
		return
	}

	// Correctly build the map first
	eventDataMap := map[string]Value{
		"name":      StringValue{Value: eventName},
		"source":    StringValue{Value: source},
		"timestamp": TimedateValue{Value: time.Now().UTC()},
		"payload":   payload,
	}
	if payload == nil {
		eventDataMap["payload"] = NilValue{}
	}

	// Correctly initialize EventValue as a map wrapper
	eventObj := EventValue{Value: eventDataMap}

	for _, handler := range handlers {
		originalScope := i.variables
		handlerScope := make(map[string]interface{})
		for k, v := range originalScope {
			handlerScope[k] = v
		}

		if handler.EventVarName != "" {
			handlerScope[handler.EventVarName] = eventObj
		}

		i.variables = handlerScope
		_, _, _, err := i.executeSteps(handler.Body, true, nil)
		i.variables = originalScope

		if err != nil {
			i.Logger().Error("Error executing 'on event' handler", "event", eventName, "error", err)
		}
	}
}
