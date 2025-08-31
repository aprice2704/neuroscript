// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Provides a helper for parsing Ask-Loop control signals.
// filename: neuroscript/pkg/aeiou/loop.go
// nlines: 40
// risk_rating: LOW

package aeiou

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

// loopControlRegex finds the first V2 LOOP magic marker and captures its JSON payload.
var loopControlRegex = regexp.MustCompile(`<<<NSENVELOPE_MAGIC_[A-F0-9]+:V2:LOOP:(.*?)>>>`)

// LoopControl holds the parsed result of an AEIOU LOOP signal.
type LoopControl struct {
	Control string `json:"control"` // "continue", "done", or "abort"
	Notes   string `json:"notes"`
	Reason  string `json:"reason"`
}

// ParseLoopControl scans a string (like the captured output from an emit stream)
// and extracts the first valid AEIOU LOOP control signal it finds.
func ParseLoopControl(output string) (*LoopControl, error) {
	matches := loopControlRegex.FindStringSubmatch(output)
	if len(matches) < 2 {
		return nil, errors.New("no valid LOOP control signal found in output")
	}

	jsonPayload := matches[1]
	var control LoopControl
	if err := json.Unmarshal([]byte(jsonPayload), &control); err != nil {
		return nil, fmt.Errorf("failed to unmarshal LOOP signal payload: %w", err)
	}

	return &control, nil
}
