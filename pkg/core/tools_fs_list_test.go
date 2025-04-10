// filename: pkg/core/tools_fs_list_test.go
package core

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// Assume newTestInterpreter and makeArgs are defined in testing_helpers.go

func TestToolListDirectory(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	testBaseDirAbs := t.TempDir()

	subDirRel := "subdir"
	file1Rel := "file1.txt"
	file2Rel := "file2.txt" // Relative to subdir
	hiddenFileRel := ".hiddenfile"

	subDirAbs := filepath.Join(testBaseDirAbs, subDirRel)
	if err := os.Mkdir(subDirAbs, 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testBaseDirAbs, file1Rel), []byte("abc"), 0644); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDirAbs, file2Rel), []byte("def"), 0644); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testBaseDirAbs, hiddenFileRel), []byte("ghi"), 0644); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	wantBaseResultMaps := []interface{}{
		map[string]interface{}{"name": ".hiddenfile", "is_dir": false},
		map[string]interface{}{"name": "file1.txt", "is_dir": false},
		map[string]interface{}{"name": "subdir", "is_dir": true},
	}
	wantSubdirResultMaps := []interface{}{
		map[string]interface{}{"name": "file2.txt", "is_dir": false},
	}

	tests := []struct {
		name           string
		pathArg        string // Path *relative* to testBaseDirAbs
		wantResult     []interface{}
		wantErrorMsg   bool
		errorContains  string
		valWantErr     bool
		valErrContains string
	}{
		{name: "List Base Dir Relative", pathArg: ".", wantResult: wantBaseResultMaps, wantErrorMsg: false, valWantErr: false},
		{name: "List Subdir Relative", pathArg: subDirRel, wantResult: wantSubdirResultMaps, wantErrorMsg: false, valWantErr: false},
		{name: "Path Not Found Relative", pathArg: "not_a_dir", wantErrorMsg: true, errorContains: "ListDirectory failed: Directory not found", valWantErr: false}, // Updated expected error
		{name: "Path Is A File Relative", pathArg: file1Rel, wantErrorMsg: true, errorContains: "not a directory", valWantErr: false},
		{name: "Path Outside Sandbox Relative", pathArg: "../outside_dir", wantErrorMsg: true, errorContains: "outside the allowed directory", valWantErr: false},
		{name: "Path Is Parent Relative", pathArg: "..", wantErrorMsg: true, errorContains: "outside the allowed directory", valWantErr: false},
		{name: "Validation Wrong Arg Count", pathArg: "", wantErrorMsg: false, valWantErr: true, valErrContains: "expected exactly 1 arguments"},
		// *** UPDATED Expected Error String ***
		{name: "Validation Wrong Arg Type", pathArg: "", wantErrorMsg: false, valWantErr: true, valErrContains: "type validation failed for argument 'path' of tool 'ListDirectory': expected string, got int"},
	}

	spec := ToolSpec{Name: "ListDirectory", Args: []ArgSpec{{Name: "path", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeSliceAny}

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
				rawArgs = makeArgs(tt.pathArg)
			}

			// --- Change CWD to temp dir for execution ---
			err := os.Chdir(testBaseDirAbs)
			if err != nil {
				t.Fatalf("Failed to Chdir to temp dir %s: %v", testBaseDirAbs, err)
			}
			defer os.Chdir(originalWD) // Ensure WD is restored *after this subtest*
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
			gotResult, toolErr := toolListDirectory(dummyInterp, convertedArgs)

			// Check error status
			expectFail := tt.wantErrorMsg
			// Failure is indicated by EITHER a non-nil Go error OR a string return value
			gotFail := (toolErr != nil) || (gotResult != nil && reflect.TypeOf(gotResult).Kind() == reflect.String)

			if gotFail != expectFail {
				t.Errorf("Failure status mismatch: gotFail=%t, wantFail=%t. toolErr=%v, gotResult=%v(%T)", gotFail, expectFail, toolErr, gotResult, gotResult)
			}

			// Check error message content if failure was expected
			if expectFail {
				var errMsg string
				if toolErr != nil {
					errMsg = toolErr.Error()
				} else if gotStr, ok := gotResult.(string); ok {
					errMsg = gotStr
				} else {
					t.Fatalf("Expected failure message string or Go error, but got %T", gotResult)
				}

				if tt.errorContains != "" && !strings.Contains(errMsg, tt.errorContains) {
					t.Errorf("Error message mismatch: got error %q, want contains %q", errMsg, tt.errorContains)
				}
				return // Stop test if failure occurred (expected or not)
			}

			// --- If no error expected ---
			gotList, isList := gotResult.([]interface{})
			if !isList {
				t.Errorf("Expected []interface{}, got %T (%v)", gotResult, gotResult)
				return
			}

			// Sort gotList for comparison
			sort.SliceStable(gotList, func(i, j int) bool {
				mapI, okI := gotList[i].(map[string]interface{})
				mapJ, okJ := gotList[j].(map[string]interface{})
				if !okI || !okJ {
					return false
				}
				nameI, _ := mapI["name"].(string)
				nameJ, _ := mapJ["name"].(string)
				return nameI < nameJ
			})
			if !reflect.DeepEqual(gotList, tt.wantResult) {
				t.Errorf("Result list mismatch:\ngot:  %#v\nwant: %#v", gotList, tt.wantResult)
			}
		})
	}
}
