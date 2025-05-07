// NeuroScript Version: 0.3.0
// File version: 0.1.2 // Corrected ErrorCode for "directory not empty"
// filename: pkg/core/tools_fs_delete.go

package core

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// toolDeleteFile implements the TOOL.DeleteFile command.
func toolDeleteFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("DeleteFile expects 1 argument, got %d", len(args)), ErrInvalidArgument)
	}
	relPath, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "DeleteFile expects a string path argument", ErrInvalidArgument)
	}
	if relPath == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "DeleteFile path cannot be empty", ErrInvalidArgument)
	}

	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		interpreter.Logger().Error("Tool: DeleteFile] Interpreter sandboxDir is empty, cannot proceed.")
		return nil, NewRuntimeError(ErrorCodeConfiguration, "interpreter sandbox directory is not set", ErrConfiguration)
	}

	absPath, secErr := SecureFilePath(relPath, sandboxRoot)
	if secErr != nil {
		interpreter.Logger().Infof("Tool: DeleteFile] Path security error for %q: %v (Sandbox Root: %s)", relPath, secErr, sandboxRoot)
		return nil, secErr
	}

	interpreter.Logger().Infof("Tool: DeleteFile] Validated path: %s. Attempting deletion.", absPath)

	err := os.Remove(absPath)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			errMsg := fmt.Sprintf("File or directory not found: %s", relPath)
			interpreter.Logger().Infof("Tool: DeleteFile] Info: %s", errMsg)
			return "OK", nil
		}

		errMsgTextLower := strings.ToLower(err.Error())
		isDirNotEmptyErr := strings.Contains(errMsgTextLower, "directory not empty")

		errMsg := fmt.Sprintf("Failed to delete '%s'", relPath)
		interpreter.Logger().Errorf("Tool: DeleteFile] Error: %s: %v", errMsg, err)

		if isDirNotEmptyErr {
			// Corrected: Use ErrorCodePreconditionFailed as "directory not empty" is a failed precondition.
			// Wrap the original OS error and the sentinel ErrCannotDelete for context.
			return nil, NewRuntimeError(ErrorCodePreconditionFailed, errMsg, errors.Join(ErrCannotDelete, err))
		}

		return nil, NewRuntimeError(ErrorCodeInternal, errMsg, err)
	}

	successMsg := fmt.Sprintf("Successfully deleted: %s", relPath)
	interpreter.Logger().Infof("Tool: DeleteFile] %s", successMsg)
	return "OK", nil
}
