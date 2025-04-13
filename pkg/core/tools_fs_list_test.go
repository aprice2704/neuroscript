// filename: pkg/core/tools_fs_list_test.go
package core

import (
	// ADDED errors import
	"fmt" // ADDED fmt import
	"os"
	"path/filepath"
	"testing"
)

// Assume testFsToolHelper is defined in testing_helpers_test.go

func TestToolListDirectory(t *testing.T) {
	interp, _ := newDefaultTestInterpreter(t) // Get interpreter for sandbox path

	// --- Test Setup Data ---
	baseDirRel := "listTest"
	subDirRel := filepath.Join(baseDirRel, "sub")
	file1Rel := filepath.Join(baseDirRel, "file1.txt")
	file2Rel := filepath.Join(subDirRel, "file2.txt")
	rootFileRel := "file_at_root.txt"
	file1Content := "hello"
	file2Content := "world"
	rootFileContent := "root"

	// --- Setup Function ---
	// *** MODIFIED: Takes sandboxRoot string argument and uses it ***
	setupListDirTest := func(sandboxRoot string) error {
		// Construct absolute paths *within* the sandbox for setup
		subDirAbs := filepath.Join(sandboxRoot, subDirRel)
		file1Abs := filepath.Join(sandboxRoot, file1Rel)
		file2Abs := filepath.Join(sandboxRoot, file2Rel)
		rootFileAbs := filepath.Join(sandboxRoot, rootFileRel)

		// Create directories using absolute paths
		if err := os.MkdirAll(subDirAbs, 0755); err != nil {
			return fmt.Errorf("setup MkdirAll failed for %s: %w", subDirAbs, err)
		}
		// Create files using absolute paths
		if err := os.WriteFile(file1Abs, []byte(file1Content), 0644); err != nil {
			return fmt.Errorf("setup WriteFile failed for %s: %w", file1Abs, err)
		}
		if err := os.WriteFile(file2Abs, []byte(file2Content), 0644); err != nil {
			return fmt.Errorf("setup WriteFile failed for %s: %w", file2Abs, err)
		}
		if err := os.WriteFile(rootFileAbs, []byte(rootFileContent), 0644); err != nil {
			return fmt.Errorf("setup WriteFile failed for %s: %w", rootFileAbs, err)
		}
		return nil
	}

	// --- Expected Size Info ---
	// Use actual content length for size verification
	file1Size := int64(len(file1Content))
	file2Size := int64(len(file2Content))
	rootFileSize := int64(len(rootFileContent))
	var dummyDirSize int64 = 4096 // Placeholder for directory size

	tests := []fsTestCase{
		{
			name:      "List root of test dir",
			toolName:  "ListDirectory",
			args:      makeArgs(baseDirRel), // Use relative path for tool argument
			setupFunc: setupListDirTest,     // Pass the setup function
			wantResult: []interface{}{ // Use relative names in expected results, actual sizes
				map[string]interface{}{"name": "file1.txt", "is_dir": false, "size": file1Size},
				map[string]interface{}{"name": "sub", "is_dir": true, "size": dummyDirSize},
			},
		},
		{
			name:      "List sub directory",
			toolName:  "ListDirectory",
			args:      makeArgs(subDirRel), // Use relative path for tool argument
			setupFunc: setupListDirTest,
			wantResult: []interface{}{
				map[string]interface{}{"name": "file2.txt", "is_dir": false, "size": file2Size},
			},
		},
		{
			name:      "List sandbox root",
			toolName:  "ListDirectory",
			args:      makeArgs("."), // Use "." for current (sandbox root)
			setupFunc: setupListDirTest,
			wantResult: []interface{}{ // Order will be sorted by helper verification
				map[string]interface{}{"name": baseDirRel, "is_dir": true, "size": dummyDirSize},
				map[string]interface{}{"name": rootFileRel, "is_dir": false, "size": rootFileSize},
			},
		},
		{
			name:          "List non-existent dir",
			toolName:      "ListDirectory",
			args:          makeArgs(filepath.Join(baseDirRel, "nonexistent")),
			setupFunc:     setupListDirTest,
			wantResult:    nil,             // No specific result expected on error
			wantToolErrIs: ErrInternalTool, // Expect wrapped OS error
			valWantErrIs:  nil,
		},
		{
			name:         "Validation_Wrong_Arg_Type",
			toolName:     "ListDirectory",
			args:         makeArgs(123),
			wantResult:   nil,
			valWantErrIs: ErrValidationTypeMismatch,
		},
		{
			name:          "Path_Outside_Sandbox",
			toolName:      "ListDirectory",
			args:          makeArgs("../outside"),
			setupFunc:     setupListDirTest,
			wantResult:    nil,
			wantToolErrIs: ErrPathViolation, // Expect path violation error from tool
			valWantErrIs:  nil,
		},
		{
			name:          "Path_Is_File",
			toolName:      "ListDirectory",
			args:          makeArgs(file1Rel), // Try to list a file
			setupFunc:     setupListDirTest,
			wantResult:    nil,
			wantToolErrIs: ErrInternalTool, // Expect wrapped OS error (e.g., "not a directory")
			valWantErrIs:  nil,
		},
	}

	for _, tt := range tests {
		// Pass interp and tt to the helper
		testFsToolHelper(t, interp, tt)
	}
}
