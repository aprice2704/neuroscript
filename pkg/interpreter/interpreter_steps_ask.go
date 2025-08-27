// NeuroScript Version: 0.7.0
// File version: 36.0.0
// Purpose: Updated to use explicit fields (e.g., ToolLoopPermitted) from the AgentModel struct instead of a generic map.
// filename: pkg/interpreter/interpreter_steps_ask.go
// nlines: 220
// risk_rating: HIGH

package interpreter

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/types"
)

const (
	loopContinueMarker = "[[loop:continue]]"
	loopDoneMarker     = "[[loop:done]]"
	loopAbortMarker    = "[[loop:abort" // Prefix match
	defaultMaxTurns    = 1
	maxTurnsCap        = 10 // A hard safety cap
)

type loopControlState int

const (
	stateContinue loopControlState = iota
	stateDone
	stateAbort
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

	agentModelObj, found := i.GetAgentModel(types.AgentModelName(agentName))
	if !found {
		return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("AgentModel '%s' is not registered", agentName), lang.ErrMapKeyNotFound).WithPosition(node.AgentModelExpr.GetPos())
	}
	agentModel, ok := agentModelObj.(types.AgentModel)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("internal error: retrieved AgentModel for '%s' is not of type types.AgentModel, but %T", agentName, agentModelObj), nil).WithPosition(node.AgentModelExpr.GetPos())
	}

	// Access loop parameters from direct, strongly-typed fields.
	toolLoopPermitted := agentModel.ToolLoopPermitted
	maxTurns := agentModel.MaxTurns
	if maxTurns <= 0 || !toolLoopPermitted {
		maxTurns = defaultMaxTurns
	}
	if maxTurns > maxTurnsCap {
		maxTurns = maxTurnsCap
	}

	// --- ASK LOOP START ---
	var finalResult lang.Value = &lang.NilValue{}
	currentOutput := initialPrompt
	var prevOutputHash string

	for turn := 1; turn <= maxTurns; turn++ {
		i.logger.Debug("Executing ask loop turn", "turn", turn, "max_turns", maxTurns)

		turnEnvelope := &aeiou.Envelope{Orchestration: currentOutput}
		composedRequest, err := turnEnvelope.Compose()
		if err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to compose turn envelope", err)
		}

		aiResp, err := callAIProvider(i, agentModel, composedRequest, step.GetPos())
		if err != nil {
			return nil, err
		}

		responseEnvelope, err := aeiou.Parse(aiResp.TextContent)
		if err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeSyntax, "failed to parse AEIOU response from model", err)
		}

		execInterp := i.clone()
		var actionEmits []string
		if err := executeAeiouTurn(execInterp, responseEnvelope, &actionEmits); err != nil {
			return nil, err
		}

		nextOutput := strings.Join(actionEmits, "\n")
		finalResult = lang.StringValue{Value: nextOutput}

		if !toolLoopPermitted {
			i.logger.Debug("Ask loop terminating: tool loop not permitted for this agent.")
			break
		}

		control := parseLoopControl(nextOutput)
		if control == stateAbort {
			i.logger.Warn("Ask loop terminating: received 'abort' marker from AI.")
			break
		}

		if control == stateDone {
			i.logger.Debug("Ask loop terminating: received 'done' marker from AI.")
			break
		}

		h := sha256.Sum256([]byte(nextOutput))
		currentHash := hex.EncodeToString(h[:])
		if currentHash == prevOutputHash {
			i.logger.Warn("Ask loop terminating: no progress detected between turns.")
			finalResult = lang.StringValue{Value: nextOutput + "\n[[loop:halt:reason=no-progress]]"}
			break
		}
		prevOutputHash = currentHash

		if control != stateContinue {
			i.logger.Debug("Ask loop terminating: no explicit 'continue' marker from AI.")
			break
		}

		if turn == maxTurns {
			i.logger.Warn("Ask loop terminating: max turns reached.")
			finalResult = lang.StringValue{Value: nextOutput + "\n[[loop:halt:reason=max-turns]]"}
			break
		}

		currentOutput = nextOutput
	}
	// --- ASK LOOP END ---

	if node.IntoTarget != nil {
		if err := i.setSingleLValue(node.IntoTarget, finalResult); err != nil {
			return nil, err
		}
	}

	return finalResult, nil
}

func parseLoopControl(output string) loopControlState {
	if strings.Contains(output, loopAbortMarker) {
		return stateAbort
	}
	if strings.Contains(output, loopDoneMarker) {
		return stateDone
	}
	if strings.Contains(output, loopContinueMarker) {
		return stateContinue
	}
	return stateDone
}

func callAIProvider(i *Interpreter, model types.AgentModel, prompt string, pos *types.Position) (*provider.AIResponse, error) {
	apiKey := ""
	if model.SecretRef != "" {
		apiKey = os.Getenv(model.SecretRef)
	}
	prov, provExists := i.GetProvider(model.Provider)
	if !provExists {
		return nil, lang.NewRuntimeError(lang.ErrorCodeProviderNotFound, fmt.Sprintf("provider '%s' for AgentModel '%s' not found", model.Provider, model.Name), nil).WithPosition(pos)
	}
	req := provider.AIRequest{ModelName: model.Model, Prompt: prompt, APIKey: apiKey}
	resp, err := prov.Chat(context.Background(), req)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeExternal, "AI provider call failed", err).WithPosition(pos)
	}
	return resp, nil
}

func executeAeiouTurn(i *Interpreter, envelope *aeiou.Envelope, emits *[]string) error {
	originalStdout := i.stdout
	r, w, _ := os.Pipe()
	i.SetStdout(w)
	defer func() {
		i.SetStdout(originalStdout)
		w.Close()
	}()
	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			*emits = append(*emits, scanner.Text())
		}
	}()

	if envelope.Implementations != "" {
		if err := executeAeiouSection(i, envelope.Implementations, "IMPLEMENTATIONS"); err != nil {
			return err
		}
	}
	if envelope.Events != "" {
		if err := executeAeiouSection(i, envelope.Events, "EVENTS"); err != nil {
			return err
		}
	}
	if envelope.Actions != "" {
		if _, err := executeAeiouActionSection(i, envelope.Actions); err != nil {
			return err
		}
	}
	return nil
}

func executeAeiouSection(i *Interpreter, content, sectionName string) error {
	parserAPI := parser.NewParserAPI(i.GetLogger())
	antlrTree, stream, err := parserAPI.ParseAndGetStream(fmt.Sprintf("aeiou_%s", sectionName), content)
	if err != nil {
		return lang.NewRuntimeError(lang.ErrorCodeSyntax, fmt.Sprintf("failed to parse %s section", sectionName), err)
	}
	builder := parser.NewASTBuilder(i.GetLogger())
	program, _, err := builder.BuildFromParseResult(antlrTree, stream)
	if err != nil {
		return lang.NewRuntimeError(lang.ErrorCodeSyntax, fmt.Sprintf("failed to build AST for %s section", sectionName), err)
	}
	return i.Load(program)
}

func executeAeiouActionSection(i *Interpreter, content string) (lang.Value, error) {
	if err := executeAeiouSection(i, content, "ACTIONS"); err != nil {
		return nil, err
	}
	return i.ExecuteCommands()
}
