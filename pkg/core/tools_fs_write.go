// NeuroScript Version: 0.3.0
// File version: 0.1.6 // Adjusted error wrapping for internal OS errors to align with test expectations.
// filename: pkg/core/tools_fs_write.go

package core

import (
	"errors" // Added for errors.Join
	"fmt"
	"os"
	"path/filepath" // For ensuring directory exists
)

// toolWriteFile implements TOOL.WriteFile
// Its ToolImplementation is now defined in tooldefs_fs.go and registered by zz_core_tools_registrar.go
func toolWriteFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("WriteFile expects 2 arguments, got %d", len(args)), ErrInvalidArgument)
	}
	filepathRelative, okPath := args[0].(string)
	content, okContent := args[1].(string)

	if !okPath {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "WriteFile expects a string filepath argument", ErrInvalidArgument)
	}
	if !okContent {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "WriteFile expects string content argument", ErrInvalidArgument)
	}

	if filepathRelative == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "WriteFile filepath cannot be empty", ErrInvalidArgument)
	}

	// Resolve and secure the path using the interpreter's FileAPI
	absPath, err := interpreter.FileAPI().ResolvePath(filepathRelative)
	if err != nil {
		// ResolvePath already creates a RuntimeError
		return nil, err
	}

	interpreter.Logger().Debugf("Tool WriteFile: Attempting to write to resolved absolute path: %s (original relative: %s)", absPath, filepathRelative)

	// Ensure the directory exists
	dir := filepath.Dir(absPath)
	if mkDirErr := os.MkdirAll(dir, 0755); mkDirErr != nil { // 0755 are typical directory permissions
		errMsg := fmt.Sprintf("failed to create directory '%s' for file '%s'", dir, filepathRelative)
		interpreter.Logger().Errorf("Tool WriteFile: %s: %v", errMsg, mkDirErr)
		// Wrap ErrInternalTool to satisfy test expectation
		return nil, NewRuntimeError(ErrorCodeInternal, errMsg, errors.Join(ErrInternalTool, mkDirErr))
	}

	// Write the file
	writeErr := os.WriteFile(absPath, []byte(content), 0644) // 0644 are typical file permissions
	if writeErr != nil {
		errMsg := fmt.Sprintf("failed writing to file '%s'", filepathRelative)
		interpreter.Logger().Errorf("Tool WriteFile: %s (resolved: '%s'): %v", errMsg, absPath, writeErr)
		// Wrap ErrInternalTool to satisfy test expectation
		return nil, NewRuntimeError(ErrorCodeInternal, errMsg, errors.Join(ErrInternalTool, writeErr))
	}

	interpreter.Logger().Infof("Tool WriteFile: Successfully wrote %d bytes to '%s'", len(content), filepathRelative)
	return "OK", nil // CRITICAL: Ensure this returns "OK"
}
