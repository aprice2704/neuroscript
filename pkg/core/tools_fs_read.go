// NeuroScript Version: 0.3.0
// File version: 0.1.3
// Refine error handling for "is a directory" and ensure correct sentinel errors.
// filename: pkg/core/tools_fs_read.go

package core

import (
	"fmt"
	"os"
	"strings" // Added for checking error messages
)

// toolReadFile implements TOOL.ReadFile
func toolReadFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("ReadFile expects 1 argument, got %d", len(args)), ErrInvalidArgument)
	}
	filepathRelative, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "ReadFile expects a string filepath argument", ErrInvalidArgument)
	}

	if filepathRelative == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "ReadFile filepath cannot be empty", ErrInvalidArgument)
	}

	// Resolve and secure the path using the interpreter's FileAPI
	absPath, err := interpreter.FileAPI().ResolvePath(filepathRelative)
	if err != nil {
		// ResolvePath already creates a RuntimeError and wraps appropriate sentinel errors (e.g., ErrPathViolation, ErrInvalidPath)
		return nil, err
	}

	interpreter.Logger().Debugf("Tool ReadFile: Reading file at resolved absolute path: %s (original: %s)", absPath, filepathRelative)

	contentBytes, readErr := os.ReadFile(absPath)
	if readErr != nil {
		interpreter.Logger().Errorf("Tool ReadFile: Error reading file '%s' (resolved: '%s'): %v", filepathRelative, absPath, readErr)
		if os.IsNotExist(readErr) {
			// For "file not found", wrap ErrFileNotFound
			return nil, NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("file not found at path '%s'", filepathRelative), ErrFileNotFound)
		}
		// Check if the error is because the path is a directory
		// This is a common way to check, though not perfectly robust across all OS/Go versions, it's standard practice.
		if strings.Contains(readErr.Error(), "is a directory") {
			errMsg := fmt.Sprintf("path '%s' is a directory, not a file", filepathRelative)
			return nil, NewRuntimeError(ErrorCodeArgMismatch, errMsg, ErrInvalidArgument)
		}
		// For other read errors, use ErrorCodeToolSpecific and wrap the original error
		return nil, NewRuntimeError(ErrorCodeToolSpecific, fmt.Sprintf("error reading file '%s': %v", filepathRelative, readErr), readErr)
	}

	interpreter.Logger().Infof("Tool ReadFile: Successfully read %d bytes from '%s'", len(contentBytes), filepathRelative)
	return string(contentBytes), nil
}
