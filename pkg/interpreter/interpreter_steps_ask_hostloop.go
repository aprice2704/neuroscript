// NeuroScript Version: 0.7.0
// File version: 16
// Purpose: FIX: Correctly chains the turn context across loop iterations, preserving the turn index.
// filename: pkg/interpreter/interpreter_steps_ask_hostloop.go
// nlines: 232
// risk_rating: HIGH

package interpreter

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/llmconn"
	"github.com/aprice2704/neuroscript/pkg/types"
	"github.com/google/uuid"
)

const (
	maxTurnsCap           = 25
	progressGuardDefaultN = 3
)

type transientKeyProvider struct {
	publicKey ed25519.PublicKey
}

func (p *transientKeyProvider) PublicKey(kid string) (ed25519.PublicKey, error) {
	return p.publicKey, nil
}

func (i *Interpreter) runAskHostLoop(pos *types.Position, agentModel *types.AgentModel, conn llmconn.Connector, initialEnvelope *aeiou.Envelope) (lang.Value, error) {
	maxTurns := agentModel.MaxTurns
	if maxTurns <= 0 {
		maxTurns = 1
	}
	if maxTurns > maxTurnsCap {
		maxTurns = maxTurnsCap
	}

	// FIX: This becomes the starting point for our context chain.
	currentCtx := i.GetTurnContext()
	sessionID, ok := currentCtx.Value(aeiou.SessionIDKey).(string)
	if !ok || sessionID == "" {
		fmt.Fprintf(os.Stderr, "[DEBUG runAskHostLoop] SID not found in context, generating new one.\n")
		sessionID = uuid.NewString()
		currentCtx = context.WithValue(currentCtx, aeiou.SessionIDKey, sessionID)
	} else {
		fmt.Fprintf(os.Stderr, "[DEBUG runAskHostLoop] Inherited SID from context: %q\n", sessionID)
	}

	progressTracker := aeiou.NewProgressTracker(progressGuardDefaultN)
	var finalResult lang.Value = &lang.NilValue{}
	turnEnvelope := initialEnvelope

	keyProvider := &transientKeyProvider{publicKey: i.transientPrivateKey.Public().(ed25519.PublicKey)}
	verifier := aeiou.NewMagicVerifier(keyProvider)
	loopController := aeiou.NewLoopController(verifier)
	replayCache := aeiou.NewReplayCache(100, 5*time.Minute)

	for turn := 1; turn <= maxTurns; turn++ {
		i.logger.Debug("--- Starting ask loop turn ---", "sid", sessionID, "turn", turn)

		turnNonce := uuid.NewString()
		hostCtx := aeiou.HostContext{
			SessionID: sessionID,
			TurnIndex: turn,
			TurnNonce: turnNonce,
			KeyID:     "transient-key-01",
		}
		// FIX: Create the next turn's context from the *previous* turn's context.
		turnCtxForProvider := context.WithValue(currentCtx, aeiou.TurnIndexKey, turn)
		turnCtxForProvider = context.WithValue(turnCtxForProvider, aeiou.TurnNonceKey, turnNonce)
		i.SetTurnContext(turnCtxForProvider)

		if i.aiTranscript != nil {
			if composedPrompt, err := turnEnvelope.Compose(); err == nil {
				transcriptHeader := fmt.Sprintf("--- PROMPT (SID: %s, Turn: %d) ---\n", sessionID, turn)
				transcriptFooter := "\n--- END PROMPT ---\n\n"
				_, _ = i.aiTranscript.Write([]byte(transcriptHeader + composedPrompt + transcriptFooter))
			}
		}

		aiResp, err := conn.Converse(turnCtxForProvider, turnEnvelope)
		if err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeExternal, "AI provider call failed", err).WithPosition(pos)
		}

		responseEnvelope, _, err := aeiou.Parse(strings.NewReader(aiResp.TextContent))
		if err != nil {
			i.logger.Warn("Failed to parse AI response envelope, attempting recovery", "sid", sessionID, "turn", turn, "error", err)
			diagnosticUserData := fmt.Sprintf(`{"error": "envelope parsing failed", "diagnostic": "%s"}`, err.Error())
			turnEnvelope = &aeiou.Envelope{
				UserData: diagnosticUserData,
				Actions:  "command endcommand",
			}
			continue
		}

		execInterp := i.clone()
		fmt.Fprintf(os.Stderr, "[DEBUG runAskHostLoop] Executing provider response in CLONE ID: %s. Parent ID was: %s\n", execInterp.id, i.id)

		var actionEmits []string
		var actionWhispers = make(map[string]lang.Value)

		actionsTrimmed := strings.TrimSpace(responseEnvelope.Actions)
		if !strings.HasPrefix(actionsTrimmed, "command") || !strings.HasSuffix(actionsTrimmed, "endcommand") {
			return nil, lang.NewRuntimeError(lang.ErrorCodeSyntax, "ACTIONS block must contain exactly one 'command...endcommand' block", nil).WithPosition(pos)
		}

		if err := executeAeiouTurn(execInterp, responseEnvelope, &actionEmits, &actionWhispers); err != nil {
			return nil, err
		}

		var lastNonEmptyLine string
		for j := len(actionEmits) - 1; j >= 0; j-- {
			if strings.TrimSpace(actionEmits[j]) != "" {
				lastNonEmptyLine = actionEmits[j]
				break
			}
		}

		if _, err := verifier.ParseAndVerify(lastNonEmptyLine, hostCtx); err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodePolicy, "The last non-empty emitted line must be a valid tool.aeiou.magic() control token.", err).WithPosition(pos)
		}

		var cleanEmits []string
		for _, emit := range actionEmits {
			if !strings.HasPrefix(emit, aeiou.TokenMarkerPrefix) {
				cleanEmits = append(cleanEmits, emit)
			}
		}
		outputBody := strings.Join(cleanEmits, "\n")
		if strings.TrimSpace(outputBody) != "" {
			finalResult = lang.StringValue{Value: outputBody}
		}

		fullOutputBody := strings.Join(actionEmits, "\n")
		scratchpadBody, _ := lang.ToString(lang.NewMapValue(actionWhispers))

		digest := aeiou.ComputeHostDigest(fullOutputBody, scratchpadBody)
		if progressTracker.CheckAndRecord(digest) {
			i.logger.Warn("Ask loop terminating: no progress detected.", "sid", sessionID, "turn", turn)
			break
		}

		decision, err := loopController.ProcessOutput(fullOutputBody, hostCtx, replayCache)
		if err != nil || decision == nil {
			i.logger.Debug("Ask loop terminating: no valid V3 control token found or error during processing.", "sid", sessionID, "turn", turn, "error", err)
			break
		}

		if decision.Action == aeiou.ActionContinue {
			if !agentModel.Tools.ToolLoopPermitted {
				return nil, lang.NewRuntimeError(lang.ErrorCodePolicy, fmt.Sprintf("agent model '%s' attempted to continue a loop but does not have 'toolLoopPermitted' grant", agentModel.Name), nil).WithPosition(pos)
			}
		} else {
			i.logger.Debug("Ask loop terminating: received signal.", "sid", sessionID, "turn", turn, "signal", decision.Action)
			break
		}

		if turn == maxTurns {
			i.logger.Warn("Ask loop terminating: max turns reached.", "sid", sessionID, "max_turns", maxTurns)
			break
		}

		turnEnvelope = &aeiou.Envelope{
			UserData:   outputBody,
			Scratchpad: scratchpadBody,
			Output:     outputBody,
			Actions:    "command endcommand",
		}
		// FIX: Update the context for the next iteration.
		currentCtx = turnCtxForProvider
	}

	i.logger.Debug("--- EXITING ask host loop ---", "sid", sessionID)
	return finalResult, nil
}
