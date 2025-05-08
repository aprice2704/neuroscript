// NeuroScript Version: 0.3.1
// File version: 0.0.2 // Replaced fmt.Errorf with NewRuntimeError using standard ErrorCodes/Sentinels.
// nlines: 110
// risk_rating: MEDIUM
// filename: pkg/core/tools_fs_walk.go
package core

import (
	"errors"
	"fmt"
	"io/fs" // Use io/fs for WalkDir and DirEntry
	"os"
	"path/filepath"
	"time"
)

// toolWalkDir recursively walks a directory, returning a list of maps,
// where each map describes a file or subdirectory found.
// Implements the WalkDir tool.
func toolWalkDir(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// --- Argument Validation ---
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("WalkDir: expected 1 argument (path string), got %d", len(args)), ErrArgumentMismatch)
	}
	relPath, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("WalkDir: path argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	if relPath == "" {
		// Treat empty path as current dir "."
		relPath = "."
		// return nil, NewRuntimeError(ErrorCodeArgMismatch, "WalkDir: path argument cannot be empty", ErrInvalidArgument)
	}

	// --- Sandbox Check ---
	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		interpreter.Logger().Error("Tool: WalkDir] Interpreter sandboxDir is empty, cannot proceed.")
		return nil, NewRuntimeError(ErrorCodeConfiguration, "WalkDir: interpreter sandbox directory is not set", ErrConfiguration)
	}

	// --- Path Security Validation ---
	absBasePath, secErr := SecureFilePath(relPath, sandboxRoot)
	if secErr != nil {
		errMsg := fmt.Sprintf("WalkDir: path security error for %q: %v", relPath, secErr)
		interpreter.Logger().Info("Tool: WalkDir] %s (Sandbox Root: %s)", errMsg, sandboxRoot)
		// Return the RuntimeError from SecureFilePath directly
		return nil, secErr
	}

	interpreter.Logger().Infof("Tool: WalkDir] Validated base path: %s (Original Relative: %q, Sandbox: %q)", absBasePath, relPath, sandboxRoot)

	// --- Check if Start Path is a Directory ---
	baseInfo, statErr := os.Stat(absBasePath)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			errMsg := fmt.Sprintf("WalkDir: start path not found '%s'", relPath)
			interpreter.Logger().Info("Tool: WalkDir] %s", errMsg)
			return nil, NewRuntimeError(ErrorCodeFileNotFound, errMsg, ErrFileNotFound)
		}
		if errors.Is(statErr, os.ErrPermission) {
			errMsg := fmt.Sprintf("WalkDir: permission denied for start path '%s'", relPath)
			return nil, NewRuntimeError(ErrorCodePermissionDenied, errMsg, ErrPermissionDenied)
		}
		errMsg := fmt.Sprintf("WalkDir: failed to stat start path '%s'", relPath)
		return nil, NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrIOFailed, statErr))
	}

	if !baseInfo.IsDir() {
		errMsg := fmt.Sprintf("WalkDir: start path '%s' is not a directory", relPath)
		interpreter.Logger().Info("Tool: WalkDir] %s", errMsg)
		return nil, NewRuntimeError(ErrorCodePathTypeMismatch, errMsg, ErrPathNotDirectory)
	}

	// --- Walk the Directory ---
	var fileInfos []interface{} // Slice of maps

	walkErr := filepath.WalkDir(absBasePath, func(currentPath string, d fs.DirEntry, walkPathErr error) error {
		// Handle error accessing the current path item
		if walkPathErr != nil {
			interpreter.Logger().Warnf("Tool: WalkDir] Error accessing %q during walk: %v. Skipping entry/subtree.", currentPath, walkPathErr)
			if errors.Is(walkPathErr, fs.ErrPermission) {
				// Option 1: Halt walk on permission error by returning a specific error
				// return NewRuntimeError(ErrorCodePermissionDenied, fmt.Sprintf("WalkDir: permission denied accessing '%s'", currentPath), ErrPermissionDenied)

				// Option 2: Skip this entry/subtree (often preferred for WalkDir)
				return nil
			}
			// For other access errors, skip the entry/subtree
			return nil
		}

		// Skip the root directory itself
		if currentPath == absBasePath {
			return nil
		}

		// Get FileInfo for the entry
		info, infoErr := d.Info()
		if infoErr != nil {
			interpreter.Logger().Warnf("Tool: WalkDir] Error getting FileInfo for %q: %v. Skipping entry.", currentPath, infoErr)
			// Decide if this error should halt the walk or just skip the entry
			// return NewRuntimeError(ErrorCodeIOFailed, fmt.Sprintf("WalkDir: failed getting FileInfo for '%s'", currentPath), errors.Join(ErrIOFailed, infoErr)) // Halts walk
			return nil // Skips entry
		}

		// Calculate path relative to the starting directory
		entryRelPath, relErr := filepath.Rel(absBasePath, currentPath)
		if relErr != nil {
			interpreter.Logger().Errorf("Tool: WalkDir] Internal error calculating relative path for %q (base %q): %v", currentPath, absBasePath, relErr)
			// This is unexpected, potentially halt the walk
			return NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("WalkDir: internal error calculating relative path for '%s'", currentPath), ErrInternal) // Halts walk
			// return nil // Skips entry
		}

		// Create map for the current entry
		entryMap := map[string]interface{}{
			"name":             d.Name(),
			"path_relative":    filepath.ToSlash(entryRelPath), // Use consistent slashes
			"is_dir":           d.IsDir(),
			"size_bytes":       info.Size(),
			"modified_unix":    info.ModTime().Unix(),
			"modified_rfc3339": info.ModTime().Format(time.RFC3339Nano),
			"mode_string":      info.Mode().String(),
		}
		fileInfos = append(fileInfos, entryMap)

		return nil // Continue walking
	})

	// --- Handle Final Error from WalkDir ---
	if walkErr != nil {
		// Check if it's a RuntimeError we returned from the callback
		var rtErr *RuntimeError
		if errors.As(walkErr, &rtErr) {
			interpreter.Logger().Errorf("Tool: WalkDir] Walk failed due to propagated error: %v", rtErr)
			return nil, rtErr // Return the specific RuntimeError
		}
		// Otherwise, it's likely an error from WalkDir itself (e.g., initial access)
		errMsg := fmt.Sprintf("WalkDir: failed walking directory '%s'", relPath)
		interpreter.Logger().Errorf("Tool: WalkDir] %s: %v", errMsg, walkErr)
		return nil, NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrIOFailed, walkErr))
	}

	interpreter.Logger().Infof("Tool: WalkDir] Walk successful", "path", relPath, "entries_found", len(fileInfos))
	return fileInfos, nil
}
