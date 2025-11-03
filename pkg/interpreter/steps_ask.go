// NeuroScript Version: 0.8.0
// File version: 79
// Purpose: Refactored to handle 'any' return value from the AEIOU service hook, fixing import cycle.
// filename: pkg/interpreter/steps_ask.go
// nlines: 172
// risk_rating: HIGH

package interpreter

import (
	"encoding/json"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/account"
	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/eval"
	"github.com/aprice2704/neuroscript/pkg/interfaces" // Added for hook
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/llmconn"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// executeAsk handles the "ask" statement.
// It now implements the AEIOU v2+ hook.
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

	var finalResult lang.Value

	// 2. --- AEIOU v2+ Service Hook ---
	// Check if an external orchestrator is registered.
	if i.hostContext != nil && i.hostContext.ServiceRegistry != nil {
		// Attempt to find and use the service.
		if reg, ok := i.hostContext.ServiceRegistry.(map[string]any); ok {
			if service, found := reg[interfaces.AeiouServiceKey]; found {
				if orchestrator, ok := service.(interfaces.AeiouOrchestrator); ok {
					// --- HOOK: Delegate to external service ---
					// We pass i.PublicAPI, which is the *api.Interpreter wrapper.
					var rawResult any
					rawResult, err = orchestrator.RunAskLoop(i.PublicAPI, agentName, initialPrompt)
					if err != nil {
						return nil, lang.WrapErrorWithPosition(err, node.GetPos(), "external AEIOU service loop failed")
					}

					// FIX: Type-assert the 'any' return value to 'lang.Value'
					finalResult, ok = rawResult.(lang.Value)
					if !ok {
						if rawResult == nil {
							finalResult = lang.NilValue{} // nil is a valid result
						} else {
							// The service returned a type that is not a lang.Value. This is a contract violation.
							return nil, lang.NewRuntimeError(lang.ErrorCodeInternal,
								fmt.Sprintf("external AEIOU service returned invalid type: got %T, want lang.Value", rawResult),
								nil).WithPosition(node.GetPos())
						}
					}
				} else {
					i.Logger().Warn("Found AeiouService, but it does not implement interfaces.AeiouOrchestrator", "type", fmt.Sprintf("%T", service))
					finalResult, err = i.executeLegacyAsk(node, agentName, initialPrompt)
				}
			} else {
				// Registry present, but no service. Fallback.
				finalResult, err = i.executeLegacyAsk(node, agentName, initialPrompt)
			}
		} else {
			i.Logger().Warn("ServiceRegistry is not a map[string]any. Falling back to legacy 'ask'.", "type", fmt.Sprintf("%T", i.hostContext.ServiceRegistry))
			finalResult, err = i.executeLegacyAsk(node, agentName, initialPrompt)
		}
	} else {
		// --- FALLBACK: No ServiceRegistry. Run legacy internal loop. ---
		finalResult, err = i.executeLegacyAsk(node, agentName, initialPrompt)
	}
	// --- End Hook Logic ---

	if err != nil {
		return nil, err // Return error from hook or legacy path
	}

	// 6. Set Final Result (identical for both paths)
	if node.IntoTarget != nil {
		if err := i.setSingleLValue(node.IntoTarget, finalResult); err != nil {
			return nil, err // Propagate LValue setting errors
		}
	}
	return finalResult, nil
}

// executeLegacyAsk contains the original v1 'ask' logic.
func (i *Interpreter) executeLegacyAsk(node *ast.AskStmt, agentName, initialPrompt string) (lang.Value, error) {
	// 2. Retrieve AgentModel and Provider
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

	// 3. Initialize LLM Connection
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

	// 5. Delegate to the Host Loop
	finalResult, err := i.runAskHostLoop(node.GetPos(), &agentModel, conn, initialEnvelope)
	if err != nil {
		if _, ok := err.(*lang.RuntimeError); !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "ask host loop failed", err).WithPosition(node.GetPos())
		}
		return nil, err
	}
	return finalResult, nil
}
