// pkg/core/tools_composite_doc_test.go
package core

import (
	"os"            // Added for file reading
	"path/filepath" // Added for file paths
	"reflect"
	"strings"
	"testing"
)

// --- Test ExtractFencedBlock ---
func TestToolExtractFencedBlock(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	spec := ToolSpec{
		Name: "ExtractFencedBlock",
		Args: []ArgSpec{
			{Name: "content", Type: ArgTypeString, Required: true},
			{Name: "block_id", Type: ArgTypeString, Required: true},
			{Name: "block_type", Type: ArgTypeString, Required: false}, // Optional
		},
		ReturnType: ArgTypeString,
	}

	// --- Load Fixture Files ---
	fixtureDir := "test_fixtures" // Assuming fixtures are in this subdirectory

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

	tests := []struct {
		name           string
		content        string        // Content loaded from fixture
		args           []interface{} // Will be adjusted to use loaded content
		want           string
		wantErr        bool // Expects error string prefix "Error:"
		errContains    string
		valWantErr     bool // Expect validation error
		valErrContains string
	}{
		// --- SUCCESS CASES using simple_blocks.md ---
		{"Simple NS Block", simpleContent, makeArgs(simpleContent, "simple-ns-block"), "DEFINE PROCEDURE Simple()\n  EMIT \"Hello from simple NS\"\nEND", false, "", false, ""},
		{"Simple PY Block", simpleContent, makeArgs(simpleContent, "simple-py-block"), `print("Hello from simple Python")`, false, "", false, ""},
		{"Simple Empty Block", simpleContent, makeArgs(simpleContent, "simple-empty-block"), "", false, "", false, ""},
		{"Simple Comment Block", simpleContent, makeArgs(simpleContent, "simple-comment-block"), "# This is a comment inside.\n-- So is this.\n\n# Even with a blank line.", false, "", false, ""},
		{"Simple NS with type match", simpleContent, makeArgs(simpleContent, "simple-ns-block", "neuroscript"), "DEFINE PROCEDURE Simple()\n  EMIT \"Hello from simple NS\"\nEND", false, "", false, ""},

		// --- SUCCESS CASES using complex_blocks.md ---
		{"Complex NS Block 1", complexContent, makeArgs(complexContent, "complex-ns-1"), "CALL TOOL.DoSomething()", false, "", false, ""},
		{"Complex PY Adjacent Block", complexContent, makeArgs(complexContent, "complex-py-adjacent"), "import os", false, "", false, ""},
		{"Metadata Only Block", complexContent, makeArgs(complexContent, "metadata-only-block"), "", false, "", false, ""}, // Expect empty content
		{"Checklist Hyphen Meta", complexContent, makeArgs(complexContent, "checklist-hyphen-meta"), "- [x] Item A\n- [ ] Item B", false, "", false, ""},

		// --- ERROR CASES using complex_blocks.md ---
		{"Block with No ID", complexContent, makeArgs(complexContent, "block-with-no-id"), "", true, "Block ID 'block-with-no-id' not found", false, ""},                // Need to target a non-existent ID
		{"Go Block Late ID (ID Not Found)", complexContent, makeArgs(complexContent, "go-block-late-id"), "", true, "Block ID 'go-block-late-id' not found", false, ""}, // ID comes after content, should not be found by current logic
		{"Unclosed Block", complexContent, makeArgs(complexContent, "unclosed-markdown-block"), "", true, "closing fence '```' not found", false, ""},

		// --- ERROR CASES using simple_blocks.md (for type mismatch etc.) ---
		{"Error ID not found (simple)", simpleContent, makeArgs(simpleContent, "nonexistent-id"), "", true, "Block ID 'nonexistent-id' not found", false, ""},
		{"Error type mismatch (simple)", simpleContent, makeArgs(simpleContent, "simple-ns-block", "python"), "", true, "type mismatch: expected 'python', got 'neuroscript'", false, ""},

		// --- VALIDATION ERROR CASES (Content doesn't matter here) ---
		{"Validation Wrong Arg Count (1)", simpleContent, makeArgs("content"), "", false, "", true, "tool 'ExtractFencedBlock' expected at least 2 arguments, but received 1"},
		{"Validation Wrong Arg Type (content)", simpleContent, makeArgs(123, "id"), "", false, "", true, "argument 'content' (index 0): expected string"}, // Still expect this error message from validation spec
		{"Validation Wrong Arg Type (block_id)", simpleContent, makeArgs("content", 123), "", false, "", true, "argument 'block_id' (index 1): expected string"},
		{"Validation Wrong Arg Type (block_type)", simpleContent, makeArgs("content", "id", 123), "", false, "", true, "argument 'block_type' (index 2): expected string"},
		// These validation OK tests don't really exercise the logic anymore, but keep for arg count check
		{"Validation OK with 2 args", simpleContent, makeArgs(simpleContent, "simple-ns-block"), "DEFINE PROCEDURE Simple()\n  EMIT \"Hello from simple NS\"\nEND", false, "", false, ""},
		{"Validation OK with 3 args", simpleContent, makeArgs(simpleContent, "simple-ns-block", "neuroscript"), "DEFINE PROCEDURE Simple()\n  EMIT \"Hello from simple NS\"\nEND", false, "", false, ""},
	}

	// Test runner loop
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// --- Argument Setup for Validation/Execution ---
			var finalArgs []interface{}
			// Use the specific content loaded for this test, adjusting args slice
			if len(tt.args) > 0 {
				tempArgs := make([]interface{}, len(tt.args))
				copy(tempArgs, tt.args)
				// Replace placeholder content arg with the actual loaded content for this test
				tempArgs[0] = tt.content
				finalArgs = tempArgs
			} else {
				finalArgs = tt.args // Handle cases like wrong arg count where tt.args might be empty/nil
			}

			// --- Validation Check ---
			convertedArgs, valErr := ValidateAndConvertArgs(spec, finalArgs)
			if (valErr != nil) != tt.valWantErr {
				t.Errorf("ValidateAndConvertArgs() error = %v, valWantErr %v", valErr, tt.valWantErr)
				// Log args passed to validation if it fails unexpectedly
				if valErr == nil && tt.valWantErr {
					t.Logf("Validation failed: Args passed = %#v", finalArgs)
				}
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

			// --- Tool Execution Check (only if validation passed) ---
			if !tt.valWantErr {
				gotInterface, toolErr := toolExtractFencedBlock(dummyInterp, convertedArgs)
				if toolErr != nil {
					t.Fatalf("toolExtractFencedBlock returned unexpected Go error: %v", toolErr)
				}

				gotStr, ok := gotInterface.(string)
				if !ok {
					t.Fatalf("toolExtractFencedBlock did not return a string result, got %T", gotInterface)
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
					// Compare successful result content, trimming whitespace for comparison robustness against fixture nuances
					// Update: Don't trim, compare raw based on previous findings
					// gotTrimmed := strings.TrimSpace(gotStr)
					// wantTrimmed := strings.TrimSpace(tt.want)
					// if gotTrimmed != wantTrimmed {
					if gotStr != tt.want {
						t.Errorf("Result mismatch:\ngot:  %q\nwant: %q", gotStr, tt.want)
					}
				}
			}
		})
	}
}

// --- Test ParseChecklist (remains the same, no fixture needed yet) ---
func TestToolParseChecklist(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	spec := ToolSpec{
		Name:       "ParseChecklist",
		Args:       []ArgSpec{{Name: "content", Type: ArgTypeString, Required: true}},
		ReturnType: ArgTypeSliceAny,
	}

	tests := []struct {
		name           string
		contentArg     string
		want           []interface{}
		wantErr        bool
		errContains    string
		valWantErr     bool
		valErrContains string
	}{
		{
			name:       "Simple checklist",
			contentArg: "- [ ] Task 1\n- [x] Task 2\n  - [ ] Indented ignored\n# Comment ignored\n- [X] Task 3 CAPS X",
			want: []interface{}{
				map[string]interface{}{"text": "Task 1", "status": "pending"},
				map[string]interface{}{"text": "Task 2", "status": "done"},
				map[string]interface{}{"text": "Task 3 CAPS X", "status": "done"},
			},
			wantErr: false, errContains: "", valWantErr: false, valErrContains: "",
		},
		{"Empty content", "", []interface{}{}, false, "", false, ""},
		{"No valid items", "# Just a comment\nSome other text", []interface{}{}, false, "", false, ""},
		{
			name:       "Spaces around markers",
			contentArg: " - [ ]   Spaces around \n- [x] No spaces",
			want: []interface{}{
				map[string]interface{}{"text": "No spaces", "status": "done"},
			},
			wantErr: false, errContains: "", valWantErr: false, valErrContains: "",
		},
		{"Invalid format ignored", "- [] Invalid format\n- [x] Valid item", []interface{}{
			map[string]interface{}{"text": "Valid item", "status": "done"},
		}, false, "", false, ""},
		{"Validation Wrong Arg Type", "", nil, false, "", true, "expected string, but received type int"},
		{"Validation Wrong Arg Count", "", nil, false, "", true, "tool 'ParseChecklist' expected exactly 1 arguments, but received 0"},
	}

	// Test runner loop remains the same...
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rawArgs []interface{}
			if tt.name == "Validation Wrong Arg Count" {
				rawArgs = makeArgs()
			} else if tt.name == "Validation Wrong Arg Type" {
				rawArgs = makeArgs(123)
			} else {
				rawArgs = makeArgs(tt.contentArg)
			}

			convertedArgs, valErr := ValidateAndConvertArgs(spec, rawArgs)
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
			if valErr != nil && !tt.valWantErr {
				t.Fatalf("ValidateAndConvertArgs() unexpected validation error: %v", valErr)
			}

			if !tt.valWantErr {
				gotInterface, toolErr := toolParseChecklist(dummyInterp, convertedArgs)
				if toolErr != nil {
					if !tt.wantErr {
						t.Fatalf("toolParseChecklist returned unexpected Go error: %v", toolErr)
					}
					if tt.wantErr && tt.errContains != "" && !strings.Contains(toolErr.Error(), tt.errContains) {
						t.Errorf("toolParseChecklist Go error mismatch. Expected contains %q, got: %v", tt.errContains, toolErr)
					}
					return
				}
				if errStr, isErrStr := gotInterface.(string); isErrStr {
					if !tt.wantErr {
						t.Errorf("Expected success, but got error string: %q", errStr)
					} else if tt.errContains != "" && !strings.Contains(errStr, tt.errContains) {
						t.Errorf("Expected error string containing %q, got: %q", tt.errContains, errStr)
					}
					return
				}
				if tt.wantErr {
					t.Errorf("Expected error, but got successful-looking result type %T: %v", gotInterface, gotInterface)
					return
				}
				gotList, ok := gotInterface.([]interface{})
				if !ok {
					t.Fatalf("toolParseChecklist did not return []interface{}, got %T", gotInterface)
				}
				if !reflect.DeepEqual(gotList, tt.want) {
					t.Errorf("Result mismatch:\ngot:  %#v\nwant: %#v", gotList, tt.want)
				}
			}
		})
	}
}

// Helper funcs (newTestInterpreter, makeArgs) remain the same (assume they are in testing_helpers.go)
