// NeuroScript Version: 0.3.1
// File version: 0.0.2 // Corrected lang.NewRuntimeError calls with standard ErrorCodes/Sentinels. Corrected error return values.
// nlines: 88
// risk_rating: HIGH
// filename: pkg/tool/fs/tools_fs_move.go
package fs

import (
	"errors"
	"fmt"
	"os"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// toolMoveFile moves or renames a file or directory within the sandbox.
// Implements the MoveFile tool.
func toolMoveFile(interpreter *neurogo.Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("MoveFile: expected 2 arguments (source_path, destination_path), got %d", len(args)), lang.ErrArgumentMismatch)
	}

	sourcePathRel, okSrc := args[0].(string)
	destPathRel, okDest := args[1].(string)

	if !okSrc {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("MoveFile: source_path argument must be a string, got %T", args[0]), lang.ErrInvalidArgument)
	}
	if !okDest {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("MoveFile: destination_path argument must be a string, got %T", args[1]), lang.ErrInvalidArgument)
	}
	if sourcePathRel == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "MoveFile: source_path cannot be empty", lang.ErrInvalidArgument)
	}
	if destPathRel == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "MoveFile: destination_path cannot be empty", lang.ErrInvalidArgument)
	}
	if sourcePathRel == destPathRel {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "MoveFile: source and destination paths cannot be the same", lang.ErrInvalidArgument)
	}

	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		interpreter.Logger().Error("Tool: MoveFile] Interpreter sandboxDir is empty, cannot proceed.")
		return nil, lang.NewRuntimeError(lang.ErrorCodeConfiguration, "MoveFile: interpreter sandbox directory is not set", lang.ErrConfiguration)
	}

	absSource, errSource := security.SecureFilePath(sourcePathRel, sandboxRoot)
	if errSource != nil {
		interpreter.Logger().Infof("Tool: MoveFile] Invalid source path '%s': %v", sourcePathRel, errSource)
		// Return the RuntimeError directly
		return nil, errSource
	}

	absDest, errDest := security.SecureFilePath(destPathRel, sandboxRoot)
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
		var rtErr *lang.RuntimeError
		if errors.Is(srcStatErr, os.ErrNotExist) {
			errMsg = fmt.Sprintf("MoveFile: source path '%s' does not exist", sourcePathRel)
			rtErr = lang.NewRuntimeError(lang.ErrorCodeFileNotFound, errMsg, lang.ErrFileNotFound)
		} else if errors.Is(srcStatErr, os.ErrPermission) {
			errMsg = fmt.Sprintf("MoveFile: permission denied checking source path '%s'", sourcePathRel)
			rtErr = lang.NewRuntimeError(lang.ErrorCodePermissionDenied, errMsg, lang.ErrPermissionDenied)
		} else {
			errMsg = fmt.Sprintf("MoveFile: error checking source path '%s'", sourcePathRel)
			rtErr = lang.NewRuntimeError(lang.ErrorCodeIOFailed, errMsg, errors.Join(lang.ErrIOFailed, srcStatErr))
		}
		interpreter.Logger().Warnf("Tool: MoveFile] Source check failed: %s: %v", errMsg, srcStatErr)
		return nil, rtErr	// Return nil value and the runtime error
	}

	// Check if destination *already exists*
	_, destStatErr := os.Stat(absDest)
	if destStatErr == nil {
		// Destination exists, this is usually an error for Rename/Move
		errMsg := fmt.Sprintf("MoveFile: destination path '%s' already exists", destPathRel)
		interpreter.Logger().Warnf("Tool: MoveFile] Error: %s (resolved: %s)", errMsg, absDest)
		// Use ErrorCodePathExists
		return nil, lang.NewRuntimeError(lang.ErrorCodePathExists, errMsg, lang.ErrPathExists)
	} else if !errors.Is(destStatErr, os.ErrNotExist) {
		// Error stating destination path (e.g., permission error on parent dir)
		errMsg := fmt.Sprintf("MoveFile: error checking destination path '%s'", destPathRel)
		interpreter.Logger().Errorf("Tool: MoveFile] %s (resolved: %s): %v", errMsg, absDest, destStatErr)
		if errors.Is(destStatErr, os.ErrPermission) {
			return nil, lang.NewRuntimeError(lang.ErrorCodePermissionDenied, errMsg, lang.ErrPermissionDenied)
		}
		return nil, lang.NewRuntimeError(lang.ErrorCodeIOFailed, errMsg, errors.Join(lang.ErrIOFailed, destStatErr))
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
		return nil, lang.NewRuntimeError(lang.ErrorCodeIOFailed, errMsg, errors.Join(lang.ErrIOFailed, renameErr))
	}

	// Success
	successMsg := fmt.Sprintf("Successfully moved/renamed '%s' to '%s'", sourcePathRel, destPathRel)
	interpreter.Logger().Infof("Tool: MoveFile] %s", successMsg)
	// Return the success map as specified in tooldefs_fs.go (ReturnType: ArgTypeMap)
	return map[string]interface{}{"message": successMsg, "error": nil}, nil
}