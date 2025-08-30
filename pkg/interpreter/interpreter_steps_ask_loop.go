// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Refactored: Contains loop control logic and constants for the 'ask' statement.
// filename: pkg/interpreter/interpreter_steps_ask_loop.go
// nlines: 48
// risk_rating: LOW

package interpreter

import (
	"strings"
)

const (
	loopControlDone      = `"control":"done"`
	loopControlAbort     = `"control":"abort"`
	loopControlPrefix    = `<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:LOOP:`
	defaultMaxTurns      = 1
	maxTurnsCap          = 10 // A hard safety cap
	bootstrapCapsuleName = "capsule/bootstrap/1"
)

type loopControlState int

const (
	stateContinue loopControlState = iota
	stateDone
	stateAbort
)

func parseLoopControl(signal string) loopControlState {
	if signal == "" {
		return stateDone // Default to done if no signal is found
	}
	if strings.Contains(signal, loopControlAbort) {
		return stateAbort
	}
	if strings.Contains(signal, loopControlDone) {
		return stateDone
	}
	// If it's a loop signal and not abort or done, it must be continue.
	return stateContinue
}

func extractEmits(actionEmits []string) (string, string) {
	var resultLines []string
	var loopControlSignal string
	for _, line := range actionEmits {
		if strings.HasPrefix(line, loopControlPrefix) {
			loopControlSignal = line
		} else {
			resultLines = append(resultLines, line)
		}
	}
	return strings.Join(resultLines, "\n"), loopControlSignal
}
