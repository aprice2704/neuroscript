// pkg/neurodata/blocks2/blocks_evil_test.go
package blocks

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings" // Import strings package
	"testing"
)

// TestExtractAllAndMetadataEvil tests ExtractAll and LookForMetadata using evil_blocks.md.
// It expects an ambiguity error right after the last valid block.
func TestExtractAllAndMetadataEvil(t *testing.T) {
	// Construct the path to the fixture file
	// Assumes fixtureDir constant is available (or redefine it)
	// const fixtureDir = "test_fixtures"
	evilFixturePath := filepath.Join(fixtureDir, "evil_blocks.md")

	// Read the content of the fixture file
	evilContentBytes, errEvil := os.ReadFile(evilFixturePath)
	if errEvil != nil {
		// Get current working directory for context in error message
		cwd, _ := os.Getwd()
		t.Fatalf("Failed to read evil fixture file %s (CWD: %s): %v", evilFixturePath, cwd, errEvil)
	}
	evilContent := string(evilContentBytes)

	// --- Expected Results for ExtractAll (Blocks BEFORE the ambiguity error on line 44) ---
	wantBlocksBeforeError := []FencedBlock{
		{ // Block 0
			LanguageID: "", // No language ID specified
			RawContent: "# id: block-at-start\nContent at start.",
			StartLine:  4,
			EndLine:    7,
		},
		{ // Block 1
			LanguageID: "yaml",
			// Includes the nested fences as raw content
			RawContent: "# id: fences-inside\nkey: value\nexample: |\n  ```bash\n  echo \"This inner fence should be captured.\"\n  ```\nanother_key: true",
			StartLine:  10,
			EndLine:    18,
		},
		{ // Block 2
			LanguageID: "neuroscript-v1.1",
			RawContent: "# id: weird-tag\nDEFINE PROCEDURE TestWeirdTag()\n  EMIT \"Tag test\"\nEND",
			StartLine:  21,
			EndLine:    26,
		},
		{ // Block 3
			LanguageID: "", // No language ID specified
			RawContent: "", // Empty block - content becomes "" after trimming blank lines
			StartLine:  29,
			EndLine:    30,
		},
		{ // Block 4
			LanguageID: "", // No language ID specified
			RawContent: "", // Whitespace-only block content becomes "" after trimming blank lines logic
			StartLine:  33,
			EndLine:    37,
		},
		{ // Block 5
			LanguageID: "javascript",
			RawContent: "# id: block-at-eof\nconsole.log(\"EOF\");",
			StartLine:  40,
			EndLine:    43, // Ends exactly at EOF in the content, before the ambiguous line 44
		},
		// The ambiguity error occurs on line 44, immediately after block 5 closes.
	}

	// --- Expected Metadata for the successfully extracted blocks ---
	wantMetadata := []map[string]string{
		{"id": "block-at-start"}, // Block 0
		{"id": "fences-inside"},  // Block 1
		{"id": "weird-tag"},      // Block 2
		{},                       // Block 3 (Empty)
		{},                       // Block 4 (Whitespace only - no metadata expected)
		{"id": "block-at-eof"},   // Block 5
	}

	// --- Run ExtractAll ---
	t.Logf("Running ExtractAll on %s (expecting ambiguity error)...", evilFixturePath)
	gotBlocks, err := ExtractAll(evilContent)

	// 1. Check for the EXPECTED error
	expectErr := true                                        // We expect an error in this test case
	wantErrContains := "ambiguous fence detected on line 44" // Specific error expected
	if err == nil {
		if expectErr {
			t.Fatalf("ExtractAll succeeded unexpectedly. Expected error containing %q because of ambiguous fence at EOF.", wantErrContains)
		}
	} else {
		if !expectErr {
			t.Fatalf("ExtractAll failed unexpectedly: %v", err)
		}
		if !strings.Contains(err.Error(), wantErrContains) {
			t.Fatalf("ExtractAll returned error, but expected message containing %q, got: %v", wantErrContains, err)
		} else {
			t.Logf("ExtractAll correctly returned expected error: %v", err)
		}
	}

	// 2. Compare the blocks extracted BEFORE the error occurred
	if !reflect.DeepEqual(gotBlocks, wantBlocksBeforeError) {
		t.Errorf("Mismatch in blocks extracted before the expected error from %s", evilFixturePath)
		t.Logf("Got %d blocks, Want %d blocks", len(gotBlocks), len(wantBlocksBeforeError))

		// Print detailed comparison
		maxLen := len(gotBlocks)
		if len(wantBlocksBeforeError) > maxLen {
			maxLen = len(wantBlocksBeforeError)
		}
		for i := 0; i < maxLen; i++ {
			var gotBlock FencedBlock
			if i < len(gotBlocks) {
				gotBlock = gotBlocks[i]
			}
			var wantBlock FencedBlock
			if i < len(wantBlocksBeforeError) {
				wantBlock = wantBlocksBeforeError[i]
			}

			if !reflect.DeepEqual(gotBlock, wantBlock) {
				t.Errorf("--- ExtractAll Block %d Mismatch ---", i)
				t.Errorf("  Got : LangID=%q, Start=%d, End=%d, Content=\n---\n%q\n---", gotBlock.LanguageID, gotBlock.StartLine, gotBlock.EndLine, gotBlock.RawContent)
				t.Errorf("  Want: LangID=%q, Start=%d, End=%d, Content=\n---\n%q\n---", wantBlock.LanguageID, wantBlock.StartLine, wantBlock.EndLine, wantBlock.RawContent)
			}
		}
		t.Logf("\nFull Got Blocks Before Error:\n%#v\n", gotBlocks)
		t.Logf("\nFull Want Blocks Before Error:\n%#v\n", wantBlocksBeforeError)
		// Fail if blocks before error don't match
		t.FailNow()
	} else {
		t.Logf("Blocks extracted before error match expected blocks.")
	}

	// --- Run LookForMetadata on the successfully extracted blocks ---
	t.Logf("Running LookForMetadata on the %d successfully extracted blocks...", len(gotBlocks))
	if len(gotBlocks) != len(wantMetadata) {
		t.Fatalf("Test setup error: Number of successfully extracted blocks (%d) does not match number of expected metadata maps (%d)", len(gotBlocks), len(wantMetadata))
	}

	for i, block := range gotBlocks {
		t.Run(fmt.Sprintf("Metadata_Block_%d", i), func(t *testing.T) {
			if i >= len(wantMetadata) {
				t.Fatalf("Internal test error: trying to access wantMetadata index %d which is out of bounds (len %d)", i, len(wantMetadata))
			}
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
