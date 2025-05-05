// pkg/neurodata/blocks/blocks_complex_test.go
package blocks

import (
	"errors" // Keep for error check
	"io"
	"os"
	"path/filepath"
	"strings" // Keep for error check
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// Assume fixtureDir is defined in blocks_helpers.go
// Assume helper functions minInt, compareBlockSlices are defined in blocks_helpers.go

var logger *adapters.SlogAdapter

func init() {
	logger, _ = adapters.NewSimpleSlogAdapter(os.Stderr, logging.LogLevelDebug)
}

func TestExtractAllAndMetadataComplex(t *testing.T) {
	complexFixturePath := filepath.Join(fixtureDir, "complex_blocks.md")
	complexContentBytes, errComplex := os.ReadFile(complexFixturePath)
	if errComplex != nil {
		cwd, _ := os.Getwd()
		t.Fatalf("Failed to read complex fixture file %s (CWD: %s): %v", complexFixturePath, cwd, errComplex)
	}
	complexContent := string(complexContentBytes)

	// --- UPDATED Expected Results based on user request ---
	wantBlocks := []FencedBlock{
		{ // Block 1 (User Index)
			LanguageID: "neuroscript",
			RawContent: "CALL TOOL.DoSomething()", // Content as extracted between fences
			StartLine:  5,                         // Line of opening ```neuroscript
			EndLine:    7,                         // Line of closing ```
			Metadata:   map[string]string{"id": "complex-ns-1"},
		},
		{ // Block 2 (User Index)
			LanguageID: "python",
			RawContent: "import os",
			StartLine:  9,  // Line of opening ```python
			EndLine:    11, // Line of closing ```
			Metadata:   map[string]string{"id": "complex-py-adjacent"},
		},
		{ // Block 3 (User Index)
			LanguageID: "text",
			RawContent: "", // Empty content
			StartLine:  16, // Line of opening ```text
			EndLine:    17, // Line of closing ```
			Metadata:   map[string]string{"id": "metadata-only-block", "version": "1.1"},
		},
		{ // Block 4 (User Index)
			LanguageID: "javascript",
			RawContent: "console.log(\"No ID here\");",
			StartLine:  20,                  // Line of opening ```javascript
			EndLine:    22,                  // Line of closing ```
			Metadata:   map[string]string{}, // Expecting empty map for "None"
		},
		{ // Block 5 (User Index)
			LanguageID: "go",
			RawContent: "package main\nimport \"fmt\"\nfunc main(){ fmt.Println(\"Go!\") }", // Joined lines
			StartLine:  28,                                                                  // Line of opening ```go
			EndLine:    32,                                                                  // Line of closing ```
			Metadata:   map[string]string{"version": "0.2", "id": "go-block-late-id"},
		},
		{ // Block 6 (User Index)
			LanguageID: "neurodata-checklist",
			RawContent: "- [x] Item A\n- [ ] Item B", // Joined lines, no trailing newline from joiner
			StartLine:  37,                           // Line of opening ```neurodata-checklist
			EndLine:    40,                           // Line of closing ```
			Metadata:   map[string]string{},          // Expecting empty map for "None"
		},
		// Unclosed block is omitted
	}
	// --- END UPDATED Expected Results ---

	expectError := false // No error expected for extraction itself
	errorContains := ""

	t.Run("Complex Fixture Extraction (Ignore Unclosed)", func(t *testing.T) {
		t.Logf("Running ExtractAll on %s (expecting NO error, ignore unclosed block)...", complexFixturePath)
		gotBlocks, err := ExtractAll(complexContent, logger)

		// Check for Unexpected Error
		if expectError {
			// This branch is unlikely to be hit now
			if err == nil {
				t.Fatalf("ExtractAll succeeded unexpectedly. Expected error containing %q.", errorContains)
			}
			if errorContains != "" && !strings.Contains(err.Error(), errorContains) {
				t.Fatalf("ExtractAll returned error, but expected message containing %q, got: %v", errorContains, err)
			}
		} else {
			// Expect NO error from ExtractAll itself (warning for unclosed is logged, not returned)
			if err != nil {
				// Allow specific non-fatal errors if necessary, otherwise fail
				if !errors.Is(err, io.EOF) { // Example: Allow io.EOF if scanner has issues at end
					t.Fatalf("ExtractAll failed unexpectedly: %v\nGot blocks: %#v", err, gotBlocks)
				} else {
					t.Logf("ExtractAll returned non-fatal error: %v", err) // Log allowed errors
				}
			} else {
				t.Logf("ExtractAll successful (returned nil error). Found %d blocks.", len(gotBlocks))
			}
		}

		// Compare Blocks using the helper
		compareBlockSlices(t, gotBlocks, wantBlocks, complexFixturePath)

		if t.Failed() {
			t.Logf("Test failed during block comparison.")
		} else {
			t.Logf("Block comparison successful.")
		}

	}) // End t.Run
}
