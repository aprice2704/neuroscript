// filename: pkg/core/tools_fs_stat_test.go
package core // Changed from core_test to access unexported toolStat

import (
	"errors" // For io.Discard
	// For log.New
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
	// Removed testify imports
)

// Helper to create a test interpreter instance (adapt if you have a standard test setup)
// Uses standard logger and correct NewInterpreter signature.
func createTestInterpreterWithSandbox(t *testing.T, sandboxDir string) *Interpreter {
	t.Helper()

	// Use correct NewInterpreter signature
	interpreter, _ := NewDefaultTestInterpreter(t)
	// Set sandbox directory manually
	interpreter.sandboxDir = sandboxDir
	// Register tools (assuming this is needed - adapt if registration happens elsewhere)
	// It might be better if NewInterpreter handled registration or had a dedicated method.
	// Let's assume core tools registration happens externally or is not strictly needed
	// for calling the tool function directly if the interpreter state is minimal.
	// Consider adding RegisterCoreTools(interpreter.ToolRegistry()) if required by toolStat.
	return interpreter
}

// Helper to create a temporary file with content
func createTempFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	// Use t.Fatalf for setup errors
	if err != nil {
		t.Fatalf("Failed to create temp file '%s': %v", filePath, err)
	}
	return filePath
}

// Helper to create a temporary directory
func createTempDir(t *testing.T, parentDir, dirName string) string {
	t.Helper()
	dirPath := filepath.Join(parentDir, dirName)
	err := os.Mkdir(dirPath, 0755)
	// Use t.Fatalf for setup errors
	if err != nil {
		t.Fatalf("Failed to create temp dir '%s': %v", dirPath, err)
	}
	return dirPath
}

func TestToolStat(t *testing.T) {
	// Test setup: Create a temporary sandbox directory
	sandboxDir := t.TempDir()
	interpreter := createTestInterpreterWithSandbox(t, sandboxDir)

	// --- Create test files and directories ---
	testFileName := "test_file.txt"
	testFileContent := "hello world"
	_ = createTempFile(t, sandboxDir, testFileName, testFileContent) // Ignore path, use rel path

	testDirName := "test_subdir"
	_ = createTempDir(t, sandboxDir, testDirName)

	// --- Test Cases ---
	testCases := []struct {
		name          string
		relPath       string                 // Relative path passed to the tool
		expectResult  bool                   // Whether to expect a non-nil map result
		expectedMap   map[string]interface{} // Expected values if expectResult is true (only checks key fields)
		expectNilErr  bool                   // Expect error returned by Go function to be nil? (Covers os.IsNotExist case)
		expectedGoErr error                  // Expected specific Go error type (using errors.Is) if expectNilErr is false
	}{
		{
			name:         "Happy Path - Existing File",
			relPath:      testFileName,
			expectResult: true,
			expectedMap: map[string]interface{}{
				"name":   testFileName,
				"path":   testFileName,
				"size":   int64(len(testFileContent)),
				"is_dir": false,
			},
			expectNilErr: true,
		},
		{
			name:         "Happy Path - Existing Directory",
			relPath:      testDirName,
			expectResult: true,
			expectedMap: map[string]interface{}{
				"name":   testDirName,
				"path":   testDirName,
				"is_dir": true,
			},
			expectNilErr: true,
		},
		{
			name:         "Happy Path - Current Directory",
			relPath:      ".",
			expectResult: true,
			expectedMap: map[string]interface{}{
				"name":   filepath.Base(sandboxDir), // Name should be the sandbox dir base name
				"path":   ".",
				"is_dir": true,
			},
			expectNilErr: true,
		},
		{
			name:          "Unhappy Path - Non-existent File",
			relPath:       "non_existent_file.txt",
			expectResult:  false, // Expect nil map
			expectNilErr:  true,  // Expect nil error (per implementation)
			expectedGoErr: nil,   // Explicitly nil
		},
		{
			name:          "Unhappy Path - Outside Sandbox",
			relPath:       "../outside_file.txt", // Attempts to go outside
			expectResult:  false,
			expectNilErr:  false,
			expectedGoErr: ErrPathViolation, // Expect security error
		},
		{
			name:          "Unhappy Path - Absolute Path",
			relPath:       filepath.Join(sandboxDir, "abs_path_test.txt"), // Generate an absolute path
			expectResult:  false,
			expectNilErr:  false,
			expectedGoErr: ErrPathViolation, // Expect security error (rejects absolute)
		},
		{
			name:          "Unhappy Path - Empty Path",
			relPath:       "",
			expectResult:  false,
			expectNilErr:  false,
			expectedGoErr: ErrPathViolation, // SecureFilePath rejects empty
		},
		{
			name:          "Unhappy Path - Null Byte Path",
			relPath:       "file\x00withnull.txt",
			expectResult:  false,
			expectNilErr:  false,
			expectedGoErr: ErrNullByteInArgument, // Use the error from errors.go
		},
	}

	// --- Run Tests ---
	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Prepare arguments for the tool function
			args := []interface{}{tc.relPath}

			// Execute the tool function directly (changed from TestingInvokeTool)
			result, err := toolStat(interpreter, args) // Direct call

			// Check Go error against expectations (Using standard testing)
			if tc.expectNilErr {
				if err != nil {
					t.Errorf("Expected Go error to be nil, but got: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected Go error to be non-nil, but got nil") // Use Fatalf if subsequent checks depend on non-nil error
				}
				if tc.expectedGoErr != nil {
					if !errors.Is(err, tc.expectedGoErr) {
						t.Errorf("Expected Go error wrapping %v, but got %v (type %T)", tc.expectedGoErr, err, err)
					}
				}
			}

			// Check result against expectations (Using standard testing)
			if tc.expectResult {
				if result == nil {
					t.Fatalf("Expected result map to be non-nil, but got nil") // Use Fatalf as map checks depend on non-nil result
				}
				resultMap, ok := result.(map[string]interface{})
				if !ok {
					t.Fatalf("Expected result to be a map[string]interface{}, but got type %T", result)
				}

				// Check key fields in the map
				for key, expectedValue := range tc.expectedMap {
					actualValue, keyExists := resultMap[key]
					if !keyExists {
						t.Errorf("Expected key '%s' missing in result map", key)
						continue // Skip value check if key missing
					}

					// Special handling for size if directory (can vary)
					if key == "size" && tc.expectedMap["is_dir"] == true {
						if _, isInt64 := actualValue.(int64); !isInt64 {
							t.Errorf("Expected size to be int64 for dir, but got type %T", actualValue)
						}
					} else if key != "mod_time" { // Don't check mod_time for exact match
						if !reflect.DeepEqual(expectedValue, actualValue) {
							t.Errorf("Value mismatch for key '%s': expected '%v' (%T), got '%v' (%T)", key, expectedValue, expectedValue, actualValue, actualValue)
						}
					} else {
						// Check mod_time format and approximate time
						modTimeStr, isString := actualValue.(string)
						if !isString {
							t.Errorf("Expected mod_time to be a string, but got type %T", actualValue)
						} else {
							_, parseErr := time.Parse(time.RFC3339, modTimeStr)
							if parseErr != nil {
								t.Errorf("mod_time (%s) should be in RFC3339 format, parse error: %v", modTimeStr, parseErr)
							}
						}
					}

				}
				// Check that path is always the input relative path
				pathVal, pathExists := resultMap["path"]
				if !pathExists {
					t.Errorf("Expected key 'path' missing in result map")
				} else if pathVal != tc.relPath {
					t.Errorf("Result map 'path' expected '%s', got '%v'", tc.relPath, pathVal)
				}

			} else {
				if result != nil {
					t.Errorf("Expected result to be nil, but got: %v", result)
				}
			}
		})
	}
}
