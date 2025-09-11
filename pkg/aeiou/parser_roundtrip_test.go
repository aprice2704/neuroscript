// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: A strict round-trip test designed to fail by catching subtle parsing errors where trailing newlines are improperly handled, leading to data loss on re-composition.
// filename: aeiou/parser_strict_roundtrip_test.go
// nlines: 35
// risk_rating: LOW

package aeiou

import (
	"strings"
	"testing"
)

func TestParseComposeStrictRoundTrip(t *testing.T) {
	// This test is designed to fail if the parser does not perfectly
	// preserve content during a round trip, which is what the failing
	// interpreter test indicates is happening.
	t.Run("Strict round trip after compose must preserve content exactly", func(t *testing.T) {
		// 1. Create a known-good envelope.
		originalEnv := &Envelope{
			UserData:   `{"subject":"round-trip-test"}`,
			Scratchpad: "private notes",
			Output:     "public output",
			Actions:    "command emit 'hello' endcommand",
		}

		// 2. Compose it, which adds newlines between sections.
		composedString, err := originalEnv.Compose()
		if err != nil {
			t.Fatalf("Initial Compose() failed unexpectedly: %v", err)
		}

		// 3. Parse it back. The bug is here: the parser is expected to
		// incorrectly handle the trailing newline on the ACTIONS section.
		parsedEnv, _, err := Parse(strings.NewReader(composedString))
		if err != nil {
			t.Fatalf("Parse() of composed string failed: %v", err)
		}

		// 4. This assertion is expected to fail.
		if *originalEnv != *parsedEnv {
			t.Fatalf("Round-trip failure: parsed envelope does not match original.\n- want: %+v\n- got:  %+v", originalEnv, parsedEnv)
		}
	})
}
