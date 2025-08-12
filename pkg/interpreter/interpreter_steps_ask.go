// NeuroScript Version: 0.6.0
// File version: 22.0.0
// Purpose: Corrects all type mismatch errors by adding explicit type conversions for types.AgentModelName.
// filename: pkg/interpreter/interpreter_steps_ask.go
// nlines: 160
// risk_rating: HIGH

package interpreter

import (
	"context"
	"fmt"
	"os"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/runtime"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// executeAsk handles the "ask" statement.
func (i *Interpreter) executeAsk(step ast.Step) (lang.Value, error) {
	if step.AskStmt == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "ask step is missing its AskStmt node", nil).WithPosition(step.GetPos())
	}
	node := step.AskStmt

	// 1. Evaluate AgentModel and Prompt expressions
	agentModelVal, err := i.evaluate.Expression(node.AgentModelExpr)
	if err != nil {
		return nil, lang.WrapErrorWithPosition(err, node.AgentModelExpr.GetPos(), "evaluating agent model for ask")
	}
	agentName, _ := lang.ToString(agentModelVal)

	promptVal, err := i.evaluate.Expression(node.PromptExpr)
	if err != nil {
		return nil, lang.WrapErrorWithPosition(err, node.PromptExpr.GetPos(), "evaluating prompt for ask")
	}
	prompt, _ := lang.ToString(promptVal)

	// 2. Get AgentModel configuration and perform policy validation
	// FIX: Explicitly convert the string agentName to types.AgentModelName for the lookup.
	agentModel, found := i.GetAgentModel(types.AgentModelName(agentName))
	if !found {
		return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("AgentModel '%s' is not registered", agentName), lang.ErrMapKeyNotFound).WithPosition(node.AgentModelExpr.GetPos())
	}

	// Construct the security envelope from the AgentModel's config for policy check.
	envelope := runtime.AgentModelEnvelope{
		// FIX: Explicitly convert the types.AgentModelName to a string for the envelope.
		Name:           string(agentModel.Name),
		Hosts:          []string{agentModel.BaseURL},
		SecretEnvKeys:  []string{agentModel.SecretRef},
		BudgetCurrency: agentModel.BudgetCurrency,
	}

	if i.ExecPolicy != nil {
		if err := i.ExecPolicy.ValidateAgentModelEnvelope(envelope); err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodePolicy, fmt.Sprintf("ask statement rejected by policy: %v", err), err).WithPosition(step.GetPos())
		}
	}

	// 3. Resolve the API key *after* the policy check has passed.
	var apiKey string
	if agentModel.SecretRef != "" {
		apiKey = os.Getenv(agentModel.SecretRef)
		if apiKey == "" {
			i.logger.Warn("AgentModel secret reference resolved to an empty string", "secret_ref", agentModel.SecretRef)
		}
	}

	// 4. Evaluate 'with' options and build the AI request
	opts := map[string]lang.Value{}
	if node.WithOptions != nil {
		optsVal, err := i.evaluate.Expression(node.WithOptions)
		if err != nil {
			return nil, lang.WrapErrorWithPosition(err, node.WithOptions.GetPos(), "evaluating 'with' options for ask")
		}
		if m, ok := optsVal.(*lang.MapValue); ok {
			opts = m.Value
		}
	}

	req := provider.AIRequest{
		ModelName: agentModel.Model,
		Prompt:    prompt,
		APIKey:    apiKey,
	}
	if temp, ok := opts["temperature"]; ok {
		req.Temperature, _ = lang.ToFloat64(temp)
	}

	// 5. Get the provider and execute the call
	prov, provExists := i.state.providers[agentModel.Provider]
	if !provExists {
		return nil, lang.NewRuntimeError(lang.ErrorCodeProviderNotFound, fmt.Sprintf("provider '%s' for AgentModel '%s' not found", agentModel.Provider, agentName), nil).WithPosition(step.GetPos())
	}

	ctx := context.Background()
	resp, err := prov.Chat(ctx, req)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeExternal, "AI provider call failed", err).WithPosition(step.GetPos())
	}

	// 6. Account for resource usage after a successful call
	if i.ExecPolicy != nil && agentModel.BudgetCurrency != "" {
		costInCents := 25 // Placeholder: This cost should come from the provider response eventually
		if err := i.ExecPolicy.Grants.CheckPerCallBudget(agentModel.BudgetCurrency, costInCents); err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodePolicy, "ask call exceeds per-call budget", err).WithPosition(step.GetPos())
		}
		_ = i.ExecPolicy.Grants.ChargeBudget(agentModel.BudgetCurrency, costInCents)
	}

	// 7. Assign result and return
	responseVal, wrapErr := lang.Wrap(resp.TextContent)
	if wrapErr != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to wrap AI response", wrapErr).WithPosition(step.GetPos())
	}

	if node.IntoTarget != nil {
		if err := i.setSingleLValue(node.IntoTarget, responseVal); err != nil {
			return nil, err
		}
	}

	return responseVal, nil
}
