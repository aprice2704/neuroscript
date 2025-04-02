// pkg/core/tools_gofmt_test.go
package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	// "os/exec" // No longer needed
)

// Assume newDummyInterpreter and makeArgs helpers are defined or copied here

func TestToolGoFmt(t *testing.T) {
	// No longer need to check for 'gofmt' command

	dummyInterp := newDummyInterpreter()
	// Use CWD for test files now, with cleanup
	testDir := "gofmt_test_files_final" // Use different name
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test dir %s: %v", testDir, err)
	}
	t.Cleanup(func() { os.RemoveAll(testDir) })

	tests := []struct {
		name                 string
		initialContent       string
		wantFormattedContent string
		wantErrorSubstring   string
		wantSuccess          bool
		wantErr              bool
		errContains          string
	}{
		{name: "Needs Formatting", initialContent: "package main\nfunc main () {println(\"hello\")}", wantFormattedContent: "package main\n\nfunc main() { println(\"hello\") }\n", wantSuccess: true, wantErr: false},
		{name: "Already Formatted", initialContent: "package main\n\nfunc main() {\n\tprintln(\"ok\")\n}\n", wantFormattedContent: "package main\n\nfunc main() {\n\tprintln(\"ok\")\n}\n", wantSuccess: true, wantErr: false},
		// Updated wantErrorSubstring for go/format parse error
		{name: "Parse Error", initialContent: "package main\nfunc main() { println(\"hello\n}", wantFormattedContent: "package main\nfunc main() { println(\"hello\n}", wantErrorSubstring: "string literal not terminated", wantSuccess: false, wantErr: false},
		{name: "Validation Wrong Arg Count", wantErr: true, errContains: "tool 'GoFmt' expected exactly 1 arguments, but received 0"},
		{name: "Validation Wrong Arg Type", wantErr: true, errContains: "tool 'GoFmt' argument 'filepath' (index 0): expected string, but received type int"},
		// Path error test uses relative path now
		{name: "Path Error (secureFilePath relative)", initialContent: "package main", wantFormattedContent: "", wantSuccess: false, wantErr: false, errContains: "outside the allowed directory"},
	}

	spec := ToolSpec{Name: "GoFmt", Args: []ArgSpec{{Name: "filepath", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeAny}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// --- Test setup: Use relative paths from CWD ---
			var rawArgs []interface{}
			relativeFilePath := "" // Path passed to tool

			if tt.name == "Validation Wrong Arg Count" {
				rawArgs = makeArgs()
			} else if tt.name == "Validation Wrong Arg Type" {
				rawArgs = makeArgs(123)
			} else if tt.name == "Path Error (secureFilePath relative)" {
				relativeFilePath = "../invalid_gofmt_path.go" // Invalid RELATIVE path
				rawArgs = makeArgs(relativeFilePath)
			} else {
				// Prepare file relative to CWD
				relativeFilePath = filepath.Join(testDir, "test.go") // Relative path
				err := os.WriteFile(relativeFilePath, []byte(tt.initialContent), 0644)
				if err != nil {
					t.Fatalf("Failed write: %v", err)
				}
				// Pass the RELATIVE path to the tool
				rawArgs = makeArgs(relativeFilePath)
			}

			// --- Validation Step ---
			_, valErr := ValidateAndConvertArgs(spec, rawArgs)
			if (valErr != nil) != tt.wantErr {
				t.Errorf("ValidateAndConvertArgs() error = %v, wantErr %v", valErr, tt.wantErr)
				return
			}
			if tt.wantErr {
				if tt.errContains != "" && (valErr == nil || !strings.Contains(valErr.Error(), tt.errContains)) {
					t.Errorf("ValidateAndConvertArgs() expected error containing %q, got: %v", tt.errContains, valErr)
				}
				return
			}

			// --- Tool Execution Step ---
			gotResult, toolErr := toolGoFmt(dummyInterp, rawArgs)
			if toolErr != nil {
				t.Fatalf("toolGoFmt() returned unexpected Go error: %v", toolErr)
			}
			gotMap, ok := gotResult.(map[string]interface{})
			if !ok {
				t.Fatalf("toolGoFmt() did not return a map, got %T", gotResult)
			}

			// --- Assertions ---
			gotSuccess, _ := gotMap["success"].(bool)
			gotContent, _ := gotMap["formatted_content"].(string)
			gotError, _ := gotMap["error"].(string)
			if gotSuccess != tt.wantSuccess {
				t.Errorf("Success mismatch: got %v, want %v", gotSuccess, tt.wantSuccess)
			}
			if tt.wantSuccess {
				if gotContent != tt.wantFormattedContent {
					t.Errorf("Formatted content mismatch:\ngot:\n%s\nwant:\n%s", gotContent, tt.wantFormattedContent)
				}
				if gotError != "" {
					t.Errorf("Expected empty error string on success, got: %q", gotError)
				}
			} else {
				expectedErrSubstr := tt.wantErrorSubstring
				if tt.name == "Path Error (secureFilePath relative)" {
					expectedErrSubstr = tt.errContains
				}
				if expectedErrSubstr != "" {
					if !strings.Contains(gotError, expectedErrSubstr) {
						t.Errorf("Error message mismatch: got %q, want contains %q", gotError, expectedErrSubstr)
					}
				} else if gotError == "" {
					t.Errorf("Expected non-empty error string on failure, got empty")
				}
				if tt.initialContent != "" && strings.Contains(tt.name, "Parse Error") {
					if gotContent != tt.initialContent {
						t.Errorf("Expected original content on parse error, got:\n%s\nOriginal:\n%s", gotContent, tt.initialContent)
					}
				}
			}
		})
	}
}

// --- Helper Functions ---
// Removed
