// NeuroScript Version: 0.3.1
// File version: 0.1.2
// Remove unused mapErrMsg variable. Update expected errors.
// nlines: 170
// risk_rating: LOW
// filename: pkg/core/tools_fs_move_test.go
package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	// "strings" // Import strings if needed for error message checks
)

// --- MoveFile Validation Tests ---
func TestToolMoveFileValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		// Corrected: Expect ErrValidationRequiredArgMissing when 'source_path' is missing
		{Name: "Wrong_Arg_Count_(None)", InputArgs: MakeArgs(), ExpectedError: ErrValidationRequiredArgMissing},
		// Corrected: Expect ErrValidationRequiredArgMissing when 'destination_path' is missing
		{Name: "Wrong_Arg_Count_(One)", InputArgs: MakeArgs("src"), ExpectedError: ErrValidationRequiredArgMissing},
		{Name: "Wrong_Arg_Count_(Three)", InputArgs: MakeArgs("src", "dest", "extra"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil_First_Arg", InputArgs: MakeArgs(nil, "dest"), ExpectedError: ErrValidationRequiredArgNil}, // source_path required
		{Name: "Nil_Second_Arg", InputArgs: MakeArgs("src", nil), ExpectedError: ErrValidationRequiredArgNil}, // destination_path required
		{Name: "Wrong_First_Arg_Type", InputArgs: MakeArgs(123, "dest"), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong_Second_Arg_Type", InputArgs: MakeArgs("src", 456), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Correct_Args", InputArgs: MakeArgs("source.txt", "destination.txt"), ExpectedError: nil},
	}
	runValidationTestCases(t, "FS.Move", testCases)
}

// --- MoveFile Functional Tests ---
func TestToolMoveFileFunctional(t *testing.T) {
	// Use t.TempDir for sandboxed filesystem operations
	sandboxDir := t.TempDir()
	interp := NewTestInterpreterWithSandbox(t, sandboxDir) // Assumes a helper like this exists or create one

	// --- Test Setup Helper ---
	createTestFile := func(relativePath, content string) string {
		t.Helper()
		absPath := filepath.Join(sandboxDir, relativePath)
		// Ensure parent directory exists for the test file
		parentDir := filepath.Dir(absPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			t.Fatalf("Failed to create parent directory %s for test file: %v", parentDir, err)
		}
		err := os.WriteFile(absPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", absPath, err)
		}
		return relativePath // return relative path for use in tool calls
	}

	// --- Test Cases ---
	testCases := []struct {
		name           string
		sourcePath     string // Relative path for tool arg
		destPath       string // Relative path for tool arg
		setupFunc      func() // Optional setup specific to this case (e.g., create source/dest)
		expectErr      bool   // Expect a Go error from the tool func itself
		expectMapError bool   // Expect the returned map["error"] to be non-nil
		// expectedErrMsgSubstring string    // Optional: Check if map error message contains this text
		checkFunc func(t *testing.T) // Optional function to verify filesystem state
	}{
		{
			name:           "Success: Rename file",
			sourcePath:     createTestFile("old.txt", "content1"),
			destPath:       "new.txt",
			expectErr:      false,
			expectMapError: false,
			checkFunc: func(t *testing.T) {
				if _, err := os.Stat(filepath.Join(sandboxDir, "old.txt")); !errors.Is(err, os.ErrNotExist) {
					t.Errorf("Source file old.txt still exists after successful move")
				}
				if _, err := os.Stat(filepath.Join(sandboxDir, "new.txt")); err != nil {
					t.Errorf("Destination file new.txt not found after successful move: %v", err)
				}
			},
		},
		{
			name: "Success: Move file into existing subdir",
			setupFunc: func() {
				createTestFile("move_me.txt", "content2")
				os.Mkdir(filepath.Join(sandboxDir, "subdir"), 0755)
			},
			sourcePath:     "move_me.txt",
			destPath:       "subdir/moved.txt",
			expectErr:      false,
			expectMapError: false,
			checkFunc: func(t *testing.T) {
				if _, err := os.Stat(filepath.Join(sandboxDir, "move_me.txt")); !errors.Is(err, os.ErrNotExist) {
					t.Errorf("Source file move_me.txt still exists after successful move")
				}
				if _, err := os.Stat(filepath.Join(sandboxDir, "subdir/moved.txt")); err != nil {
					t.Errorf("Destination file subdir/moved.txt not found after successful move: %v", err)
				}
			},
		},
		{
			name:           "Fail: Source does not exist",
			sourcePath:     "nonexistent_source.txt",
			destPath:       "any_dest.txt",
			expectErr:      true, // Expect Go error because os.Stat fails
			expectMapError: true,
			checkFunc: func(t *testing.T) {
				// Ensure destination wasn't created
				if _, err := os.Stat(filepath.Join(sandboxDir, "any_dest.txt")); !errors.Is(err, os.ErrNotExist) {
					t.Errorf("Destination file should not exist when source is missing")
				}
			},
		},
		{
			name: "Fail: Destination exists",
			setupFunc: func() {
				createTestFile("src_exists.txt", "content3")
				createTestFile("dest_exists.txt", "content4")
			},
			sourcePath:     "src_exists.txt",
			destPath:       "dest_exists.txt",
			expectErr:      true, // Expect Go error because destination exists
			expectMapError: true,
			checkFunc: func(t *testing.T) {
				// Ensure source wasn't moved
				if _, err := os.Stat(filepath.Join(sandboxDir, "src_exists.txt")); err != nil {
					t.Errorf("Source file should still exist when destination exists")
				}
			},
		},
		{
			name:           "Fail: Path outside sandbox (Source)",
			sourcePath:     "../outside_src.txt", // Attempt to go up
			destPath:       "dest.txt",
			expectErr:      true, // Expect Go error from SecureFilePath
			expectMapError: true,
		},
		{
			name:           "Fail: Path outside sandbox (Destination)",
			sourcePath:     createTestFile("valid_src.txt", "content5"),
			destPath:       "../outside_dest.txt", // Attempt to go up
			expectErr:      true,                  // Expect Go error from SecureFilePath
			expectMapError: true,
			checkFunc: func(t *testing.T) {
				// Ensure source wasn't moved
				if _, err := os.Stat(filepath.Join(sandboxDir, "valid_src.txt")); err != nil {
					t.Errorf("Source file should still exist when destination is invalid")
				}
			},
		},
		{
			name:           "Fail: Empty Source Path",
			sourcePath:     "",
			destPath:       "some_dest.txt",
			expectErr:      true, // Expect Go error from tool func validation (likely ErrInvalidArgument)
			expectMapError: true, // Tool func should return error in map
		},
		{
			name:           "Fail: Empty Destination Path",
			sourcePath:     createTestFile("another_valid_src.txt", "content6"),
			destPath:       "",
			expectErr:      true, // Expect Go error from tool func validation (likely ErrInvalidArgument)
			expectMapError: true, // Tool func should return error in map
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run setup if defined
			if tc.setupFunc != nil {
				tc.setupFunc()
			}

			// Call the tool function
			resultIntf, err := toolMoveFile(interp, MakeArgs(tc.sourcePath, tc.destPath))

			// Check for Go-level error
			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected a Go error, but got nil. Result: %+v", resultIntf)
				} else {
					t.Logf("Got expected Go error: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected Go error to be nil, but got: %v", err)
				}
			}

			// Check the returned map (if tool func didn't return Go error directly)
			if err == nil {
				resultMap, ok := resultIntf.(map[string]interface{})
				if !ok {
					t.Fatalf("Expected result to be map[string]interface{}, got %T", resultIntf)
				}
				mapErrVal := resultMap["error"]
				// mapErrMsg, _ := mapErrVal.(string) // Removed unused variable

				if tc.expectMapError {
					if mapErrVal == nil {
						t.Errorf("Expected map[\"error\"] to be non-nil, but got nil")
					} else {
						t.Logf("Got expected error in map: %v", mapErrVal)
					}
				} else {
					if mapErrVal != nil {
						t.Errorf("Expected map[\"error\"] to be nil, but got: %v (%T)", mapErrVal, mapErrVal)
					}
				}
				// Additionally check for specific error messages if needed, e.g.,
				// if tc.expectMapError && tc.expectedErrMsgSubstring != "" {
				// 	// Use mapErrMsg here if uncommented
				// 	if !strings.Contains(mapErrMsg, tc.expectedErrMsgSubstring) {
				// 		t.Errorf("Expected map error message to contain %q, but got %q", tc.expectedErrMsgSubstring, mapErrMsg)
				// 	}
				// }

			} else {
				// If a Go error was returned, the map might be nil or contain the error again
				t.Logf("Tool function returned Go error (as expected or unexpected): %v", err)
				if resultIntf != nil {
					resultMap, ok := resultIntf.(map[string]interface{})
					if ok && resultMap != nil && resultMap["error"] != nil {
						t.Logf("Map returned alongside Go error also contains error: %v", resultMap["error"])
					}
				}
			}

			// Run filesystem checks if defined
			if tc.checkFunc != nil {
				tc.checkFunc(t)
			}
		})
	}
}

// Helper function to create a test interpreter with a sandbox directory
// Replace with your actual test setup if different
func NewTestInterpreterWithSandbox(t *testing.T, sandboxDir string) *Interpreter {
	t.Helper()
	// Create a minimal interpreter for testing
	interp, _ := NewDefaultTestInterpreter(t) // Use nil logger or a test logger
	// Ensure the sandbox directory is set correctly *after* potential registration
	err := interp.SetSandboxDir(sandboxDir)
	if err != nil {
		t.Fatalf("Failed to set sandbox dir in test helper: %v", err)
	}
	return interp
}
