// NeuroScript Version: 0.7.0
// File version: 4
// Purpose: Corrects the token selection logic to properly handle "last-wins".
// filename: aeiou/loop.go
// nlines: 138
// risk_rating: HIGH

package aeiou

import (
	"strings"
)

// Decision represents the outcome of a turn after processing the agent's output.
type Decision struct {
	Action LoopAction
	Notes  string
	Reason string
	Lints  []Lint
}

// candidateToken is a private struct to hold a valid token and its location.
type candidateToken struct {
	payload *TokenPayload
	lineNum int
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
// It also detects and reports non-fatal lints, such as text after the token.
func (lc *LoopController) ProcessOutput(output string, hostCtx HostContext, replayCache *ReplayCache) (*Decision, error) {
	var candidates []candidateToken
	lineNum := 0

	outputLines := strings.Split(output, "\n")

	for i, line := range outputLines {
		lineNum = i + 1
		if !strings.HasPrefix(line, TokenMarkerPrefix) {
			continue
		}

		payload, err := lc.verifier.ParseAndVerify(line, hostCtx)
		if err != nil {
			// Ignore invalid tokens as per spec (they are just inert text).
			continue
		}

		if err := replayCache.CheckAndAdd(payload.JTI); err != nil {
			// This is a validly signed token that has been replayed. Treat as inert.
			continue
		}

		candidates = append(candidates, candidateToken{payload: payload, lineNum: lineNum})
	}

	if len(candidates) == 0 {
		return nil, nil // No valid, non-replayed token found.
	}

	// Apply "precedence + last-wins" rule.
	winner := selectDecision(candidates)

	decision := &Decision{
		Action: winner.payload.Payload.Action,
		Notes:  string(winner.payload.Payload.Telemetry),
		Reason: string(winner.payload.Payload.Request),
	}

	// Check for post-token text lint.
	for i := winner.lineNum; i < len(outputLines); i++ {
		if strings.TrimSpace(outputLines[i]) != "" {
			decision.Lints = append(decision.Lints, Lint{
				Code:    LintCodePostTokenText,
				Message: "extraneous text found after the chosen control token",
			})
			break // Only need to report the lint once.
		}
	}

	return decision, nil
}

// getActionPrecedence returns a numerical weight for a given action.
// Higher numbers have higher precedence.
func getActionPrecedence(action LoopAction) int {
	switch action {
	case ActionAbort:
		return 3
	case ActionDone:
		return 2
	case ActionContinue:
		return 1
	default:
		return 0
	}
}

// selectDecision applies the precedence rules (abort > done > continue)
// and last-wins for ties.
func selectDecision(candidates []candidateToken) candidateToken {
	best := candidates[0]

	for i := 1; i < len(candidates); i++ {
		current := candidates[i]

		currentBestPrecedence := getActionPrecedence(best.payload.Payload.Action)
		newPrecedence := getActionPrecedence(current.payload.Payload.Action)

		// If the new token has higher or equal precedence, it becomes the new best.
		// This correctly implements "last-wins" for ties because we are iterating
		// in the order the tokens appeared.
		if newPrecedence >= currentBestPrecedence {
			best = current
		}
	}

	return best
}
