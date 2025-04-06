// pkg/neurodata/blocks/blocks_test.go
package blocks

import (
	"os"
	"path/filepath"
	"reflect" // Keep reflect for DeepEqual on list results
	"strings"
	"testing"

	// Import core for tool testing types if needed by tool tests
	"[github.com/aprice2704/neuroscript/pkg/core](https://www.google.com/search?q=https://github.com/aprice2704/neuroscript/pkg/core)" // Needed for Interpreter context in tool tests
)

// --- Test ExtractAllFencedBlocks (Moved from core/embed_blocks_test.go) ---
// Create test fixtures subdirectory within this package if needed, or adjust path
const fixtureDir = "../core/test_fixtures" // Adjust path relative to this file

func TestExtractAllFencedBlocksSimple(t *testing.T) {
	// ... test logic remains the same, using fixtureDir ...
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
		"# id: simple-empty-block", // Expect empty string here
		"# id: simple-comment-block\n# This is a comment inside.\n-- So is this.\n\n# Even with a blank line.",
	}

	gotBlocks, err := ExtractAllFencedBlocks(simpleContent)
	if err != nil {
		t.Fatalf("ExtractAllFencedBlocks failed unexpectedly for simple fixture: %v", err)
	}
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

func TestExtractAllFencedBlocksComplex(t *testing.T) {
	// ... test logic remains the same, using fixtureDir ...
	complexFixturePath := filepath.Join(fixtureDir, "complex_blocks.md")
	complexContentBytes, errComplex := os.ReadFile(complexFixturePath)
	if errComplex != nil {
		cwd, _ := os.Getwd()
		t.Fatalf("Failed to read complex fixture file %s (CWD: %s): %v", complexFixturePath, cwd, errComplex)
	}
	complexContent := string(complexContentBytes)

	wantBlocksBeforeError := []string{
		"# id: complex-ns-1\nCALL TOOL.DoSomething()", // Only the first block
	}
	wantErrMsgContains := "ambiguous fence pattern: line 8 starts with '```' immediately after a previous block closed at line 7"

	gotBlocks, err := ExtractAllFencedBlocks(complexContent)

	if err == nil {
		t.Fatalf("TestExtractAllFencedBlocksComplex: Expected an error containing %q, but got nil", wantErrMsgContains)
	}
	if !strings.Contains(err.Error(), wantErrMsgContains) {
		t.Errorf("TestExtractAllFencedBlocksComplex: Expected error message containing %q, but got: %v", wantErrMsgContains, err)
	}
	if !reflect.DeepEqual(gotBlocks, wantBlocksBeforeError) {
		t.Errorf("TestExtractAllFencedBlocksComplex: Mismatch in blocks extracted before the expected error.")
		// ... (detailed diff logic as before) ...
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

func TestExtractAllFencedBlocksEvil(t *testing.T) {
	// ... test logic remains the same, using fixtureDir ...
	evilFixturePath := filepath.Join(fixtureDir, "evil_blocks.md") // Ensure this file has a newline after the final ```
	evilContentBytes, errEvil := os.ReadFile(evilFixturePath)
	if errEvil != nil {
		cwd, _ := os.Getwd()
		t.Fatalf("Failed to read evil fixture file %s (CWD: %s): %v", evilFixturePath, cwd, errEvil)
	}
	evilContent := string(evilContentBytes)

	wantErr := true
	wantErrMsgContains := "ambiguous fence pattern: line 44 starts with '```' immediately after a previous block closed at line 43"

	gotBlocks, err := ExtractAllFencedBlocks(evilContent) // Removed debug param

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
	}
}

func TestExtractAllFencedBlocksWithBlankLines(t *testing.T) {
	// ... test logic remains the same ...
	content := "```\n\nActual content line 1.\nActual content line 2.\n\n```"
	wantBlocks := []string{"\nActual content line 1.\nActual content line 2.\n"} // Includes internal blank lines
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
	wantBlocksSingleBlank := []string{""} // Just a blank line inside becomes "" content
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
	wantBlocksEmpty := []string{""} // Empty content.
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

// --- Test TOOL.ExtractFencedBlock (Moved from core/tools_composite_doc_test.go) ---
func TestToolExtractFencedBlockByID(t *testing.T) {
	// Need a core.Interpreter for context
	dummyInterp := core.NewInterpreter(nil) // Use core.NewInterpreter

	spec := core.ToolSpec{ // Use core types
		Name: "ExtractFencedBlock",
		Args: []core.ArgSpec{
			{Name: "content", Type: core.ArgTypeString, Required: true},
			{Name: "block_id", Type: core.ArgTypeString, Required: true},
			{Name: "block_type", Type: core.ArgTypeString, Required: false}, // Optional
		},
		ReturnType: core.ArgTypeString,
	}

	// --- Load Fixture Files ---
	simpleFixturePath := filepath.Join(fixtureDir, "simple_blocks.md")
	complexFixturePath := filepath.Join(fixtureDir, "complex_blocks.md")
	simpleContentBytes, errSimple := os.ReadFile(simpleFixturePath)
	if errSimple != nil {
		t.Fatalf("Failed to read simple fixture file %s: %v", simpleFixturePath, errSimple)
	}
	simpleContent := string(simpleContentBytes)
	complexContentBytes, errComplex := os.ReadFile(complexFixturePath)
	if errComplex != nil {
		t.Fatalf("Failed to read complex fixture file %s: %v", complexFixturePath, errComplex)
	}
	complexContent := string(complexContentBytes)
	// --- End Load Fixture Files ---

	// Helper for arg validation and conversion (using core function)
	makeToolArgs := func(args ...interface{}) []interface{} {
		if args == nil {
			return []interface{}{}
		}
		return args
	}

	tests := []struct {
		name           string
		content        string        // Content loaded from fixture
		args           []interface{} // Tool arguments (content placeholder)
		want           string
		wantErr        bool // Expects error string prefix "Error:" from tool logic
		errContains    string
		valWantErr     bool // Expect validation error from ValidateAndConvertArgs
		valErrContains string
	}{
		// --- SUCCESS CASES using simple_blocks.md ---
		{"Simple NS Block", simpleContent, makeToolArgs(simpleContent, "simple-ns-block"), "DEFINE PROCEDURE Simple()\n  EMIT \"Hello from simple NS\"\nEND", false, "", false, ""},
		{"Simple PY Block", simpleContent, makeToolArgs(simpleContent, "simple-py-block"), `print("Hello from simple Python")`, false, "", false, ""},
		{"Simple Empty Block", simpleContent, makeToolArgs(simpleContent, "simple-empty-block"), "", false, "", false, ""},
		{"Simple Comment Block", simpleContent, makeToolArgs(simpleContent, "simple-comment-block"), "# This is a comment inside.\n-- So is this.\n\n# Even with a blank line.", false, "", false, ""},
		{"Simple NS with type match", simpleContent, makeToolArgs(simpleContent, "simple-ns-block", "neuroscript"), "DEFINE PROCEDURE Simple()\n  EMIT \"Hello from simple NS\"\nEND", false, "", false, ""},
		{"Simple Match First Block", simpleContent, makeToolArgs(simpleContent, ""), "DEFINE PROCEDURE Simple()\n  EMIT \"Hello from simple NS\"\nEND", false, "", false, ""}, // Empty ID matches first

		// --- SUCCESS CASES using complex_blocks.md ---
		{"Complex NS Block 1", complexContent, makeToolArgs(complexContent, "complex-ns-1"), "CALL TOOL.DoSomething()", false, "", false, ""},
		{"Complex PY Adjacent Block", complexContent, makeToolArgs(complexContent, "complex-py-adjacent"), "import os", false, "", false, ""},
		{"Metadata Only Block", complexContent, makeToolArgs(complexContent, "metadata-only-block"), "", false, "", false, ""},
		{"Checklist Hyphen Meta", complexContent, makeToolArgs(complexContent, "checklist-hyphen-meta"), "- [x] Item A\n- [ ] Item B", false, "", false, ""},
		{"Complex Match First Block", complexContent, makeToolArgs(complexContent, ""), "CALL TOOL.DoSomething()", false, "", false, ""}, // Empty ID matches first

		// --- ERROR CASES using complex_blocks.md ---
		{"Block with No ID (Not Found)", complexContent, makeToolArgs(complexContent, "block-with-no-id"), "", true, "Block ID 'block-with-no-id' not found", false, ""},
		{"Go Block Late ID (Not Found by Tool)", complexContent, makeToolArgs(complexContent, "go-block-late-id"), "", true, "Block ID 'go-block-late-id' not found", false, ""},
		{"Unclosed Block", complexContent, makeToolArgs(complexContent, "unclosed-markdown-block"), "", true, "closing fence '```' not found", false, ""},

		// --- ERROR CASES using simple_blocks.md ---
		{"Error ID not found (simple)", simpleContent, makeToolArgs(simpleContent, "nonexistent-id"), "", true, "Block ID 'nonexistent-id' not found", false, ""},
		{"Error type mismatch (simple)", simpleContent, makeToolArgs(simpleContent, "simple-ns-block", "python"), "", true, "type mismatch: expected 'python', got 'neuroscript'", false, ""},
		{"Error match first type mismatch", simpleContent, makeToolArgs(simpleContent, "", "python"), "", true, "type mismatch: expected 'python', got 'neuroscript'", false, ""}, // Match first, wrong type

		// --- VALIDATION ERROR CASES ---
		{"Validation Wrong Arg Count (1)", simpleContent, makeToolArgs("content"), "", false, "", true, "tool 'ExtractFencedBlock' expected at least 2 arguments"},
		{"Validation Wrong Arg Type (content)", simpleContent, makeToolArgs(123, "id"), "", false, "", true, "argument 'content' (index 0): expected string"},
		{"Validation Wrong Arg Type (block_id)", simpleContent, makeToolArgs("content", 123), "", false, "", true, "argument 'block_id' (index 1): expected string"},
		{"Validation Wrong Arg Type (block_type)", simpleContent, makeToolArgs("content", "id", 123), "", false, "", true, "argument 'block_type' (index 2): expected string"},
	}

	// Test runner loop
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// --- Argument Setup ---
			var finalArgs []interface{}
			if len(tt.args) > 0 {
				tempArgs := make([]interface{}, len(tt.args))
				copy(tempArgs, tt.args)
				tempArgs[0] = tt.content // Replace placeholder content
				finalArgs = tempArgs
			} else {
				finalArgs = tt.args
			}

			// --- Validation Check ---
			convertedArgs, valErr := core.ValidateAndConvertArgs(spec, finalArgs) // Use core validator
			if (valErr != nil) != tt.valWantErr {
				t.Errorf("ValidateAndConvertArgs() error = %v, valWantErr %v", valErr, tt.valWantErr)
				return
			}
			if tt.valWantErr {
				if tt.valErrContains != "" && (valErr == nil || !strings.Contains(valErr.Error(), tt.valErrContains)) {
					t.Errorf("ValidateAndConvertArgs() expected error containing %q, got: %v", tt.valErrContains, valErr)
				}
				return
			}
			if !tt.valWantErr && valErr != nil {
				t.Fatalf("ValidateAndConvertArgs() unexpected validation error: %v. Args: %#v", valErr, finalArgs)
			}

			// --- Tool Execution Check ---
			if !tt.valWantErr {
				gotInterface, toolErr := toolExtractFencedBlockByID(dummyInterp, convertedArgs)
				if toolErr != nil {
					t.Fatalf("toolExtractFencedBlockByID returned unexpected Go error: %v", toolErr)
				}

				gotStr, ok := gotInterface.(string)
				if !ok {
					t.Fatalf("toolExtractFencedBlockByID did not return a string result, got %T", gotInterface)
				}

				isReturnedError := strings.HasPrefix(gotStr, "Error:")

				if isReturnedError != tt.wantErr {
					t.Errorf("Expected error state mismatch. wantErr=%v but returned string is: %q", tt.wantErr, gotStr)
				}

				if tt.wantErr {
					if !isReturnedError {
						t.Errorf("Expected error string prefix 'Error:', but got success string: %q", gotStr)
					}
					if tt.errContains != "" && !strings.Contains(gotStr, tt.errContains) {
						t.Errorf("Expected error string containing %q, got: %q", tt.errContains, gotStr)
					}
				} else {
					if isReturnedError {
						t.Errorf("Expected success, but got error string: %q", gotStr)
					}
					if gotStr != tt.want {
						t.Errorf("Result mismatch:\ngot:  %q\nwant: %q", gotStr, tt.want)
					}
				}
			}
		})
	}
}
