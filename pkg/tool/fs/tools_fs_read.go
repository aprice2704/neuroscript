// NeuroScript Version: 0.3.1
// File version: 0.0.3 // Add explicit empty path check. Handle "is a directory" error.
// nlines: 70 // Approximate
// risk_rating: MEDIUM
// filename: pkg/tool/fs/tools_fs_read.go
package fs

import (
	"errors"
	"fmt"
	"os"
	"strings" // For checking "is a directory" error string

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/security"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// toolReadFile implements the TOOL.ReadFile command.
func toolReadFile(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("ReadFile: expected 1 argument (filepath), got %d", len(args)), lang.ErrArgumentMismatch)
	}

	relPath, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("ReadFile: filepath argument must be a string, got %T", args[0]), lang.ErrInvalidArgument)
	}

	// *** ADDED: Explicit check for empty path ***
	if relPath == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "ReadFile: filepath argument cannot be empty", lang.ErrInvalidArgument)
	}

	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		interpreter.GetLogger().Error("Tool: ReadFile] Interpreter sandboxDir is empty, cannot proceed.")
		return nil, lang.NewRuntimeError(lang.ErrorCodeConfiguration, "ReadFile: interpreter sandbox directory is not set", lang.ErrConfiguration)
	}

	// Use ResolveAndSecurePath which handles various security checks
	absPath, secErr := security.ResolveAndSecurePath(relPath, sandboxRoot)
	if secErr != nil {
		interpreter.GetLogger().Warn("Tool: ReadFile path validation failed", "relative_path", relPath, "sandbox_root", sandboxRoot, "error", secErr)
		return "", secErr // Return empty string and the error
	}

	interpreter.GetLogger().Debug("Tool: ReadFile attempting to read", "validated_path", absPath, "original_relative_path", relPath, "sandbox_root", sandboxRoot)

	// Read the file content
	contentBytes, err := os.ReadFile(absPath)
	if err != nil {
		// Handle specific errors
		if errors.Is(err, os.ErrNotExist) {
			errMsg := fmt.Sprintf("ReadFile: file not found '%s'", relPath)
			interpreter.GetLogger().Debug(errMsg)
			return "", lang.NewRuntimeError(lang.ErrorCodeFileNotFound, errMsg, lang.ErrFileNotFound) // Return empty string and error
		}
		if errors.Is(err, os.ErrPermission) {
			errMsg := fmt.Sprintf("ReadFile: permission denied for '%s'", relPath)
			interpreter.GetLogger().Warn(errMsg)
			return "", lang.NewRuntimeError(lang.ErrorCodePermissionDenied, errMsg, lang.ErrPermissionDenied) // Return empty string and error
		}

		// *** ADDED: Check for "is a directory" error ***
		if strings.Contains(err.Error(), "is a directory") {
			errMsg := fmt.Sprintf("ReadFile: path '%s' is a directory, not a file", relPath)
			interpreter.GetLogger().Debug(errMsg)
			return "", lang.NewRuntimeError(lang.ErrorCodePathTypeMismatch, errMsg, lang.ErrPathNotFile)
		}

		// Handle other potential I/O errors
		errMsg := fmt.Sprintf("ReadFile: failed to read file '%s'", relPath)
		interpreter.GetLogger().Error(errMsg, "error", err)
		return "", lang.NewRuntimeError(lang.ErrorCodeIOFailed, errMsg, errors.Join(lang.ErrIOFailed, err))
	}

	// Success
	content := string(contentBytes)
	interpreter.GetLogger().Debug("Tool: ReadFile successful", "file_path", relPath, "bytes_read", len(contentBytes))
	return content, nil
}
