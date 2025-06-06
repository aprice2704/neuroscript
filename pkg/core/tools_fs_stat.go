// NeuroScript Version: 0.3.1
// File version: 0.0.4 // Add explicit empty path check.
// nlines: 77
// risk_rating: LOW
// filename: pkg/core/tools_fs_stat.go
package core

import (
	"errors" // Required for errors.Is
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// toolStat gets information about a file or directory within the sandbox.
func toolStat(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// --- Argument Validation ---
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("StatPath: expected 1 argument (path), got %d", len(args)), ErrArgumentMismatch)
	}
	relPath, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("StatPath: path argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	// --- ADDED: Explicit check for empty path BEFORE resolving ---
	if relPath == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "StatPath: path argument cannot be empty", ErrInvalidArgument)
	}

	// --- Sandbox Check ---
	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		interpreter.Logger().Error("Tool: StatPath] Interpreter sandboxDir is empty, cannot proceed.")
		return nil, NewRuntimeError(ErrorCodeConfiguration, "StatPath: interpreter sandbox directory is not set", ErrConfiguration)
	}

	// --- Path Security Validation ---
	// ResolveAndSecurePath handles validation (absolute, traversal, null bytes, empty)
	absPathToStat, secErr := ResolveAndSecurePath(relPath, sandboxRoot)
	if secErr != nil {
		interpreter.Logger().Debug("Tool: StatPath] Path validation failed", "error", secErr.Error(), "path", relPath)
		return nil, secErr // Return the *RuntimeError directly
	}

	interpreter.Logger().Debug("Tool: StatPath attempting to stat validated path", "validated_path", absPathToStat, "original_path", relPath)

	// --- Stat Path ---
	info, statErr := os.Stat(absPathToStat)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			errMsg := fmt.Sprintf("StatPath: path not found '%s'", relPath)
			interpreter.Logger().Debug("Tool: StatPath] %s", errMsg)
			return nil, NewRuntimeError(ErrorCodeFileNotFound, errMsg, ErrFileNotFound)
		}
		if errors.Is(statErr, os.ErrPermission) {
			errMsg := fmt.Sprintf("StatPath: permission denied for '%s'", relPath)
			interpreter.Logger().Warn("Tool: StatPath] %s", errMsg)
			return nil, NewRuntimeError(ErrorCodePermissionDenied, errMsg, ErrPermissionDenied)
		}
		errMsg := fmt.Sprintf("StatPath: failed to stat path '%s'", relPath)
		interpreter.Logger().Error("Tool: StatPath] %s: %v", errMsg, statErr)
		return nil, NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrIOFailed, statErr))
	}

	// --- Success: Construct Result Map ---
	resultMap := map[string]interface{}{
		"name":             info.Name(),
		"path":             filepath.ToSlash(relPath),
		"size_bytes":       info.Size(), // Use size_bytes key
		"is_dir":           info.IsDir(),
		"modified_unix":    info.ModTime().Unix(),
		"modified_rfc3339": info.ModTime().Format(time.RFC3339Nano),
		"mode_string":      info.Mode().String(),
		"mode_perm":        fmt.Sprintf("%04o", info.Mode().Perm()),
	}

	interpreter.Logger().Debug("Tool: StatPath] Stat successful", "path", relPath)
	return resultMap, nil
}
