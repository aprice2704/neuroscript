// filename: pkg/core/tools_fs_dirs_test.go
package core

import (
	// Keep errors
	"fmt"           // Keep fmt
	"os"            // Keep os
	"path/filepath" // Keep filepath
	"testing"
)

// Assume testFsToolHelper is defined in tools_fs_helpers_test.go

func TestToolMkdir(t *testing.T) {
	interp, _ := newDefaultTestInterpreter(t) // Get interpreter for sandbox path

	// --- Test Setup Data ---
	newDirPathRel := "newDir"
	existingDirPathRel := "existingDir"
	nestedPathRel := filepath.Join("parentDir", "childDir")
	filePathRel := "existingFile.txt" // To test creating dir where file exists

	// --- Setup Function ---
	// *** MODIFIED: Takes sandboxRoot string argument and uses it ***
	setupMkdirTest := func(sandboxRoot string) error {
		// Construct absolute paths within the sandbox for setup
		existingDirAbs := filepath.Join(sandboxRoot, existingDirPathRel)
		fileAbs := filepath.Join(sandboxRoot, filePathRel)

		// Create existing directory using absolute path
		if err := os.Mkdir(existingDirAbs, 0755); err != nil && !os.IsExist(err) {
			return fmt.Errorf("setup Mkdir failed for %s: %w", existingDirAbs, err)
		}
		// Create existing file using absolute path
		if err := os.WriteFile(fileAbs, []byte("i am a file"), 0644); err != nil {
			return fmt.Errorf("setup WriteFile failed for %s: %w", fileAbs, err)
		}
		return nil
	}

	tests := []fsTestCase{
		{
			name:       "Create New Directory",
			toolName:   "Mkdir",
			args:       makeArgs(newDirPathRel),
			setupFunc:  setupMkdirTest, // Setup existing stuff
			wantResult: "OK",
			// Verification should ideally happen in helper or a verifyFunc
		},
		{
			name:       "Create Nested Directories",
			toolName:   "Mkdir",
			args:       makeArgs(nestedPathRel),
			setupFunc:  setupMkdirTest,
			wantResult: "OK",
		},
		{
			name:       "Create Existing Directory", // MkdirAll is idempotent
			toolName:   "Mkdir",
			args:       makeArgs(existingDirPathRel),
			setupFunc:  setupMkdirTest,
			wantResult: "OK",
		},
		{
			name:          "Create Directory Where File Exists",
			toolName:      "Mkdir",
			args:          makeArgs(filePathRel), // Path of existing file
			setupFunc:     setupMkdirTest,
			wantResult:    fmt.Sprintf("Mkdir failed for '%s': mkdir %s: not a directory", filePathRel, filepath.Join(interp.sandboxDir, filePathRel)), // Expect specific error message
			wantToolErrIs: ErrCannotCreateDir,
		},
		{
			name:         "Validation_Wrong_Arg_Type",
			toolName:     "Mkdir",
			args:         makeArgs(12345),
			valWantErrIs: ErrValidationTypeMismatch,
		},
		{
			name:          "Path_Outside_Sandbox",
			toolName:      "Mkdir",
			args:          makeArgs("../someDir"),
			setupFunc:     setupMkdirTest, // Setup existing stuff
			wantResult:    fmt.Sprintf("Mkdir path error for '../someDir': %s: relative path '../someDir' resolves to '%s' which is outside the allowed directory '%s'", ErrPathViolation.Error(), filepath.Clean(filepath.Join(interp.sandboxDir, "../someDir")), interp.sandboxDir),
			wantToolErrIs: ErrPathViolation,
		},
		{
			name:         "Validation_Missing_Arg",
			toolName:     "Mkdir",
			args:         makeArgs(),
			valWantErrIs: ErrValidationArgCount,
		},
	}

	for _, tt := range tests {
		testFsToolHelper(t, interp, tt)
		// Add manual verification step if testFsToolHelper doesn't cover it
		if tt.wantToolErrIs == nil && tt.valWantErrIs == nil {
			// Verify directory was actually created using absolute path
			dirPathAbs := filepath.Join(interp.sandboxDir, tt.args[0].(string))
			info, err := os.Stat(dirPathAbs)
			if err != nil {
				t.Errorf("Test '%s': Failed to stat expected directory '%s': %v", tt.name, dirPathAbs, err)
			} else if !info.IsDir() {
				t.Errorf("Test '%s': Expected path '%s' to be a directory, but it's not.", tt.name, dirPathAbs)
			}
		}
	}
}
