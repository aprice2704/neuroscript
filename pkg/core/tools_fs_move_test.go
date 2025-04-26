// filename: pkg/core/tools_fs_move_test.go
package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// --- MoveFile Validation Tests ---
func TestToolMoveFileValidation(t *testing.T) {
	testCases := []ValidationTestCase{
		{Name: "Wrong Arg Count (None)", InputArgs: MakeArgs(), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong Arg Count (One)", InputArgs: MakeArgs("src"), ExpectedError: ErrValidationArgCount},
		{Name: "Wrong Arg Count (Three)", InputArgs: MakeArgs("src", "dest", "extra"), ExpectedError: ErrValidationArgCount},
		{Name: "Nil First Arg", InputArgs: MakeArgs(nil, "dest"), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Nil Second Arg", InputArgs: MakeArgs("src", nil), ExpectedError: ErrValidationRequiredArgNil},
		{Name: "Wrong First Arg Type", InputArgs: MakeArgs(123, "dest"), ExpectedError: ErrValidationTypeMismatch},
		{Name: "Wrong Second Arg Type", InputArgs: MakeArgs("src", 456), ExpectedError: ErrValidationTypeMismatch},
		// --- REMOVED: Empty path checks are done inside tool func, not ValidateAndConvertArgs ---
		// {Name: "Empty First Arg", InputArgs: MakeArgs("", "dest"), ExpectedError: ErrValidationArgValue}, // Tool func validates empty path
		// {Name: "Empty Second Arg", InputArgs: MakeArgs("src", ""), ExpectedError: ErrValidationArgValue}, // Tool func validates empty path
		// --- END REMOVED ---
		{Name: "Correct Args", InputArgs: MakeArgs("source.txt", "destination.txt"), ExpectedError: nil},
		// Note: Path security validation happens inside the tool function
	}
	runValidationTestCases(t, "MoveFile", testCases)
}

// --- MoveFile Functional Tests ---
func TestToolMoveFileFunctional(t *testing.T) {
	// Use t.TempDir for sandboxed filesystem operations
	sandboxDir := t.TempDir()
	interp := newTestInterpreterWithSandbox(t, sandboxDir) // Assumes a helper like this exists or create one

	// --- Test Setup Helper ---
	createTestFile := func(relativePath, content string) string {
		t.Helper()
		absPath := filepath.Join(sandboxDir, relativePath)
		err := os.WriteFile(absPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", absPath, err)
		}
		return relativePath // return relative path for use in tool calls
	}

	// --- Test Cases ---
	testCases := []struct {
		name           string
		sourcePath     string             // Relative path for tool arg
		destPath       string             // Relative path for tool arg
		setupFunc      func()             // Optional setup specific to this case (e.g., create source/dest)
		expectErr      bool               // Expect a Go error from the tool func itself
		expectMapError bool               // Expect the returned map["error"] to be non-nil
		checkFunc      func(t *testing.T) // Optional function to verify filesystem state
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
		// Add functional tests for empty paths if desired
		{
			name:           "Fail: Empty Source Path",
			sourcePath:     "",
			destPath:       "some_dest.txt",
			expectErr:      true,  // Expect Go error from tool func validation
			expectMapError: false, // Map shouldn't even be returned if Go error occurs
		},
		{
			name:           "Fail: Empty Destination Path",
			sourcePath:     createTestFile("another_valid_src.txt", "content6"),
			destPath:       "",
			expectErr:      true, // Expect Go error from tool func validation
			expectMapError: false,
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
					t.Errorf("Expected a Go error, but got nil")
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
				mapErr := resultMap["error"]
				if tc.expectMapError {
					if mapErr == nil {
						t.Errorf("Expected map[\"error\"] to be non-nil, but got nil")
					} else {
						// Optionally check error message contains expected substring? Be careful with exact matches.
						t.Logf("Got expected error in map: %v", mapErr)
					}
				} else {
					if mapErr != nil {
						t.Errorf("Expected map[\"error\"] to be nil, but got: %v", mapErr)
					}
				}
			} else {
				// If a Go error was returned, the map might be nil or contain the error again
				// We prioritize checking the Go error tc.expectErr
				t.Logf("Tool function returned Go error (as expected or unexpected): %v", err)
				if resultIntf != nil {
					// Check if the map also contains the error if present
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
func newTestInterpreterWithSandbox(t *testing.T, sandboxDir string) *Interpreter {
	t.Helper()
	// Create a minimal interpreter for testing
	interp, _ := NewDefaultTestInterpreter(t) // Use nil logger or a test logger
	interp.sandboxDir = sandboxDir
	// Register necessary tools if validation relies on them (unlikely for MoveFile validation itself)
	// RegisterCoreTools(interp.toolRegistry) // Maybe not needed for just validation tests
	return interp
}

// Helper to create arguments easily (assuming it exists from other tests)
// func MakeArgs(args ...interface{}) []interface{} {
// 	return args
// }
