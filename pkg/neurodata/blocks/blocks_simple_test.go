// pkg/neurodata/blocks/blocks_simple_test.go
package blocks

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const fixtureDir = "test_fixtures"

func TestExtractAllAndMetadataSimple(t *testing.T) {
	simpleFixturePath := filepath.Join(fixtureDir, "simple_blocks.md")
	simpleContentBytes, errSimple := os.ReadFile(simpleFixturePath)
	if errSimple != nil {
		cwd, _ := os.Getwd()
		t.Fatalf("Failed to read simple fixture file %s (CWD: %s): %v", simpleFixturePath, cwd, errSimple)
	}
	simpleContent := string(simpleContentBytes)

	// --- Expected Results (RawContent includes :: and comment lines, trimmed) ---
	wantBlocks := []FencedBlock{
		{
			LanguageID: "neuroscript",
			RawContent: ":: id: simple-ns-block\n:: version: 1.0\nDEFINE PROCEDURE Simple()\n  EMIT \"Hello from simple NS\"\nEND",
			StartLine:  5, EndLine: 11,
		},
		{
			LanguageID: "python",
			RawContent: ":: id: simple-py-block\n:: version: 0.1\nprint(\"Hello from simple Python\")",
			StartLine:  15, EndLine: 19,
		},
		{
			LanguageID: "text",
			RawContent: ":: id: simple-empty-block",
			StartLine:  21, EndLine: 23,
		},
		{
			LanguageID: "text",
			RawContent: ":: id: simple-comment-block\n# This is a comment inside.\n-- So is this.\n\n# Even with a blank line.",
			StartLine:  25, EndLine: 31,
		},
	}

	// --- Expected Metadata (Unchanged) ---
	wantMetadata := []map[string]string{
		{"id": "simple-ns-block", "version": "1.0"},
		{"id": "simple-py-block", "version": "0.1"},
		{"id": "simple-empty-block"},
		{"id": "simple-comment-block"},
	}

	t.Logf("Running ExtractAll on %s...", simpleFixturePath)
	testLogger := log.New(io.Discard, "[TEST-SIMPLE] ", 0)
	gotBlocks, err := ExtractAll(simpleContent, testLogger)

	if err != nil {
		t.Fatalf("ExtractAll failed unexpectedly for simple fixture: %v", err)
	}

	// Compare Blocks
	if !reflect.DeepEqual(gotBlocks, wantBlocks) {
		t.Errorf("Mismatch in extracted blocks from %s", simpleFixturePath)
		// (Detailed print logic omitted for brevity, but remains the same)
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
				t.Errorf("  Got : LangID=%q, Start=%d, End=%d, Content=\n---\n%s\n---", gotBlock.LanguageID, gotBlock.StartLine, gotBlock.EndLine, gotBlock.RawContent)
				t.Errorf("  Want: LangID=%q, Start=%d, End=%d, Content=\n---\n%s\n---", wantBlock.LanguageID, wantBlock.StartLine, wantBlock.EndLine, wantBlock.RawContent)
			}
		}
		t.Fatalf("Block extraction failed, cannot proceed to metadata tests.")
	} else {
		t.Logf("ExtractAll successful for simple fixture.")
	}

	// Compare Metadata
	t.Logf("Running LookForMetadata on extracted blocks...")
	if len(gotBlocks) != len(wantMetadata) {
		t.Fatalf("Test setup error: Number of extracted blocks (%d) does not match number of expected metadata maps (%d)", len(gotBlocks), len(wantMetadata))
	}
	for i, block := range gotBlocks {
		t.Run(fmt.Sprintf("Metadata_Block_%d", i), func(t *testing.T) {
			expectedMeta := wantMetadata[i]
			gotMeta, metaErr := LookForMetadata(block.RawContent)
			if metaErr != nil {
				t.Errorf("LookForMetadata returned unexpected error for block %d: %v", i, metaErr)
				return
			}
			if !reflect.DeepEqual(gotMeta, expectedMeta) {
				t.Errorf("Metadata mismatch for block %d (LangID: %q, StartLine: %d)", i, block.LanguageID, block.StartLine)
				t.Errorf("  Got : %#v", gotMeta)
				t.Errorf("  Want: %#v", expectedMeta)
				t.Errorf("  Raw Content Searched:\n---\n%s\n---", block.RawContent)
			}
		})
	}
}
