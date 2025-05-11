// NeuroScript Version: 0.3.1
// File version: 0.0.3 // Add explicit empty path check. Handle "is a directory" error.
// nlines: 70 // Approximate
// risk_rating: MEDIUM
// filename: pkg/core/tools_fs_read.go
package core

import (
	"errors"
	"fmt"
	"os"
	"strings" // For checking "is a directory" error string
)

// toolReadFile implements the TOOL.ReadFile command.
func toolReadFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("ReadFile: expected 1 argument (filepath), got %d", len(args)), ErrArgumentMismatch)
	}

	relPath, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("ReadFile: filepath argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}

	// *** ADDED: Explicit check for empty path ***
	if relPath == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "ReadFile: filepath argument cannot be empty", ErrInvalidArgument)
	}

	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		interpreter.Logger().Error("Tool: ReadFile] Interpreter sandboxDir is empty, cannot proceed.")
		return nil, NewRuntimeError(ErrorCodeConfiguration, "ReadFile: interpreter sandbox directory is not set", ErrConfiguration)
	}

	// Use ResolveAndSecurePath which handles various security checks
	absPath, secErr := ResolveAndSecurePath(relPath, sandboxRoot)
	if secErr != nil {
		interpreter.Logger().Warn("Tool: ReadFile path validation failed", "relative_path", relPath, "sandbox_root", sandboxRoot, "error", secErr)
		return "", secErr // Return empty string and the error
	}

	interpreter.Logger().Debug("Tool: ReadFile attempting to read", "validated_path", absPath, "original_relative_path", relPath, "sandbox_root", sandboxRoot)

	// Read the file content
	contentBytes, err := os.ReadFile(absPath)
	if err != nil {
		// Handle specific errors
		if errors.Is(err, os.ErrNotExist) {
			errMsg := fmt.Sprintf("ReadFile: file not found '%s'", relPath)
			interpreter.Logger().Debug(errMsg)
			return "", NewRuntimeError(ErrorCodeFileNotFound, errMsg, ErrFileNotFound) // Return empty string and error
		}
		if errors.Is(err, os.ErrPermission) {
			errMsg := fmt.Sprintf("ReadFile: permission denied for '%s'", relPath)
			interpreter.Logger().Warn(errMsg)
			return "", NewRuntimeError(ErrorCodePermissionDenied, errMsg, ErrPermissionDenied) // Return empty string and error
		}

		// *** ADDED: Check for "is a directory" error ***
		// This error isn't standard, so check the message content
		// Note: This might be OS-dependent.
		if strings.Contains(err.Error(), "is a directory") {
			errMsg := fmt.Sprintf("ReadFile: path '%s' is a directory, not a file", relPath)
			interpreter.Logger().Debug(errMsg)
			// Use ErrPathNotFile sentinel error
			return "", NewRuntimeError(ErrorCodePathTypeMismatch, errMsg, ErrPathNotFile) // Return empty string and error
		}

		// Handle other potential I/O errors
		errMsg := fmt.Sprintf("ReadFile: failed to read file '%s'", relPath)
		interpreter.Logger().Error(errMsg, "error", err)
		return "", NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrIOFailed, err)) // Return empty string and error
	}

	// Success
	content := string(contentBytes)
	interpreter.Logger().Debug("Tool: ReadFile successful", "file_path", relPath, "bytes_read", len(contentBytes))
	return content, nil
}
