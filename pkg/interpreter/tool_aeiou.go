// NeuroScript Version: 0.8.0
// File version: 10
// Purpose: Adds extensive DEBUG output to trace context propagation.
// filename: pkg/interpreter/tool_aeiou.go
// nlines: 132
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
		Func:       makeMagicToolFunc(magicTool),
		IsInternal: true,
	}
	_, err := r.RegisterTool(impl)
	return err
}

func makeMagicToolFunc(magicTool *aeiou.MagicTool) tool.ToolFunc {
	return func(rt tool.Runtime, args []interface{}) (interface{}, error) {
		fmt.Fprintf(os.Stderr, "[DEBUG] makeMagicToolFunc: Entered. Runtime type is %T\n", rt) // DEBUG

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

		// --- THIS IS THE FIX ---
		// Internal tools must get the context via the TurnContextProvider interface,
		// which is implemented by both the internal interpreter and the public wrapper.
		ctxProvider, ok := rt.(TurnContextProvider)
		if !ok {
			fmt.Fprintf(os.Stderr, "[DEBUG] makeMagicToolFunc: FATAL! Runtime %T does not implement TurnContextProvider.\n", rt) // DEBUG
			return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "runtime does not implement TurnContextProvider", nil)
		}
		fmt.Fprintf(os.Stderr, "[DEBUG] makeMagicToolFunc: Runtime implements TurnContextProvider.\n") // DEBUG
		turnCtx := ctxProvider.GetTurnContext()
		if turnCtx == nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] makeMagicToolFunc: FATAL! GetTurnContext() returned nil.\n") // DEBUG
			return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "turn context was nil", nil)
		}
		fmt.Fprintf(os.Stderr, "[DEBUG] makeMagicToolFunc: Received context %p\n", turnCtx) // DEBUG

		hostCtx, err := getHostContext(turnCtx)
		// --- END FIX ---

		if err != nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] makeMagicToolFunc: getHostContext failed: %v\n", err) // DEBUG
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "[DEBUG] makeMagicToolFunc: getHostContext succeeded. SID: %s, Turn: %d\n", hostCtx.SessionID, hostCtx.TurnIndex) // DEBUG
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
	fmt.Fprintf(os.Stderr, "[DEBUG] getHostContext: Checking context %p for SID\n", ctx) // DEBUG
	sid, ok := ctx.Value(AeiouSessionIDKey).(string)
	if !ok {
		fmt.Fprintln(os.Stderr, "[DEBUG] getHostContext: AEIOU session ID not found.") // DEBUG
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "AEIOU session ID not found in turn context", nil)
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] getHostContext: Found SID %s. Checking for TurnIndex\n", sid) // DEBUG
	turn, ok := ctx.Value(AeiouTurnIndexKey).(int)
	if !ok {
		fmt.Fprintln(os.Stderr, "[DEBUG] getHostContext: AEIOU turn index not found.") // DEBUG
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "AEIOU turn index not found in turn context", nil)
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] getHostContext: Found TurnIndex %d. Checking for Nonce\n", turn) // DEBUG
	nonce, ok := ctx.Value(AeiouTurnNonceKey).(string)
	if !ok {
		fmt.Fprintln(os.Stderr, "[DEBUG] getHostContext: AEIOU turn nonce not found.") // DEBUG
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "AEIOU turn nonce not found in turn context", nil)
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] getHostContext: Found Nonce. Success.\n") // DEBUG

	return &aeiou.HostContext{
		SessionID: sid,
		TurnIndex: turn,
		TurnNonce: nonce,
	}, nil
}
