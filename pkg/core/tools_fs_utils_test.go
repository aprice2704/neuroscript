// filename: pkg/core/tools_fs_utils_test.go
package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Assume newTestInterpreter and makeArgs are defined in testing_helpers.go

// --- Tests for toolLineCountFile ---
func TestToolLineCountFile(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	testDirAbs := t.TempDir()

	file1Rel := "one_line.txt"
	file3Rel := "three_lines.txt"
	file3nlRel := "three_lines_nl.txt"
	fileEmptyRel := "empty.txt"
	notFoundRel := "not_found.txt"
	outsideRel := "../invalid_path_linecount"

	// Create test files within the temp directory
	if err := os.WriteFile(filepath.Join(testDirAbs, file1Rel), []byte("Hello"), 0644); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testDirAbs, file3Rel), []byte("Line 1\nLine 2\nLine 3"), 0644); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testDirAbs, file3nlRel), []byte("Line 1\nLine 2\nLine 3\n"), 0644); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testDirAbs, fileEmptyRel), []byte(""), 0644); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// *** Corrected struct literals to use field names consistently ***
	tests := []struct {
		name           string
		inputArg       string
		wantResult     int64
		valWantErr     bool
		valErrContains string
	}{
		{name: "File Path One Line", inputArg: file1Rel, wantResult: 1, valWantErr: false, valErrContains: ""},
		{name: "File Path Multi Line", inputArg: file3Rel, wantResult: 3, valWantErr: false, valErrContains: ""},
		{name: "File Path With Trailing NL", inputArg: file3nlRel, wantResult: 3, valWantErr: false, valErrContains: ""},
		{name: "File Path Empty", inputArg: fileEmptyRel, wantResult: 0, valWantErr: false, valErrContains: ""},
		{name: "File Path Not Found", inputArg: notFoundRel, wantResult: -1, valWantErr: false, valErrContains: ""},
		{name: "File Path Invalid (Outside CWD)", inputArg: outsideRel, wantResult: -1, valWantErr: false, valErrContains: ""},
		{name: "Raw String One Line (Rejected as Path)", inputArg: "Hello", wantResult: -1, valWantErr: false, valErrContains: ""}, // toolLineCountFile expects path
		{name: "Raw String Empty (Rejected as Path)", inputArg: "", wantResult: -1, valWantErr: false, valErrContains: ""},         // Empty string fails SecureFilePath
		{name: "Validation Wrong Arg Type", inputArg: "", valWantErr: true, valErrContains: "expected string, but received type int"},
		{name: "Validation Wrong Arg Count", inputArg: "", valWantErr: true, valErrContains: "expected exactly 1 arguments"},
	}
	// *** End struct literal correction ***

	spec := ToolSpec{Name: "LineCountFile", Args: []ArgSpec{{Name: "filepath", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeInt}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get original CWD: %v", err)
	}
	t.Cleanup(func() { os.Chdir(originalWD) })

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rawArgs []interface{}
			if tt.name == "Validation Wrong Arg Count" {
				rawArgs = makeArgs()
			} else if tt.name == "Validation Wrong Arg Type" {
				rawArgs = makeArgs(123)
			} else {
				rawArgs = makeArgs(tt.inputArg)
			}

			err := os.Chdir(testDirAbs)
			if err != nil {
				t.Fatalf("Failed to Chdir to temp dir %s: %v", testDirAbs, err)
			}
			defer os.Chdir(originalWD)

			convertedArgs, valErr := ValidateAndConvertArgs(spec, rawArgs)
			if (valErr != nil) != tt.valWantErr {
				t.Errorf("Validate err=%v, wantErr %v", valErr, tt.valWantErr)
				return
			}
			if tt.valWantErr {
				if tt.valErrContains != "" && (valErr == nil || !strings.Contains(valErr.Error(), tt.valErrContains)) {
					t.Errorf("Validate expected err %q, got: %v", tt.valErrContains, valErr)
				}
				return
			}
			if valErr != nil && !tt.valWantErr {
				t.Fatalf("Validate unexpected err: %v", valErr)
			}

			gotResult, toolErr := toolLineCountFile(dummyInterp, convertedArgs)
			if toolErr != nil {
				t.Fatalf("toolLineCountFile unexpected Go err: %v", toolErr)
			}
			gotInt, ok := gotResult.(int64)
			if !ok {
				t.Fatalf("Expected int64 result, got %T (%v)", gotResult, gotResult)
			}
			if gotInt != tt.wantResult {
				t.Errorf("Result mismatch: got %d, want %d", gotInt, tt.wantResult)
			}
		})
	}
}

// --- Tests for toolSanitizeFilename ---
func TestToolSanitizeFilename(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	// *** Corrected struct literals to use field names consistently ***
	tests := []struct {
		name           string
		inputArg       string
		wantResult     string
		valWantErr     bool
		valErrContains string
	}{
		{name: "Simple Alpha", inputArg: "FileName", wantResult: "FileName", valWantErr: false, valErrContains: ""},
		{name: "With Spaces", inputArg: "File Name With Spaces", wantResult: "File_Name_With_Spaces", valWantErr: false, valErrContains: ""},
		{name: "With Slashes", inputArg: "path/to/file", wantResult: "path_to_file", valWantErr: false, valErrContains: ""},
		{name: "With Backslashes", inputArg: `win\path\file`, wantResult: "win_path_file", valWantErr: false, valErrContains: ""},
		{name: "Mixed Separators", inputArg: "a / b \\ c d", wantResult: "a_b_c_d", valWantErr: false, valErrContains: ""},
		{name: "Invalid Chars", inputArg: "File*Name?<>|:", wantResult: "FileName", valWantErr: false, valErrContains: ""},
		{name: "Leading/Trailing Dots/Underscores", inputArg: "._-_file_._-", wantResult: "file", valWantErr: false, valErrContains: ""},
		{name: "Multiple Separators", inputArg: "a___b--c..d", wantResult: "a_b-c_d", valWantErr: false, valErrContains: ""},
		{name: "Long Name", inputArg: strings.Repeat("a", 150), wantResult: strings.Repeat("a", 100), valWantErr: false, valErrContains: ""},
		{name: "Long Name with Separator", inputArg: strings.Repeat("a", 60) + "_" + strings.Repeat("b", 60), wantResult: strings.Repeat("a", 60), valWantErr: false, valErrContains: ""},
		{name: "Reserved Name CON", inputArg: "CON", wantResult: "CON_", valWantErr: false, valErrContains: ""},
		{name: "Reserved Name con.txt", inputArg: "con.txt", wantResult: "con.txt_", valWantErr: false, valErrContains: ""},
		{name: "Empty Input", inputArg: "", wantResult: "default_sanitized_name", valWantErr: false, valErrContains: ""},
		{name: "Just Dots/Slashes", inputArg: "../../../..", wantResult: "default_sanitized_name", valWantErr: false, valErrContains: ""},
		{name: "Starts with Number", inputArg: "123File", wantResult: "123File", valWantErr: false, valErrContains: ""},
		{name: "With Extension", inputArg: "report.final.docx", wantResult: "report.final.docx", valWantErr: false, valErrContains: ""},
		{name: "Ends with Dot", inputArg: "filename.", wantResult: "filename", valWantErr: false, valErrContains: ""},
		{name: "Validation Wrong Arg Type", inputArg: "", valWantErr: true, valErrContains: "expected string, but received type int"},
		{name: "Validation Wrong Arg Count", inputArg: "", valWantErr: true, valErrContains: "expected exactly 1 arguments"},
	}
	// *** End struct literal correction ***

	spec := ToolSpec{Name: "SanitizeFilename", Args: []ArgSpec{{Name: "name", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rawArgs []interface{}
			if tt.name == "Validation Wrong Arg Count" {
				rawArgs = makeArgs()
			} else if tt.name == "Validation Wrong Arg Type" {
				rawArgs = makeArgs(123)
			} else {
				rawArgs = makeArgs(tt.inputArg)
			}

			convertedArgs, valErr := ValidateAndConvertArgs(spec, rawArgs)
			if (valErr != nil) != tt.valWantErr {
				t.Errorf("Validate err=%v, wantErr %v", valErr, tt.valWantErr)
				return
			}
			if tt.valWantErr {
				if tt.valErrContains != "" && (valErr == nil || !strings.Contains(valErr.Error(), tt.valErrContains)) {
					t.Errorf("Validate expected err %q, got: %v", tt.valErrContains, valErr)
				}
				return
			}
			if valErr != nil && !tt.valWantErr {
				t.Fatalf("Validate unexpected err: %v", valErr)
			}

			gotResult, toolErr := toolSanitizeFilename(dummyInterp, convertedArgs)
			if toolErr != nil {
				t.Fatalf("toolSanitizeFilename unexpected Go err: %v", toolErr)
			}
			gotStr, ok := gotResult.(string)
			if !ok {
				t.Fatalf("Expected string result, got %T (%v)", gotResult, gotResult)
			}
			if gotStr != tt.wantResult {
				t.Errorf("Result mismatch:\ngot:  %q\nwant: %q", gotStr, tt.wantResult)
			}
		})
	}
}
