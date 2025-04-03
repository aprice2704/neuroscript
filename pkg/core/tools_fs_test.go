// pkg/core/tools_fs_test.go
package core

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	// Needed for LineCount test file creation
)

// --- Tests for toolListDirectory --- (Existing tests remain the same)
func TestToolListDirectory(t *testing.T) {
	dummyInterp := newDummyInterpreter()
	testBaseDir := "list_dir_test_files_temp"
	err := os.MkdirAll(testBaseDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create base test dir %s: %v", testBaseDir, err)
	}
	t.Cleanup(func() { os.RemoveAll(testBaseDir) })
	subDir := filepath.Join(testBaseDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	err = os.WriteFile(filepath.Join(testBaseDir, "file1.txt"), []byte("abc"), 0644)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	err = os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("def"), 0644)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	err = os.WriteFile(filepath.Join(testBaseDir, ".hiddenfile"), []byte("ghi"), 0644)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	tests := []struct {
		name          string
		pathArg       string
		wantResult    []string
		wantErrorMsg  bool
		errorContains string
		valWantErr    bool
	}{
		{name: "List Base Test Dir", pathArg: testBaseDir, wantResult: []string{".hiddenfile", "file1.txt", "subdir/"}, wantErrorMsg: false, valWantErr: false},
		{name: "List Subdir", pathArg: filepath.Join(testBaseDir, "subdir"), wantResult: []string{"file2.txt"}, wantErrorMsg: false, valWantErr: false},
		{name: "List CWD (.)", pathArg: ".", wantResult: nil, wantErrorMsg: false, valWantErr: false},
		{name: "Path Not Found", pathArg: filepath.Join(testBaseDir, "not_a_dir"), wantErrorMsg: true, errorContains: "ListDirectory read error", valWantErr: false},
		{name: "Path Is A File", pathArg: filepath.Join(testBaseDir, "file1.txt"), wantErrorMsg: true, errorContains: "not a directory", valWantErr: false},
		{name: "Path Outside CWD (Relative)", pathArg: "../some_other_dir_list", wantErrorMsg: true, errorContains: "outside the allowed directory", valWantErr: false},
		{name: "Path Resolves to Parent (Outside core)", pathArg: "..", wantErrorMsg: true, errorContains: "outside the allowed directory", valWantErr: false}, // Corrected expectation
		{name: "Validation Wrong Arg Count", pathArg: "", wantErrorMsg: false, valWantErr: true, errorContains: "tool 'ListDirectory' expected exactly 1 arguments, but received 0"},
		{name: "Validation Wrong Arg Type", pathArg: "", wantErrorMsg: false, valWantErr: true, errorContains: "tool 'ListDirectory' argument 'path' (index 0): expected string, but received type int"},
	}
	spec := ToolSpec{Name: "ListDirectory", Args: []ArgSpec{{Name: "path", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeSliceString}
	originalWD, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(originalWD) }) // Restore WD
	t.Logf("Running ListDirectory tests from CWD: %s", originalWD)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rawArgs []interface{}
			if tt.name == "Validation Wrong Arg Count" {
				rawArgs = makeArgs()
			} else if tt.name == "Validation Wrong Arg Type" {
				rawArgs = makeArgs(123)
			} else {
				rawArgs = makeArgs(tt.pathArg)
			}
			_, valErr := ValidateAndConvertArgs(spec, rawArgs)
			if (valErr != nil) != tt.valWantErr {
				t.Errorf("Validate err=%v, wantErr %v", valErr, tt.valWantErr)
				return
			}
			if tt.valWantErr {
				if tt.errorContains != "" && (valErr == nil || !strings.Contains(valErr.Error(), tt.errorContains)) {
					t.Errorf("Validate expected err %q, got: %v", tt.errorContains, valErr)
				}
				return
			}
			if valErr != nil && !tt.valWantErr {
				t.Fatalf("Validate unexpected err: %v", valErr)
			}
			gotResult, toolErr := toolListDirectory(dummyInterp, rawArgs)
			if toolErr != nil {
				t.Fatalf("toolListDirectory unexpected Go err: %v", toolErr)
			}
			gotList, isList := gotResult.([]string)
			gotStr, isStr := gotResult.(string)
			if tt.wantErrorMsg {
				if !isStr {
					t.Errorf("Expected err string, got %T (%v)", gotResult, gotResult)
					return
				}
				if tt.errorContains != "" && !strings.Contains(gotStr, tt.errorContains) {
					t.Errorf("Result err mismatch: got %q, want contains %q", gotStr, tt.errorContains)
				}
			} else {
				if !isList {
					t.Errorf("Expected []string, got %T (%v)", gotResult, gotResult)
					return
				}
				if tt.name == "List CWD (.)" {
					if gotList == nil {
						t.Error("Expected non-nil slice for CWD")
					}
					t.Logf("List CWD (.) returned: %v", gotList)
				} else {
					sort.Strings(gotList)
					if !reflect.DeepEqual(gotList, tt.wantResult) {
						t.Errorf("Result list mismatch:\ngot:  %v\nwant: %v", gotList, tt.wantResult)
					}
				}
			}
		})
	}
}

// --- Tests for toolLineCount ---
func TestToolLineCount(t *testing.T) {
	dummyInterp := newDummyInterpreter()
	// Setup test files
	testDir := "linecount_test_files_temp"
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(testDir) })

	// Create files with different line endings and content
	filePath1 := filepath.Join(testDir, "one_line.txt")
	filePath3 := filepath.Join(testDir, "three_lines.txt")
	filePath3nl := filepath.Join(testDir, "three_lines_nl.txt")
	filePathEmpty := filepath.Join(testDir, "empty.txt")

	_ = os.WriteFile(filePath1, []byte("Hello"), 0644)
	_ = os.WriteFile(filePath3, []byte("Line 1\nLine 2\nLine 3"), 0644)
	_ = os.WriteFile(filePath3nl, []byte("Line 1\nLine 2\nLine 3\n"), 0644) // Trailing newline
	_ = os.WriteFile(filePathEmpty, []byte(""), 0644)

	tests := []struct {
		name           string
		inputArg       string // Input to the tool (path or raw string)
		wantResult     int64  // Expected line count (-1 for expected error return)
		valWantErr     bool
		valErrContains string
	}{
		// Raw String Inputs
		{name: "Raw String One Line", inputArg: "Hello", wantResult: 1, valWantErr: false},
		{name: "Raw String Multi Line", inputArg: "Hello\nWorld\nTest", wantResult: 3, valWantErr: false},
		{name: "Raw String With Trailing NL", inputArg: "Hello\nWorld\n", wantResult: 2, valWantErr: false}, // Trailing newline doesn't count as extra line
		{name: "Raw String Empty", inputArg: "", wantResult: 0, valWantErr: false},
		{name: "Raw String Just Newline", inputArg: "\n", wantResult: 1, valWantErr: false}, // One line, but empty first line
		{name: "Raw String Just Newlines", inputArg: "\n\n\n", wantResult: 3, valWantErr: false},

		// File Path Inputs
		{name: "File Path One Line", inputArg: filePath1, wantResult: 1, valWantErr: false},
		{name: "File Path Multi Line", inputArg: filePath3, wantResult: 3, valWantErr: false},
		{name: "File Path With Trailing NL", inputArg: filePath3nl, wantResult: 3, valWantErr: false}, // Trailing newline doesn't count here either
		{name: "File Path Empty", inputArg: filePathEmpty, wantResult: 0, valWantErr: false},
		{name: "File Path Not Found", inputArg: filepath.Join(testDir, "not_found.txt"), wantResult: -1, valWantErr: false}, // Tool should return -1 on read error
		{name: "File Path Invalid (Outside)", inputArg: "../invalid_path_linecount", wantResult: -1, valWantErr: false},     // Tool should return -1 on security error

		// Validation Errors
		{name: "Validation Wrong Arg Type", inputArg: "", valWantErr: true, valErrContains: "expected string, but received type int"},
		{name: "Validation Wrong Arg Count", inputArg: "", valWantErr: true, valErrContains: "expected exactly 1 arguments, but received 0"},
	}

	spec := ToolSpec{Name: "LineCount", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeInt}
	originalWD, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(originalWD) }) // Restore WD

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
			gotResult, toolErr := toolLineCount(dummyInterp, convertedArgs)
			if toolErr != nil {
				t.Fatalf("toolLineCount unexpected Go err: %v", toolErr)
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

// --- Tests for toolReadFile, toolWriteFile, toolSanitizeFilename (Add if needed) ---
