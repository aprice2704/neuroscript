// filename: pkg/core/tools_go_ast_path_helper_test.go
package goast

import (

	// Required for io.Discard
	"log"
	"os"
	"path"          // Use standard 'path' for cleaning expected import paths
	"path/filepath" // Use filepath for OS-specific test setup paths
	"runtime"
	"testing"
)

// Assume testWriter is defined in universal_test_helpers.go or similar
// type testWriter struct { t *testing.T }
// func (tw testWriter) Write(p []byte) (n int, err error) { ... }

// TestDebugCalculateCanonicalPath focuses solely on testing the path calculation logic.
func TestDebugCalculateCanonicalPath(t *testing.T) {
	// Use test logger that writes to t.Logf
	testLogger := log.New(testWriter{t}, "[PATH HELPER TEST] ", log.Ltime|log.Lmicroseconds)

	// Define common base paths for tests, ensuring they are absolute
	// Use TempDir to guarantee writable paths and cleanup
	baseTmpDir := t.TempDir()
	moduleRootDir, err := filepath.Abs(filepath.Join(baseTmpDir, "project")) // e.g., /tmp/TestDebug.../project
	if err != nil {
		t.Fatalf("Failed to get absolute path for test module root: %v", err)
	}
	// Ensure the root directory exists for filepath.Rel to work correctly
	if err := os.MkdirAll(moduleRootDir, 0755); err != nil {
		t.Fatalf("Failed to create test module root dir '%s': %v", moduleRootDir, err)
	}

	// Helper to create subdirectories for tests
	mustCreateDir := func(path string) {
		t.Helper()
		// Create path relative to the specific moduleRootDir for this test run
		absPath := filepath.Join(moduleRootDir, path)
		if err := os.MkdirAll(absPath, 0755); err != nil {
			t.Fatalf("Failed to create test dir '%s': %v", absPath, err)
		}
	}

	testCases := []struct {
		name          string
		modulePath    string // Value from go.mod
		moduleRootDir string // Absolute path to dir containing go.mod
		dirPathRel    string // Relative path *from moduleRootDir* to the directory being processed
		expectedPath  string // Expected canonical Go import path (using forward slashes)
		expectError   bool   // Expect an error from the helper itself (e.g., bad inputs)
	}{
		{
			name:          "Subdirectory",
			modulePath:    "example.com/mymodule",
			moduleRootDir: moduleRootDir,
			dirPathRel:    filepath.Join("pkg", "subpkg"), // Use filepath.Join for OS compatibility
			expectedPath:  "example.com/mymodule/pkg/subpkg",
			expectError:   false,
		},
		{
			name:          "Root Directory",
			modulePath:    "example.com/mymodule",
			moduleRootDir: moduleRootDir,
			dirPathRel:    ".", // Represents the module root directory itself
			expectedPath:  "example.com/mymodule",
			expectError:   false,
		},
		{
			name:          "Nested Subdirectory",
			modulePath:    "example.com/mymodule",
			moduleRootDir: moduleRootDir,
			dirPathRel:    filepath.Join("internal", "util", "helpers"),
			expectedPath:  "example.com/mymodule/internal/util/helpers",
			expectError:   false,
		},
		{
			name:          "Module path with hyphens",
			modulePath:    "my-module/pkg",
			moduleRootDir: moduleRootDir,
			dirPathRel:    "sub-dir", // Keep simple name
			expectedPath:  "my-module/pkg/sub-dir",
			expectError:   false,
		},
		{
			name:          "Original Failing Case Sub1",
			modulePath:    "testtool",
			moduleRootDir: moduleRootDir,                                   // e.g., /tmp/TestDebug.../project
			dirPathRel:    filepath.Join("testtool", "refactored", "sub1"), // Correct relative path
			expectedPath:  "testtool/refactored/sub1",                      // EXPECTED CORRECT PATH
			expectError:   false,
		},
		{
			name:          "Original Failing Case Refactored Base",
			modulePath:    "testtool",
			moduleRootDir: moduleRootDir,
			dirPathRel:    filepath.Join("testtool", "refactored"), // Correct relative path
			expectedPath:  "testtool/refactored",
			expectError:   false,
		},
		//core.Error cases for helper inputs
		{
			name:          "Error: Empty module path",
			modulePath:    "",
			moduleRootDir: moduleRootDir,
			dirPathRel:    "pkg",
			expectError:   true, // Helper should return error
		},
		{
			name:          "Error: Empty module root dir",
			modulePath:    "example.com/mod",
			moduleRootDir: "", // Invalid input
			dirPathRel:    "pkg",
			expectError:   true, // Helper should return error
		},
		{
			name:          "Error: Empty dir path",
			modulePath:    "example.com/mod",
			moduleRootDir: moduleRootDir,
			dirPathRel:    "",   // Invalid input
			expectError:   true, // Helper should return error
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Construct absolute dirPath for the helper call based on test case's moduleRootDir
			absDirPath := filepath.Join(tc.moduleRootDir, tc.dirPathRel)
			if tc.dirPathRel != "." && tc.dirPathRel != "" { // Only create if it's not the root or empty
				mustCreateDir(tc.dirPathRel) // Create using relative path from root
			}

			// Call the helper function
			gotPath, err := debugCalculateCanonicalPath(tc.modulePath, tc.moduleRootDir, absDirPath, testLogger)

			// Check error expectation
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				} else {
					t.Logf("Got expected error: %v", err)
				}
				// Don't check path if error was expected
				return
			}
			// If no error expected, fail if one occurred
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check path only if no error was expected
			// Normalize expected path just in case (using 'path' package for Go paths)
			cleanExpected := path.Clean(tc.expectedPath)
			if gotPath != cleanExpected {
				t.Errorf("Canonical path mismatch:\n  Got Path:  %q\n  Want Path: %q\n  (ModulePath: %q, ModuleRootDir: %q, DirPathRel: %q, AbsDirPath: %q)",
					gotPath, cleanExpected, tc.modulePath, tc.moduleRootDir, tc.dirPathRel, absDirPath)
			} else {
				t.Logf("Canonical path matched: %q", gotPath)
			}

			// Explicitly check the debug log output in the test runner's logs (-v flag)
			t.Logf("--- End Test Case (%s) ---", tc.name)
		})
	}

	// Add a specific test case for Windows path separators if on Windows
	if runtime.GOOS == "windows" {
		t.Run("Windows Path Separators", func(t *testing.T) {
			modulePathWin := "example.com/winmod"
			// Use TempDir for reliable base path
			winTestDir := t.TempDir()
			moduleRootDirWin := filepath.Join(winTestDir, "project")
			dirPathRelWin := `pkg\sub` // Relative path with backslashes
			absDirPathWin := filepath.Join(moduleRootDirWin, dirPathRelWin)
			expectedPathWin := "example.com/winmod/pkg/sub" // Expected slash path

			if err := os.MkdirAll(absDirPathWin, 0755); err != nil {
				t.Fatalf("Failed to create mock windows dir structure: %v", err)
			}

			gotPath, err := debugCalculateCanonicalPath(modulePathWin, moduleRootDirWin, absDirPathWin, testLogger)

			if err != nil {
				t.Errorf("Unexpected error on Windows path test: %v", err)
			}
			cleanExpected := path.Clean(expectedPathWin)
			if gotPath != cleanExpected {
				t.Errorf("Windows path conversion mismatch:\n  Got:  %q\n  Want: %q\n (ModulePath: %q, ModuleRootDir: %q, AbsDirPath: %q)",
					gotPath, cleanExpected, modulePathWin, moduleRootDirWin, absDirPathWin)
			} else {
				t.Logf("Windows path conversion matched: %q", gotPath)
			}
			t.Logf("--- End Test Case (Windows) ---")
		})
	}
}
