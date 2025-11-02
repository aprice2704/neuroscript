// NeuroScript Version: 0.8.0
// File version: 77
// Purpose: Fixed compile error by passing a 'string' (not types.AgentModelName) to AgentModels().Get().
// filename: pkg/interpreter/steps_ask.go
// nlines: 139
// risk_rating: HIGH

package interpreter

import (
	"encoding/json"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/account"
	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/eval"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/llmconn" // Restored dependency

	// "github.com/aprice2704/neuroscript/pkg/provider" // No longer directly needed here
	"github.com/aprice2704/neuroscript/pkg/types"
)

// executeAsk handles the "ask" statement.
func (i *Interpreter) executeAsk(step ast.Step) (lang.Value, error) {
	if step.AskStmt == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "ask step is missing its AskStmt node", nil).WithPosition(step.GetPos())
	}
	node := step.AskStmt

	// 1. Evaluate Agent and Prompt Expressions
	agentModelVal, err := eval.Expression(i, node.AgentModelExpr)
	if err != nil {
		return nil, lang.WrapErrorWithPosition(err, node.AgentModelExpr.GetPos(), "evaluating agent model for ask")
	}
	agentName, _ := lang.ToString(agentModelVal)

	promptVal, err := eval.Expression(i, node.PromptExpr)
	if err != nil {
		return nil, lang.WrapErrorWithPosition(err, node.PromptExpr.GetPos(), "evaluating prompt for ask")
	}
	initialPrompt, _ := lang.ToString(promptVal)

	// 2. Retrieve AgentModel and Provider
	// FIX: Use string 'agentName' directly
	agentModelObj, found := i.AgentModels().Get(agentName)
	if !found {
		return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("AgentModel '%s' is not registered", agentName), nil).WithPosition(node.AgentModelExpr.GetPos())
	}
	agentModel, ok := agentModelObj.(types.AgentModel)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("internal error: retrieved AgentModel for '%s' is not of type types.AgentModel, but %T", agentName, agentModelObj), nil).WithPosition(node.GetPos())
	}

	prov, provExists := i.GetProvider(agentModel.Provider)
	if !provExists {
		return nil, lang.NewRuntimeError(lang.ErrorCodeProviderNotFound, fmt.Sprintf("provider '%s' for AgentModel '%s' not found", agentModel.Provider, agentModel.Name), nil).WithPosition(node.GetPos())
	}

	// 2a. Resolve Account and Inject API Key
	if agentModel.AccountName != "" {
		accountObj, found := i.Accounts().Get(agentModel.AccountName)
		if !found {
			return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("account '%s' specified by AgentModel '%s' not found", agentModel.AccountName, agentModel.Name), nil).WithPosition(node.GetPos())
		}
		acc, ok := accountObj.(account.Account)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("retrieved object for account '%s' is not of type account.Account, but %T", agentModel.AccountName, accountObj), nil).WithPosition(node.GetPos())
		}
		if acc.APIKey == "" {
			return nil, lang.NewRuntimeError(lang.ErrorCodeConfiguration, fmt.Sprintf("account '%s' is missing or has an empty 'api_key' in its configuration", agentModel.AccountName), nil).WithPosition(node.GetPos())
		}
		agentModel.APIKey = acc.APIKey // Inject the API key into the model copy
	}

	// 3. Initialize LLM Connection (FIXED)
	conn, err := llmconn.New(&agentModel, prov, i.hostContext.Emitter)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeConfiguration, "failed to create LLM connection", err).WithPosition(node.GetPos())
	}

	// 4. Construct Initial V3 Envelope
	var userdataPayload string
	var jsonData map[string]interface{}
	if json.Unmarshal([]byte(initialPrompt), &jsonData) == nil {
		userdataPayload = initialPrompt
	} else {
		payloadMap := map[string]interface{}{
			"subject": "ask",
			"fields":  map[string]interface{}{"prompt": initialPrompt},
		}
		jsonBytes, err := json.Marshal(payloadMap)
		if err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to marshal simple prompt into JSON for USERDATA", err).WithPosition(node.PromptExpr.GetPos())
		}
		userdataPayload = string(jsonBytes)
	}
	initialEnvelope := &aeiou.Envelope{
		UserData: userdataPayload,
		Actions:  "command endcommand", // Start with empty actions
	}

	// 5. Delegate to the Host Loop - Pass the llmconn.Connector
	finalResult, err := i.runAskHostLoop(node.GetPos(), &agentModel, conn, initialEnvelope)
	if err != nil {
		// --- FIX: Wrap the error to ensure it's a RuntimeError ---
		if _, ok := err.(*lang.RuntimeError); !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "ask host loop failed", err).WithPosition(node.GetPos())
		}
		// --- End Fix ---
		return nil, err
	}

	// 6. Set Final Result
	if node.IntoTarget != nil {
		if err := i.setSingleLValue(node.IntoTarget, finalResult); err != nil {
			return nil, err // Propagate LValue setting errors
		}
	}
	return finalResult, nil
}
