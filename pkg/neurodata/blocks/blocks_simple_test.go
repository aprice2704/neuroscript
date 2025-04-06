// pkg/neurodata/blocks2/blocks_simple_test.go
package blocks

import (
	"fmt" // Import fmt for error messages
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Define the path to the test fixtures directory relative to the test file
const fixtureDir = "test_fixtures"

// TestExtractAllAndMetadataSimple tests ExtractAll and LookForMetadata using simple_blocks.md.
func TestExtractAllAndMetadataSimple(t *testing.T) {
	// Construct the path to the fixture file
	simpleFixturePath := filepath.Join(fixtureDir, "simple_blocks.md")

	// Read the content of the fixture file
	simpleContentBytes, errSimple := os.ReadFile(simpleFixturePath)
	if errSimple != nil {
		// Get current working directory for context in error message
		cwd, _ := os.Getwd()
		t.Fatalf("Failed to read simple fixture file %s (CWD: %s): %v", simpleFixturePath, cwd, errSimple)
	}
	simpleContent := string(simpleContentBytes)

	// --- Expected Results for ExtractAll ---
	wantBlocks := []FencedBlock{
		{
			LanguageID: "neuroscript",
			RawContent: "# id: simple-ns-block\n# version: 1.0\nDEFINE PROCEDURE Simple()\n  EMIT \"Hello from simple NS\"\nEND",
			StartLine:  5,
			EndLine:    11,
		},
		{
			LanguageID: "python",
			RawContent: "# id: simple-py-block\n# version: 0.1\nprint(\"Hello from simple Python\")",
			StartLine:  15,
			EndLine:    19,
		},
		{
			LanguageID: "text",
			RawContent: "# id: simple-empty-block",
			StartLine:  21,
			EndLine:    23,
		},
		{
			LanguageID: "text",
			RawContent: "# id: simple-comment-block\n# This is a comment inside.\n-- So is this.\n\n# Even with a blank line.",
			StartLine:  25,
			EndLine:    31,
		},
	}

	// --- Expected Results for LookForMetadata (per block index) ---
	wantMetadata := []map[string]string{
		{"id": "simple-ns-block", "version": "1.0"}, // Block 0
		{"id": "simple-py-block", "version": "0.1"}, // Block 1
		{"id": "simple-empty-block"},                // Block 2
		{"id": "simple-comment-block"},              // Block 3
	}

	// --- Run ExtractAll ---
	t.Logf("Running ExtractAll on %s...", simpleFixturePath)
	gotBlocks, err := ExtractAll(simpleContent)

	// 1. Check ExtractAll errors
	if err != nil {
		t.Fatalf("ExtractAll failed unexpectedly for simple fixture: %v", err)
	}

	// 2. Compare ExtractAll results
	if !reflect.DeepEqual(gotBlocks, wantBlocks) {
		t.Errorf("Mismatch in extracted blocks from %s", simpleFixturePath)
		t.Logf("Got %d blocks, Want %d blocks", len(gotBlocks), len(wantBlocks))
		// Print detailed comparison (same as before)
		maxLen := len(gotBlocks)
		if len(wantBlocks) > maxLen {
			maxLen = len(wantBlocks)
		}
		for i := 0; i < maxLen; i++ {
			var gotBlock FencedBlock
			if i < len(gotBlocks) {
				gotBlock = gotBlocks[i]
			}
			var wantBlock FencedBlock
			if i < len(wantBlocks) {
				wantBlock = wantBlocks[i]
			}
			if !reflect.DeepEqual(gotBlock, wantBlock) {
				t.Errorf("--- ExtractAll Block %d Mismatch ---", i)
				t.Errorf("  Got : LangID=%q, Start=%d, End=%d, Content=\n---\n%q\n---", gotBlock.LanguageID, gotBlock.StartLine, gotBlock.EndLine, gotBlock.RawContent)
				t.Errorf("  Want: LangID=%q, Start=%d, End=%d, Content=\n---\n%q\n---", wantBlock.LanguageID, wantBlock.StartLine, wantBlock.EndLine, wantBlock.RawContent)
			}
		}
		// Fail fast if block extraction is wrong, as metadata tests depend on it
		t.Fatalf("Block extraction failed, cannot proceed to metadata tests.")
	} else {
		t.Logf("ExtractAll successful for simple fixture.")
	}

	// --- Run LookForMetadata on each extracted block ---
	t.Logf("Running LookForMetadata on extracted blocks...")
	if len(gotBlocks) != len(wantMetadata) {
		t.Fatalf("Test setup error: Number of extracted blocks (%d) does not match number of expected metadata maps (%d)", len(gotBlocks), len(wantMetadata))
	}

	for i, block := range gotBlocks {
		t.Run(fmt.Sprintf("Metadata_Block_%d", i), func(t *testing.T) {
			expectedMeta := wantMetadata[i]
			gotMeta, metaErr := LookForMetadata(block.RawContent)

			// Check for unexpected errors from LookForMetadata
			if metaErr != nil {
				t.Errorf("LookForMetadata returned unexpected error for block %d: %v", i, metaErr)
				return // Skip comparison if error occurred
			}

			// Compare the metadata map
			if !reflect.DeepEqual(gotMeta, expectedMeta) {
				t.Errorf("Metadata mismatch for block %d (LangID: %q, StartLine: %d)", i, block.LanguageID, block.StartLine)
				t.Errorf("  Got : %#v", gotMeta)
				t.Errorf("  Want: %#v", expectedMeta)
				t.Errorf("  Raw Content Searched:\n---\n%s\n---", block.RawContent)
			}
		})
	}
}
