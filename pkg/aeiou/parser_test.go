// NeuroScript Version: 0.7.0
// File version: 7
// Purpose: Updates tests for new lint-reporting parser signature.
// filename: aeiou/parser_test.go
// nlines: 184
// risk_rating: MEDIUM

package aeiou

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expectedEnv   *Envelope
		expectedLints []Lint
		expectErrIs   error
	}{
		{
			name: "Minimal valid envelope",
			input: strings.Join([]string{
				Wrap(SectionStart),
				Wrap(SectionUserData),
				`{"subject":"test"}`,
				Wrap(SectionActions),
				`command { emit "ok" }`,
				Wrap(SectionEnd),
			}, "\n"),
			expectedEnv: &Envelope{
				UserData: `{"subject":"test"}`,
				Actions:  `command { emit "ok" }`,
			},
		},
		{
			name: "Full valid envelope with optional sections",
			input: strings.Join([]string{
				Wrap(SectionStart),
				Wrap(SectionUserData),
				"user data content",
				Wrap(SectionScratchpad),
				"scratchpad content",
				Wrap(SectionOutput),
				"output content",
				Wrap(SectionActions),
				"actions content",
				Wrap(SectionEnd),
			}, "\n"),
			expectedEnv: &Envelope{
				UserData:   "user data content",
				Scratchpad: "scratchpad content",
				Output:     "output content",
				Actions:    "actions content",
			},
		},
		{
			name: "Duplicate section is ignored and linted",
			input: strings.Join([]string{
				Wrap(SectionStart),
				Wrap(SectionUserData),
				"first user data",
				Wrap(SectionActions),
				"actions content",
				Wrap(SectionUserData),
				"second user data (should be ignored)",
				Wrap(SectionEnd),
			}, "\n"),
			expectedEnv: &Envelope{
				UserData: "first user data",
				Actions:  "actions content",
			},
			expectedLints: []Lint{
				{Code: LintCodeDuplicateSection, Message: "duplicate section 'USERDATA' ignored (first instance is used)"},
			},
		},
		{
			name: "Missing required USERDATA section",
			input: strings.Join([]string{
				Wrap(SectionStart),
				Wrap(SectionActions),
				"actions content",
				Wrap(SectionEnd),
			}, "\n"),
			expectErrIs: ErrSectionMissing,
		},
		{
			name: "Missing required ACTIONS section",
			input: strings.Join([]string{
				Wrap(SectionStart),
				Wrap(SectionUserData),
				"user data content",
				Wrap(SectionEnd),
			}, "\n"),
			expectErrIs: ErrSectionMissing,
		},
		{
			name: "Sections out of order",
			input: strings.Join([]string{
				Wrap(SectionStart),
				Wrap(SectionActions),
				"actions content",
				Wrap(SectionUserData),
				"user data content",
				Wrap(SectionEnd),
			}, "\n"),
			expectErrIs: ErrSectionOrder,
		},
		{
			name:        "Missing END marker",
			input:       Wrap(SectionStart),
			expectErrIs: ErrMarkerInvalid,
		},
		{
			name:        "Missing START marker",
			input:       Wrap(SectionEnd),
			expectErrIs: ErrMarkerInvalid,
		},
		{
			name: "Section exceeds MaxSectionSize",
			input: strings.Join([]string{
				Wrap(SectionStart),
				Wrap(SectionUserData),
				"small",
				Wrap(SectionActions),
				strings.Repeat("a", MaxSectionSize+1),
				Wrap(SectionEnd),
			}, "\n"),
			expectErrIs: ErrPayloadTooLarge,
		},
		{
			name: "Total envelope exceeds MaxEnvelopeSize",
			input: strings.Join([]string{
				Wrap(SectionStart),
				Wrap(SectionUserData),
				strings.Repeat("a", MaxSectionSize-100),
				Wrap(SectionScratchpad),
				strings.Repeat("b", MaxSectionSize-100),
				Wrap(SectionOutput),
				strings.Repeat("c", 201),
				Wrap(SectionActions),
				"action",
				Wrap(SectionEnd),
			}, "\n"),
			expectErrIs: ErrPayloadTooLarge,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := strings.NewReader(tc.input)
			parsedEnv, lints, err := Parse(r)

			if tc.expectErrIs != nil {
				if !errors.Is(err, tc.expectErrIs) {
					if !errors.Is(err, ErrPayloadTooLarge) {
						t.Fatalf("Parse() expected error target %v, got %v", tc.expectErrIs, err)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse() failed unexpectedly: %v", err)
			}

			if !reflect.DeepEqual(tc.expectedEnv, parsedEnv) {
				t.Errorf("Parse() mismatch:\n- want: %+v\n- got:  %+v", tc.expectedEnv, parsedEnv)
			}
			if !reflect.DeepEqual(tc.expectedLints, lints) {
				t.Errorf("Parse() lints mismatch:\n- want: %+v\n- got:  %+v", tc.expectedLints, lints)
			}

			// Round Trip Test
			if tc.expectedEnv != nil {
				composedString, err := parsedEnv.Compose()
				if err != nil {
					t.Fatalf("Compose() failed unexpectedly: %v", err)
				}

				reParsedEnv, _, err := Parse(strings.NewReader(composedString))
				if err != nil {
					t.Fatalf("Failed to re-parse composed payload: %v", err)
				}

				if !reflect.DeepEqual(parsedEnv, reParsedEnv) {
					t.Errorf("Round trip mismatch:\n- original: %+v\n- re-parsed: %+v", parsedEnv, reParsedEnv)
				}
			}
		})
	}
}
