// pkg/core/tools_composite_doc_test.go
package core

import (
	"reflect"
	"strings"
	"testing"
)

// --- Test ExtractFencedBlock ---
func TestToolExtractFencedBlock(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	// ** FIX: Added Required: true to content and block_id args **
	spec := ToolSpec{
		Name: "ExtractFencedBlock",
		Args: []ArgSpec{
			{Name: "content", Type: ArgTypeString, Required: true},
			{Name: "block_id", Type: ArgTypeString, Required: true},
			{Name: "block_type", Type: ArgTypeString, Required: false}, // Optional
		},
		ReturnType: ArgTypeString,
	}

	// Test content definitions remain the same...
	testContentBase := `
Some text before.

` + "```" + `neuroscript
# id: block-1
# version: 1.0
DEFINE PROCEDURE Test1()
END
` + "```" + `

More text.

` + "```" + `neurodata-checklist
# version: 0.1
# id: checklist-abc

- [ ] Item 1
- [x] Item 2
` + "```" + `

` + "```" + `python
-- id: py-block
-- version: 0.1
print("hello")
` + "```" + `

` + "```" + `neuroscript
# id: block-no-close
DEFINE PROCEDURE Test2()
` // Missing closing fence

	testContentWithEmpty := testContentBase + "\n```text\n# id: empty-block\n```\n"
	testContentWithComment := testContentWithEmpty + "\n```text\n# id: comment-block\n# A comment\n  \n-- Another\n```\n"
	testContentForValidation := "\n```text\n# id: id\nValid Content\n```\n"

	tests := []struct {
		name           string
		args           []interface{}
		want           string
		wantErr        bool
		errContains    string
		valWantErr     bool
		valErrContains string
	}{
		// --- SUCCESS CASES ---
		{"Extract neuroscript ok", makeArgs(testContentWithComment, "block-1"), "DEFINE PROCEDURE Test1()\nEND", false, "", false, ""},
		{"Extract checklist ok", makeArgs(testContentWithComment, "checklist-abc"), "- [ ] Item 1\n- [x] Item 2", false, "", false, ""},
		{"Extract python ok (using -- id)", makeArgs(testContentWithComment, "py-block"), `print("hello")`, false, "", false, ""},
		{"Extract empty block", makeArgs(testContentWithComment, "empty-block"), "", false, "", false, ""},
		{"Extract comment-only block", makeArgs(testContentWithComment, "comment-block"), "# A comment\n  \n-- Another", false, "", false, ""},
		{"Extract neuroscript with type match", makeArgs(testContentWithComment, "block-1", "neuroscript"), "DEFINE PROCEDURE Test1()\nEND", false, "", false, ""},
		{"Extract python with type match", makeArgs(testContentWithComment, "py-block", "python"), `print("hello")`, false, "", false, ""},

		// --- ERROR CASES ---
		{"Error ID not found", makeArgs(testContentWithComment, "nonexistent-id"), "", true, "Block ID 'nonexistent-id' not found", false, ""},
		{"Error type mismatch", makeArgs(testContentWithComment, "block-1", "python"), "", true, "type mismatch: expected 'python', got 'neuroscript'", false, ""},
		{"Error no closing fence", makeArgs(testContentWithComment, "block-no-close"), "", true, "closing fence '```' not found", false, ""}, // Test for the EOF fix

		// --- VALIDATION ERROR CASES ---
		// ** FIX: Corrected expected error message based on Required: true **
		{"Validation Wrong Arg Count (1)", makeArgs("content"), "", false, "", true, "tool 'ExtractFencedBlock' expected at least 2 arguments, but received 1"},
		{"Validation Wrong Arg Type (content)", makeArgs(123, "id"), "", false, "", true, "argument 'content' (index 0): expected string"},
		{"Validation Wrong Arg Type (block_id)", makeArgs("content", 123), "", false, "", true, "argument 'block_id' (index 1): expected string"},
		{"Validation Wrong Arg Type (block_type)", makeArgs("content", "id", 123), "", false, "", true, "argument 'block_type' (index 2): expected string"},
		{"Validation OK with 2 args", makeArgs(testContentForValidation, "id"), "Valid Content", false, "", false, ""},
		{"Validation OK with 3 args", makeArgs(testContentForValidation, "id", "text"), "Valid Content", false, "", false, ""},
	}

	// Test runner loop remains the same...
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			convertedArgs, valErr := ValidateAndConvertArgs(spec, tt.args)

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
				t.Fatalf("ValidateAndConvertArgs() unexpected validation error: %v. Args: %v", valErr, tt.args)
			}

			if !tt.valWantErr {
				gotInterface, toolErr := toolExtractFencedBlock(dummyInterp, convertedArgs)
				if toolErr != nil {
					t.Fatalf("toolExtractFencedBlock returned unexpected Go error: %v", toolErr)
				}
				gotStr, ok := gotInterface.(string)
				if !ok {
					t.Fatalf("toolExtractFencedBlock did not return a string, got %T", gotInterface)
				}

				if tt.wantErr {
					if !strings.HasPrefix(gotStr, "Error:") {
						t.Errorf("Expected error string prefix 'Error:', got: %q", gotStr)
					}
					if tt.errContains != "" && !strings.Contains(gotStr, tt.errContains) {
						t.Errorf("Expected error string containing %q, got: %q", tt.errContains, gotStr)
					}
				} else {
					if strings.HasPrefix(gotStr, "Error:") {
						t.Errorf("Expected success, but got error string: %q", gotStr)
					}
					if !reflect.DeepEqual(gotStr, tt.want) {
						t.Errorf("Result mismatch:\ngot:  %q\nwant: %q", gotStr, tt.want)
					}
				}
			}
		})
	}
}

// --- Test ParseChecklist ---
func TestToolParseChecklist(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	// ** FIX: Added Required: true **
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
			// ** FIX: Corrected expected output based on stricter regex **
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
			contentArg: " - [ ]   Spaces around \n- [x] No spaces", // First line has leading space
			// ** FIX: Corrected expected output based on stricter regex **
			want: []interface{}{
				map[string]interface{}{"text": "No spaces", "status": "done"}, // Only second line should match
			},
			wantErr: false, errContains: "", valWantErr: false, valErrContains: "",
		},
		{"Invalid format ignored", "- [] Invalid format\n- [x] Valid item", []interface{}{
			map[string]interface{}{"text": "Valid item", "status": "done"},
		}, false, "", false, ""},
		{"Validation Wrong Arg Type", "", nil, false, "", true, "expected string, but received type int"},
		// ** FIX: Corrected expected error message based on Required: true **
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
					t.Fatalf("toolParseChecklist returned unexpected Go error: %v", toolErr)
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
					t.Errorf("Expected error string, but got type %T: %v", gotInterface, gotInterface)
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
