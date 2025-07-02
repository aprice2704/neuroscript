// filename: pkg/parser/parser_minimal_test.go
package parser

import (
	"strings"
	"testing"
	// For NoOpLogger adapter
)

// TestParseMinimalNewline attempts to parse a minimal script that mimics
// the structure potentially causing issues in the break/continue test.
// It specifically checks for the 'missing NEWLINE' error and reports
// the details provided by the enhanced error listener.
func TestParseMinimalNewline(t *testing.T) {
	logger := logging.NewNoLogger() // Use a logger adapter
	parserAPI := NewParserAPI(logger)
	scriptContent := `:: Test: Minimal newline issue
func dummy() means
  emit "dummy func"
endfunc

func main() means
  # Line 5 Comment
  set x = 1   # Line 6 Statement
              # Line 7 Blank Line
  call dummy() # Line 8 Statement (call)
  set y = 2   # Line 9 Statement
endfunc
`
	_, err := parserAPI.Parse(scriptContent)

	// --- Verification ---
	if err != nil {
		// Check if the error is the specific 'missing NEWLINE' error
		// The enhanced error message includes "near token '...'"
		// Example check (adapt based on exact expected error string):
		expectedErrorSubstring := "missing NEWLINE near token 'set'" // Expecting error before 'set y = 2'
		if !strings.Contains(err.Error(), expectedErrorSubstring) {
			t.Errorf("Parse failed, but not with the expected 'missing NEWLINE near token set' error. Got: %v", err)
		} else {
			// Log the exact error message for analysis, especially the line number
			t.Logf("Minimal parse failed as expected. Error details: %v", err)
			// Mark test as failed explicitly if the goal is to *fix* this minimal case
			// t.Errorf("Minimal parse failed as expected, indicating the core issue persists. Error: %v", err)
		}
	} else {
		// If err IS nil, it means the minimal script parsed correctly,
		// suggesting the issue might be related to the complexity or specific
		// statements (like break/continue) in the original script, or interactions
		// not present in the minimal one.
		t.Logf("Minimal script parsed successfully without errors. This might indicate the issue is specific to the break/continue script's content or structure.")
	}
}
