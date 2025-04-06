// neuroscript/pkg/neurodata/blocks/blocks_extract_test.go
package blocks

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestExtractAllFencedBlocksSimple tests basic block extraction with various content types.
// It focuses on verifying the raw capture logic, including metadata lines.
func TestExtractAllFencedBlocksSimple(t *testing.T) {
	simpleFixturePath := filepath.Join(fixtureDir, "simple_blocks.md")
	simpleContentBytes, errSimple := os.ReadFile(simpleFixturePath)
	if errSimple != nil {
		cwd, _ := os.Getwd()
		t.Fatalf("Failed to read simple fixture file %s (CWD: %s): %v", simpleFixturePath, cwd, errSimple)
	}
	simpleContent := string(simpleContentBytes)

	// Expectations based on RAW capture, including metadata lines
	wantBlocks := []string{
		"# id: simple-ns-block\n# version: 1.0\nDEFINE PROCEDURE Simple()\n  EMIT \"Hello from simple NS\"\nEND",
		"# id: simple-py-block\n# version: 0.1\nprint(\"Hello from simple Python\")",
		"# id: simple-empty-block", // Metadata is the content for raw capture
		"# id: simple-comment-block\n# This is a comment inside.\n-- So is this.\n\n# Even with a blank line.",
	}

	// Call the function under test
	gotBlocks, err := ExtractAllFencedBlocks(simpleContent)
	if err != nil {
		t.Fatalf("ExtractAllFencedBlocks failed unexpectedly for simple fixture: %v", err)
	}

	// Compare results
	if !reflect.DeepEqual(gotBlocks, wantBlocks) {
		t.Errorf("Mismatch in extracted blocks from simple_blocks.md")
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

// TestExtractAllFencedBlocksComplex tests extraction behavior with potentially ambiguous or edge-case fences.
// It specifically checks for the expected error when fences are immediately adjacent.
func TestExtractAllFencedBlocksComplex(t *testing.T) {
	complexFixturePath := filepath.Join(fixtureDir, "complex_blocks.md")
	complexContentBytes, errComplex := os.ReadFile(complexFixturePath)
	if errComplex != nil {
		cwd, _ := os.Getwd()
		t.Fatalf("Failed to read complex fixture file %s (CWD: %s): %v", complexFixturePath, cwd, errComplex)
	}
	complexContent := string(complexContentBytes)

	// Expect extraction up to the point of ambiguity, then an error
	wantBlocksBeforeError := []string{"# id: complex-ns-1\nCALL TOOL.DoSomething()"}
	wantErrMsgContains := "ambiguous fence pattern: line 8 starts with '```' immediately after a previous block closed at line 7"

	gotBlocks, err := ExtractAllFencedBlocks(complexContent)

	// Check for the expected error
	if err == nil {
		t.Fatalf("TestExtractAllFencedBlocksComplex: Expected an error containing %q, but got nil", wantErrMsgContains)
	}
	if !strings.Contains(err.Error(), wantErrMsgContains) {
		t.Errorf("TestExtractAllFencedBlocksComplex: Expected error message containing %q, but got: %v", wantErrMsgContains, err)
	}

	// Check that blocks *before* the error were extracted correctly
	if !reflect.DeepEqual(gotBlocks, wantBlocksBeforeError) {
		t.Errorf("TestExtractAllFencedBlocksComplex: Mismatch in blocks extracted before the expected error.")
		t.Logf("Got %d blocks before error, Want %d blocks", len(gotBlocks), len(wantBlocksBeforeError))
		// Detailed block comparison (similar to TestExtractAllFencedBlocksSimple) can be added here if needed.
	}
}

// TestExtractAllFencedBlocksEvil tests extraction with potentially tricky fence placements, like at the start/end of file.
// It primarily focuses on error detection for malformed structures (like ambiguous fences).
func TestExtractAllFencedBlocksEvil(t *testing.T) {
	evilFixturePath := filepath.Join(fixtureDir, "evil_blocks.md")
	evilContentBytes, errEvil := os.ReadFile(evilFixturePath)
	if errEvil != nil {
		cwd, _ := os.Getwd()
		t.Fatalf("Failed to read evil fixture file %s (CWD: %s): %v", evilFixturePath, cwd, errEvil)
	}
	evilContent := string(evilContentBytes)

	wantErr := true                                                                                                      // This fixture is designed to have an error
	wantErrMsgContains := "ambiguous fence pattern: line 44 starts with '```' immediately after a previous block closed" // Example specific error, adjust if needed

	gotBlocks, err := ExtractAllFencedBlocks(evilContent)

	if !wantErr {
		if err != nil {
			t.Fatalf("TestExtractAllFencedBlocksEvil: Expected no error, but got: %v", err)
		}
	} else {
		if err == nil {
			t.Fatalf("TestExtractAllFencedBlocksEvil: Expected an error containing %q, but got nil. Got blocks: %+v", wantErrMsgContains, gotBlocks)
		}
		if wantErrMsgContains != "" && !strings.Contains(err.Error(), wantErrMsgContains) {
			t.Errorf("TestExtractAllFencedBlocksEvil: Expected error message containing %q, but got: %v", wantErrMsgContains, err)
		}
		// Optionally, check the content of gotBlocks before the error occurred if relevant.
	}
}

// TestExtractAllFencedBlocksWithBlankLines tests raw capture behavior with blank lines within fenced blocks.
func TestExtractAllFencedBlocksWithBlankLines(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantBlocks []string
		wantErr    bool
	}{
		{
			name:       "Blank lines inside",
			content:    "```\n\nActual content line 1.\nActual content line 2.\n\n```",
			wantBlocks: []string{"\nActual content line 1.\nActual content line 2.\n"}, // Raw capture keeps internal blanks
			wantErr:    false,
		},
		{
			name:       "Single blank line inside",
			content:    "```\n\n```",
			wantBlocks: []string{""}, // Single blank line inside becomes empty string after join
			wantErr:    false,
		},
		{
			name:       "Empty block",
			content:    "```\n```",
			wantBlocks: []string{""}, // No content between fences
			wantErr:    false,
		},
		{
			name:       "No leading/trailing blank inside",
			content:    "```\nActual content line 1.\nActual content line 2.\n```",
			wantBlocks: []string{"Actual content line 1.\nActual content line 2."},
			wantErr:    false,
		},
		{
			name:       "Only fence", // Test case for just fences
			content:    "```\n```",
			wantBlocks: []string{""},
			wantErr:    false,
		},
		{
			name:       "Block at EOF without newline",
			content:    "```\nEOF Block",
			wantBlocks: []string{"EOF Block"}, // Should capture content
			wantErr:    true,                  // But should report unclosed block error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBlocks, err := ExtractAllFencedBlocks(tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractAllFencedBlocks() error = %v, wantErr %v", err, tt.wantErr)
				// Log content and result for debugging errors
				if err != nil {
					t.Logf("Content causing error:\n---\n%s\n---", tt.content)
					t.Logf("Blocks extracted before error: %#v", gotBlocks)
				}
				return
			}
			// Only compare block content if no error was expected
			if !tt.wantErr {
				if !reflect.DeepEqual(gotBlocks, tt.wantBlocks) {
					t.Errorf("ExtractAllFencedBlocks() mismatch:\n Got: %#v\n Want: %#v", gotBlocks, tt.wantBlocks)
				}
			} else {
				// If error was expected, check if blocks before error match (if applicable)
				// For the EOF case, we expect the content to be captured *before* the error is returned.
				if tt.name == "Block at EOF without newline" {
					if !reflect.DeepEqual(gotBlocks, tt.wantBlocks) {
						t.Errorf("ExtractAllFencedBlocks() mismatch before EOF error:\n Got: %#v\n Want: %#v", gotBlocks, tt.wantBlocks)
					}
				}
				// Optionally check error message content here if needed.
			}
		})
	}
}
