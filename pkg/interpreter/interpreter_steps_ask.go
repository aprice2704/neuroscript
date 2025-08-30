// NeuroScript Version: 0.7.0
// File version: 48.0.0
// Purpose: Refactored to correctly use the llmconn.Connector interface, fixing the ask loop logic and test failures.
// filename: pkg/interpreter/interpreter_steps_ask.go
// nlines: 191
// risk_rating: HIGH

package interpreter

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/llmconn"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// executeAsk handles the "ask" statement by orchestrating the AEIOU protocol,
// with a robust, policy-gated, multi-turn loop.
func (i *Interpreter) executeAsk(step ast.Step) (lang.Value, error) {
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

	agentModelObj, found := i.AgentModels().Get(types.AgentModelName(agentName))
	if !found {
		return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("AgentModel '%s' is not registered", agentName), nil).WithPosition(node.AgentModelExpr.GetPos())
	}
	agentModel, ok := agentModelObj.(types.AgentModel)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("internal error: retrieved AgentModel for '%s' is not of type types.AgentModel, but %T", agentName, agentModelObj), nil).WithPosition(node.AgentModelExpr.GetPos())
	}

	// Get the provider for the agent model.
	prov, provExists := i.GetProvider(agentModel.Provider)
	if !provExists {
		return nil, lang.NewRuntimeError(lang.ErrorCodeProviderNotFound, fmt.Sprintf("provider '%s' for AgentModel '%s' not found", agentModel.Provider, agentModel.Name), nil).WithPosition(node.GetPos())
	}

	// ** ARCHITECTURAL FIX **
	// Instantiate the LLMConn to manage the conversation state and provider interaction.
	conn, err := llmconn.New(&agentModel, prov)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeConfiguration, "failed to create LLM connection", err).WithPosition(node.GetPos())
	}

	toolLoopPermitted := agentModel.Tools.ToolLoopPermitted
	maxTurns := agentModel.MaxTurns
	if maxTurns <= 0 || !toolLoopPermitted {
		maxTurns = 1
	}
	if maxTurns > maxTurnsCap {
		maxTurns = maxTurnsCap
	}

	// --- ASK LOOP START ---
	var finalResult lang.Value = &lang.NilValue{}
	var prevOutputHash string

	// Initial envelope with the user's prompt in the Orchestration section.
	turnEnvelope := &aeiou.Envelope{
		Orchestration: initialPrompt,
		// In a real host, UserData would be populated with richer context.
		// For now, Orchestration carries the primary goal.
	}

	for turn := 1; turn <= maxTurns; turn++ {
		i.logger.Debug("Executing ask loop turn", "turn", turn)

		aiResp, err := conn.Converse(context.Background(), turnEnvelope)
		if err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeExternal, "AI provider call failed", err).WithPosition(node.GetPos())
		}

		i.logger.Debug("Raw AI Response", "turn", turn, "response", aiResp.TextContent)

		responseEnvelope, err := aeiou.RobustParse(aiResp.TextContent)
		if err != nil {
			i.logger.Debug("Failed to parse AEIOU response, treating as raw final answer", "error", err)
			finalResult = lang.StringValue{Value: strings.TrimSpace(aiResp.TextContent)}
			break // Exit the loop with the raw response.
		}

		i.logger.Debug("Successfully parsed AEIOU envelope", "actions", responseEnvelope.Actions)

		execInterp := i.clone()
		var actionEmits []string
		actionWhispers := make(map[string]lang.Value)
		if err := executeAeiouTurn(execInterp, responseEnvelope, &actionEmits, &actionWhispers); err != nil {
			return nil, err
		}

		// The final result is the combined output of all non-signal emits.
		nextOutput, loopControlSignal := extractEmits(actionEmits)
		finalResult = lang.StringValue{Value: nextOutput}

		// If no loop is permitted, we are done after the first turn.
		if !toolLoopPermitted {
			i.logger.Debug("Ask loop terminating: tool loop not permitted.")
			break
		}

		// Check the loop control signal to see if we should continue.
		control, err := aeiou.ParseLoopControl(loopControlSignal)
		if err != nil {
			// If there's no valid LOOP signal, the conversation is over.
			i.logger.Debug("Ask loop terminating: no valid LOOP signal found.")
			break
		}

		if control.Control == "abort" || control.Control == "done" {
			i.logger.Debug("Ask loop terminating: received 'done' or 'abort' signal.", "signal", control.Control)
			break
		}

		// Anti-Stall Guard: Check if the AI is making progress.
		h := sha256.Sum256([]byte(nextOutput))
		currentHash := hex.EncodeToString(h[:])
		if turn > 1 && currentHash == prevOutputHash {
			i.logger.Warn("Ask loop terminating: no progress detected between turns.")
			break
		}
		prevOutputHash = currentHash

		// If we've hit the turn limit, stop.
		if turn == maxTurns {
			i.logger.Warn("Ask loop terminating: max turns reached.")
			break
		}

		// Prepare the envelope for the next turn. The output of this turn becomes
		// the primary context (Orchestration/Output) for the next.
		turnEnvelope = &aeiou.Envelope{
			Header:        responseEnvelope.Header,
			Orchestration: nextOutput,
			// TODO: Populate scratchpad from 'actionWhispers'
		}
	}
	// --- ASK LOOP END ---

	i.logger.Debug("Final Result Assigned to Variable", "result", fmt.Sprintf("%#v", finalResult))

	if node.IntoTarget != nil {
		if err := i.setSingleLValue(node.IntoTarget, finalResult); err != nil {
			return nil, err
		}
	}

	return finalResult, nil
}
