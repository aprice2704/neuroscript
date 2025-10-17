// NeuroScript Version: 0.8.0
// File version: 25
// Purpose: Replaces undefined 'lang.ErrorCodeFormat' with 'lang.ErrorCodeSyntax' to fix compiler error.
// filename: pkg/interpreter/steps_ask_hostloop.go
// nlines: 188
// risk_rating: HIGH

package interpreter

import (
	"context"
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/llmconn" // Restored dependency

	// "github.com/aprice2704/neuroscript/pkg/provider" // No longer needed here
	"github.com/aprice2704/neuroscript/pkg/types"
	"github.com/google/uuid"
)

const (
	// maxTurnsCap is a hard safety limit on the number of turns in an ask loop.
	maxTurnsCap = 25
	// progressGuardDefaultN is the number of consecutive identical digests that trigger a HALT.
	progressGuardDefaultN = 3
	// loopDoneSignal is the simple string an AI can emit to signal task completion.
	loopDoneSignal = "<<<LOOP:DONE>>>"
)

// runAskHostLoop orchestrates the turn-by-turn execution of an AEIOU v3 conversation.
// It now accepts the llmconn.Connector again.
func (i *Interpreter) runAskHostLoop(pos *types.Position, agentModel *types.AgentModel, conn llmconn.Connector, initialEnvelope *aeiou.Envelope) (lang.Value, error) {
	maxTurns := agentModel.MaxTurns
	if maxTurns <= 0 {
		maxTurns = 1 // Default to one-shot if not specified or invalid
	}
	if maxTurns > maxTurnsCap {
		i.Logger().Warn("AgentModel max_turns exceeds cap", "agent", agentModel.Name, "requested", agentModel.MaxTurns, "capped_at", maxTurnsCap)
		maxTurns = maxTurnsCap
	}

	sessionID := uuid.NewString()
	progressTracker := aeiou.NewProgressTracker(progressGuardDefaultN)
	var finalResult lang.Value = &lang.NilValue{} // Initialize final result
	turnEnvelope := initialEnvelope

	for turn := 1; turn <= maxTurns; turn++ {
		i.Logger().Debug("--- Starting ask loop turn ---", "sid", sessionID, "turn", turn)

		turnNonce := uuid.NewString()

		// --- Context Propagation ---
		baseCtx := i.GetTurnContext() // Get context possibly set by caller (e.g., api.RunProcedure)
		// Add session/turn info for this specific loop iteration
		turnCtxForLLM := context.WithValue(baseCtx, AeiouSessionIDKey, sessionID)
		turnCtxForLLM = context.WithValue(turnCtxForLLM, AeiouTurnIndexKey, turn)
		turnCtxForLLM = context.WithValue(turnCtxForLLM, AeiouTurnNonceKey, turnNonce)
		// We use turnCtxForLLM for the Converse call.
		fmt.Printf("[DEBUG] runAskHostLoop: Turn %d. Context for LLM call %p (derived from base %p).\n", turn, turnCtxForLLM, baseCtx)

		// --- Transcript Logging ---
		if i.hostContext.AITranscript != nil {
			if composedPrompt, err := turnEnvelope.Compose(); err == nil {
				transcriptHeader := fmt.Sprintf("--- PROMPT (SID: %s, Turn: %d) ---\n", sessionID, turn)
				transcriptFooter := "\n--- END PROMPT ---\n\n"
				_, _ = i.hostContext.AITranscript.Write([]byte(transcriptHeader + composedPrompt + transcriptFooter))
			} else {
				i.Logger().Error("Failed to compose envelope for transcript", "sid", sessionID, "turn", turn, "error", err)
			}
		}

		// --- Call AI Provider via Connector ---
		// The conn.Converse method handles interaction with the underlying provider.Chat
		aiResp, err := conn.Converse(turnCtxForLLM, turnEnvelope)
		if err != nil {
			// Converse wraps provider errors; return the wrapped error directly.
			// No need for lang.NewRuntimeError here, Converse should return one if appropriate.
			return nil, err // Return error from Converse (could be provider error, context canceled, etc.)
		}

		// --- Parse AI Response ---
		responseEnvelope, _, err := aeiou.Parse(strings.NewReader(aiResp.TextContent))
		if err != nil {
			// Handle parse failure - potentially log and try to continue if robust handling is needed,
			// or fail the loop. For now, we fail.
			i.Logger().Error("Failed to parse AI response envelope", "sid", sessionID, "turn", turn, "error", err, "raw_response", aiResp.TextContent)
			// THE FIX: Changed lang.ErrorCodeFormat (which doesn't exist) to lang.ErrorCodeSyntax.
			return nil, lang.NewRuntimeError(lang.ErrorCodeSyntax, "failed to parse AI response envelope", err).WithPosition(pos)
			/* // Recovery attempt (optional):
			   diagnosticUserData := fmt.Sprintf(`{"error": "envelope parsing failed", "diagnostic": "%s"}`, err.Error())
			   turnEnvelope = &aeiou.Envelope{ UserData: diagnosticUserData, Actions: "command endcommand" }
			   continue
			*/
		}

		// --- Execute ACTIONS Block ---
		execInterp := i.fork()
		// CRITICAL: The forked interpreter needs the correct context for THIS turn
		// so that tools called within the ACTIONS block get the right context.
		execInterp.SetTurnContext(turnCtxForLLM) // Pass the turn-specific context
		fmt.Printf("[DEBUG] runAskHostLoop: Turn %d. Set context on EXEC interpreter %s. Ctx %p.\n", turn, execInterp.id, turnCtxForLLM)

		var actionEmits []string // Capture emits as strings
		var actionWhispers = make(map[string]lang.Value)

		actionsTrimmed := strings.TrimSpace(responseEnvelope.Actions)
		// Allow empty ACTIONS block
		if actionsTrimmed != "" && (!strings.HasPrefix(actionsTrimmed, "command") || !strings.HasSuffix(actionsTrimmed, "endcommand")) {
			return nil, lang.NewRuntimeError(lang.ErrorCodeSyntax, "ACTIONS block must contain exactly one 'command...endcommand' block or be empty", nil).WithPosition(pos)
		}

		// Execute the AI's commands in the sandboxed interpreter
		if err := executeAeiouTurn(execInterp, responseEnvelope, &actionEmits, &actionWhispers); err != nil {
			// Propagate errors from executing the ACTIONS block (e.g., tool errors, syntax errors)
			return nil, err
		}

		// --- Process Emits for Loop Control and Result ---
		var cleanEmits []string
		var loopDone bool
		for _, emit := range actionEmits {
			trimmedEmit := strings.TrimSpace(emit)
			if trimmedEmit == loopDoneSignal {
				loopDone = true
				i.Logger().Debug("<<<LOOP:DONE>>> signal detected", "sid", sessionID, "turn", turn)
				continue // Don't include the signal in the final output or next prompt
			}
			// Filter out any vestigial AEIOU v2 tokens just in case
			if !strings.HasPrefix(trimmedEmit, aeiou.TokenMarkerPrefix) {
				cleanEmits = append(cleanEmits, emit) // Keep the original emit, not trimmed
			}
		}
		outputBody := strings.Join(cleanEmits, "\n")
		// The final result of the 'ask' statement is the *last* non-empty outputBody produced.
		// Update finalResult only if there's actual content emitted this turn.
		// Allow empty string as a valid final result if the AI emits nothing then DONE.
		if len(actionEmits) > 0 || loopDone { // Consider turn valid if anything happened (emit or DONE)
			finalResult = lang.StringValue{Value: outputBody}
		}

		// --- Progress Guard ---
		fullOutputBody := strings.Join(actionEmits, "\n") // Use all emits for digest
		scratchpadBody, _ := lang.ToString(lang.NewMapValue(actionWhispers))
		digest := aeiou.ComputeHostDigest(fullOutputBody, scratchpadBody)
		if progressTracker.CheckAndRecord(digest) {
			i.Logger().Warn("Ask loop terminating: no progress detected.", "sid", sessionID, "turn", turn, "last_digest", digest)
			break // Terminate due to lack of progress
		}

		// --- Termination Checks ---
		if loopDone {
			i.Logger().Debug("Ask loop terminating: AI emitted DONE signal.", "sid", sessionID, "turn", turn)
			break // Terminate because AI signaled completion
		}

		if turn == maxTurns {
			i.Logger().Debug("Ask loop terminating: max turns reached.", "sid", sessionID, "max_turns", maxTurns)
			break // Terminate due to reaching turn limit
		}

		// --- Prepare for Next Turn ---
		// UserData for the next turn is the clean output from this turn.
		turnEnvelope = &aeiou.Envelope{
			UserData:   outputBody,
			Scratchpad: scratchpadBody,
			Output:     outputBody,           // Output section mirrors UserData for next prompt
			Actions:    "command endcommand", // Default empty actions for next turn
		}
	} // End of turn loop

	i.Logger().Debug("--- EXITING ask host loop ---", "sid", sessionID)
	// Return the result accumulated from the last valid turn.
	return finalResult, nil
}
