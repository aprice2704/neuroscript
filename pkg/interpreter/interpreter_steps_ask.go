// NeuroScript Version: 0.7.2
// File version: 70
// Purpose: Auto-wraps simple string prompts in the 'ask' statement into the required AEIOU v3 JSON format.
// filename: pkg/interpreter/interpreter_steps_ask.go
// nlines: 125
// risk_rating: MEDIUM

package interpreter

import (
	"encoding/json"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/account"
	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/llmconn"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// executeAsk handles the "ask" statement. It sets up the AEIOU v3 host loop
// and delegates the core turn-based execution to runAskHostLoop.
func (i *Interpreter) executeAsk(step ast.Step) (lang.Value, error) {
	if step.AskStmt == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "ask step is missing its AskStmt node", nil).WithPosition(step.GetPos())
	}
	node := step.AskStmt

	// 1. Evaluate Agent and Prompt Expressions
	agentModelVal, err := i.evaluate.Expression(node.AgentModelExpr)
	if err != nil {
		return nil, lang.WrapErrorWithPosition(err, node.AgentModelExpr.GetPos(), "evaluating agent model for ask")
	}
	agentName, _ := lang.ToString(agentModelVal)

	promptVal, err := i.evaluate.Expression(node.PromptExpr)
	if err != nil {
		return nil, lang.WrapErrorWithPosition(err, node.PromptExpr.GetPos(), "evaluating prompt for ask")
	}
	initialPrompt, _ := lang.ToString(promptVal)

	// 2. Retrieve AgentModel and Provider
	agentModelObj, found := i.AgentModels().Get(types.AgentModelName(agentName))
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

	// 2a. Resolve Account and Inject API Key into the AgentModel for this call
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
		agentModel.APIKey = acc.APIKey
	}

	// 3. Initialize LLM Connection
	conn, err := llmconn.New(&agentModel, prov, i.emitter)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeConfiguration, "failed to create LLM connection", err).WithPosition(node.GetPos())
	}

	// 4. Construct Initial V3 Envelope & Handle Prompt Wrapping
	//    This block implements the required auto-wrapping of simple string prompts
	//    to conform to the AEIOU v3 protocol.
	var userdataPayload string
	var jsonData map[string]interface{}
	if json.Unmarshal([]byte(initialPrompt), &jsonData) == nil {
		// The prompt is already a valid JSON object string. Use it directly.
		userdataPayload = initialPrompt
	} else {
		// The prompt is a simple string. Wrap it in the standard JSON structure.
		payloadMap := map[string]interface{}{
			"subject": "ask",
			"fields": map[string]interface{}{
				"prompt": initialPrompt,
			},
		}
		jsonBytes, err := json.Marshal(payloadMap)
		if err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to marshal simple prompt into JSON for USERDATA", err).WithPosition(node.PromptExpr.GetPos())
		}
		userdataPayload = string(jsonBytes)
	}

	initialEnvelope := &aeiou.Envelope{
		UserData: userdataPayload,
		Actions:  "command endcommand", // A valid, empty actions block is required.
	}

	// 5. Delegate to the Host Loop
	finalResult, err := i.runAskHostLoop(node.GetPos(), &agentModel, conn, initialEnvelope)
	if err != nil {
		return nil, err // The loop will return a fully formed RuntimeError
	}

	// 6. Set Final Result
	if node.IntoTarget != nil {
		if err := i.setSingleLValue(node.IntoTarget, finalResult); err != nil {
			return nil, err
		}
	}

	return finalResult, nil
}
