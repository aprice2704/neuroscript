// NeuroScript Version: 0.3.1
// File version: 0.0.8 // Fix compareStatResults signature mismatch.
// nlines: 195 // Approximate
// risk_rating: LOW
// filename: pkg/core/tools_fs_stat_test.go
package core

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// --- Keep Helpers (createTempFile, createTempDir) ---
func createTempFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	parentDir := filepath.Dir(filepath.Join(dir, filename))
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		t.Fatalf("Failed to create parent dir '%s' for temp file '%s': %v", parentDir, filename, err)
	}
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file '%s': %v", filePath, err)
	}
	return filePath
}

func createTempDir(t *testing.T, parentDir, dirName string) string {
	t.Helper()
	dirPath := filepath.Join(parentDir, dirName)
	err := os.Mkdir(dirPath, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("Failed to create temp dir '%s': %v", dirPath, err)
	}
	return dirPath
}

func TestToolStat(t *testing.T) {
	// Shared setup data (relative paths)
	testFileName := "test_file.txt"
	testFileContent := "hello world"
	testDirName := "test_subdir"

	// --- Test Cases ---
	testCases := []fsTestCase{
		{
			name:     "Happy Path - Existing File",
			toolName: "StatPath",
			args:     MakeArgs(testFileName),
			wantResult: map[string]interface{}{
				"name":             testFileName,
				"path":             testFileName,
				"size_bytes":       int64(len(testFileContent)),
				"is_dir":           false,
				"mode_string":      "-rw-r--r--",
				"mode_perm":        "0644",
				"modified_unix":    int64(0), // Placeholder
				"modified_rfc3339": "",       // Placeholder
			},
		},
		{
			name:     "Happy Path - Existing Directory",
			toolName: "StatPath",
			args:     MakeArgs(testDirName),
			wantResult: map[string]interface{}{
				"name":             testDirName,
				"path":             testDirName,
				"is_dir":           true,
				"size_bytes":       int64(0), // Placeholder
				"mode_string":      "drwxr-xr-x",
				"mode_perm":        "0755",
				"modified_unix":    int64(0),
				"modified_rfc3339": "",
			},
		},
		{
			name:     "Happy Path - Current Directory",
			toolName: "StatPath",
			args:     MakeArgs("."),
			wantResult: map[string]interface{}{
				"path":             ".",
				"is_dir":           true,
				"size_bytes":       int64(0), // Placeholder
				"mode_string":      "drwxr-xr-x",
				"mode_perm":        "0755",
				"modified_unix":    int64(0),
				"modified_rfc3339": "",
			},
		},
		{
			name:          "Unhappy Path - Non-existent File",
			toolName:      "StatPath",
			args:          MakeArgs("non_existent_file.txt"),
			wantResult:    "path not found 'non_existent_file.txt'",
			wantToolErrIs: ErrFileNotFound,
		},
		{
			name:          "Unhappy Path - Outside Sandbox",
			toolName:      "StatPath",
			args:          MakeArgs("../outside_file.txt"),
			wantResult:    "path resolves outside allowed directory",
			wantToolErrIs: ErrPathViolation,
		},
		{
			name:          "Unhappy Path - Absolute Path",
			toolName:      "StatPath",
			args:          MakeArgs("/abs/path/test.txt"),
			wantResult:    "must be relative, not absolute",
			wantToolErrIs: ErrPathViolation,
		},
		{
			name:          "Unhappy Path - Empty Path",
			toolName:      "StatPath",
			args:          MakeArgs(""),
			wantResult:    "path argument cannot be empty",
			wantToolErrIs: ErrInvalidArgument,
		},
		{
			name:          "Unhappy Path - Null Byte Path",
			toolName:      "StatPath",
			args:          MakeArgs("file\x00withnull.txt"),
			wantResult:    "path contains null byte",
			wantToolErrIs: ErrNullByteInArgument,
		},
		{
			name:          "Validation_Wrong_Arg_Type",
			toolName:      "StatPath",
			args:          MakeArgs(123),
			wantResult:    "path argument must be a string",
			wantToolErrIs: ErrInvalidArgument,
		},
		{
			name:          "Validation_Nil_Arg",
			toolName:      "StatPath",
			args:          MakeArgs(nil),
			wantResult:    "path argument must be a string",
			wantToolErrIs: ErrInvalidArgument,
		},
		{
			name:          "Validation_Missing_Arg",
			toolName:      "StatPath",
			args:          MakeArgs(),
			wantResult:    "expected 1 argument",
			wantToolErrIs: ErrArgumentMismatch,
		},
	}

	// Custom comparison for Stat results
	// *** Ensure signature matches compareFuncType ***
	compareStatResults := func(t *testing.T, tc fsTestCase, expected, actual interface{}) {
		t.Helper()
		actualMap, okA := actual.(map[string]interface{})
		if !okA {
			t.Fatalf("Actual result is not map[string]interface{}, got %T", actual)
		}
		expectedMap, okE := expected.(map[string]interface{})
		if !okE {
			if _, ok := expected.(string); ok && tc.wantToolErrIs != nil {
				t.Logf("Expected result was error string, skipping map comparison.")
				return
			}
			t.Fatalf("Expected result is not map[string]interface{}, got %T", expected)
		}

		// Extract relative path from tc.args
		relPath := ""
		if len(tc.args) > 0 {
			relPath, _ = tc.args[0].(string)
		}

		// Fields to compare directly
		fieldsToCompare := []string{"path", "is_dir", "mode_perm"}
		if name, ok := expectedMap["name"]; ok {
			if nameStr, okStr := name.(string); okStr && nameStr != "" {
				fieldsToCompare = append(fieldsToCompare, "name")
			}
		}
		if isDirVal, ok := actualMap["is_dir"].(bool); ok && !isDirVal {
			fieldsToCompare = append(fieldsToCompare, "size_bytes")
		}

		for _, key := range fieldsToCompare {
			wantVal, wantOk := expectedMap[key]
			gotVal, gotOk := actualMap[key]

			// Use relPath (derived from tc) for special handling
			if key == "name" && relPath == "." && !wantOk {
				if !gotOk {
					t.Errorf("Mandatory key %q missing from actual map for '.' path", key)
				} else {
					t.Logf("Ignoring potentially missing 'name' in expectedMap for '.' path test")
					continue
				}
			}

			if !wantOk {
				t.Errorf("Expected key %q missing in expectedMap", key)
				continue
			}
			if !gotOk {
				t.Errorf("Expected key %q missing in actual result map", key)
				continue
			}
			if !reflect.DeepEqual(wantVal, gotVal) {
				if key == "mode_string" {
					// ... (mode_string special handling as before) ...
					if isDir, ok := actualMap["is_dir"].(bool); ok {
						gotStr, _ := gotVal.(string)
						prefix := "-"
						if isDir {
							prefix = "d"
						}
						if !strings.HasPrefix(gotStr, prefix) {
							t.Errorf("Mismatch for key %q prefix: Got %q, Want prefix %q (Full Want: %#v)", key, gotStr, prefix, wantVal)
						} else {
							t.Logf("Ignoring potential mode_string variance beyond type prefix for key %q:\n  Got:  %#v (%T)\n  Want: %#v (%T)", key, gotVal, gotVal, wantVal, wantVal)
						}
						continue
					}
				}
				t.Errorf("Mismatch for key %q:\n  Got:  %#v (%T)\n  Want: %#v (%T)", key, gotVal, gotVal, wantVal, wantVal)
			}
		}

		// Check presence and basic validity of dynamic fields
		dynamicFields := []string{"name", "size_bytes", "modified_unix", "modified_rfc3339", "mode_string"}
		for _, key := range dynamicFields {
			val, ok := actualMap[key]
			if !ok {
				// Use relPath for special handling
				if key == "name" && relPath == "." {
					continue
				}
				t.Errorf("Mandatory dynamic key %q missing from actual result map", key)
				continue
			}
			if key == "modified_unix" && (val == int64(0) || val == nil) {
				t.Errorf("modified_unix seems invalid: %#v", val)
			}
			if key == "modified_rfc3339" && (val == "" || val == nil) {
				t.Errorf("modified_rfc3339 seems invalid: %#v", val)
			}
			// Use relPath for special handling
			if key == "name" && relPath != "." && (val == "" || val == nil) {
				t.Errorf("name seems invalid: %#v", val)
			}
		}
	}

	// Run tests using the helper
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interp, currentSandbox := NewDefaultTestInterpreter(t)
			if strings.Contains(tc.name, "Existing File") {
				createTempFile(t, currentSandbox, testFileName, testFileContent)
			}
			if strings.Contains(tc.name, "Existing Directory") || strings.Contains(tc.name, "Current Directory") {
				createTempDir(t, currentSandbox, testDirName)
			}
			// *** Ensure the call passes the function matching the type ***
			testFsToolHelperWithCompare(t, interp, tc, compareStatResults)
		})
	}
}
