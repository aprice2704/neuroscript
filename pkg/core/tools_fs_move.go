// NeuroScript Version: 0.3.1
// File version: 0.0.2 // Corrected NewRuntimeError calls with standard ErrorCodes/Sentinels. Corrected error return values.
// nlines: 88
// risk_rating: HIGH
// filename: pkg/core/tools_fs_move.go
package core

import (
	"errors"
	"fmt"
	"os"
)

// toolMoveFile moves or renames a file or directory within the sandbox.
// Implements the MoveFile tool.
func toolMoveFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("MoveFile: expected 2 arguments (source_path, destination_path), got %d", len(args)), ErrArgumentMismatch)
	}

	sourcePathRel, okSrc := args[0].(string)
	destPathRel, okDest := args[1].(string)

	if !okSrc {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("MoveFile: source_path argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	if !okDest {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("MoveFile: destination_path argument must be a string, got %T", args[1]), ErrInvalidArgument)
	}
	if sourcePathRel == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "MoveFile: source_path cannot be empty", ErrInvalidArgument)
	}
	if destPathRel == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "MoveFile: destination_path cannot be empty", ErrInvalidArgument)
	}
	if sourcePathRel == destPathRel {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "MoveFile: source and destination paths cannot be the same", ErrInvalidArgument)
	}

	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		interpreter.Logger().Error("Tool: MoveFile] Interpreter sandboxDir is empty, cannot proceed.")
		return nil, NewRuntimeError(ErrorCodeConfiguration, "MoveFile: interpreter sandbox directory is not set", ErrConfiguration)
	}

	absSource, errSource := SecureFilePath(sourcePathRel, sandboxRoot)
	if errSource != nil {
		interpreter.Logger().Infof("Tool: MoveFile] Invalid source path '%s': %v", sourcePathRel, errSource)
		// Return the RuntimeError directly
		return nil, errSource
	}

	absDest, errDest := SecureFilePath(destPathRel, sandboxRoot)
	if errDest != nil {
		interpreter.Logger().Infof("Tool: MoveFile] Invalid destination path '%s': %v", destPathRel, errDest)
		// Return the RuntimeError directly
		return nil, errDest
	}

	interpreter.Logger().Infof("Tool: MoveFile] Validated paths: Source '%s' (abs: '%s'), Dest '%s' (abs: '%s')", sourcePathRel, absSource, destPathRel, absDest)

	// Check if source exists before trying to move
	_, srcStatErr := os.Stat(absSource)
	if srcStatErr != nil {
		errMsg := ""
		var rtErr *RuntimeError
		if errors.Is(srcStatErr, os.ErrNotExist) {
			errMsg = fmt.Sprintf("MoveFile: source path '%s' does not exist", sourcePathRel)
			rtErr = NewRuntimeError(ErrorCodeFileNotFound, errMsg, ErrFileNotFound)
		} else if errors.Is(srcStatErr, os.ErrPermission) {
			errMsg = fmt.Sprintf("MoveFile: permission denied checking source path '%s'", sourcePathRel)
			rtErr = NewRuntimeError(ErrorCodePermissionDenied, errMsg, ErrPermissionDenied)
		} else {
			errMsg = fmt.Sprintf("MoveFile: error checking source path '%s'", sourcePathRel)
			rtErr = NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrIOFailed, srcStatErr))
		}
		interpreter.Logger().Warnf("Tool: MoveFile] Source check failed: %s: %v", errMsg, srcStatErr)
		return nil, rtErr // Return nil value and the runtime error
	}

	// Check if destination *already exists*
	_, destStatErr := os.Stat(absDest)
	if destStatErr == nil {
		// Destination exists, this is usually an error for Rename/Move
		errMsg := fmt.Sprintf("MoveFile: destination path '%s' already exists", destPathRel)
		interpreter.Logger().Warnf("Tool: MoveFile] Error: %s (resolved: %s)", errMsg, absDest)
		// Use ErrorCodePathExists
		return nil, NewRuntimeError(ErrorCodePathExists, errMsg, ErrPathExists)
	} else if !errors.Is(destStatErr, os.ErrNotExist) {
		// Error stating destination path (e.g., permission error on parent dir)
		errMsg := fmt.Sprintf("MoveFile: error checking destination path '%s'", destPathRel)
		interpreter.Logger().Errorf("Tool: MoveFile] %s (resolved: %s): %v", errMsg, absDest, destStatErr)
		if errors.Is(destStatErr, os.ErrPermission) {
			return nil, NewRuntimeError(ErrorCodePermissionDenied, errMsg, ErrPermissionDenied)
		}
		return nil, NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrIOFailed, destStatErr))
	}
	// Destination does not exist (or stat failed with os.ErrNotExist), which is what we want.

	// Attempt the move/rename operation
	interpreter.Logger().Infof("Tool: MoveFile] Attempting rename/move: '%s' -> '%s'", absSource, absDest)
	renameErr := os.Rename(absSource, absDest)
	if renameErr != nil {
		errMsg := fmt.Sprintf("MoveFile: failed to move/rename '%s' to '%s'", sourcePathRel, destPathRel)
		interpreter.Logger().Errorf("Tool: MoveFile] Error: %s: %v", errMsg, renameErr)
		// Check for specific OS errors if needed (e.g., cross-device link), otherwise treat as general I/O
		// Use ErrorCodeIOFailed
		return nil, NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrIOFailed, renameErr))
	}

	// Success
	successMsg := fmt.Sprintf("Successfully moved/renamed '%s' to '%s'", sourcePathRel, destPathRel)
	interpreter.Logger().Infof("Tool: MoveFile] %s", successMsg)
	// Return the success map as specified in tooldefs_fs.go (ReturnType: ArgTypeMap)
	return map[string]interface{}{"message": successMsg, "error": nil}, nil
}
