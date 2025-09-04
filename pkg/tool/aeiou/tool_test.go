// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Provides unit tests for the new `tool.aeiou.ComposeEnvelope` tool.
// filename: pkg/tool/aeiou/tool_test.go
// nlines: 97
// risk_rating: LOW

package aeiou

import (
	"errors"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
)

func TestEnvelopeToolFunc(t *testing.T) {
	testCases := []struct {
		name                   string
		args                   []any
		expectedStringContains []string
		expectedOrder          []string
		expectErrIs            error
	}{
		{
			name: "Success - Minimal envelope",
			args: []any{`{"sub":"test"}`, `command emit 'ok' endcommand`, "", ""},
			expectedStringContains: []string{
				"<<<NSENV:V3:START>>>",
				"<<<NSENV:V3:USERDATA>>>",
				`{"sub":"test"}`,
				"<<<NSENV:V3:ACTIONS>>>",
				`command emit 'ok' endcommand`,
				"<<<NSENV:V3:END>>>",
			},
			expectErrIs: nil,
		},
		{
			name: "Success - Full envelope",
			args: []any{`{"sub":"test"}`, "actions", "scratch", "output"},
			expectedStringContains: []string{
				"<<<NSENV:V3:USERDATA>>>",
				"<<<NSENV:V3:SCRATCHPAD>>>",
				"scratch",
				"<<<NSENV:V3:OUTPUT>>>",
				"output",
				"<<<NSENV:V3:ACTIONS>>>",
				"actions",
			},
			expectedOrder: []string{ // Verify canonical order
				"<<<NSENV:V3:USERDATA>>>",
				"<<<NSENV:V3:SCRATCHPAD>>>",
				"<<<NSENV:V3:OUTPUT>>>",
				"<<<NSENV:V3:ACTIONS>>>",
			},
			expectErrIs: nil,
		},
		{
			name: "Success - Empty actions gets default command block",
			args: []any{`{"sub":"test"}`, "", "", ""},
			expectedStringContains: []string{
				"<<<NSENV:V3:ACTIONS>>>",
				"command\nendcommand",
			},
			expectErrIs: nil,
		},
		{
			name:        "Fail - Missing required userdata",
			args:        []any{"", "actions", "", ""},
			expectErrIs: aeiou.ErrSectionMissing,
		},
		{
			name:        "Robustness - Non-string args for required fields",
			args:        []any{123, 456, nil, nil},
			expectErrIs: aeiou.ErrSectionMissing, // Fails because userdata becomes "" after type assertion
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := envelopeToolFunc(nil, tc.args) // Runtime is not used, so nil is fine.

			if tc.expectErrIs != nil {
				if !errors.Is(err, tc.expectErrIs) {
					t.Errorf("Expected error wrapping [%v], but got: %v", tc.expectErrIs, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			resultStr, ok := result.(string)
			if !ok {
				t.Fatalf("Expected result to be a string, but got %T", result)
			}

			for _, substr := range tc.expectedStringContains {
				if !strings.Contains(resultStr, substr) {
					t.Errorf("Expected result to contain %q, but it did not.\nResult:\n%s", substr, resultStr)
				}
			}

			if len(tc.expectedOrder) > 0 {
				lastIndex := -1
				for _, sectionMarker := range tc.expectedOrder {
					currentIndex := strings.Index(resultStr, sectionMarker)
					if currentIndex < lastIndex {
						t.Errorf("Section %q appeared out of order.", sectionMarker)
					}
					lastIndex = currentIndex
				}
			}
		})
	}
}
