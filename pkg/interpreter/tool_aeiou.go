// NeuroScript Version: 0.8.0
// File version: 6
// Purpose: Registers the AEIOU v3 magic token tool.
// filename: pkg/interpreter/tool_aeiou.go
// nlines: 115
// risk_rating: HIGH

package interpreter

import (
	"context"
	"encoding/json"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func registerAeiouTools(r tool.ToolRegistry, magicTool *aeiou.MagicTool) error {
	impl := tool.ToolImplementation{
		Spec: tool.ToolSpec{
			Name:        "magic",
			Group:       "tool.aeiou",
			Description: "Generates a signed AEIOU v3 control token.",
			Args: []tool.ArgSpec{
				{Name: "kind", Type: tool.ArgTypeString, Description: "The AEIOU control kind (e.g., 'LOOP').", Required: true},
				{Name: "params", Type: tool.ArgTypeMap, Description: "A map of parameters for the token (e.g., action, reason).", Required: true},
			},
			ReturnType: tool.ArgTypeString,
			ReturnHelp: "The signed control token string.",
		},
		Func: makeMagicToolFunc(magicTool),
	}
	_, err := r.RegisterTool(impl)
	return err
}

func makeMagicToolFunc(magicTool *aeiou.MagicTool) tool.ToolFunc {
	return func(rt tool.Runtime, args []interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "expected 2 arguments: kind and params", nil)
		}

		kindStr, ok := args[0].(string)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, "argument 1 'kind' must be a string", nil)
		}

		paramsMap, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, "argument 2 'params' must be a map", nil)
		}

		agentPayload, err := mapToControlPayload(paramsMap)
		if err != nil {
			return nil, err
		}

		interp, ok := rt.(*Interpreter)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "runtime is not an interpreter", nil)
		}
		hostCtx, err := getHostContext(interp.turnCtx)
		if err != nil {
			return nil, err
		}
		hostCtx.KeyID = "transient-key-01"
		hostCtx.TTL = 120 // 2 minute default TTL

		return magicTool.MintMagicToken(aeiou.ControlKind(kindStr), *agentPayload, *hostCtx)
	}
}

func mapToControlPayload(params map[string]interface{}) (*aeiou.ControlPayload, error) {
	var payload aeiou.ControlPayload

	actionVal, ok := params["action"]
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, "params map must contain an 'action' key", nil)
	}
	actionStr, ok := actionVal.(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "'action' key must be a string", nil)
	}
	payload.Action = aeiou.LoopAction(actionStr)

	if requestVal, ok := params["request"]; ok {
		jsonBytes, err := json.Marshal(requestVal)
		if err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to marshal 'request' to JSON", err)
		}
		payload.Request = jsonBytes
	}

	if telemetryVal, ok := params["telemetry"]; ok {
		jsonBytes, err := json.Marshal(telemetryVal)
		if err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to marshal 'telemetry' to JSON", err)
		}
		payload.Telemetry = jsonBytes
	}

	return &payload, nil
}

func getHostContext(ctx context.Context) (*aeiou.HostContext, error) {
	sid, ok := ctx.Value(aeiouSessionIDKey).(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "AEIOU session ID not found in turn context", nil)
	}
	turn, ok := ctx.Value(aeiouTurnIndexKey).(int)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "AEIOU turn index not found in turn context", nil)
	}
	nonce, ok := ctx.Value(aeiouTurnNonceKey).(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "AEIOU turn nonce not found in turn context", nil)
	}

	return &aeiou.HostContext{
		SessionID: sid,
		TurnIndex: turn,
		TurnNonce: nonce,
	}, nil
}
