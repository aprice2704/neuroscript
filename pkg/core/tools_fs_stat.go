// NeuroScript Version: 0.3.1
// File version: 0.0.1 // Added std arg validation & NewRuntimeError calls with standard ErrorCodes/Sentinels.
// nlines: 73
// risk_rating: LOW
// filename: pkg/core/tools_fs_stat.go
package core

import (
	"errors" // Required for errors.Is
	"fmt"
	"os"
	"path/filepath" // Added for path normalization in result map
	"time"
)

// toolStat gets information about a file or directory within the sandbox.
// Implements the StatPath tool.
func toolStat(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// --- Argument Validation ---
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("StatPath: expected 1 argument (path), got %d", len(args)), ErrArgumentMismatch)
	}
	relPath, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("StatPath: path argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	if relPath == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "StatPath: path cannot be empty", ErrInvalidArgument)
	}

	// --- Sandbox Check ---
	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		interpreter.Logger().Error("Tool: StatPath] Interpreter sandboxDir is empty, cannot proceed.")
		return nil, NewRuntimeError(ErrorCodeConfiguration, "StatPath: interpreter sandbox directory is not set", ErrConfiguration)
	}

	// --- Path Security Validation ---
	absPathToStat, secErr := SecureFilePath(relPath, sandboxRoot)
	if secErr != nil {
		errMsg := fmt.Sprintf("StatPath: path security error for '%s': %v", relPath, secErr)
		interpreter.Logger().Info("Tool: StatPath] %s (Sandbox Root: %s)", errMsg, sandboxRoot)
		// Return the RuntimeError from SecureFilePath directly
		return nil, secErr
	}

	interpreter.Logger().Debug("Tool: StatPath attempting to stat validated path", "validated_path", absPathToStat, "original_path", relPath)

	// --- Stat Path ---
	info, statErr := os.Stat(absPathToStat)
	if statErr != nil {
		// Handle file not found
		if errors.Is(statErr, os.ErrNotExist) {
			errMsg := fmt.Sprintf("StatPath: path not found '%s'", relPath)
			interpreter.Logger().Info("Tool: StatPath] %s", errMsg)
			// Return specific error code and sentinel
			return nil, NewRuntimeError(ErrorCodeFileNotFound, errMsg, ErrFileNotFound)
		}
		// Handle permission denied
		if errors.Is(statErr, os.ErrPermission) {
			errMsg := fmt.Sprintf("StatPath: permission denied for '%s'", relPath)
			interpreter.Logger().Warn("Tool: StatPath] %s", errMsg)
			return nil, NewRuntimeError(ErrorCodePermissionDenied, errMsg, ErrPermissionDenied)
		}
		// Handle other I/O errors
		errMsg := fmt.Sprintf("StatPath: failed to stat path '%s'", relPath)
		interpreter.Logger().Error("Tool: StatPath] %s: %v", errMsg, statErr)
		return nil, NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrIOFailed, statErr))
	}

	// --- Success: Construct Result Map ---
	resultMap := map[string]interface{}{
		"name":             info.Name(),                             // Base name of the file/dir
		"path":             filepath.ToSlash(relPath),               // Original relative path requested (normalized)
		"size_bytes":       info.Size(),                             // int64
		"is_dir":           info.IsDir(),                            // bool
		"modified_unix":    info.ModTime().Unix(),                   // int64
		"modified_rfc3339": info.ModTime().Format(time.RFC3339Nano), // string
		"mode_string":      info.Mode().String(),                    // string (e.g., "-rw-r--r--")
		"mode_perm":        fmt.Sprintf("%04o", info.Mode().Perm()), // string (e.g., "0644")
		// "abs_path":      absPathToStat, // Maybe don't expose absolute path? Stick to relative.
	}

	interpreter.Logger().Info("Tool: StatPath] Stat successful", "path", relPath)
	return resultMap, nil
}
