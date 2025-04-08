// pkg/neurodata/blocks/blocks_complex_test.go
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

func TestExtractAllAndMetadataComplex(t *testing.T) {
	// Assume fixtureDir is defined in simple_test.go
	complexFixturePath := filepath.Join(fixtureDir, "complex_blocks.md")
	complexContentBytes, errComplex := os.ReadFile(complexFixturePath)
	if errComplex != nil {
		cwd, _ := os.Getwd()
		t.Fatalf("Failed to read complex fixture file %s (CWD: %s): %v", complexFixturePath, cwd, errComplex)
	}
	complexContent := string(complexContentBytes)

	// --- Expected Results (Only block before ambiguity on line 8, RawContent fixed) ---
	wantBlocksBeforeError := []FencedBlock{
		{
			LanguageID: "neuroscript",
			RawContent: ":: id: complex-ns-1\nCALL TOOL.DoSomething()",
			StartLine:  4, EndLine: 7,
		},
	}

	// --- Expected Metadata (Only Block 0) ---
	wantMetadata := []map[string]string{
		{"id": "complex-ns-1"},
	}

	t.Logf("Running ExtractAll on %s (expecting ambiguous fence error)...", complexFixturePath)
	testLogger := log.New(io.Discard, "[TEST-COMPLEX] ", 0)
	gotBlocks, err := ExtractAll(complexContent, testLogger)

	// 1. Check for the EXPECTED error (Ambiguous Fence on line 8)
	wantErrContains := "ambiguous fence detected on line 8"
	if err == nil {
		t.Errorf("ExtractAll succeeded unexpectedly. Expected error containing %q.", wantErrContains)
		t.Logf("Got blocks on unexpected success:\n%#v", gotBlocks)
	} else if !strings.Contains(err.Error(), wantErrContains) {
		t.Errorf("ExtractAll returned error, but expected message containing %q, got: %v", wantErrContains, err)
		t.Logf("Got blocks on error:\n%#v", gotBlocks)
	} else {
		t.Logf("ExtractAll correctly returned expected error: %v", err)
	}

	// 2. Compare blocks extracted BEFORE the error
	if !reflect.DeepEqual(gotBlocks, wantBlocksBeforeError) {
		t.Errorf("Mismatch in blocks extracted before the expected error from %s", complexFixturePath)
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
