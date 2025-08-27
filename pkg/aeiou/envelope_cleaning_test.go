// NeuroScript Version: 0.3.0
// File version: 7
// Purpose: Updates cleaning tests to ensure they work with the V2 block-based parsing protocol.
// filename: neuroscript/pkg/aeiou/envelope_cleaning_test.go
// nlines: 70
// risk_rating: LOW

package aeiou

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestParseWithCleaningV2(t *testing.T) {
	startMarker, _ := Wrap(SectionStart, nil)
	endMarker, _ := Wrap(SectionEnd, nil)
	actionsMarker, _ := Wrap(SectionActions, nil)

	testCases := []struct {
		name         string
		nastyPayload string
		expected     *Envelope
		expectErrIs  error
	}{
		{
			name: "Removes null bytes in V2",
			nastyPayload: fmt.Sprintf("%s\n%s\ncommand { \x00emit \"hello\" }\n%s",
				startMarker, actionsMarker, endMarker),
			expected: &Envelope{
				Actions: "command { emit \"hello\" }",
			},
		},
		{
			name: "Fails on invalid UTF-8 in V2",
			nastyPayload: fmt.Sprintf("%s\n%s\nHere is invalid UTF-8: \xff\xfe\xfd\n%s",
				startMarker, actionsMarker, endMarker),
			expectErrIs: lang.ErrInvalidUTF8,
		},
		{
			name: "Error on oversized payload in V2",
			nastyPayload: fmt.Sprintf("%s\n%s\n%s\n%s",
				startMarker, actionsMarker, strings.Repeat("a", 20*1024*1024), endMarker),
			expectErrIs: lang.ErrResourceExhaustion,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use RobustParse because it's the primary entrypoint that handles extraction
			parsed, err := RobustParse(tc.nastyPayload)

			if tc.expectErrIs != nil {
				if !errors.Is(err, tc.expectErrIs) {
					t.Fatalf("Expected error target %v, got %v", tc.expectErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse() failed unexpectedly: %v", err)
			}

			// For these tests, we don't care about the header
			parsed.Header = nil

			if !reflect.DeepEqual(tc.expected, parsed) {
				t.Errorf("Parse() with cleaning mismatch:\n- want: %+v\n- got:  %+v", tc.expected, parsed)
			}
		})
	}
}
