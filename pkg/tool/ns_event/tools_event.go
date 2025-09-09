// NeuroScript Version: 0.7.0
// File version: 8
// Purpose: Implements tool functions for handling ns standard events, with proper structured error returns.
// filename: pkg/tool/ns_event/tools_event.go
// nlines: 200
// risk_rating: MEDIUM
package ns_event

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/aprice2704/neuroscript/pkg/json_lite"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func toolComposeEvent(rt tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) < 2 || len(args) > 4 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("Compose: expected 2 to 4 arguments, got %d", len(args)), lang.ErrArgumentMismatch)
	}

	kind, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "argument 'kind' must be a string", lang.ErrInvalidArgument)
	}
	payload, ok := args[1].(map[string]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "argument 'payload' must be a map[string]interface{}", lang.ErrInvalidArgument)
	}

	id := ""
	if len(args) > 2 && args[2] != nil {
		var ok bool
		id, ok = args[2].(string)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, "argument 'id' must be a string", lang.ErrInvalidArgument)
		}
	}
	if id == "" {
		id = uuid.New().String()
	}

	agentID := ""
	if len(args) > 3 && args[3] != nil {
		var ok bool
		agentID, ok = args[3].(string)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, "argument 'agent_id' must be a string", lang.ErrInvalidArgument)
		}
	}

	eventEnvelope := map[string]interface{}{
		"ID":      id,
		"Kind":    kind,
		"AgentID": agentID,
		"TS":      time.Now().UnixNano(),
		"Payload": payload,
	}

	eventObject := map[string]interface{}{
		"payload": []interface{}{eventEnvelope},
	}

	shape, err := json_lite.ParseShape(fdmEventShape)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal,
			fmt.Sprintf("internal error: canonical event shape is invalid: %v", err), err)
	}

	if err := shape.Validate(eventObject, nil); err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal,
			fmt.Sprintf("internal error: composed event failed validation against its own shape: %v", err), err)
	}

	return eventObject, nil
}

// getFirstEventEnvelope safely extracts the first event map from the nested structure.
func getFirstEventEnvelope(eventObject map[string]interface{}) (map[string]interface{}, bool) {
	payload, ok := eventObject["payload"]
	if !ok {
		return nil, false
	}

	payloadList, ok := payload.([]interface{})
	if !ok || len(payloadList) == 0 {
		return nil, false
	}

	firstEvent, ok := payloadList[0].(map[string]interface{})
	if !ok {
		return nil, false
	}

	return firstEvent, true
}

func toolGetPayload(rt tool.Runtime, args []interface{}) (interface{}, error) {
	eventObject, ok := args[0].(map[string]interface{})
	if !ok {
		return lang.NewMapValue(nil), nil
	}

	firstEvent, ok := getFirstEventEnvelope(eventObject)
	if !ok {
		return lang.NewMapValue(nil), nil
	}

	payload, ok := firstEvent["Payload"].(map[string]interface{})
	if !ok {
		return lang.NewMapValue(nil), nil
	}

	return payload, nil
}

func toolGetAllPayloads(rt tool.Runtime, args []interface{}) (interface{}, error) {
	allPayloads := make([]interface{}, 0)

	eventObject, ok := args[0].(map[string]interface{})
	if !ok {
		return allPayloads, nil
	}

	payload, ok := eventObject["payload"]
	if !ok {
		return allPayloads, nil
	}

	payloadList, ok := payload.([]interface{})
	if !ok {
		return allPayloads, nil
	}

	for _, item := range payloadList {
		if eventEnvelope, ok := item.(map[string]interface{}); ok {
			if p, ok := eventEnvelope["Payload"].(map[string]interface{}); ok {
				allPayloads = append(allPayloads, p)
			}
		}
	}

	return allPayloads, nil
}

func toolGetID(rt tool.Runtime, args []interface{}) (interface{}, error) {
	eventObject, ok := args[0].(map[string]interface{})
	if !ok {
		return "", nil
	}

	firstEvent, ok := getFirstEventEnvelope(eventObject)
	if !ok {
		return "", nil
	}

	id, _ := firstEvent["ID"].(string)
	return id, nil
}

func toolGetKind(rt tool.Runtime, args []interface{}) (interface{}, error) {
	eventObject, ok := args[0].(map[string]interface{})
	if !ok {
		return "", nil
	}

	firstEvent, ok := getFirstEventEnvelope(eventObject)
	if !ok {
		return "", nil
	}

	kind, _ := firstEvent["Kind"].(string)
	return kind, nil
}

func toolGetTimestamp(rt tool.Runtime, args []interface{}) (interface{}, error) {
	eventObject, ok := args[0].(map[string]interface{})
	if !ok {
		return int64(0), nil
	}

	firstEvent, ok := getFirstEventEnvelope(eventObject)
	if !ok {
		return int64(0), nil
	}

	tsFloat, okFloat := firstEvent["TS"].(float64)
	if okFloat {
		return int64(tsFloat), nil
	}
	tsInt, okInt := firstEvent["TS"].(int64)
	if okInt {
		return tsInt, nil
	}
	return int64(0), nil
}
