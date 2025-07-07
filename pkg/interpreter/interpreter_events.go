// NeuroScript Version: 0.5.2
// File version: 15
// Purpose: Correctly uses a sandboxed clone for event handlers as per design spec.
// filename: pkg/interpreter/interpreter_events.go
// nlines: 60
// risk_rating: HIGH

package interpreter

import (
	"time"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func (i *Interpreter) EmitEvent(eventName string, source string, payload lang.Value) {
	i.eventManager.eventHandlersMu.RLock()
	handlers := i.eventManager.eventHandlers[eventName]
	i.eventManager.eventHandlersMu.RUnlock()

	if len(handlers) == 0 {
		return
	}

	eventDataMap := map[string]lang.Value{
		lang.EventKeyName:    lang.StringValue{Value: eventName},
		lang.EventKeySource:  lang.StringValue{Value: source},
		"timestamp":          lang.TimedateValue{Value: time.Now().UTC()},
		lang.EventKeyPayload: payload,
	}
	if payload == nil {
		eventDataMap[lang.EventKeyPayload] = &lang.NilValue{}
	}
	eventObj := lang.EventValue{Value: eventDataMap}

	for _, handler := range handlers {
		// FIX: Use the sandboxing clone to execute the handler. This ensures
		// that the handler has read-only access to globals and that any
		// state changes it makes are discarded.
		handlerInterpreter := i.CloneForEventHandler()

		if handler.EventVarName != "" {
			handlerInterpreter.SetVariable(handler.EventVarName, eventObj)
		}

		_, _, _, err := handlerInterpreter.executeSteps(handler.Body, true, nil)

		if err != nil {
			i.Logger().Error("Error executing 'on event' handler", "event", eventName, "error", err)
		}
	}
}
