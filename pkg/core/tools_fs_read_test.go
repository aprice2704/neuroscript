// filename: pkg/core/tools_fs_read_test.go
package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Assume newTestInterpreter and makeArgs are defined in testing_helpers.go

func TestToolReadFile(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	testDirAbs := t.TempDir()

	content1 := "Hello Reader!"
	content2 := "Line1\nLine2"
	file1Rel := "read_test1.txt"
	file2Rel := filepath.Join("subdir", "read_test2.txt")
	notFoundRel := "not_exists.txt"
	outsideRel := "../secrets.txt"

	subDirAbs := filepath.Join(testDirAbs, "subdir")
	if err := os.MkdirAll(subDirAbs, 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testDirAbs, file1Rel), []byte(content1), 0644); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDirAbs, "read_test2.txt"), []byte(content2), 0644); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	tests := []struct {
		name           string
		inputArg       string // Path *relative* to testDirAbs
		wantResult     string
		wantToolError  bool
		wantResultFail bool
		valWantErr     bool
		valErrContains string
	}{
		{name: "Read OK Simple", inputArg: file1Rel, wantResult: content1, wantResultFail: false, valWantErr: false},
		{name: "Read OK Subdir", inputArg: file2Rel, wantResult: content2, wantResultFail: false, valWantErr: false},
		{name: "Read File Not Found", inputArg: notFoundRel, wantResult: "File not found", wantResultFail: true, valWantErr: false},
		{name: "Read Path Outside", inputArg: outsideRel, wantResult: "path error", wantResultFail: true, valWantErr: false}, // Expect error string from SecureFilePath
		// Validation Errors
		// *** UPDATED Expected Error String ***
		{name: "Validation Wrong Arg Type", inputArg: "", valWantErr: true, valErrContains: "type validation failed for argument 'filepath' of tool 'ReadFile': expected string, got int"},
		{name: "Validation Wrong Arg Count", inputArg: "", valWantErr: true, valErrContains: "expected exactly 1 arguments"},
	}

	spec := ToolSpec{Name: "ReadFile", Args: []ArgSpec{{Name: "filepath", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}

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
				rawArgs = makeArgs(123) // Pass int instead of string
			} else {
				rawArgs = makeArgs(tt.inputArg)
			} // Pass relative path

			// --- Change CWD for execution context ---
			err := os.Chdir(testDirAbs)
			if err != nil {
				t.Fatalf("Failed to Chdir to temp dir %s: %v", testDirAbs, err)
			}
			defer os.Chdir(originalWD)
			// --- End CWD Change ---

			// Validation
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

			// Execution
			gotResult, toolErr := toolReadFile(dummyInterp, convertedArgs)
			if (toolErr != nil) != tt.wantToolError {
				t.Fatalf("toolReadFile Go error = %v, wantToolError %v", toolErr, tt.wantToolError)
			}
			gotStr, ok := gotResult.(string)
			if !ok {
				t.Fatalf("Expected string result, got %T (%v)", gotResult, gotResult)
			}

			// Check result content
			if tt.wantResultFail {
				if !strings.Contains(gotStr, tt.wantResult) {
					t.Errorf("Expected result error string containing %q, got %q", tt.wantResult, gotStr)
				}
			} else {
				if gotStr != tt.wantResult {
					t.Errorf("Result mismatch:\ngot:  %q\nwant: %q", gotStr, tt.wantResult)
				}
			}
		})
	}
}
