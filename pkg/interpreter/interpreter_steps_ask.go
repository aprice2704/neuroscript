// NeuroScript Version: 0.7.0
// File version: 60
// Purpose: Corrected the initial envelope creation to include a minimal 'Actions' block, fixing the 'envelope missing required section' test failures.
// filename: pkg/interpreter/interpreter_steps_ask.go
// nlines: 101
// risk_rating: HIGH
package interpreter

import (
	"encoding/json"
	"fmt"

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

	// 3. Initialize LLM Connection
	conn, err := llmconn.New(&agentModel, prov)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeConfiguration, "failed to create LLM connection", err).WithPosition(node.GetPos())
	}

	// 4. Construct Initial V3 Envelope
	userDataBytes, err := json.Marshal(map[string]interface{}{"goal": initialPrompt})
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to marshal initial prompt to JSON for USERDATA", err).WithPosition(node.GetPos())
	}

	initialEnvelope := &aeiou.Envelope{
		UserData: string(userDataBytes),
		Actions:  "command endcommand", // FIX: Add minimal valid Actions block
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
