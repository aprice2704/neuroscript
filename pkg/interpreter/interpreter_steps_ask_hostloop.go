// NeuroScript Version: 0.7.0
// File version: 11
// Purpose: Replaced the faulty key provider with one that correctly uses the interpreter's public key, fixing the cryptographic panic, and refined the final result logic to exclude the control token.
// filename: pkg/interpreter/interpreter_steps_ask_hostloop.go
// nlines: 187
// risk_rating: HIGH

package interpreter

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"strings"
	"time"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/llmconn"
	"github.com/aprice2704/neuroscript/pkg/types"
	"github.com/google/uuid"
)

const (
	// maxTurnsCap is a hard safety limit on the number of turns in an ask loop.
	maxTurnsCap = 25
	// progressGuardDefaultN is the number of consecutive identical digests that trigger a HALT.
	progressGuardDefaultN = 3
)

// transientKeyProvider provides the public key derived from the interpreter's
// transient private key, allowing the verifier to check token signatures.
type transientKeyProvider struct {
	publicKey ed25519.PublicKey
}

func (p *transientKeyProvider) PublicKey(kid string) (ed25519.PublicKey, error) {
	// For now, we only have one transient key. A real implementation would
	// check the kid against a key registry.
	return p.publicKey, nil
}

// runAskHostLoop orchestrates the turn-by-turn execution of an AEIOU v3 conversation.
func (i *Interpreter) runAskHostLoop(pos *types.Position, agentModel *types.AgentModel, conn llmconn.Connector, initialEnvelope *aeiou.Envelope) (lang.Value, error) {
	// 1. Initialize Host Loop Controller State
	maxTurns := agentModel.MaxTurns
	if maxTurns <= 0 {
		maxTurns = 1
	}
	if maxTurns > maxTurnsCap {
		i.logger.Warn("AgentModel max_turns exceeds cap", "agent", agentModel.Name, "requested", agentModel.MaxTurns, "capped_at", maxTurnsCap)
		maxTurns = maxTurnsCap
	}

	sessionID := uuid.NewString()
	progressTracker := aeiou.NewProgressTracker(progressGuardDefaultN)
	var finalResult lang.Value = &lang.NilValue{}
	turnEnvelope := initialEnvelope

	// Correctly initialize the verifier with the public key.
	keyProvider := &transientKeyProvider{publicKey: i.transientPrivateKey.Public().(ed25519.PublicKey)}
	verifier := aeiou.NewMagicVerifier(keyProvider)
	loopController := aeiou.NewLoopController(verifier)
	replayCache := aeiou.NewReplayCache(100, 5*time.Minute)

	// 2. --- AEIOU v3 HOST LOOP ---
	for turn := 1; turn <= maxTurns; turn++ {
		i.logger.Debug("--- Starting ask loop turn ---", "sid", sessionID, "turn", turn)

		// A. Create and set the context for this specific turn.
		turnNonce := uuid.NewString()
		hostCtx := aeiou.HostContext{
			SessionID: sessionID,
			TurnIndex: turn,
			TurnNonce: turnNonce,
			KeyID:     "transient-key-01", // Match the hardcoded KID in the tool
		}
		turnCtxForInterpreter := context.WithValue(context.Background(), aeiouSessionIDKey, sessionID)
		turnCtxForInterpreter = context.WithValue(turnCtxForInterpreter, aeiouTurnIndexKey, turn)
		turnCtxForInterpreter = context.WithValue(turnCtxForInterpreter, aeiouTurnNonceKey, turnNonce)
		i.setTurnContext(turnCtxForInterpreter)

		// B. Log transcript if configured
		if i.aiTranscript != nil {
			if composedPrompt, err := turnEnvelope.Compose(); err == nil {
				transcriptHeader := fmt.Sprintf("--- PROMPT (SID: %s, Turn: %d) ---\n", sessionID, turn)
				transcriptFooter := "\n--- END PROMPT ---\n\n"
				_, _ = i.aiTranscript.Write([]byte(transcriptHeader + composedPrompt + transcriptFooter))
			}
		}

		// C. Call the LLM
		aiResp, err := conn.Converse(turnCtxForInterpreter, turnEnvelope)
		if err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeExternal, "AI provider call failed", err).WithPosition(pos)
		}

		responseEnvelope, _, err := aeiou.Parse(strings.NewReader(aiResp.TextContent))
		if err != nil {
			isOneShotAgent := maxTurns == 1 && !agentModel.Tools.ToolLoopPermitted
			if isOneShotAgent {
				i.logger.Debug("Response is not an envelope; treating as final answer for one-shot agent.")
				finalResult = lang.StringValue{Value: aiResp.TextContent}
				break
			}
			return nil, lang.NewRuntimeError(lang.ErrorCodeSyntax, "failed to parse AEIOU envelope from AI response", err).WithPosition(pos)
		}

		// D. Execute ACTIONS and capture outputs
		execInterp := i.clone()
		var actionEmits []string
		var actionWhispers = make(map[string]lang.Value)
		if err := executeAeiouTurn(execInterp, responseEnvelope, &actionEmits, &actionWhispers); err != nil {
			return nil, err
		}

		// Filter out the magic token from the final result.
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

		// E. Mandatory Progress Guard
		digest := aeiou.ComputeHostDigest(fullOutputBody, scratchpadBody)
		if progressTracker.CheckAndRecord(digest) {
			i.logger.Warn("Ask loop terminating: no progress detected.", "sid", sessionID, "turn", turn)
			break
		}

		// F. Make Control Decision using the V3 LoopController
		decision, err := loopController.ProcessOutput(fullOutputBody, hostCtx, replayCache)
		if err != nil || decision == nil {
			i.logger.Debug("Ask loop terminating: no valid V3 control token found or error during processing.", "sid", sessionID, "turn", turn, "error", err)
			break // Halt
		}

		if decision.Action == aeiou.ActionContinue {
			if !agentModel.Tools.ToolLoopPermitted {
				return nil, lang.NewRuntimeError(lang.ErrorCodePolicy, fmt.Sprintf("agent model '%s' attempted to continue a loop but does not have 'toolLoopPermitted' grant", agentModel.Name), nil).WithPosition(pos)
			}
		} else {
			i.logger.Debug("Ask loop terminating: received signal.", "sid", sessionID, "turn", turn, "signal", decision.Action)
			break // Done or Abort
		}

		if turn == maxTurns {
			i.logger.Warn("Ask loop terminating: max turns reached.", "sid", sessionID, "max_turns", maxTurns)
			break
		}

		// G. Prepare Envelope for Next Turn
		turnEnvelope = &aeiou.Envelope{
			UserData:   fullOutputBody, // The next turn gets the full output, including the token
			Scratchpad: scratchpadBody,
			Actions:    "command endcommand",
		}
	}

	i.logger.Debug("--- EXITING ask host loop ---", "sid", sessionID)
	return finalResult, nil
}
