// filename: pkg/core/tools_fs_write_test.go
package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Assume newTestInterpreter and makeArgs are defined in testing_helpers.go

func TestToolWriteFile(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	testDirAbs := t.TempDir()

	newFileRel := "newfile.txt"
	subFileRel := filepath.Join("newdir", "subfile.log")
	overwriteRel := "overwrite.txt"
	outsideRel := "../cannot_write.txt"

	tests := []struct {
		name            string
		pathArg         string // Relative path to use
		contentArg      string
		wantResult      string // "OK" or error message substring
		wantResultFail  bool
		wantFileContent string // Expected content after write, if successful
		valWantErr      bool
		valErrContains  string
	}{
		{name: "Write OK New File", pathArg: newFileRel, contentArg: "Content Line 1", wantResult: "OK", wantResultFail: false, wantFileContent: "Content Line 1", valWantErr: false},
		{name: "Write OK Subdir", pathArg: subFileRel, contentArg: "Log Data", wantResult: "OK", wantResultFail: false, wantFileContent: "Log Data", valWantErr: false},
		{name: "Write OK Overwrite", pathArg: overwriteRel, contentArg: "New Content", wantResult: "OK", wantResultFail: false, wantFileContent: "New Content", valWantErr: false},
		{name: "Write Path Outside Sandbox", pathArg: outsideRel, contentArg: "data", wantResult: "path error", wantResultFail: true, valWantErr: false}, // SecureFilePath check within tool
		{name: "Validation Missing Content", pathArg: newFileRel, contentArg: "", valWantErr: true, valErrContains: "expected exactly 2 arguments", wantResultFail: true},
		{name: "Validation Wrong Path Type", pathArg: "", contentArg: "data", valWantErr: true, valErrContains: "expected string, but received type int", wantResultFail: true},
		{name: "Validation Wrong Content Type", pathArg: newFileRel, contentArg: "", valWantErr: true, valErrContains: "expected string, but received type bool", wantResultFail: true},
	}

	// Setup for overwrite test
	overwritePathAbs := filepath.Join(testDirAbs, overwriteRel)
	if err := os.WriteFile(overwritePathAbs, []byte("Old Content"), 0644); err != nil {
		t.Fatalf("Setup failed writing overwrite file: %v", err)
	}

	spec := ToolSpec{Name: "WriteFile", Args: []ArgSpec{
		{Name: "filepath", Type: ArgTypeString, Required: true},
		{Name: "content", Type: ArgTypeString, Required: true},
	}, ReturnType: ArgTypeString}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get original CWD: %v", err)
	}
	t.Cleanup(func() { os.Chdir(originalWD) })

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rawArgs []interface{}
			if tt.name == "Validation Missing Content" {
				rawArgs = makeArgs(tt.pathArg)
			} else if tt.name == "Validation Wrong Path Type" {
				rawArgs = makeArgs(123, tt.contentArg)
			} else if tt.name == "Validation Wrong Content Type" {
				rawArgs = makeArgs(tt.pathArg, true)
			} else {
				rawArgs = makeArgs(tt.pathArg, tt.contentArg)
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
			gotResult, toolErr := toolWriteFile(dummyInterp, convertedArgs)
			if toolErr != nil {
				t.Fatalf("toolWriteFile unexpected Go error = %v", toolErr)
			}
			gotStr, ok := gotResult.(string)
			if !ok {
				t.Fatalf("Expected string result, got %T (%v)", gotResult, gotResult)
			}

			// Check result string ("OK" or error message)
			if tt.wantResultFail {
				if !strings.Contains(gotStr, tt.wantResult) {
					t.Errorf("Expected result error string containing %q, got %q", tt.wantResult, gotStr)
				}
			} else { // Expect "OK"
				if gotStr != "OK" {
					t.Errorf("Expected result 'OK', got %q", gotStr)
				}
				// Verify file content only on success
				// Construct absolute path for reading back the file, relative to where Chdir placed us
				fullPathWritten := filepath.Join(testDirAbs, tt.pathArg) // Use absolute base
				fileContentBytes, readErr := os.ReadFile(fullPathWritten)
				if readErr != nil {
					t.Errorf("Failed to read back written file %s: %v", fullPathWritten, readErr)
				} else if string(fileContentBytes) != tt.wantFileContent {
					t.Errorf("Written file content mismatch for %s:\ngot:  %q\nwant: %q", tt.pathArg, string(fileContentBytes), tt.wantFileContent)
				}
			}
		})
	}
}
