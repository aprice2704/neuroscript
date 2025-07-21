// NeuroScript Version: 0.6.0
// File version: 8
// Purpose: Corrected the 'evil' test case to be syntactically valid by adding a statement to an empty function block.
// filename: pkg/nsfmt/reconstructor_test.go
// nlines: 160
// risk_rating: MEDIUM

package nsfmt

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// assertRoundTrip is a helper that parses, reconstructs, and compares.
func assertRoundTrip(t *testing.T, source string) {
	t.Helper()

	// 1. Normalize the source string to have consistent newlines and trim leading/trailing whitespace.
	normalizedSource := strings.TrimSpace(strings.ReplaceAll(source, "\r\n", "\n"))

	// 2. Parse the source into an AST. This is now a two-step process.
	logger := logging.NewTestLogger(t)
	parserAPI := parser.NewParserAPI(logger)

	// 2a. Get the raw ANTLR tree and token stream from the parser.
	antlrTree, tokenStream, pErr := parserAPI.ParseAndGetStream("roundtrip_test.ns", normalizedSource)
	if pErr != nil {
		t.Fatalf("ParseAndGetStream failed unexpectedly: %v", pErr)
	}

	// 2b. Use the AST Builder to walk the ANTLR tree and create our application-specific AST.
	builder := parser.NewASTBuilder(logger)
	program, _, bErr := builder.BuildFromParseResult(antlrTree, tokenStream)
	if bErr != nil {
		t.Fatalf("AST Builder failed unexpectedly: %v", bErr)
	}
	tree := &ast.Tree{Root: program}

	// 3. Reconstruct the source from the AST.
	reconstructed, err := Reconstruct(tree)
	if err != nil {
		t.Fatalf("Reconstruct failed: %v", err)
	}
	normalizedReconstructed := strings.TrimSpace(strings.ReplaceAll(reconstructed, "\r\n", "\n"))

	// 4. Compare and provide a detailed diff on failure.
	if normalizedSource != normalizedReconstructed {
		t.Errorf("Round trip failed for test case '%s'. Source and reconstructed text do not match.", t.Name())
		originalLines := strings.Split(normalizedSource, "\n")
		reconstructedLines := strings.Split(normalizedReconstructed, "\n")
		maxLines := len(originalLines)
		if len(reconstructedLines) > maxLines {
			maxLines = len(reconstructedLines)
		}

		t.Log("--- DIFF ---")
		t.Logf("%-5s %-40s | %-40s", "LINE", "WANT", "GOT")
		t.Log(strings.Repeat("-", 88))
		for i := 0; i < maxLines; i++ {
			wantLine := ""
			if i < len(originalLines) {
				wantLine = originalLines[i]
			}
			gotLine := ""
			if i < len(reconstructedLines) {
				gotLine = reconstructedLines[i]
			}
			marker := " "
			if wantLine != gotLine {
				marker = "!"
			}
			t.Logf("%s %-4d %-40s | %-40s", marker, i+1, wantLine, gotLine)
		}
		t.FailNow()
	}
}

func TestSimpleRoundTrip(t *testing.T) {
	source := `
# This is a file-level comment.

func main() means
	# This is a leading comment for the set statement.
	set x = 1 # This is a trailing comment.
endfunc
`
	assertRoundTrip(t, source)
}

func TestRoundTripMultipleStatements(t *testing.T) {
	source := `
func main() means
	set x = 1
	set name = "world" # A trailing comment
	# A leading comment
	set val = nil
endfunc
`
	assertRoundTrip(t, source)
}

func TestRoundTripMultipleProcedures(t *testing.T) {
	source := `
# File header comment.

func first() means
	set a = 1
endfunc


# Comment between functions.

func second() means
	set b = 2
endfunc
`
	assertRoundTrip(t, source)
}

func TestRoundTripLayoutAndComments(t *testing.T) {
	source := `
func complex() means
	# Block comment
	# for the first statement.
	set x = 1


	set y = 2 # Trailing comment for y.


	# Leading comment for z.
	set z = 3
endfunc
`
	assertRoundTrip(t, source)
}

func TestRoundTripEvilLayout(t *testing.T) {
	source := `
# Top comment line 1
# Top comment line 2

func main() means
	# Leading comment for 'a'
	set a = 1 # Trailing for 'a'

	set b = 2

	# Multi-line
	# leading comment
	# for 'c'
	set c = 3 # Trailing for 'c'
endfunc


# Comment between funcs
func empty() means
	# Just a comment inside.
	# And another.
	set _ = nil
endfunc
`
	assertRoundTrip(t, source)
}
