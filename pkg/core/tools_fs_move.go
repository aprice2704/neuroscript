// NeuroScript Version: 0.3.0
// File version: 0.1.2 // Corrected ReturnType and success value for functional test
// filename: pkg/core/tools_fs_move.go

package core

import (
	"errors"
	"fmt"
	"os"
	// "path/filepath" // Not strictly needed if os.Rename handles cross-dir within sandbox well
)

// toolMoveFile implements the TOOL.MoveFile command.
func toolMoveFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("MoveFile expects 2 arguments (source, destination), got %d", len(args)), ErrInvalidArgument)
	}
	sourcePathRel, okSrc := args[0].(string)
	destPathRel, okDest := args[1].(string)

	if !okSrc {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "MoveFile source path must be a string", ErrInvalidArgument)
	}
	if !okDest {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "MoveFile destination path must be a string", ErrInvalidArgument)
	}
	if sourcePathRel == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "MoveFile source path cannot be empty", ErrInvalidArgument)
	}
	if destPathRel == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "MoveFile destination path cannot be empty", ErrInvalidArgument)
	}
	if sourcePathRel == destPathRel {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "MoveFile source and destination paths cannot be the same", ErrInvalidArgument)
	}

	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		interpreter.Logger().Error("Tool: MoveFile] Interpreter sandboxDir is empty, cannot proceed.")
		return nil, NewRuntimeError(ErrorCodeConfiguration, "interpreter sandbox directory is not set", ErrConfiguration)
	}

	absSource, errSource := SecureFilePath(sourcePathRel, sandboxRoot)
	if errSource != nil {
		interpreter.Logger().Infof("Tool: MoveFile] Invalid source path '%s': %v", sourcePathRel, errSource)
		// Return map with error for consistency with test expectations on error
		return map[string]interface{}{"error": errSource.Error()}, errSource
	}

	absDest, errDest := SecureFilePath(destPathRel, sandboxRoot)
	if errDest != nil {
		interpreter.Logger().Infof("Tool: MoveFile] Invalid destination path '%s': %v", destPathRel, errDest)
		return map[string]interface{}{"error": errDest.Error()}, errDest
	}

	interpreter.Logger().Infof("Tool: MoveFile] Validated paths: Source '%s' (abs: '%s'), Dest '%s' (abs: '%s')", sourcePathRel, absSource, destPathRel, absDest)

	_, srcStatErr := os.Stat(absSource)
	if srcStatErr != nil {
		errMsg := ""
		var rtErr *RuntimeError
		if errors.Is(srcStatErr, os.ErrNotExist) {
			errMsg = fmt.Sprintf("source path '%s' does not exist", sourcePathRel)
			rtErr = NewRuntimeError(ErrorCodeKeyNotFound, errMsg, srcStatErr)
		} else {
			errMsg = fmt.Sprintf("error checking source path '%s'", sourcePathRel)
			rtErr = NewRuntimeError(ErrorCodeInternal, errMsg, srcStatErr)
		}
		interpreter.Logger().Infof("Tool: MoveFile] Error: %s (resolved: %s)", errMsg, absSource)
		return map[string]interface{}{"error": errMsg}, rtErr
	}

	_, destStatErr := os.Stat(absDest)
	if destStatErr == nil {
		errMsg := fmt.Sprintf("destination path '%s' already exists", destPathRel)
		interpreter.Logger().Infof("Tool: MoveFile] Error: %s (resolved: %s)", errMsg, absDest)
		return map[string]interface{}{"error": errMsg}, NewRuntimeError(ErrorCodePreconditionFailed, errMsg, ErrCannotCreateDir)
	} else if !errors.Is(destStatErr, os.ErrNotExist) {
		errMsg := fmt.Sprintf("error checking destination path '%s'", destPathRel)
		interpreter.Logger().Errorf("Tool: MoveFile] %s (resolved: %s): %v", errMsg, absDest, destStatErr)
		return map[string]interface{}{"error": errMsg}, NewRuntimeError(ErrorCodeInternal, errMsg, destStatErr)
	}

	interpreter.Logger().Infof("Tool: MoveFile] Attempting rename/move: '%s' -> '%s'", absSource, absDest)
	renameErr := os.Rename(absSource, absDest)
	if renameErr != nil {
		errMsg := fmt.Sprintf("failed to move/rename '%s' to '%s'", sourcePathRel, destPathRel)
		interpreter.Logger().Errorf("Tool: MoveFile] Error: %s: %v", errMsg, renameErr)
		return map[string]interface{}{"error": errMsg}, NewRuntimeError(ErrorCodeInternal, errMsg, renameErr)
	}

	successMsg := fmt.Sprintf("Successfully moved/renamed '%s' to '%s'", sourcePathRel, destPathRel)
	interpreter.Logger().Infof("Tool: MoveFile] %s", successMsg)
	// Corrected: Return a map on success as expected by the functional test
	return map[string]interface{}{"message": successMsg, "error": nil}, nil
}
