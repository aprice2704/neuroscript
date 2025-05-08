// NeuroScript Version: 0.3.1
// File version: 0.0.2 // Corrected NewRuntimeError calls with standard ErrorCodes/Sentinels.
// nlines: 67
// risk_rating: HIGH
// filename: pkg/core/tools_fs_write.go
package core

import (
	"errors" // Required for errors.Join
	"fmt"
	"os"
	"path/filepath"
)

// toolWriteFile writes content to a specified file within the sandbox.
// It creates parent directories if they don't exist.
// Returns "OK" on success, or an error on failure.
// Its ToolImplementation is defined in tooldefs_fs.go.
func toolWriteFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return "", NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("WriteFile: expected 2 arguments (filepath, content), got %d", len(args)), ErrArgumentMismatch)
	}
	filePath, pathOk := args[0].(string)
	content, contentOk := args[1].(string)

	if !pathOk {
		// Using ErrorCodeType for wrong type, wrapping ErrInvalidArgument
		return "", NewRuntimeError(ErrorCodeType, fmt.Sprintf("WriteFile: filepath argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	if !contentOk {
		// Using ErrorCodeType for wrong type, wrapping ErrInvalidArgument
		return "", NewRuntimeError(ErrorCodeType, fmt.Sprintf("WriteFile: content argument must be a string, got %T", args[1]), ErrInvalidArgument)
	}

	if filePath == "" {
		// Empty path is treated as an invalid argument value.
		return "", NewRuntimeError(ErrorCodeArgMismatch, "WriteFile: filepath cannot be empty", ErrInvalidArgument)
	}

	sandboxRoot := interpreter.SandboxDir()
	absPath, secErr := SecureFilePath(filePath, sandboxRoot)
	if secErr != nil {
		interpreter.Logger().Warn("Tool: WriteFile path validation failed", "relative_path", filePath, "sandbox_root", sandboxRoot, "error", secErr)
		// Directly return the RuntimeError from SecureFilePath
		return "", secErr
	}

	interpreter.Logger().Debug("Tool: WriteFile attempting to write", "validated_path", absPath, "original_relative_path", filePath, "sandbox_root", sandboxRoot)

	// Create parent directories if they don't exist
	dir := filepath.Dir(absPath)
	if mkDirErr := os.MkdirAll(dir, 0755); mkDirErr != nil { // Permissions 0755 are common for directories
		errMsg := fmt.Sprintf("WriteFile: could not create directories for '%s'", filePath)
		interpreter.Logger().Error("Tool: WriteFile MkdirAll failed", "path", dir, "error", mkDirErr)
		// Use ErrorCodeIOFailed as it's an OS-level I/O issue, wrap ErrCannotCreateDir sentinel for context.
		return "", NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrCannotCreateDir, mkDirErr))
	}

	// Write the file
	writeErr := os.WriteFile(absPath, []byte(content), 0644) // Permissions 0644 are common for files
	if writeErr != nil {
		errMsg := fmt.Sprintf("WriteFile: could not write to file '%s'", filePath)
		interpreter.Logger().Error("Tool: WriteFile failed", "path", absPath, "error", writeErr)
		// Use ErrorCodeIOFailed and wrap ErrIOFailed sentinel + original error.
		return "", NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrIOFailed, writeErr))
	}

	interpreter.Logger().Info("Tool: WriteFile successful", "file_path", filePath, "bytes_written", len(content))
	// Return "OK" string literal on success as defined in tooldefs_fs.go spec.
	return "OK", nil
}
