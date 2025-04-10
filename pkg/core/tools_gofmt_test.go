// pkg/core/tools_gofmt_test.go
package core

import (
	// "os" // No longer needed for file ops here
	// "path/filepath" // No longer needed here
	"strings"
	"testing"
	// "os/exec" // No longer needed
)

// Assume newTestInterpreter and makeArgs helpers are defined or copied here
// func newDefaultTestInterpreter() *Interpreter { return NewInterpreter(nil) }
// func makeArgs(vals ...interface{}) []interface{} { return vals }

func TestToolGoFmt(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()

	// Define content directly
	needsFormattingContent := "package main\nfunc main () {println(\"hello\")}"
	alreadyFormattedContent := "package main\n\nfunc main() {\n\tprintln(\"ok\")\n}\n"
	parseErrorContent := "package main\nfunc main() { println(\"hello\n}"
	formattedNeedsFormatting := "package main\n\nfunc main() { println(\"hello\") }\n"

	tests := []struct {
		name                 string
		inputContent         string // Use content directly
		wantFormattedContent string // Expected content *in the result map*
		wantErrorSubstring   string // Expected error string *in the result map*
		wantSuccess          bool   // Expected success field in the result map
		valWantErr           bool   // Expect validation error?
		valErrContains       string // Validation error substring
	}{
		{name: "Needs Formatting", inputContent: needsFormattingContent, wantFormattedContent: formattedNeedsFormatting, wantSuccess: true, valWantErr: false},
		{name: "Already Formatted", inputContent: alreadyFormattedContent, wantFormattedContent: alreadyFormattedContent, wantSuccess: true, valWantErr: false},
		{name: "Parse Error", inputContent: parseErrorContent, wantFormattedContent: parseErrorContent, wantErrorSubstring: "string literal not terminated", wantSuccess: false, valWantErr: false}, // Expect original content on error
		// Validation tests
		{name: "Validation Wrong Arg Count", inputContent: "", valWantErr: true, valErrContains: "tool 'GoFmt' expected exactly 1 arguments, but received 0"},
		// *** UPDATED Expected Error String ***
		{name: "Validation Wrong Arg Type", inputContent: "", valWantErr: true, valErrContains: "type validation failed for argument 'content' of tool 'GoFmt': expected string, got int"},
		// Removed Path Error test as tool no longer takes path
	}

	spec := ToolSpec{Name: "GoFmt", Args: []ArgSpec{{Name: "content", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeAny}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// --- Test setup: Prepare args ---
			var rawArgs []interface{}
			if tt.name == "Validation Wrong Arg Count" {
				rawArgs = makeArgs()
			} else if tt.name == "Validation Wrong Arg Type" {
				rawArgs = makeArgs(123) // Pass wrong type (int)
			} else {
				rawArgs = makeArgs(tt.inputContent) // Pass content string
			}

			// --- Validation Step ---
			convertedArgs, valErr := ValidateAndConvertArgs(spec, rawArgs) // Use convertedArgs
			if (valErr != nil) != tt.valWantErr {
				t.Errorf("ValidateAndConvertArgs() error = %v, valWantErr %v", valErr, tt.valWantErr)
				return
			}
			if tt.valWantErr {
				if tt.valErrContains != "" && (valErr == nil || !strings.Contains(valErr.Error(), tt.valErrContains)) {
					t.Errorf("ValidateAndConvertArgs() expected error containing %q, got: %v", tt.valErrContains, valErr)
				}
				return // Stop if validation failed as expected
			}
			if valErr != nil && !tt.valWantErr {
				t.Fatalf("ValidateAndConvertArgs() unexpected validation error: %v", valErr)
			}

			// --- Tool Execution Step ---
			gotResult, toolErr := toolGoFmt(dummyInterp, convertedArgs) // Use convertedArgs
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

			// Check content and error message based on expected success
			if tt.wantSuccess {
				if gotContent != tt.wantFormattedContent {
					// Use %q for quoting to make whitespace differences obvious
					t.Errorf("Formatted content mismatch:\ngot:  %q\nwant: %q", gotContent, tt.wantFormattedContent)
				}
				if gotError != "" {
					t.Errorf("Expected empty error string on success, got: %q", gotError)
				}
			} else { // Expected failure (e.g., parse error)
				if tt.wantErrorSubstring != "" {
					if !strings.Contains(gotError, tt.wantErrorSubstring) {
						t.Errorf("Error message mismatch: got %q, want contains %q", gotError, tt.wantErrorSubstring)
					}
				} else if gotError == "" {
					t.Errorf("Expected non-empty error string on failure, got empty")
				}
				// Check if original content was returned on error
				if gotContent != tt.inputContent {
					t.Errorf("Expected original content on format error, got:\n%q\nOriginal:\n%q", gotContent, tt.inputContent)
				}
			}
		})
	}
}
