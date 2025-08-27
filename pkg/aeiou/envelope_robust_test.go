// NeuroScript Version: 0.3.0
// File version: 4
// Purpose: Adds a final fuzzy test to ensure broken/incomplete magic markers are ignored during robust parsing.
// filename: neuroscript/pkg/aeiou/envelope_robust_test.go
// nlines: 102
// risk_rating: LOW

package aeiou

import (
	"strings"
	"testing"
)

func TestRobustParse_V2_Extraction(t *testing.T) {
	startMarker, _ := Wrap(SectionStart, nil)
	endMarker, _ := Wrap(SectionEnd, nil)
	actionsMarker, _ := Wrap(SectionActions, nil)
	orchMarker, _ := Wrap(SectionOrchestration, nil)

	// --- Test Case 1: Standard noisy input ---
	t.Run("Extracts first of multiple envelopes", func(t *testing.T) {
		noisyInput := `
Preamble to ignore.
` + startMarker + `
` + orchMarker + `
prompt1
` + actionsMarker + `
action1
` + endMarker + `
Text between envelopes to ignore.
` + startMarker + `
` + orchMarker + `
prompt2
` + endMarker + `
Postamble to ignore.
`
		expectedActions := "action1"
		expectedOrchestration := "prompt1"

		env, err := RobustParse(noisyInput)
		if err != nil {
			t.Fatalf("RobustParse failed unexpectedly: %v", err)
		}
		if strings.TrimSpace(env.Actions) != expectedActions {
			t.Errorf("ACTIONS mismatch. Got %q, want %q", env.Actions, expectedActions)
		}
		if strings.TrimSpace(env.Orchestration) != expectedOrchestration {
			t.Errorf("ORCHESTRATION mismatch. Got %q, want %q", env.Orchestration, expectedOrchestration)
		}
	})

	// --- Test Case 2: Fuzzy - Multiple START markers ---
	t.Run("Fuzzy: Handles multiple START markers", func(t *testing.T) {
		fuzzyInput := `
` + startMarker + `
` + startMarker + `
` + actionsMarker + `
action from first valid envelope
` + endMarker + `
` + endMarker + `
`
		expectedActions := "action from first valid envelope"

		env, err := RobustParse(fuzzyInput)
		if err != nil {
			t.Fatalf("RobustParse with multiple STARTs failed: %v", err)
		}
		if strings.TrimSpace(env.Actions) != expectedActions {
			t.Errorf("ACTIONS mismatch. Got %q, want %q", env.Actions, expectedActions)
		}
	})

	// --- Test Case 3: Fuzzy - Broken markers ---
	t.Run("Fuzzy: Ignores broken markers", func(t *testing.T) {
		fuzzyInput := `
` + startMarker + `
<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:ACTIONS
` + actionsMarker + `
actual action
` + endMarker + `
`
		expectedActions := "actual action"
		env, err := RobustParse(fuzzyInput)
		if err != nil {
			t.Fatalf("RobustParse with broken markers failed: %v", err)
		}
		if strings.TrimSpace(env.Actions) != expectedActions {
			t.Errorf("ACTIONS mismatch. Got %q, want %q", env.Actions, expectedActions)
		}
	})
}
