// NeuroScript Version: 0.6.0
// File version: 16.0.0
// Purpose: Corrected logic to properly extract agent, prompt, and options from the generic fields of the ast.Step struct, aligning with the current AST builder's output.
// filename: pkg/interpreter/interpreter_steps_ask.go
// nlines: 115
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
	// FIX: The AST builder places ask components into the generic Step fields.
	// We must extract them from there instead of a dedicated AskStmt field.
	if len(step.Values) < 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "ask statement requires at least 2 arguments (agent, prompt)", nil).WithPosition(step.GetPos())
	}

	agentModelExpr := step.Values[0]
	promptExpr := step.Values[1]
	var withOptionsExpr ast.Expression
	if len(step.Values) > 2 {
		withOptionsExpr = step.Values[2]
	}
	var intoTarget *ast.LValueNode
	if len(step.LValues) > 0 {
		intoTarget = step.LValues[0]
	}

	// 1. Evaluate AgentModel name and Prompt expressions
	agentModelVal, err := i.evaluate.Expression(agentModelExpr)
	if err != nil {
		return nil, err
	}
	agentName, isString := lang.ToString(agentModelVal)
	if !isString || agentName == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "first 'ask' argument (AgentModel name) must be a non-empty string", nil).WithPosition(agentModelExpr.GetPos())
	}

	promptVal, err := i.evaluate.Expression(promptExpr)
	if err != nil {
		return nil, err
	}
	prompt, isString := lang.ToString(promptVal)
	if !isString {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "second 'ask' argument (prompt) must be a string", nil).WithPosition(promptExpr.GetPos())
	}

	// 2. Get the AgentModel configuration from the registry
	agentModel, found := i.GetAgentModel(agentName)
	if !found {
		return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("AgentModel '%s' is not registered", agentName), lang.ErrMapKeyNotFound).WithPosition(agentModelExpr.GetPos())
	}

	// 3. Get the corresponding AI provider from the centralized state
	i.state.providersMu.RLock()
	aiProvider, providerFound := i.state.providers[agentModel.Provider]
	i.state.providersMu.RUnlock()

	if !providerFound {
		return nil, lang.NewRuntimeError(lang.ErrorCodeConfig, fmt.Sprintf("AI provider '%s' for AgentModel '%s' is not registered", agentModel.Provider, agentName), nil).WithPosition(step.GetPos())
	}

	// 4. Evaluate 'with' options and merge with AgentModel config
	req := provider.AIRequest{
		AgentModelName: agentModel.Name,
		ProviderName:   agentModel.Provider,
		ModelName:      agentModel.Model,
		BaseURL:        agentModel.BaseURL,
		APIKey:         agentModel.APIKey,
		Prompt:         prompt,
		Temperature:    agentModel.Temperature,
	}

	if withOptionsExpr != nil {
		optionsVal, err := i.evaluate.Expression(withOptionsExpr)
		if err != nil {
			return nil, err
		}
		optionsMap, ok := optionsVal.(*lang.MapValue)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, "'with' clause must be a map", nil).WithPosition(withOptionsExpr.GetPos())
		}

		if tempVal, ok := optionsMap.Value["temperature"]; ok {
			if temp, isFloat := lang.ToFloat64(tempVal); isFloat {
				req.Temperature = temp
			}
		}
	}

	// 5. Execute the request via the provider
	ctx := context.Background()
	resp, err := aiProvider.Chat(ctx, req)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeExternal, fmt.Sprintf("AI provider call failed: %v", err), err).WithPosition(step.GetPos())
	}

	// 6. Assign the result to the 'into' variable, if present
	resultVal, wrapErr := lang.Wrap(resp.TextContent)
	if wrapErr != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("failed to wrap AI response: %v", wrapErr), wrapErr).WithPosition(step.GetPos())
	}

	if intoTarget != nil {
		if err := i.setSingleLValue(intoTarget, resultVal); err != nil {
			return nil, err
		}
	}

	return resultVal, nil
}
