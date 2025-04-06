// pkg/neurodata/blocks2/blocks_complex_test.go
package blocks

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings" // Import strings package
	"testing"
)

// TestExtractAllAndMetadataComplex tests ExtractAll and LookForMetadata using complex_blocks.md.
// It specifically expects ExtractAll to return an AMBIGUITY error after the first block.
func TestExtractAllAndMetadataComplex(t *testing.T) {
	// Construct the path to the fixture file
	// Assumes fixtureDir constant is available (or redefine it)
	// const fixtureDir = "test_fixtures" // Uncomment if running this test in isolation
	complexFixturePath := filepath.Join(fixtureDir, "complex_blocks.md")

	// Read the content of the fixture file
	complexContentBytes, errComplex := os.ReadFile(complexFixturePath)
	if errComplex != nil {
		// Get current working directory for context in error message
		cwd, _ := os.Getwd()
		t.Fatalf("Failed to read complex fixture file %s (CWD: %s): %v", complexFixturePath, cwd, errComplex)
	}
	complexContent := string(complexContentBytes)

	// --- Expected Results for ExtractAll (ONLY the block BEFORE the ambiguity error) ---
	wantBlocksBeforeError := []FencedBlock{
		{
			LanguageID: "neuroscript",
			RawContent: "# id: complex-ns-1\nCALL TOOL.DoSomething()",
			StartLine:  4, // Line numbers based on complex_blocks.md content
			EndLine:    7,
		},
		// The ambiguity error occurs on line 8, so only the first block is expected.
	}

	// --- Expected Metadata for the successfully extracted block(s) ---
	wantMetadata := []map[string]string{
		{"id": "complex-ns-1"}, // Block 0
	}

	// --- Run ExtractAll ---
	t.Logf("Running ExtractAll on %s (expecting ambiguity error)...", complexFixturePath)
	gotBlocks, err := ExtractAll(complexContent)

	// 1. Check for the EXPECTED error
	//wantErr := true
	// Match the error message precisely from the previous run's output
	wantErrContains := "ambiguous fence detected on line 8, immediately following block closed on line 7"
	if err == nil {
		t.Fatalf("ExtractAll succeeded unexpectedly. Expected error containing %q because of ambiguous fence.", wantErrContains)
	}
	if !strings.Contains(err.Error(), wantErrContains) {
		t.Fatalf("ExtractAll returned error, but expected message containing %q, got: %v", wantErrContains, err)
	} else {
		t.Logf("ExtractAll correctly returned expected error: %v", err)
	}

	// 2. Compare the blocks extracted BEFORE the error occurred
	if !reflect.DeepEqual(gotBlocks, wantBlocksBeforeError) {
		t.Errorf("Mismatch in blocks extracted before the expected error from %s", complexFixturePath)
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
	// This should only run if the error was correct AND the blocks extracted before it were correct.
	t.Logf("Running LookForMetadata on the %d successfully extracted blocks...", len(gotBlocks))
	if len(gotBlocks) != len(wantMetadata) {
		// This check might be redundant given the DeepEqual above, but good for clarity
		t.Fatalf("Test setup error: Number of successfully extracted blocks (%d) does not match number of expected metadata maps (%d)", len(gotBlocks), len(wantMetadata))
	}

	for i, block := range gotBlocks {
		// Since wantBlocksBeforeError only has one entry, i will only be 0 here if the test passed the previous check.
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
