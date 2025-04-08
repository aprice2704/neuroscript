// pkg/neurodata/blocks/blocks_evil_test.go
package blocks

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Removed fixtureDir definition

func TestExtractAllAndMetadataEvil(t *testing.T) {
	// Assume fixtureDir is defined in simple_test.go
	evilFixturePath := filepath.Join(fixtureDir, "evil_blocks.md")
	evilContentBytes, errEvil := os.ReadFile(evilFixturePath)
	if errEvil != nil {
		cwd, _ := os.Getwd()
		t.Fatalf("Failed to read evil fixture file %s (CWD: %s): %v", evilFixturePath, cwd, errEvil)
	}
	evilContent := string(evilContentBytes)

	// --- Expected Results (RawContent includes :: lines, trimmed) ---
	wantBlocksBeforeError := []FencedBlock{
		{ // Block 0
			LanguageID: "",
			RawContent: ":: id: block-at-start\nContent at start.",
			StartLine:  4, EndLine: 7,
		},
		{ // Block 1
			LanguageID: "yaml",
			RawContent: ":: id: fences-inside\nkey: value\nexample: |\n  ```bash\n  echo \"This inner fence should be captured.\"\n  ```\nanother_key: true",
			StartLine:  10, EndLine: 18,
		},
		{ // Block 2
			LanguageID: "neuroscript-v1.1",
			RawContent: ":: id: weird-tag\nDEFINE PROCEDURE TestWeirdTag()\n  EMIT \"Tag test\"\nEND",
			StartLine:  21, EndLine: 26,
		},
		{ // Block 3
			LanguageID: "",
			RawContent: ":: id: empty-block-with-meta",
			StartLine:  29, EndLine: 31,
		},
		{ // Block 4
			LanguageID: "",
			RawContent: ":: id: whitespace-block-meta", // Inner whitespace preserved until final trim
			StartLine:  34, EndLine: 40,
		},
		{ // Block 5
			LanguageID: "javascript",
			RawContent: ":: id: block-at-eof\nconsole.log(\"EOF\");",
			StartLine:  43, EndLine: 46,
		},
	}

	// --- Expected Metadata (Unchanged) ---
	wantMetadata := []map[string]string{
		{"id": "block-at-start"},
		{"id": "fences-inside"},
		{"id": "weird-tag"},
		{"id": "empty-block-with-meta"},
		{"id": "whitespace-block-meta"},
		{"id": "block-at-eof"},
	}

	t.Logf("Running ExtractAll on %s (expecting ambiguity error)...", evilFixturePath)
	testLogger := log.New(io.Discard, "[TEST-EVIL] ", 0)
	gotBlocks, err := ExtractAll(evilContent, testLogger)

	// 1. Check for the EXPECTED error (Ambiguity on line 47)
	expectErr := true
	wantErrContains := "ambiguous fence detected on line 47"
	if err == nil {
		if expectErr {
			t.Errorf("ExtractAll succeeded unexpectedly. Expected error containing %q.", wantErrContains)
		}
	} else {
		if !expectErr {
			t.Errorf("ExtractAll failed unexpectedly: %v", err)
		}
		if !strings.Contains(err.Error(), wantErrContains) {
			t.Errorf("ExtractAll returned error, but expected message containing %q, got: %v", wantErrContains, err)
		} else {
			t.Logf("ExtractAll correctly returned expected error: %v", err)
		}
	}

	// 2. Compare blocks extracted BEFORE the error
	if !reflect.DeepEqual(gotBlocks, wantBlocksBeforeError) {
		t.Errorf("Mismatch in blocks extracted before the expected error from %s", evilFixturePath)
		// (Detailed print logic omitted for brevity)
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
				t.Errorf("  Got : LangID=%q, Start=%d, End=%d, Content=\n---\n%s\n---", gotBlock.LanguageID, gotBlock.StartLine, gotBlock.EndLine, gotBlock.RawContent)
				t.Errorf("  Want: LangID=%q, Start=%d, End=%d, Content=\n---\n%s\n---", wantBlock.LanguageID, wantBlock.StartLine, wantBlock.EndLine, wantBlock.RawContent)
			}
		}
		t.Logf("\nFull Got Blocks Before Error:\n%#v\n", gotBlocks)
		t.Logf("\nFull Want Blocks Before Error:\n%#v\n", wantBlocksBeforeError)

	} else {
		t.Logf("Blocks extracted before error match expected blocks.")
	}

	// --- Run LookForMetadata ---
	t.Logf("Running LookForMetadata on the %d extracted blocks...", len(gotBlocks))
	if len(gotBlocks) != len(wantMetadata) {
		t.Errorf("Test setup error: Number of extracted blocks (%d) does not match number of expected metadata maps (%d)", len(gotBlocks), len(wantMetadata))
	}
	checkLen := len(gotBlocks)
	if len(wantMetadata) < checkLen {
		checkLen = len(wantMetadata)
	}
	for i := 0; i < checkLen; i++ {
		block := gotBlocks[i]
		t.Run(fmt.Sprintf("Metadata_Block_%d", i), func(t *testing.T) {
			expectedMeta := wantMetadata[i]
			gotMeta, metaErr := LookForMetadata(block.RawContent) // Uses updated RawContent

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
