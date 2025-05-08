// NeuroScript Version: 0.3.1
// File version: 0.0.4 // Return detailed success message string on success.
// nlines: 80 // Approximate
// risk_rating: HIGH // Writes files
// filename: pkg/core/tools_fs_write.go
package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// toolWriteFile implements TOOL.WriteFile.
// It creates parent directories if they don't exist.
func toolWriteFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("WriteFile: expected 2 arguments (filepath, content), got %d", len(args)), ErrArgumentMismatch)
	}

	relPath, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("WriteFile: filepath argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	contentArg := args[1] // Handle nil explicitly below
	content := ""         // Default to empty string

	// Allow nil content, treat as empty string
	if contentArg == nil {
		content = ""
	} else if contentStr, okStr := contentArg.(string); okStr {
		content = contentStr
	} else {
		// If not nil and not string, it's an error
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("WriteFile: content argument must be a string or nil, got %T", args[1]), ErrInvalidArgument)
	}

	if relPath == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "WriteFile: filepath argument cannot be empty", ErrInvalidArgument)
	}

	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		interpreter.Logger().Error("Tool: WriteFile] Interpreter sandboxDir is empty, cannot proceed.")
		return nil, NewRuntimeError(ErrorCodeConfiguration, "WriteFile: interpreter sandbox directory is not set", ErrConfiguration)
	}

	absPath, secErr := ResolveAndSecurePath(relPath, sandboxRoot)
	if secErr != nil {
		interpreter.Logger().Warn("Tool: WriteFile path validation failed", "relative_path", relPath, "sandbox_root", sandboxRoot, "error", secErr)
		return "", secErr
	}

	interpreter.Logger().Debug("Tool: WriteFile attempting to write", "validated_path", absPath, "original_relative_path", relPath, "sandbox_root", sandboxRoot)
	parentDir := filepath.Dir(absPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		errMsg := fmt.Sprintf("WriteFile: failed to create parent directory for '%s'", relPath)
		interpreter.Logger().Error(errMsg, "error", err)
		return "", NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrCannotCreateDir, err))
	}

	err := os.WriteFile(absPath, []byte(content), 0644)
	if err != nil {
		errMsg := fmt.Sprintf("WriteFile: failed to write file '%s'", relPath)
		interpreter.Logger().Error(errMsg, "error", err)
		if errors.Is(err, os.ErrPermission) {
			return "", NewRuntimeError(ErrorCodePermissionDenied, errMsg, ErrPermissionDenied)
		}
		info, statErr := os.Stat(absPath)
		if statErr == nil && info.IsDir() {
			errMsg = fmt.Sprintf("WriteFile: path '%s' exists and is a directory", relPath)
			interpreter.Logger().Info(errMsg)
			return "", NewRuntimeError(ErrorCodePathTypeMismatch, errMsg, ErrPathNotFile)
		}
		return "", NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrIOFailed, err))
	}

	// Success
	bytesWritten := len([]byte(content))
	// *** ENSURE detailed success message is returned ***
	successMsg := fmt.Sprintf("Successfully wrote %d bytes to %s", bytesWritten, relPath)
	interpreter.Logger().Info("Tool: WriteFile successful", "file_path", relPath, "bytes_written", bytesWritten)
	return successMsg, nil // Return the formatted string
}
