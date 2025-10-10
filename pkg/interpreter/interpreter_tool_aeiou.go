// NeuroScript Version: 0.8.0
// File version: 8
// Purpose: Updates AEIOU tool to get the turn context via the interpreter's GetTurnContext method.
// filename: pkg/interpreter/interpreter_tool_aeiou.go
// nlines: 121
// risk_rating: HIGH

package interpreter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

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
		hostCtx, err := getHostContext(interp.GetTurnContext())
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
	// --- MORE DEBUGGING ---
	if ctx == nil {
		fmt.Fprintf(os.Stderr, "[DEBUG getHostContext] Received a NIL context!\n")
	} else {
		sid, sidOK := ctx.Value(aeiou.SessionIDKey).(string)
		turn, turnOK := ctx.Value(aeiou.TurnIndexKey).(int)
		nonce, nonceOK := ctx.Value(aeiou.TurnNonceKey).(string)
		fmt.Fprintf(os.Stderr, "[DEBUG getHostContext] Context received. SID OK: %t (val: %q), Turn OK: %t (val: %d), Nonce OK: %t (val: %q)\n", sidOK, sid, turnOK, turn, nonceOK, nonce)
	}
	// --- END DEBUGGING ---

	sid, ok := ctx.Value(aeiou.SessionIDKey).(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "AEIOU session ID not found in turn context", nil)
	}
	turn, ok := ctx.Value(aeiou.TurnIndexKey).(int)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "AEIOU turn index not found in turn context", nil)
	}
	nonce, ok := ctx.Value(aeiou.TurnNonceKey).(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "AEIOU turn nonce not found in turn context", nil)
	}

	return &aeiou.HostContext{
		SessionID: sid,
		TurnIndex: turn,
		TurnNonce: nonce,
	}, nil
}
