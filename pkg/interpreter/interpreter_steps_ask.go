// NeuroScript Version: 0.6.0
// File version: 15.0.0
// Purpose: Corrected to use the AskStmt field from the updated ast.Step struct.
// filename: pkg/interpreter/interpreter_steps_ask.go
// nlines: 105
// risk_rating: HIGH

package interpreter

import (
	"context"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// executeAsk handles the execution of an 'ask' statement step.
// It orchestrates evaluating arguments, finding the correct AgentModel and AIProvider,
// making the call, and assigning the result.
func (i *Interpreter) executeAsk(step ast.Step) (lang.Value, error) {
	if step.AskStmt == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "ask step is missing its AskStmt node", nil).WithPosition(step.GetPos())
	}
	node := step.AskStmt

	// 1. Evaluate AgentModel name and Prompt expressions
	agentModelVal, err := i.evaluate.Expression(node.AgentModelExpr)
	if err != nil {
		return nil, err
	}
	agentName, isString := lang.ToString(agentModelVal)
	if !isString || agentName == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "first 'ask' argument (AgentModel name) must be a non-empty string", nil).WithPosition(node.GetPos())
	}

	promptVal, err := i.evaluate.Expression(node.PromptExpr)
	if err != nil {
		return nil, err
	}
	prompt, isString := lang.ToString(promptVal)
	if !isString {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "second 'ask' argument (prompt) must be a string", nil).WithPosition(node.GetPos())
	}

	// 2. Get the AgentModel configuration from the registry
	agentModel, found := i.GetAgentModel(agentName)
	if !found {
		return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("AgentModel '%s' is not registered", agentName), lang.ErrMapKeyNotFound).WithPosition(node.GetPos())
	}

	// 3. Get the corresponding AI provider from the centralized state
	i.state.providersMu.RLock()
	aiProvider, providerFound := i.state.providers[agentModel.Provider]
	i.state.providersMu.RUnlock()

	if !providerFound {
		return nil, lang.NewRuntimeError(lang.ErrorCodeConfig, fmt.Sprintf("AI provider '%s' for AgentModel '%s' is not registered", agentModel.Provider, agentName), nil).WithPosition(node.GetPos())
	}

	// 4. Evaluate 'with' options and merge with AgentModel config
	req := provider.AIRequest{
		AgentModelName: agentModel.Name,
		ProviderName:   agentModel.Provider,
		ModelName:      agentModel.Model,
		BaseURL:        agentModel.BaseURL,
		APIKey:         agentModel.APIKey,
		Prompt:         prompt,
		Temperature:    agentModel.Temperature, // Default from AgentModel
	}

	if node.WithOptions != nil {
		optionsVal, err := i.evaluate.Expression(node.WithOptions)
		if err != nil {
			return nil, err
		}
		optionsMap, ok := optionsVal.(*lang.MapValue)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, "'with' clause must be a map", nil).WithPosition(node.GetPos())
		}

		// Override with values from the 'with' map
		if tempVal, ok := optionsMap.Value["temperature"]; ok {
			if temp, isFloat := lang.ToFloat64(tempVal); isFloat {
				req.Temperature = temp
			}
		}
		// ... handle other overrides like 'stream', 'timeout', etc. ...
	}

	// 5. Execute the request via the provider
	ctx := context.Background() // In a real scenario, this would be cancellable
	resp, err := aiProvider.Chat(ctx, req)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeExternal, fmt.Sprintf("AI provider call failed: %v", err), err).WithPosition(node.GetPos())
	}

	// 6. Assign the result to the 'into' variable, if present
	resultVal, wrapErr := lang.Wrap(resp.TextContent)
	if wrapErr != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("failed to wrap AI response: %v", wrapErr), wrapErr).WithPosition(node.GetPos())
	}

	if node.IntoTarget != nil {
		if err := i.setSingleLValue(node.IntoTarget, resultVal); err != nil {
			return nil, err
		}
	}

	return resultVal, nil
}
