// NeuroScript Version: 0.3.1
// File version: 0.0.2 // Corrected lang.NewRuntimeError calls with standard ErrorCodes/Sentinels.
// nlines: 62
// risk_rating: HIGH
// filename: pkg/tool/fs/tools_fs_delete.go
package fs

import (
	"errors"
	"fmt"
	"os"
	"strings"	// Keep for "directory not empty" check if needed, though errors.Is might be better if a specific error exists.

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// toolDeleteFile implements the TOOL.DeleteFile command.
// It deletes a file or an *empty* directory.
func toolDeleteFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("DeleteFile: expected 1 argument (path), got %d", len(args)), ErrArgumentMismatch)
	}
	relPath, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(ErrorCodeType, fmt.Sprintf("DeleteFile: path argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	if relPath == "" {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "DeleteFile: path cannot be empty", ErrInvalidArgument)
	}

	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		interpreter.Logger().Error("Tool: DeleteFile] Interpreter sandboxDir is empty, cannot proceed.")
		return nil, lang.NewRuntimeError(ErrorCodeConfiguration, "DeleteFile: interpreter sandbox directory is not set", ErrConfiguration)
	}

	absPath, secErr := SecureFilePath(relPath, sandboxRoot)
	if secErr != nil {
		interpreter.Logger().Infof("Tool: DeleteFile] Path security error for %q: %v (Sandbox Root: %s)", relPath, secErr, sandboxRoot)
		return nil, secErr	// SecureFilePath returns RuntimeError
	}

	interpreter.Logger().Infof("Tool: DeleteFile] Validated path: %s. Attempting deletion.", absPath)

	// Attempt removal
	err := os.Remove(absPath)

	if err != nil {
		// If the file/dir doesn't exist, treat it as success (idempotent delete)
		if errors.Is(err, os.ErrNotExist) {
			errMsg := fmt.Sprintf("Path not found, nothing to delete: %s", relPath)
			interpreter.Logger().Infof("Tool: DeleteFile] Info: %s", errMsg)
			return "OK", nil	// Return "OK" as per spec
		}

		// Check if it's a "directory not empty" error
		// Note: This check might vary slightly across OSes. Go doesn't have a standard os.ErrDirNotEmpty.
		// Using string check is common but potentially brittle.
		errMsgTextLower := ""
		if err != nil {
			errMsgTextLower = strings.ToLower(err.Error())
		}
		isDirNotEmptyErr := strings.Contains(errMsgTextLower, "directory not empty") || strings.Contains(errMsgTextLower, "not empty")	// Common variations

		errMsg := fmt.Sprintf("Failed to delete '%s'", relPath)
		interpreter.Logger().Errorf("Tool: DeleteFile] Error: %s: %v", errMsg, err)

		if isDirNotEmptyErr {
			// Use ErrorCodePreconditionFailed for "directory not empty"
			return nil, lang.NewRuntimeError(ErrorCodePreconditionFailed, errMsg+": directory not empty", errors.Join(ErrCannotDelete, err))
		}

		// Check for permission errors specifically
		if errors.Is(err, os.ErrPermission) {
			return nil, lang.NewRuntimeError(ErrorCodePermissionDenied, errMsg, ErrPermissionDenied)
		}

		// For other errors, use ErrorCodeIOFailed
		return nil, lang.NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrIOFailed, err))
	}

	successMsg := fmt.Sprintf("Successfully deleted: %s", relPath)
	interpreter.Logger().Infof("Tool: DeleteFile] %s", successMsg)
	// Return "OK" string literal on success
	return "OK", nil
}