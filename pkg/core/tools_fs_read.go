// NeuroScript Version: 0.3.1
// File version: 0.0.2 // Corrected NewRuntimeError calls with standard ErrorCodes/Sentinels.
// nlines: 56
// risk_rating: MEDIUM
// filename: pkg/core/tools_fs_read.go
package core

import (
	"errors" // Required for errors.Is and errors.Join
	"fmt"
	"os"
	// Keep for checking "is a directory" error potentially
)

// toolReadFile reads the entire content of a specified file within the sandbox.
// Returns the file content as a string, or an empty string and error on failure.
func toolReadFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return "", NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("ReadFile: expected 1 argument (filepath), got %d", len(args)), ErrArgumentMismatch)
	}
	filePath, ok := args[0].(string)
	if !ok {
		// Using ErrorCodeType for wrong type, wrapping ErrInvalidArgument
		return "", NewRuntimeError(ErrorCodeType, fmt.Sprintf("ReadFile: filepath argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	if filePath == "" {
		// Empty path is treated as an invalid argument value.
		return "", NewRuntimeError(ErrorCodeArgMismatch, "ReadFile: filepath cannot be empty", ErrInvalidArgument)
	}

	sandboxRoot := interpreter.SandboxDir()
	absPath, secErr := SecureFilePath(filePath, sandboxRoot)
	if secErr != nil {
		interpreter.Logger().Warn("Tool: ReadFile path validation failed", "relative_path", filePath, "sandbox_root", sandboxRoot, "error", secErr)
		// Directly return the RuntimeError from SecureFilePath
		return "", secErr
	}

	interpreter.Logger().Debug("Tool: ReadFile attempting to read", "validated_path", absPath, "original_relative_path", filePath, "sandbox_root", sandboxRoot)

	// Check if the path is actually a directory *before* trying to read
	info, statErr := os.Stat(absPath)
	if statErr != nil {
		// Handle stat errors (like not found, permission) before read attempt
		if errors.Is(statErr, os.ErrNotExist) {
			return "", NewRuntimeError(ErrorCodeFileNotFound, fmt.Sprintf("ReadFile: file not found '%s'", filePath), ErrFileNotFound)
		}
		if errors.Is(statErr, os.ErrPermission) {
			return "", NewRuntimeError(ErrorCodePermissionDenied, fmt.Sprintf("ReadFile: permission denied for '%s'", filePath), ErrPermissionDenied)
		}
		// Other stat errors
		return "", NewRuntimeError(ErrorCodeIOFailed, fmt.Sprintf("ReadFile: failed to stat path '%s'", filePath), errors.Join(ErrIOFailed, statErr))
	}
	if info.IsDir() {
		// Use the specific code and sentinel for path type mismatch
		return "", NewRuntimeError(ErrorCodePathTypeMismatch, fmt.Sprintf("ReadFile: path '%s' is a directory, not a file", filePath), ErrPathNotFile)
	}

	// Proceed with reading the file
	contentBytes, readErr := os.ReadFile(absPath)
	if readErr != nil {
		// Should not happen if Stat succeeded, but handle defensively
		interpreter.Logger().Warn("Tool: ReadFile os.ReadFile failed unexpectedly after Stat succeeded", "path", filePath, "error", readErr)
		if errors.Is(readErr, os.ErrPermission) { // Check permissions again in case they changed
			return "", NewRuntimeError(ErrorCodePermissionDenied, fmt.Sprintf("ReadFile: permission denied for '%s'", filePath), ErrPermissionDenied)
		}
		// General I/O error during read
		return "", NewRuntimeError(ErrorCodeIOFailed, fmt.Sprintf("ReadFile: error reading file '%s'", filePath), errors.Join(ErrIOFailed, readErr))
	}

	interpreter.Logger().Info("Tool: ReadFile successful", "file_path", filePath, "bytes_read", len(contentBytes))
	return string(contentBytes), nil
}
