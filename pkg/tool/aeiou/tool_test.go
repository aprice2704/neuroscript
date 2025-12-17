// :: product: FDM/NS
// :: majorVersion: 0
// :: fileVersion: 2
// :: description: Updated expected strings to AEIOU V4 markers.
// :: latestChange: Changed all NSENV:V3 expectations to NSENV:V4.
// :: filename: pkg/tool/aeiou/tool_test.go
// :: serialization: go
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
				"<<<NSENV:V4:START>>>",
				"<<<NSENV:V4:USERDATA>>>",
				`{"sub":"test"}`,
				"<<<NSENV:V4:ACTIONS>>>",
				`command emit 'ok' endcommand`,
				"<<<NSENV:V4:END>>>",
			},
			expectErrIs: nil,
		},
		{
			name: "Success - Full envelope",
			args: []any{`{"sub":"test"}`, "actions", "scratch", "output"},
			expectedStringContains: []string{
				"<<<NSENV:V4:USERDATA>>>",
				"<<<NSENV:V4:SCRATCHPAD>>>",
				"scratch",
				"<<<NSENV:V4:OUTPUT>>>",
				"output",
				"<<<NSENV:V4:ACTIONS>>>",
				"actions",
			},
			expectedOrder: []string{ // Verify canonical order
				"<<<NSENV:V4:USERDATA>>>",
				"<<<NSENV:V4:SCRATCHPAD>>>",
				"<<<NSENV:V4:OUTPUT>>>",
				"<<<NSENV:V4:ACTIONS>>>",
			},
			expectErrIs: nil,
		},
		{
			name: "Success - Empty actions gets default command block",
			args: []any{`{"sub":"test"}`, "", "", ""},
			expectedStringContains: []string{
				"<<<NSENV:V4:ACTIONS>>>",
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
