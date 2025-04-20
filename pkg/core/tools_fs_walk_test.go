// filename: pkg/core/tools_fs_walk_test.go
package core // Use package core to access unexported toolWalkDir

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort" // For sorting results for comparison
	"testing"
)

// --- Test Helpers (local to this file) ---

// createWalkTestInterpreter creates a minimal interpreter setup for walk tests.
func createWalkTestInterpreter(t *testing.T, sandboxDir string) *Interpreter {
	t.Helper()
	// Use correct NewInterpreter signature found in interpreter.go
	interpreter, _ := newDefaultTestInterpreter(t)
	// Set sandbox directory directly
	interpreter.sandboxDir = sandboxDir
	// No need to register all tools if calling toolWalkDir directly
	return interpreter
}

// createWalkTempFile creates a temporary file.
func createWalkTempFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file '%s': %v", filePath, err)
	}
	return filePath
}

// createWalkTempDir creates a temporary directory, including parents.
func createWalkTempDir(t *testing.T, parentDir, dirName string) string {
	t.Helper()
	dirPath := filepath.Join(parentDir, dirName)
	err := os.MkdirAll(dirPath, 0755) // Use MkdirAll for nested structure
	if err != nil {
		t.Fatalf("Failed to create temp dir '%s': %v", dirPath, err)
	}
	return dirPath
}

// assertWalkResultsEqual compares slices of maps, sorting by path first.
func assertWalkResultsEqual(t *testing.T, name string, expected, actual []map[string]interface{}) {
	t.Helper()

	// Sort function
	sortFn := func(slice []map[string]interface{}) {
		sort.SliceStable(slice, func(i, j int) bool {
			pathI, okI := slice[i]["path"].(string)
			pathJ, okJ := slice[j]["path"].(string)
			if !okI || !okJ {
				t.Errorf("%s: Found non-string or missing 'path' key during sort comparison: item i=%#v, item j=%#v", name, slice[i], slice[j])
				return false // Cannot compare reliably
			}
			return pathI < pathJ
		})
	}

	sortFn(expected)
	sortFn(actual)

	if !reflect.DeepEqual(expected, actual) {
		// Basic error message
		t.Errorf("%s: Walk result mismatch (-want +got). Lengths: Want=%d, Got=%d", name, len(expected), len(actual))

		// More detailed diff (limited to presence and path key for brevity)
		maxLen := len(expected)
		if len(actual) > maxLen {
			maxLen = len(actual)
		}
		for i := 0; i < maxLen; i++ {
			var ePath, aPath interface{} = "<missing>", "<missing>"
			if i < len(expected) {
				ePath = expected[i]["path"]
			}
			if i < len(actual) {
				aPath = actual[i]["path"]
			}
			if !reflect.DeepEqual(ePath, aPath) || (i >= len(expected) || i >= len(actual)) || !reflect.DeepEqual(expected[i], actual[i]) {
				t.Errorf("  Mismatch near index %d:\n    Want Path: %v -> %#v\n    Got Path:  %v -> %#v", i, ePath, expected[i], aPath, actual[i])
			}
		}
	}
}

// --- Test Suite ---

func TestToolWalkDir(t *testing.T) {
	sandboxDir := t.TempDir()
	interpreter := createWalkTestInterpreter(t, sandboxDir)

	// --- Create Test Structure ---
	createWalkTempFile(t, sandboxDir, "root_file1.txt", "r1")
	createWalkTempFile(t, sandboxDir, ".hiddenfile", "hidden") // Hidden file
	subDir1 := createWalkTempDir(t, sandboxDir, "subdir1")
	createWalkTempFile(t, subDir1, "sub1_fileA.txt", "s1a")
	createWalkTempDir(t, sandboxDir, "subdir2_empty")            // Empty subdir
	nestedDir := createWalkTempDir(t, subDir1, ".nested_hidden") // Hidden subdir
	createWalkTempFile(t, nestedDir, "nested_file.dat", "nd")
	// File to test 'not a directory' error
	notADirFile := "not_a_dir.txt"
	createWalkTempFile(t, sandboxDir, notADirFile, "file")

	// --- Test Cases ---
	testCases := []struct {
		name           string
		relPath        string                   // Path passed to the tool
		expectedResult []map[string]interface{} // Expected slice of maps if successful (ModTime/Size omitted for simplicity)
		wantNilResult  bool                     // True if expected result should be nil (e.g., path not found)
		expectedGoErr  error                    // Expected specific Go error type (using errors.Is)
	}{
		{
			name:    "Walk Root",
			relPath: ".",
			expectedResult: []map[string]interface{}{
				// Files first due to sorting
				{"name": ".hiddenfile", "path": ".hiddenfile", "isDir": false, "size": int64(6)},
				{"name": "not_a_dir.txt", "path": "not_a_dir.txt", "isDir": false, "size": int64(4)},
				{"name": "root_file1.txt", "path": "root_file1.txt", "isDir": false, "size": int64(2)},
				// Then dirs
				{"name": "subdir1", "path": "subdir1", "isDir": true, "size": int64(0)}, // Placeholder size
				{"name": ".nested_hidden", "path": "subdir1/.nested_hidden", "isDir": true, "size": int64(0)},
				{"name": "nested_file.dat", "path": "subdir1/.nested_hidden/nested_file.dat", "isDir": false, "size": int64(2)},
				{"name": "sub1_fileA.txt", "path": "subdir1/sub1_fileA.txt", "isDir": false, "size": int64(3)},
				{"name": "subdir2_empty", "path": "subdir2_empty", "isDir": true, "size": int64(0)},
			},
		},
		{
			name:    "Walk Subdir1",
			relPath: "subdir1",
			expectedResult: []map[string]interface{}{
				{"name": ".nested_hidden", "path": ".nested_hidden", "isDir": true, "size": int64(0)},
				{"name": "nested_file.dat", "path": ".nested_hidden/nested_file.dat", "isDir": false, "size": int64(2)},
				{"name": "sub1_fileA.txt", "path": "sub1_fileA.txt", "isDir": false, "size": int64(3)},
			},
		},
		{
			name:           "Walk Empty Subdir",
			relPath:        "subdir2_empty",
			expectedResult: []map[string]interface{}{}, // Expect empty list
		},
		{
			name:          "Path is a file",
			relPath:       notADirFile,
			expectedGoErr: ErrInvalidArgument,
		},
		{
			name:          "Path not found",
			relPath:       "nonexistent_dir",
			wantNilResult: true, // Expect nil result, nil error
			expectedGoErr: nil,
		},
		{
			name:          "Path outside sandbox",
			relPath:       "../other",
			expectedGoErr: ErrPathViolation,
		},
		{
			name:          "Path with null byte",
			relPath:       "subdir\x001",
			expectedGoErr: ErrNullByteInArgument,
		},
		{
			name:          "Empty Path",
			relPath:       "",
			expectedGoErr: ErrInvalidArgument, // Explicit check in toolWalkDir
		},
	}

	// --- Run Tests ---
	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			args := []interface{}{tc.relPath}
			// Call tool function directly
			result, err := toolWalkDir(interpreter, args)

			// Check Go error
			if tc.expectedGoErr != nil {
				if err == nil {
					t.Fatalf("Expected Go error wrapping %v, but got nil", tc.expectedGoErr)
				}
				if !errors.Is(err, tc.expectedGoErr) {
					t.Fatalf("Expected Go error wrapping %v, but got %v (type %T)", tc.expectedGoErr, err, err)
				}
				return // Don't check result if error was expected
			}
			if err != nil {
				// No error was expected
				t.Fatalf("Unexpected Go error: %v", err)
			}

			// Check nil result expectation
			if tc.wantNilResult {
				if result != nil {
					t.Fatalf("Expected result to be nil, but got: %v", result)
				}
				return // Don't check map content if nil was expected
			}

			// Check result type and content
			if result == nil {
				t.Fatalf("Expected non-nil result (slice of maps), but got nil")
			}
			resultSlice, ok := result.([]map[string]interface{})
			if !ok {
				t.Fatalf("Expected result type []map[string]interface{}, but got %T", result)
			}

			// Prepare expected results (clean out ModTime and flexible size for dirs)
			expectedResultClean := make([]map[string]interface{}, len(tc.expectedResult))
			for i, item := range tc.expectedResult {
				cleanItem := make(map[string]interface{})
				for k, v := range item {
					if k != "modTime" {
						if k == "size" {
							// If dir size is 0 in expected, don't add it, let comparison handle it
							if isDir, _ := item["isDir"].(bool); isDir && v == int64(0) {
								continue
							}
						}
						cleanItem[k] = v
					}
				}
				expectedResultClean[i] = cleanItem
			}

			// Clean actual results (remove ModTime, check dir size type)
			actualResultClean := make([]map[string]interface{}, len(resultSlice))
			for i, item := range resultSlice {
				cleanItem := make(map[string]interface{})
				for k, v := range item {
					if k != "modTime" {
						if k == "size" {
							// Check type only for size
							if _, isInt64 := v.(int64); !isInt64 {
								t.Errorf("Item %d ('%s'): size expected int64, got %T", i, item["path"], v)
							}
							// If dir, only copy size if it wasn't 0 in expected map
							if isDir, _ := item["isDir"].(bool); isDir {
								if _, expectedSizeExists := expectedResultClean[i]["size"]; expectedSizeExists {
									cleanItem[k] = v // Copy size only if expected had non-zero size
								}
							} else {
								cleanItem[k] = v // Copy size for files
							}

						} else {
							cleanItem[k] = v // Copy other keys
						}
					} else {
						// Optional: Check modTime format if desired
						if _, isString := v.(string); !isString {
							t.Errorf("Item %d ('%s'): modTime expected string, got %T", i, item["path"], v)
						}
						// _, parseErr := time.Parse(time.RFC3339, modTimeStr) // Example check
					}
				}
				actualResultClean[i] = cleanItem
			}

			// Compare cleaned results after sorting
			assertWalkResultsEqual(t, tc.name, expectedResultClean, actualResultClean)
		})
	}
}
