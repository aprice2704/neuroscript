// NeuroScript Version: 0.7.0
// File version: 56
// Purpose: Correctly prepends the bootstrap capsule to the orchestration prompt.
// filename: pkg/interpreter/interpreter_steps_ask.go
// nlines: 200+
// risk_rating: HIGH
package interpreter

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/llmconn"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// executeAsk handles the "ask" statement by orchestrating the AEIOU protocol,
// with a robust, policy-gated, multi-turn loop.
func (i *Interpreter) executeAsk(step ast.Step) (lang.Value, error) {
	i.logger.Debug("--- ENTERING executeAsk ---")
	if step.AskStmt == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "ask step is missing its AskStmt node", nil).WithPosition(step.GetPos())
	}
	node := step.AskStmt

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

	bootstrap, ok := capsule.Get(bootstrapCapsuleName)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "bootstrap capsule not found", nil).WithPosition(node.GetPos())
	}

	agentModelObj, found := i.AgentModels().Get(types.AgentModelName(agentName))
	if !found {
		return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("AgentModel '%s' is not registered", agentName), nil).WithPosition(node.AgentModelExpr.GetPos())
	}
	agentModel, ok := agentModelObj.(types.AgentModel)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("internal error: retrieved AgentModel for '%s' is not of type types.AgentModel, but %T", agentName, agentModelObj), nil).WithPosition(node.AgentModelExpr.GetPos())
	}

	prov, provExists := i.GetProvider(agentModel.Provider)
	if !provExists {
		return nil, lang.NewRuntimeError(lang.ErrorCodeProviderNotFound, fmt.Sprintf("provider '%s' for AgentModel '%s' not found", agentModel.Provider, agentModel.Name), nil).WithPosition(node.GetPos())
	}

	conn, err := llmconn.New(&agentModel, prov)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeConfiguration, "failed to create LLM connection", err).WithPosition(node.GetPos())
	}

	maxTurns := agentModel.MaxTurns
	if maxTurns <= 0 {
		maxTurns = 1
	}
	if maxTurns > maxTurnsCap {
		maxTurns = maxTurnsCap
	}

	var finalResult lang.Value = &lang.NilValue{}
	var prevOutputHash string

	// Combine the bootstrap instructions with the user's actual prompt.
	fullPrompt := fmt.Sprintf("%s\n\n%s", bootstrap.Content, initialPrompt)

	turnEnvelope := &aeiou.Envelope{
		Orchestration: fullPrompt,
	}

	for turn := 1; turn <= maxTurns; turn++ {
		i.logger.Debug("--- Starting ask loop turn ---", "turn", turn)

		aiResp, err := conn.Converse(context.Background(), turnEnvelope)
		if err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeExternal, "AI provider call failed", err).WithPosition(node.GetPos())
		}

		// DEBUG: Print the raw response from the AI provider.
		fmt.Printf("--- RAW AI RESPONSE ---\n%s\n-----------------------\n", aiResp.TextContent)

		responseEnvelope, err := aeiou.RobustParse(aiResp.TextContent)
		if err != nil {
			isOneShotAgent := maxTurns == 1 && !agentModel.Tools.ToolLoopPermitted
			if isOneShotAgent {
				i.logger.Debug("Response is not an envelope; treating as final answer for one-shot agent.")
				finalResult = lang.StringValue{Value: aiResp.TextContent}
				break
			}
			return nil, lang.NewRuntimeError(lang.ErrorCodeSyntax, "failed to parse AEIOU envelope from AI response", err).WithPosition(node.GetPos())
		}

		execInterp := i.clone()
		var actionEmits []string
		var actionWhispers = make(map[string]lang.Value)
		if err := executeAeiouTurn(execInterp, responseEnvelope, &actionEmits, &actionWhispers); err != nil {
			return nil, err
		}

		nextOutput, loopControlSignal := extractEmits(actionEmits)
		finalResult = lang.StringValue{Value: nextOutput}

		control, err := aeiou.ParseLoopControl(loopControlSignal)
		if err != nil {
			i.logger.Debug("Ask loop terminating: no valid LOOP signal found.")
			break
		}

		if control.Control == "continue" {
			if !agentModel.Tools.ToolLoopPermitted {
				return nil, lang.NewRuntimeError(lang.ErrorCodePolicy, fmt.Sprintf("agent model '%s' attempted to continue a loop but does not have 'toolLoopPermitted' grant", agentName), nil).WithPosition(node.GetPos())
			}
		} else {
			i.logger.Debug("Ask loop terminating: received 'done' or 'abort' signal.", "signal", control.Control)
			break
		}

		h := sha256.Sum256([]byte(nextOutput))
		currentHash := hex.EncodeToString(h[:])
		if turn > 1 && currentHash == prevOutputHash {
			i.logger.Warn("Ask loop terminating: no progress detected between turns.")
			break
		}
		prevOutputHash = currentHash

		if turn == maxTurns {
			i.logger.Warn("Ask loop terminating: max turns reached.")
			break
		}

		turnEnvelope = &aeiou.Envelope{
			Header:        responseEnvelope.Header,
			Orchestration: nextOutput,
		}
	}

	if node.IntoTarget != nil {
		if err := i.setSingleLValue(node.IntoTarget, finalResult); err != nil {
			return nil, err
		}
	}

	i.logger.Debug("--- EXITING executeAsk ---")
	return finalResult, nil
}
