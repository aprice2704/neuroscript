// NeuroScript Version: 0.8.0
// File version: 13
// Purpose: Updates the AEIOU host loop to use HostContext for logging and AI transcripts.
// filename: pkg/interpreter/steps_ask_hostloop.go
// nlines: 220
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
	return p.publicKey, nil
}

// runAskHostLoop orchestrates the turn-by-turn execution of an AEIOU v3 conversation.
func (i *Interpreter) runAskHostLoop(pos *types.Position, agentModel *types.AgentModel, conn llmconn.Connector, initialEnvelope *aeiou.Envelope) (lang.Value, error) {
	maxTurns := agentModel.MaxTurns
	if maxTurns <= 0 {
		maxTurns = 1
	}
	if maxTurns > maxTurnsCap {
		i.Logger().Warn("AgentModel max_turns exceeds cap", "agent", agentModel.Name, "requested", agentModel.MaxTurns, "capped_at", maxTurnsCap)
		maxTurns = maxTurnsCap
	}

	sessionID := uuid.NewString()
	progressTracker := aeiou.NewProgressTracker(progressGuardDefaultN)
	var finalResult lang.Value = &lang.NilValue{}
	turnEnvelope := initialEnvelope

	keyProvider := &transientKeyProvider{publicKey: i.transientPrivateKey.Public().(ed25519.PublicKey)}
	verifier := aeiou.NewMagicVerifier(keyProvider)
	loopController := aeiou.NewLoopController(verifier)
	replayCache := aeiou.NewReplayCache(100, 5*time.Minute)

	for turn := 1; turn <= maxTurns; turn++ {
		i.Logger().Debug("--- Starting ask loop turn ---", "sid", sessionID, "turn", turn)

		turnNonce := uuid.NewString()
		hostCtx := aeiou.HostContext{
			SessionID: sessionID,
			TurnIndex: turn,
			TurnNonce: turnNonce,
			KeyID:     "transient-key-01",
		}
		turnCtxForInterpreter := context.WithValue(context.Background(), aeiouSessionIDKey, sessionID)
		turnCtxForInterpreter = context.WithValue(turnCtxForInterpreter, aeiouTurnIndexKey, turn)
		turnCtxForInterpreter = context.WithValue(turnCtxForInterpreter, aeiouTurnNonceKey, turnNonce)
		i.setTurnContext(turnCtxForInterpreter)

		if i.hostContext.AITranscript != nil {
			if composedPrompt, err := turnEnvelope.Compose(); err == nil {
				transcriptHeader := fmt.Sprintf("--- PROMPT (SID: %s, Turn: %d) ---\n", sessionID, turn)
				transcriptFooter := "\n--- END PROMPT ---\n\n"
				_, _ = i.hostContext.AITranscript.Write([]byte(transcriptHeader + composedPrompt + transcriptFooter))
			}
		}

		aiResp, err := conn.Converse(turnCtxForInterpreter, turnEnvelope)
		if err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeExternal, "AI provider call failed", err).WithPosition(pos)
		}

		responseEnvelope, _, err := aeiou.Parse(strings.NewReader(aiResp.TextContent))
		if err != nil {
			i.Logger().Warn("Failed to parse AI response envelope, attempting recovery", "sid", sessionID, "turn", turn, "error", err)
			diagnosticUserData := fmt.Sprintf(`{"error": "envelope parsing failed", "diagnostic": "%s"}`, err.Error())
			turnEnvelope = &aeiou.Envelope{
				UserData: diagnosticUserData,
				Actions:  "command endcommand",
			}
			continue
		}

		execInterp := i.fork()
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
			i.Logger().Warn("Ask loop terminating: no progress detected.", "sid", sessionID, "turn", turn)
			break
		}

		decision, err := loopController.ProcessOutput(fullOutputBody, hostCtx, replayCache)
		if err != nil || decision == nil {
			i.Logger().Debug("Ask loop terminating: no valid V3 control token found or error during processing.", "sid", sessionID, "turn", turn, "error", err)
			break // Halt
		}

		if decision.Action == aeiou.ActionContinue {
			if !agentModel.Tools.ToolLoopPermitted {
				return nil, lang.NewRuntimeError(lang.ErrorCodePolicy, fmt.Sprintf("agent model '%s' attempted to continue a loop but does not have 'toolLoopPermitted' grant", agentModel.Name), nil).WithPosition(pos)
			}
		} else {
			i.Logger().Debug("Ask loop terminating: received signal.", "sid", sessionID, "turn", turn, "signal", decision.Action)
			break // Done or Abort
		}

		if turn == maxTurns {
			i.Logger().Warn("Ask loop terminating: max turns reached.", "sid", sessionID, "max_turns", maxTurns)
			break
		}

		turnEnvelope = &aeiou.Envelope{
			UserData:   outputBody,
			Scratchpad: scratchpadBody,
			Output:     outputBody,
			Actions:    "command endcommand",
		}
	}

	i.Logger().Debug("--- EXITING ask host loop ---", "sid", sessionID)
	return finalResult, nil
}
