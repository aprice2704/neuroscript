// pkg/core/extract_blocks_test.go
package blocks

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	// Imports needed by ExtractAllFencedBlocks if it were here
	// "bufio"
	// "fmt"
	// "regexp"
)

// --- Test ExtractAllFencedBlocks with simple_blocks.md ---
func TestExtractAllFencedBlocksSimple(t *testing.T) {
	fixtureDir := "test_fixtures"
	simpleFixturePath := filepath.Join(fixtureDir, "simple_blocks.md")
	simpleContentBytes, errSimple := os.ReadFile(simpleFixturePath)
	if errSimple != nil {
		cwd, _ := os.Getwd()
		t.Fatalf("Failed to read simple fixture file %s (CWD: %s): %v", simpleFixturePath, cwd, errSimple)
	}
	simpleContent := string(simpleContentBytes)

	wantBlocks := []string{
		"# id: simple-ns-block\n# version: 1.0\nDEFINE PROCEDURE Simple()\n  EMIT \"Hello from simple NS\"\nEND",
		"# id: simple-py-block\n# version: 0.1\nprint(\"Hello from simple Python\")",
		"# id: simple-empty-block",
		"# id: simple-comment-block\n# This is a comment inside.\n-- So is this.\n\n# Even with a blank line.",
	}

	// Assuming ExtractAllFencedBlocks is the clean version without debug prints now
	gotBlocks, err := ExtractAllFencedBlocks(simpleContent)
	if err != nil {
		t.Fatalf("ExtractAllFencedBlocks failed unexpectedly for simple fixture: %v", err)
	}
	if !reflect.DeepEqual(gotBlocks, wantBlocks) {
		t.Errorf("Mismatch in extracted blocks from simple_blocks.md")
		// Provide detailed diff for easier debugging
		t.Logf("Got %d blocks, Want %d blocks", len(gotBlocks), len(wantBlocks))
		maxLen := len(gotBlocks)
		if len(wantBlocks) > maxLen {
			maxLen = len(wantBlocks)
		}
		for i := 0; i < maxLen; i++ {
			got := ""
			if i < len(gotBlocks) {
				got = gotBlocks[i]
			}
			want := ""
			if i < len(wantBlocks) {
				want = wantBlocks[i]
			}
			if got != want {
				t.Errorf("Block %d mismatch:\nGot:\n---\n%s\n---\nWant:\n---\n%s\n---", i, got, want)
			}
		}
	}
}

// --- Test ExtractAllFencedBlocks with complex_blocks.md ---
func TestExtractAllFencedBlocksComplex(t *testing.T) {
	fixtureDir := "test_fixtures"
	complexFixturePath := filepath.Join(fixtureDir, "complex_blocks.md")
	complexContentBytes, errComplex := os.ReadFile(complexFixturePath)
	if errComplex != nil {
		cwd, _ := os.Getwd()
		t.Fatalf("Failed to read complex fixture file %s (CWD: %s): %v", complexFixturePath, cwd, errComplex)
	}
	complexContent := string(complexContentBytes)

	// ** CORRECTED: Expect only blocks before the *actual* error (ambiguity at line 8) **
	wantBlocksBeforeError := []string{
		"# id: complex-ns-1\nCALL TOOL.DoSomething()", // Only the first block is parsed before ambiguity
	}
	// ** CORRECTED: Expect ambiguity error at line 8, not EOF error **
	wantErrMsgContains := "ambiguous fence pattern: line 8 starts with '```' immediately after a previous block closed at line 7"

	// Assuming ExtractAllFencedBlocks is the clean version without debug prints now
	gotBlocks, err := ExtractAllFencedBlocks(complexContent)

	// Check for the expected error
	if err == nil {
		// ** CORRECTED: Updated failure message **
		t.Fatalf("TestExtractAllFencedBlocksComplex: Expected an error containing %q, but got nil", wantErrMsgContains)
	}
	if !strings.Contains(err.Error(), wantErrMsgContains) {
		// ** CORRECTED: Updated failure message **
		t.Errorf("TestExtractAllFencedBlocksComplex: Expected error message containing %q, but got: %v", wantErrMsgContains, err)
	}

	// Check the blocks extracted *before* the error occurred
	if !reflect.DeepEqual(gotBlocks, wantBlocksBeforeError) {
		// ** CORRECTED: Updated failure message **
		t.Errorf("TestExtractAllFencedBlocksComplex: Mismatch in blocks extracted before the expected error.")
		// Provide detailed diff for easier debugging
		t.Logf("Got %d blocks before error, Want %d blocks", len(gotBlocks), len(wantBlocksBeforeError))
		maxLen := len(gotBlocks)
		if len(wantBlocksBeforeError) > maxLen {
			maxLen = len(wantBlocksBeforeError)
		}
		for i := 0; i < maxLen; i++ {
			got := ""
			if i < len(gotBlocks) {
				got = gotBlocks[i]
			}
			want := ""
			if i < len(wantBlocksBeforeError) {
				want = wantBlocksBeforeError[i]
			}
			if got != want {
				t.Errorf("Block %d mismatch:\nGot:\n---\n%s\n---\nWant:\n---\n%s\n---", i, got, want)
			}
		}
	}
}

// --- Test ExtractAllFencedBlocks with evil_blocks.md ---
// This test was modified previously to ignore block content and only check the error.
// Keeping it that way as requested.
func TestExtractAllFencedBlocksEvil(t *testing.T) {
	fixtureDir := "test_fixtures"
	evilFixturePath := filepath.Join(fixtureDir, "evil_blocks.md") // Ensure this file has a newline after the final ```
	evilContentBytes, errEvil := os.ReadFile(evilFixturePath)
	if errEvil != nil {
		cwd, _ := os.Getwd()
		t.Fatalf("Failed to read evil fixture file %s (CWD: %s): %v", evilFixturePath, cwd, errEvil)
	}
	evilContent := string(evilContentBytes)

	wantErr := true // Expect an error due to ambiguity at line 44
	// Expected error message and line number
	wantErrMsgContains := "ambiguous fence pattern: line 44 starts with '```' immediately after a previous block closed at line 43"

	// Assuming ExtractAllFencedBlocks is the clean version without debug prints now
	gotBlocks, err := ExtractAllFencedBlocks(evilContent)

	// --- Assertions for Error Case (Only checking error, not blocks) ---
	if !wantErr {
		// This branch is unlikely now, but kept for structure
		if err != nil {
			t.Fatalf("TestExtractAllFencedBlocksEvil: Expected no error, but got: %v", err)
		}
	} else { // This branch executes because wantErr is true
		if err == nil {
			// Use %+v to potentially get more detail if gotBlocks is complex
			t.Fatalf("TestExtractAllFencedBlocksEvil: Expected an error containing %q, but got nil. Got blocks: %+v", wantErrMsgContains, gotBlocks)
		}
		if wantErrMsgContains != "" && !strings.Contains(err.Error(), wantErrMsgContains) {
			t.Errorf("TestExtractAllFencedBlocksEvil: Expected error message containing %q, but got: %v", wantErrMsgContains, err)
		}
		// Block content check is commented out as requested previously
	}
}

// --- Test ExtractAllFencedBlocksWithBlankLines ---
func TestExtractAllFencedBlocksWithBlankLines(t *testing.T) {
	content := "```\n\nActual content line 1.\nActual content line 2.\n\n```"
	// Capture includes the leading/trailing blank lines *within* the fences
	wantBlocks := []string{"\nActual content line 1.\nActual content line 2.\n"}

	// Assuming ExtractAllFencedBlocks is the clean version without debug prints now
	gotBlocks, err := ExtractAllFencedBlocks(content)
	if err != nil {
		t.Fatalf("ExtractAllFencedBlocks failed unexpectedly: %v", err)
	}
	if len(gotBlocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(gotBlocks))
	}
	if !reflect.DeepEqual(gotBlocks, wantBlocks) {
		t.Errorf("Mismatch (Test 1):\n Got:\n%q\n Want:\n%q", gotBlocks[0], wantBlocks[0])
	}

	contentSingleBlank := "```\n\n```"
	// ** CORRECTED: strings.Join of a single empty string is "", not "\n" **
	wantBlocksSingleBlank := []string{""}
	gotBlocksSingleBlank, errSingleBlank := ExtractAllFencedBlocks(contentSingleBlank)
	if errSingleBlank != nil {
		t.Fatalf("ExtractAllFencedBlocks failed for single blank line: %v", errSingleBlank)
	}
	if len(gotBlocksSingleBlank) != 1 {
		t.Fatalf("Expected 1 block (single blank), got %d", len(gotBlocksSingleBlank))
	}
	if !reflect.DeepEqual(gotBlocksSingleBlank, wantBlocksSingleBlank) {
		t.Errorf("Mismatch (Test 2):\n Got:\n%q\n Want:\n%q", gotBlocksSingleBlank[0], wantBlocksSingleBlank[0])
	}

	contentEmpty := "```\n```"
	wantBlocksEmpty := []string{""} // Expect empty content.
	gotBlocksEmpty, errEmpty := ExtractAllFencedBlocks(contentEmpty)
	if errEmpty != nil {
		t.Fatalf("ExtractAllFencedBlocks failed for empty block: %v", errEmpty)
	}
	if len(gotBlocksEmpty) != 1 {
		t.Fatalf("Expected 1 block (empty), got %d", len(gotBlocksEmpty))
	}
	if !reflect.DeepEqual(gotBlocksEmpty, wantBlocksEmpty) {
		t.Errorf("Mismatch (Test 3):\n Got:\n%q\n Want:\n%q", gotBlocksEmpty[0], wantBlocksEmpty[0])
	}

	contentNoLeadingBlank := "```\nActual content line 1.\nActual content line 2.\n```"
	wantBlocksNoLeadingBlank := []string{"Actual content line 1.\nActual content line 2."}
	gotBlocksNoLeadingBlank, errNoLeadingBlank := ExtractAllFencedBlocks(contentNoLeadingBlank)
	if errNoLeadingBlank != nil {
		t.Fatalf("ExtractAllFencedBlocks failed for no leading blank: %v", errNoLeadingBlank)
	}
	if len(gotBlocksNoLeadingBlank) != 1 {
		t.Fatalf("Expected 1 block (no leading blank), got %d", len(gotBlocksNoLeadingBlank))
	}
	if !reflect.DeepEqual(gotBlocksNoLeadingBlank, wantBlocksNoLeadingBlank) {
		t.Errorf("Mismatch (Test 4):\n Got:\n%q\n Want:\n%q", gotBlocksNoLeadingBlank[0], wantBlocksNoLeadingBlank[0])
	}
}
