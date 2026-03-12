// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 31
// :: description: Backported V4 features: Self-correction loop, explosive output tripwire, and split-emit fallback.
// :: latestChange: Integrated CleanLLMPayload, maxTurnBytes tripwire, and error feedback injection.
// :: filename: pkg/interpreter/steps_ask_hostloop.go
// :: serialization: go

package interpreter

import (
	"context"
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/llmconn"
	"github.com/aprice2704/neuroscript/pkg/types"
	"github.com/google/uuid"
)

const (
	maxTurnsCap           = 25
	progressGuardDefaultN = 3
	loopDoneSignal        = "<<<LOOP:DONE>>>"
	maxTurnBytes          = 524288 // 512KB limit to prevent context explosion
)

// runAskHostLoop orchestrates the turn-by-turn execution of an AEIOU v4 conversation.
func (i *Interpreter) runAskHostLoop(pos *types.Position, agentModel *types.AgentModel, conn llmconn.Connector, initialEnvelope *aeiou.Envelope) (lang.Value, error) {
	i.Logger().Info("--- Running V4 Native ask host loop (steps_ask_hostloop.go) ---")

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

	for turn := 1; turn <= maxTurns; turn++ {
		i.Logger().Debug("--- Starting ask loop turn ---", "sid", sessionID, "turn", turn)
		turnNonce := uuid.NewString()

		baseCtx := i.GetTurnContext()
		turnCtxForLLM := context.WithValue(baseCtx, AeiouSessionIDKey, sessionID)
		turnCtxForLLM = context.WithValue(turnCtxForLLM, AeiouTurnIndexKey, turn)
		turnCtxForLLM = context.WithValue(turnCtxForLLM, AeiouTurnNonceKey, turnNonce)

		if i.hostContext.AITranscript != nil {
			if composedPrompt, err := turnEnvelope.Compose(); err == nil {
				transcriptHeader := fmt.Sprintf("--- PROMPT (SID: %s, Turn: %d) ---\n", sessionID, turn)
				transcriptFooter := "\n--- END PROMPT ---\n\n"
				_, _ = i.hostContext.AITranscript.Write([]byte(transcriptHeader + composedPrompt + transcriptFooter))
			} else {
				i.Logger().Error("Failed to compose envelope for transcript", "sid", sessionID, "turn", turn, "error", err)
			}
		}

		aiResp, err := conn.Converse(turnCtxForLLM, turnEnvelope)
		if err != nil {
			if _, ok := err.(*lang.RuntimeError); !ok {
				return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "AI provider conversation failed", err).WithPosition(pos)
			}
			return nil, err
		}

		// --- Explosive Output Tripwire ---
		if len(aiResp.TextContent) > maxTurnBytes {
			i.Logger().Warn("Ask loop: Response size limit exceeded", "sid", sessionID, "turn", turn, "size", len(aiResp.TextContent))
			errorFeedback := fmt.Sprintf("SYSTEM_ERROR: Response size (%d bytes) exceeded the %d byte turn limit. Do not echo the USERDATA payload. Use tools (e.g., fs.Write) or split your emit statements across multiple turns to handle large data safely.", len(aiResp.TextContent), maxTurnBytes)

			turnEnvelope = &aeiou.Envelope{
				UserData:   turnEnvelope.UserData,
				Scratchpad: turnEnvelope.Scratchpad,
				Output:     errorFeedback,
				Actions:    "command\n  # Recover from size limit\nendcommand",
			}
			continue
		}

		// --- Parse AI Response with Defensive Self-Correction ---
		responseEnvelope, _, err := aeiou.Parse(strings.NewReader(aiResp.TextContent))
		if err != nil {
			i.Logger().Warn("Ask loop: Failed to parse AI response envelope, triggering self-correction", "sid", sessionID, "turn", turn, "error", err)
			errorFeedback := fmt.Sprintf("SYSTEM_ERROR: The previous ACTIONS block failed to parse or validate.\nError: %v\n\nEnsure you are using valid NeuroScript syntax.", err)

			turnEnvelope = &aeiou.Envelope{
				UserData:   turnEnvelope.UserData,
				Scratchpad: turnEnvelope.Scratchpad,
				Output:     errorFeedback,
				Actions:    "command\n  # Fix syntax error\nendcommand",
			}
			continue
		}

		execInterp := i.fork()
		execInterp.SetTurnContext(turnCtxForLLM)

		var actionEmits []string
		var actionWhispers = make(map[string]lang.Value)

		actionsTrimmed := strings.TrimSpace(responseEnvelope.Actions)
		if actionsTrimmed != "" && (!strings.HasPrefix(actionsTrimmed, "command") || !strings.HasSuffix(actionsTrimmed, "endcommand")) {
			// Instead of returning error, feed back to AI
			errorFeedback := "SYSTEM_ERROR: ACTIONS block must contain exactly one 'command...endcommand' block or be empty."
			turnEnvelope = &aeiou.Envelope{
				UserData:   turnEnvelope.UserData,
				Scratchpad: turnEnvelope.Scratchpad,
				Output:     errorFeedback,
				Actions:    "command\n  # Fix syntax error\nendcommand",
			}
			continue
		}

		// --- Execute ACTIONS Block with Defensive Self-Correction ---
		if err := executeAeiouTurn(execInterp, responseEnvelope, &actionEmits, &actionWhispers); err != nil {
			i.Logger().Warn("Ask loop: ACTIONS execution failed, triggering self-correction", "sid", sessionID, "turn", turn, "error", err)
			hint := ""
			if strings.Contains(err.Error(), "mismatched input 'tool'") {
				hint = " (Hint: Did you forget the 'call' keyword before a tool invocation?)"
			} else if strings.Contains(err.Error(), "mismatched input '='") {
				hint = " (Hint: Tool arguments MUST be passed as a JSON object, not Python kwargs. Use `call tool.name({\"key\":\"val\"})`)"
			}

			errorFeedback := fmt.Sprintf("SYSTEM_ERROR: The previous ACTIONS block failed during execution.\nError: %v%s\n\nEnsure you are using valid NeuroScript syntax.", err, hint)

			turnEnvelope = &aeiou.Envelope{
				UserData:   turnEnvelope.UserData,
				Scratchpad: turnEnvelope.Scratchpad,
				Output:     errorFeedback,
				Actions:    "command\n  # Fix execution error\nendcommand",
			}
			continue
		}

		// --- Process Emits for Loop Control and Result ---
		var cleanEmits []string
		var loopDone bool
		var markerExtracted string

		for _, emit := range actionEmits {
			trimmedEmit := strings.TrimSpace(emit)
			if strings.Contains(trimmedEmit, loopDoneSignal) {
				loopDone = true
				i.Logger().Debug("<<<LOOP:DONE>>> signal detected", "sid", sessionID, "turn", turn)

				cleaned := aeiou.CleanLLMPayload(trimmedEmit)
				if cleaned != "" {
					markerExtracted = cleaned
				}
				continue
			}
			// Filter out vestigial V3 tokens if they somehow appear
			if !strings.HasPrefix(trimmedEmit, "<<<NSMAG:V3") {
				cleanEmits = append(cleanEmits, emit)
			}
		}

		outputBody := strings.Join(cleanEmits, "\n")

		// If Loop Done, set result based on precedence (inline marker payload vs split-emit)
		if loopDone {
			if markerExtracted != "" {
				finalResult = lang.StringValue{Value: markerExtracted}
			} else if outputBody != "" {
				finalResult = lang.StringValue{Value: outputBody}
			} else {
				finalResult = &lang.NilValue{}
			}
			i.Logger().Debug("Ask loop terminating: AI emitted DONE signal.", "sid", sessionID, "turn", turn)
			break
		}

		// Turn considered valid if anything happened (emit) or loop continued
		if len(actionEmits) > 0 {
			finalResult = lang.StringValue{Value: outputBody}
		}

		// --- Progress Guard ---
		fullOutputBody := strings.Join(actionEmits, "\n")
		scratchpadBody, _ := lang.ToString(lang.NewMapValue(actionWhispers))
		digest := aeiou.ComputeHostDigest(fullOutputBody, scratchpadBody)
		if progressTracker.CheckAndRecord(digest) {
			i.Logger().Warn("Ask loop terminating: no progress detected.", "sid", sessionID, "turn", turn, "last_digest", digest)
			return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "ask loop stuck in repetitive state", nil).WithPosition(pos)
		}

		// --- Termination Checks ---
		if turn == maxTurns {
			i.Logger().Debug("Ask loop terminating: max turns reached.", "sid", sessionID, "max_turns", maxTurns)
			break
		}

		// --- Prepare for Next Turn ---
		turnEnvelope = &aeiou.Envelope{
			UserData:   turnEnvelope.UserData,
			Scratchpad: scratchpadBody,
			Output:     outputBody,
			Actions:    "command\n  # Execute next step\nendcommand",
		}
	}

	i.Logger().Debug("--- EXITING ask host loop ---", "sid", sessionID)
	return finalResult, nil
}
