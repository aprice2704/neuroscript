// filename: pkg/neurodata/blocks/blocks_simple_test.go
// pkg/neurodata/blocks/blocks_simple_test.go
package blocks

import (
	"os"
	"path/filepath"	// Keep reflect for the remaining test
	"testing"
)

// Assume fixtureDir is defined in blocks_helpers.go
// Assume helper functions minInt, compareBlockSlices are defined in blocks_helpers.go

func TestExtractAllAndMetadataSimple(t *testing.T) {
	simpleFixturePath := filepath.Join(fixtureDir, "simple_blocks.md")
	simpleContentBytes, errSimple := os.ReadFile(simpleFixturePath)
	if errSimple != nil {
		cwd, _ := os.Getwd()
		t.Fatalf("Failed to read simple fixture file %s (CWD: %s): %v", simpleFixturePath, cwd, errSimple)
	}
	simpleContent := string(simpleContentBytes)

	// --- Expected Results based on simple_blocks.md structure ---
	wantBlocks := []FencedBlock{
		{	// Block 0: neuroscript
			LanguageID:	"neuroscript",
			RawContent:	"DEFINE PROCEDURE Simple()\n  EMIT \"Hello from simple NS\"\nEND",
			StartLine:	7, EndLine: 11,	// Lines based on fixture
			Metadata:	map[string]string{"id": "simple-ns-block", "version": "1.0"},
		},
		{	// Block 1: python
			LanguageID:	"python",
			RawContent:	"print(\"Hello from simple Python\")",
			StartLine:	17, EndLine: 19,	// Lines based on fixture
			Metadata:	map[string]string{"id": "simple-py-block", "version": "0.1"},
		},
		{	// Block 2: empty text
			LanguageID:	"text",
			RawContent:	"",
			StartLine:	22, EndLine: 23,	// Lines based on fixture
			Metadata:	map[string]string{"id": "simple-empty-block"},
		},
		{	// Block 3: text with comments
			LanguageID:	"text",
			RawContent:	"# This is a comment inside.\n-- So is this.\n\n# Even with a blank line.",
			StartLine:	26, EndLine: 31,	// Lines based on fixture
			Metadata:	map[string]string{"id": "simple-comment-block"},
		},
	}
	// --- END Expected Results ---

	t.Run("Simple Fixture Extraction", func(t *testing.T) {
		t.Logf("Running ExtractAll on %s...", simpleFixturePath)
		// --- USE os.Stderr for logger to see debug output ---
		// To disable debug logging for this test, comment out the next line
		// and uncomment the line after it.
		// testLogger := log.New(io.Discard, "[TEST-SIMPLE] ", 0)
		// --- END Logger Change ---
		gotBlocks, err := ExtractAll(simpleContent, logger)
		if err != nil {
			// Only fail if there's an unexpected error (e.g., scanner error)
			t.Fatalf("ExtractAll failed unexpectedly for simple fixture: %v\nGot Blocks: %#v", err, gotBlocks)
		}
		// Use helper to compare blocks
		compareBlockSlices(t, gotBlocks, wantBlocks, simpleFixturePath)
		if t.Failed() {
			t.Logf("Block extraction/comparison failed (see details above).")
		} else {
			t.Logf("ExtractAll successful and block comparison passed for simple fixture.")
		}
	})	// End t.Run Simple Fixture

	// --- REMOVED t.Run("MetadataSyntaxInsideBlock", ...) ---

}	// End TestExtractAllAndMetadataSimple