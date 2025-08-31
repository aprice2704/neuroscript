// NeuroScript Version: 0.3.0
// File version: 9
// Purpose: Adds a final test case to ensure the parser handles out-of-order sections gracefully.
// filename: neuroscript/pkg/aeiou/envelope_test.go
// nlines: 224
// risk_rating: LOW

package aeiou

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestParseAndComposeV2(t *testing.T) {
	startMarker, _ := Wrap(SectionStart, nil)
	endMarker, _ := Wrap(SectionEnd, nil)
	actionsMarker, _ := Wrap(SectionActions, nil)
	orchMarker, _ := Wrap(SectionOrchestration, nil)
	headerJson := `{"proto":"NSENVELOPE","v":2,"caps":["test"]}`
	headerMarker, _ := Wrap(SectionHeader, json.RawMessage(headerJson))
	loopContinueJson := `{"control":"continue","notes":"plan ready"}`
	loopContinueMarker, _ := Wrap(SectionLoop, json.RawMessage(loopContinueJson))

	testCases := []struct {
		name             string
		inputPayload     string
		expectParseErrIs error
		expectedEnv      *Envelope
	}{
		{
			name: "Full valid V2 envelope",
			inputPayload: strings.Join([]string{
				startMarker,
				headerMarker,
				orchMarker,
				"This is the prompt.",
				actionsMarker,
				"command { emit \"action\" }",
				endMarker,
			}, "\n"),
			expectedEnv: &Envelope{
				Header: &Header{
					Proto:   "NSENVELOPE",
					Version: 2,
					Caps:    []string{"test"},
				},
				Orchestration: "This is the prompt.",
				Actions:       "command { emit \"action\" }",
			},
		},
		{
			name: "V2 with loop control json",
			inputPayload: strings.Join([]string{
				startMarker,
				actionsMarker,
				"command { emit \"" + loopContinueMarker + "\" }",
				endMarker,
			}, "\n"),
			expectedEnv: &Envelope{
				Actions: "command { emit \"" + loopContinueMarker + "\" }",
			},
		},
		{
			name: "Handles out-of-order sections",
			inputPayload: strings.Join([]string{
				startMarker,
				actionsMarker, // Actions first
				"action content",
				orchMarker, // Then orchestration
				"prompt content",
				endMarker,
			}, "\n"),
			expectedEnv: &Envelope{
				Actions:       "action content",
				Orchestration: "prompt content",
			},
		},
		{
			name:             "V2 missing start marker",
			inputPayload:     "some content",
			expectParseErrIs: ErrEnvelopeNoStart,
		},
		{
			name:             "V2 missing end marker",
			inputPayload:     startMarker,
			expectParseErrIs: ErrEnvelopeNoEnd,
		},
		{
			name: "V2 duplicate section",
			inputPayload: strings.Join([]string{
				startMarker,
				actionsMarker,
				"one",
				actionsMarker,
				"two",
				endMarker,
			}, "\n"),
			expectParseErrIs: ErrDuplicateSection,
		},
		// --- Fuzzy & Edge Case Tests ---
		{
			name: "Fuzzy: Empty envelope",
			inputPayload: strings.Join([]string{
				startMarker,
				endMarker,
			}, "\n"),
			expectedEnv: &Envelope{}, // Should parse to an empty struct
		},
		{
			name: "Fuzzy: Envelope with empty sections",
			inputPayload: strings.Join([]string{
				startMarker,
				actionsMarker,
				orchMarker,
				endMarker,
			}, "\n"),
			expectedEnv: &Envelope{},
		},
		{
			name: "Fuzzy: Malformed JSON in header",
			inputPayload: strings.Join([]string{
				startMarker,
				`<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:HEADER:{"v":2,>>>`, // Invalid JSON
				endMarker,
			}, "\n"),
			expectParseErrIs: ErrInvalidJSONHeader,
		},
		{
			name:         "Fuzzy: Weird whitespace",
			inputPayload: "  \n\n" + startMarker + "\n\n" + actionsMarker + "\n\n   some action\n" + endMarker + "\n\n",
			expectedEnv: &Envelope{
				Actions: "some action",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsedEnv, err := RobustParse(tc.inputPayload)

			if tc.expectParseErrIs != nil {
				if !errors.Is(err, tc.expectParseErrIs) {
					t.Fatalf("RobustParse() expected error target %v, got %v", tc.expectParseErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("RobustParse() failed unexpectedly: %v", err)
			}

			if tc.expectedEnv.Header == nil {
				parsedEnv.Header = nil
			}

			if !reflect.DeepEqual(tc.expectedEnv, parsedEnv) {
				t.Errorf("Parse() mismatch:\n- want: %+v\n- got:  %+v", tc.expectedEnv, parsedEnv)
			}

			// Test Compose and Round Trip if no error was expected
			composedPayload, err := parsedEnv.Compose()
			if err != nil {
				t.Fatalf("Compose() failed unexpectedly: %v", err)
			}

			reParsedEnv, err := RobustParse(composedPayload)
			if err != nil {
				t.Fatalf("Failed to re-parse composed payload: %v\nPayload:\n%s", err, composedPayload)
			}

			if tc.expectedEnv.Header == nil {
				reParsedEnv.Header = nil
			} else {
				reParsedEnv.Header.Checksum = ""
			}

			if !reflect.DeepEqual(tc.expectedEnv, reParsedEnv) {
				t.Errorf("Round trip mismatch:\n- want: %+v\n- got:  %+v", tc.expectedEnv, reParsedEnv)
			}
		})
	}
}
