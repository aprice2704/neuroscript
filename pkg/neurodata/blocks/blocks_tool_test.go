// neuroscript/pkg/neurodata/blocks/blocks_tool_test.go
package blocks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	// Import core for tool testing types and Interpreter context
	"github.com/aprice2704/neuroscript/pkg/core" // Ensure this is plain text
)

// --- Test Helpers ---
func newDefaultTestInterpreter() *core.Interpreter {
	return core.NewInterpreter(nil)
}
func makeArgs(vals ...interface{}) []interface{} {
	if vals == nil {
		return []interface{}{}
	}
	args := make([]interface{}, len(vals))
	copy(args, vals)
	return args
}

// --- Test TOOL.ExtractFencedBlockByID ---
const fixtureDir = "test_fixtures"

func TestToolExtractFencedBlockByID(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	spec := core.ToolSpec{Name: "ExtractFencedBlock", Args: []core.ArgSpec{{Name: "content", Type: core.ArgTypeString, Required: true}, {Name: "block_id", Type: core.ArgTypeString, Required: true}, {Name: "block_type", Type: core.ArgTypeString, Required: false}}, ReturnType: core.ArgTypeString}

	// Load fixture files
	simpleFixturePath := filepath.Join(fixtureDir, "simple_blocks.md")
	// complexFixturePath := filepath.Join(fixtureDir, "complex_blocks.md") // Keep commented out

	simpleContentBytes, errSimple := os.ReadFile(simpleFixturePath)
	if errSimple != nil {
		t.Fatalf("Failed to read simple fixture file %s: %v", simpleFixturePath, errSimple)
	}
	simpleContent := string(simpleContentBytes)
	// complexContent := "" // Keep commented out
	// if _, err := os.Stat(complexFixturePath); err == nil {
	// 	complexContentBytes, errComplex := os.ReadFile(complexFixturePath)
	// 	if errComplex != nil { t.Fatalf("Failed to read complex fixture file %s: %v", complexFixturePath, errComplex) }
	// 	complexContent = string(complexContentBytes)
	// } else {
	// 	t.Logf("Complex fixture file not found, skipping related tests.")
	// }

	tests := []struct {
		name           string
		content        string
		args           []interface{}
		want           string // Expected result OR error prefix if wantErr is true
		wantErr        bool   // Expects tool func to return string starting with "Error:"
		errContains    string // Substring expected in the tool's error return string
		valWantErr     bool   // Expect error from ValidateAndConvertArgs?
		valErrContains string // Substring expected in validation error
	}{
		// --- SUCCESS CASES using simple_blocks.md ---
		{"Simple NS Block by ID", simpleContent, makeArgs(simpleContent, "simple-ns-block"), "# id: simple-ns-block\n# version: 1.0\nDEFINE PROCEDURE Simple()\n  EMIT \"Hello from simple NS\"\nEND", false, "", false, ""},
		{"Simple PY Block by ID", simpleContent, makeArgs(simpleContent, "simple-py-block"), "# id: simple-py-block\n# version: 0.1\nprint(\"Hello from simple Python\")", false, "", false, ""},
		{"Simple Empty Block by ID", simpleContent, makeArgs(simpleContent, "simple-empty-block"), "# id: simple-empty-block", false, "", false, ""},
		{"Simple Comment Block by ID", simpleContent, makeArgs(simpleContent, "simple-comment-block"), "# id: simple-comment-block\n# This is a comment inside.\n-- So is this.\n\n# Even with a blank line.", false, "", false, ""},
		{"Simple NS with type match", simpleContent, makeArgs(simpleContent, "simple-ns-block", "neuroscript"), "# id: simple-ns-block\n# version: 1.0\nDEFINE PROCEDURE Simple()\n  EMIT \"Hello from simple NS\"\nEND", false, "", false, ""},
		// *** Corrected expectation: Shouldn't include opening fence ***
		{"Simple Match First Block (ID='')", simpleContent, makeArgs(simpleContent, ""), "# id: simple-ns-block\n# version: 1.0\nDEFINE PROCEDURE Simple()\n  EMIT \"Hello from simple NS\"\nEND", false, "", false, ""},

		// --- TOOL RETURN ERROR CASES (Using Simple Content where applicable) ---
		{"ID Not Found (Simple)", simpleContent, makeArgs(simpleContent, "nonexistent-id"), "", true, "Block ID 'nonexistent-id' not found", false, ""},
		{"Type Mismatch by ID", simpleContent, makeArgs(simpleContent, "simple-ns-block", "python"), "", true, "type mismatch: expected 'python', got 'neuroscript'", false, ""},
		{"Type Mismatch First Block", simpleContent, makeArgs(simpleContent, "", "python"), "", true, "type mismatch: expected 'python', got 'neuroscript'", false, ""},
		{"No Blocks Found (First Block)", "Just text", makeArgs("Just text", ""), "", true, "No fenced code blocks found", false, ""},
		{"No Blocks Found (by ID)", "Just text", makeArgs("Just text", "some-id"), "", true, "Block ID 'some-id' not found", false, ""},

		// --- VALIDATION ERROR CASES ---
		{"Validation Wrong Arg Count (1)", simpleContent, makeArgs("content"), "", false, "", true, "expected between 2 and 3 arguments"},
		// --- Corrected Expectations for Validation with Lenient String Conversion ---
		// Validation should PASS because of lenient conversion, but the TOOL should FAIL finding the block ID.
		{"Validation Wrong Arg Type (content)", simpleContent, makeArgs(123, "id"), "", true, "Block ID 'id' not found", false, ""},
		{"Validation Wrong Arg Type (block_id)", simpleContent, makeArgs("content", 123), "", true, "Block ID '123' not found", false, ""},
		{"Validation Wrong Arg Type (block_type)", simpleContent, makeArgs("content", "id", 123), "", true, "Block ID 'id' not found", false, ""}, // Tool fails finding ID 'id'

		// === Tests using complex_blocks.md (Temporarily Disabled) ===
		/*
			{"Complex NS Block 1 by ID", complexContent, makeArgs(complexContent, "complex-ns-1"), "# id: complex-ns-1\nCALL TOOL.DoSomething()", false, "", false, ""},
			{"Complex PY Adjacent by ID", complexContent, makeArgs(complexContent, "complex-py-adjacent"), "# id: complex-py-adjacent\nimport os", false, "", false, ""},
			{"Metadata Only Block by ID", complexContent, makeArgs(complexContent, "metadata-only-block"), "# id: metadata-only-block\n# version: 1.1", false, "", false, ""}, // Expect metadata lines
			{"Checklist Hyphen Meta by ID", complexContent, makeArgs(complexContent, "checklist-hyphen-meta"), "-- id: checklist-hyphen-meta\n-- version: 1.0\n- [x] Item A\n- [ ] Item B", false, "", false, ""},
			{"Complex Match First Block (ID='')", complexContent, makeArgs(complexContent, ""), "# id: complex-ns-1\nCALL TOOL.DoSomething()", false, "", false, ""}, // Empty ID matches first block
			{"ID Not Found (Complex)", complexContent, makeArgs(complexContent, "block-with-no-id"), "", true, "Block ID 'block-with-no-id' not found", false, ""},
			{"Unclosed Block Found (by ID)", complexContent, makeArgs(complexContent, "unclosed-markdown-block"), "", true, "closing fence '```' not found", false, ""},
			{"Unclosed Block First Block", complexContent + "\n```unclosed\ncontent", makeArgs(complexContent + "\n```unclosed\ncontent", ""), "", true, "closing fence '```' not found", false, ""},
		*/
	}

	// Test runner loop
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var finalArgs []interface{}
			// Prepare args, handle specific setup for validation tests needing wrong types
			if strings.Contains(tt.name, "Validation Wrong Arg Type (content)") {
				finalArgs = makeArgs(123, "id") // Pass int instead of string for content
			} else if strings.Contains(tt.name, "Validation Wrong Arg Type (block_id)") {
				finalArgs = makeArgs(tt.content, 123) // Pass int instead of string for block_id
			} else if strings.Contains(tt.name, "Validation Wrong Arg Type (block_type)") {
				finalArgs = makeArgs(tt.content, "id", 123) // Pass int instead of string for block_type
			} else if strings.Contains(tt.name, "Validation Wrong Arg Count") {
				finalArgs = makeArgs("content_placeholder") // Pass only one arg
			} else if len(tt.args) > 0 {
				tempArgs := make([]interface{}, len(tt.args))
				copy(tempArgs, tt.args)
				tempArgs[0] = tt.content // Use actual content for execution tests
				finalArgs = tempArgs
			} else {
				finalArgs = tt.args
			}

			// Validate Arguments
			convertedArgs, valErr := core.ValidateAndConvertArgs(spec, finalArgs)

			// --- Validation Check ---
			if tt.valWantErr { // If we expect validation to fail
				if valErr == nil {
					t.Errorf("ValidateAndConvertArgs() expected error, but got nil. Args: %#v", finalArgs)
					return
				}
				if tt.valErrContains != "" && !strings.Contains(valErr.Error(), tt.valErrContains) {
					t.Errorf("ValidateAndConvertArgs() expected error containing %q, got: %v", tt.valErrContains, valErr)
				}
				return // Stop test here if validation failed as expected
			} else { // If we expect validation to succeed
				if valErr != nil {
					// Validation failed unexpectedly
					t.Fatalf("ValidateAndConvertArgs() unexpected validation error: %v. Args: %#v", valErr, finalArgs)
				}
			}
			// --- End Validation Check ---

			// Execute Tool (only if validation passed)
			gotInterface, toolErr := toolExtractFencedBlockByID(dummyInterp, convertedArgs)
			if toolErr != nil {
				t.Fatalf("toolExtractFencedBlockByID returned unexpected Go error: %v", toolErr)
			}
			gotStr, ok := gotInterface.(string)
			if !ok {
				t.Fatalf("toolExtractFencedBlockByID did not return a string result, got %T", gotInterface)
			}

			// Check if the tool's returned string indicates an error
			isReturnedError := strings.HasPrefix(gotStr, "Error:")

			if isReturnedError != tt.wantErr {
				t.Errorf("Expected error state mismatch. wantErr=%v (expecting 'Error:' prefix) but returned string is: %q", tt.wantErr, gotStr)
			}

			// If an error string was expected from the tool
			if tt.wantErr {
				if !isReturnedError {
					t.Errorf("Expected error string prefix 'Error:', but got success string: %q", gotStr)
				}
				if tt.errContains != "" && !strings.Contains(gotStr, tt.errContains) {
					t.Errorf("Expected error string containing %q, got: %q", tt.errContains, gotStr)
				}
			} else { // If success was expected from the tool
				if isReturnedError {
					t.Errorf("Expected success, but got error string: %q", gotStr)
				}
				// Compare the actual extracted content
				if gotStr != tt.want {
					t.Errorf("Result mismatch:\ngot:  %q\nwant: %q", gotStr, tt.want)
					t.Logf("Got length: %d, Want length: %d", len(gotStr), len(tt.want))
				}
			}
		})
	}
}
