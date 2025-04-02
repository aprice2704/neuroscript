// pkg/core/tools_fs_test.go
package core

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// Assume newDummyInterpreter and makeArgs helpers are defined elsewhere

func TestToolListDirectory(t *testing.T) {
	dummyInterp := newDummyInterpreter()
	testBaseDir := "list_dir_test_files_final" // Use different name
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

	tests := []struct {
		name        string
		pathArg     string   // Relative path passed to tool
		wantResult  []string // Expected list of names (sorted)
		wantErr     bool     // Validation error?
		errContains string   // Substring for error message (validation or execution)
		wantPrefix  string   // Prefix for non-validation error messages
	}{
		{name: "List Base Test Dir", pathArg: testBaseDir, wantResult: []string{"file1.txt", "subdir/"}, wantErr: false},
		{name: "List Subdir", pathArg: filepath.Join(testBaseDir, "subdir"), wantResult: []string{"file2.txt"}, wantErr: false},
		{name: "List CWD (.)", pathArg: ".", wantResult: nil, wantErr: false}, // Check non-error execution
		{name: "Path Not Found", pathArg: filepath.Join(testBaseDir, "not_a_dir"), wantErr: false, wantPrefix: "ListDirectory read error", errContains: "no such file or directory"},
		{name: "Path Is A File", pathArg: filepath.Join(testBaseDir, "file1.txt"), wantErr: false, wantPrefix: "ListDirectory read error", errContains: "not a directory"},
		{name: "Path Outside CWD (Relative)", pathArg: "../some_other_dir", wantErr: false, wantPrefix: "ListDirectory path error", errContains: "outside the allowed directory"},
		{
			name:    "Validation Wrong Arg Count",
			pathArg: "", // Placeholder
			wantErr: true,
			// *** UPDATED errContains ***
			errContains: "tool 'ListDirectory' expected exactly 1 arguments, but received 0",
		},
		{
			name:    "Validation Wrong Arg Type",
			pathArg: "", // Placeholder
			wantErr: true,
			// *** UPDATED errContains ***
			errContains: "tool 'ListDirectory' argument 'path' (index 0): expected string, but received type int",
		},
	}

	spec := ToolSpec{Name: "ListDirectory", Args: []ArgSpec{{Name: "path", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeSliceString}

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

			// Validation
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

			// Execution
			gotResult, toolErr := toolListDirectory(dummyInterp, rawArgs)
			if toolErr != nil {
				t.Fatalf("toolListDirectory() returned unexpected Go error: %v", toolErr)
			}

			// Verification
			if tt.wantPrefix != "" {
				gotStr, ok := gotResult.(string)
				if !ok {
					t.Errorf("Expected error string result, got %T", gotResult)
					return
				}
				if !strings.HasPrefix(gotStr, tt.wantPrefix) {
					t.Errorf("Result mismatch: got %q, want prefix %q", gotStr, tt.wantPrefix)
				}
				if tt.errContains != "" && !strings.Contains(gotStr, tt.errContains) {
					t.Errorf("Result error message mismatch: got %q, want contains %q", gotStr, tt.errContains)
				}
			} else if tt.wantResult != nil {
				gotList, ok := gotResult.([]string)
				if !ok {
					t.Errorf("Expected []string result, got %T", gotResult)
					return
				}
				sort.Strings(gotList)
				if !reflect.DeepEqual(gotList, tt.wantResult) {
					t.Errorf("Result list mismatch:\ngot:  %v\nwant: %v", gotList, tt.wantResult)
				}
			} else if tt.name == "List CWD (.)" {
				if _, ok := gotResult.([]string); !ok {
					t.Errorf("Expected []string result for CWD, got %T (%v)", gotResult, gotResult)
				}
				t.Logf("List CWD (.) returned: %v", gotResult)
			} else {
				t.Errorf("Invalid test case setup: wantPrefix or wantResult must be specified if not wantErr")
			}
		})
	}
}
