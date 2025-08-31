// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Implements the AEIOU v3 host-side loop controller.
// filename: aeiou/loop.go
// nlines: 90
// risk_rating: HIGH

package aeiou

import (
	"bufio"
	"strings"
)

// Decision represents the outcome of a turn after processing the agent's output.
type Decision struct {
	Action LoopAction
	Notes  string
	Reason string
}

// LoopController orchestrates the host's decision-making process for each turn.
// It uses a MagicVerifier to validate control tokens from the agent's output.
type LoopController struct {
	verifier *MagicVerifier
}

// NewLoopController creates a new loop controller.
func NewLoopController(verifier *MagicVerifier) *LoopController {
	return &LoopController{verifier: verifier}
}

// ProcessOutput scans the agent's raw output, finds all valid control tokens,
// and selects a final decision based on the "precedence + last-wins" rule.
func (lc *LoopController) ProcessOutput(output string, hostCtx HostContext, replayCache *ReplayCache) (*Decision, error) {
	var validTokens []*TokenPayload

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, TokenMarkerPrefix) {
			continue
		}

		payload, err := lc.verifier.ParseAndVerify(line, hostCtx)
		if err != nil {
			// Ignore invalid tokens as per spec (they are just inert text).
			// A real host would log these failures.
			continue
		}

		// Check for replay before considering the token valid for this turn.
		if err := replayCache.CheckAndAdd(payload.JTI); err != nil {
			// This is a validly signed token that has been replayed.
			// The host should log this as a security event but treat it as inert.
			continue
		}

		validTokens = append(validTokens, payload)
	}

	if len(validTokens) == 0 {
		// No valid control token was found. The host should HALT.
		// We return a nil decision to signal this.
		return nil, nil
	}

	// Apply "precedence + last-wins" rule.
	return selectDecision(validTokens), nil
}

// selectDecision applies the precedence rules (abort > done > continue)
// and last-wins for ties.
func selectDecision(tokens []*TokenPayload) *Decision {
	var bestToken *TokenPayload

	for _, token := range tokens {
		if bestToken == nil {
			bestToken = token
			continue
		}

		// Precedence: abort > done > continue
		currentAction := bestToken.Payload.Action
		newAction := token.Payload.Action

		if newAction == ActionAbort {
			bestToken = token
		} else if newAction == ActionDone && currentAction != ActionAbort {
			bestToken = token
		} else if newAction == ActionContinue && currentAction != ActionAbort && currentAction != ActionDone {
			bestToken = token
		}
	}

	return &Decision{
		Action: bestToken.Payload.Action,
		Notes:  string(bestToken.Payload.Telemetry), // Simplification for now
		Reason: string(bestToken.Payload.Request),   // Simplification for now
	}
}
