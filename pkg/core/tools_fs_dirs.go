// NeuroScript Version: 0.3.1
// File version: 0.0.2 // Corrected NewRuntimeError calls with standard ErrorCodes/Sentinels.
// nlines: 150
// risk_rating: MEDIUM
// filename: pkg/core/tools_fs_dirs.go
package core

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// toolListDirectory lists contents of a directory.
// Returns []map[string]interface{} with keys: name, path, isDir, size, modTime
func toolListDirectory(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("ListDirectory: expected 1 or 2 arguments (path, [recursive]), got %d", len(args)), ErrArgumentMismatch)
	}
	relPath, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("ListDirectory: path argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	if relPath == "" {
		// Treat empty path as current sandbox root for listing, consistent with shell behavior.
		relPath = "."
		// return nil, NewRuntimeError(ErrorCodeArgMismatch, "ListDirectory: path argument cannot be empty", ErrInvalidArgument)
	}

	recursive := false
	if len(args) == 2 {
		if args[1] == nil {
			// Allow null for optional boolean, treat as false
			recursive = false
		} else {
			recursiveVal, okBool := args[1].(bool)
			if !okBool {
				return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("ListDirectory: recursive argument must be a boolean or null, got %T", args[1]), ErrInvalidArgument)
			}
			recursive = recursiveVal
		}
	}

	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		interpreter.Logger().Error("Tool: ListDirectory] Interpreter sandboxDir is empty, cannot proceed.")
		return nil, NewRuntimeError(ErrorCodeConfiguration, "ListDirectory: interpreter sandbox directory is not set", ErrConfiguration)
	}

	absBasePath, secErr := SecureFilePath(relPath, sandboxRoot)
	if secErr != nil {
		interpreter.Logger().Infof("Tool: ListDirectory] Path security error for %q: %v (Sandbox Root: %s)", relPath, secErr, sandboxRoot)
		return nil, secErr // SecureFilePath returns RuntimeError
	}
	interpreter.Logger().Infof("Tool: ListDirectory] Validated base path: %s (Original Relative: %q, Sandbox: %q, Recursive: %t)", absBasePath, relPath, sandboxRoot, recursive)

	baseInfo, statErr := os.Stat(absBasePath)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			interpreter.Logger().Infof("Tool: ListDirectory] Path not found %q (resolved: %s)", relPath, absBasePath)
			// Use ErrorCodeFileNotFound
			return nil, NewRuntimeError(ErrorCodeFileNotFound, fmt.Sprintf("ListDirectory: path not found '%s'", relPath), ErrFileNotFound)
		}
		if errors.Is(statErr, os.ErrPermission) {
			interpreter.Logger().Errorf("Tool: ListDirectory] Permission error stating path %q: %v", relPath, statErr)
			return nil, NewRuntimeError(ErrorCodePermissionDenied, fmt.Sprintf("ListDirectory: permission denied stating path '%s'", relPath), ErrPermissionDenied)
		}
		interpreter.Logger().Errorf("Tool: ListDirectory] Failed to stat path %q (resolved: %s): %v", relPath, absBasePath, statErr)
		return nil, NewRuntimeError(ErrorCodeIOFailed, fmt.Sprintf("ListDirectory: failed to stat path '%s'", relPath), errors.Join(ErrIOFailed, statErr))
	}

	if !baseInfo.IsDir() {
		errMsg := fmt.Sprintf("path '%s' is not a directory", relPath)
		interpreter.Logger().Infof("Tool: ListDirectory] %s (resolved: %s)", errMsg, absBasePath)
		// Use ErrorCodePathTypeMismatch and ErrPathNotDirectory
		return nil, NewRuntimeError(ErrorCodePathTypeMismatch, errMsg, ErrPathNotDirectory)
	}

	var fileInfos = make([]map[string]interface{}, 0)

	if recursive {
		interpreter.Logger().Debugf("Tool: ListDirectory] Walking recursively from %s...", absBasePath)
		walkErr := filepath.WalkDir(absBasePath, func(currentPath string, d fs.DirEntry, err error) error {
			if err != nil {
				interpreter.Logger().Warnf("Tool: ListDirectory Walk] Error accessing %q during walk: %v. Skipping entry/subtree.", currentPath, err)
				// Check for specific errors if needed (e.g., permission)
				if errors.Is(err, os.ErrPermission) {
					// Optionally, halt the walk on permission error, or just skip
					// return NewRuntimeError(ErrorCodePermissionDenied, fmt.Sprintf("ListDirectory walk: permission denied for '%s'", currentPath), ErrPermissionDenied) // Halts walk
					return nil // Skips entry
				}
				// For other errors, skip the entry
				return nil
			}
			if currentPath == absBasePath && d.IsDir() {
				return nil // Skip the root dir itself
			}

			info, infoErr := d.Info()
			if infoErr != nil {
				interpreter.Logger().Warnf("Tool: ListDirectory Walk] Error getting FileInfo for %q: %v. Skipping entry.", currentPath, infoErr)
				return nil // Skip this entry
			}

			entryRelPath, relErr := filepath.Rel(absBasePath, currentPath)
			if relErr != nil {
				interpreter.Logger().Warnf("Tool: ListDirectory Walk] Error calculating relative path for %q (base %q): %v. Skipping entry.", currentPath, absBasePath, relErr)
				return nil // Skip this entry
			}

			entryMap := map[string]interface{}{
				"name":    d.Name(),
				"path":    filepath.ToSlash(entryRelPath),
				"isDir":   d.IsDir(),
				"size":    info.Size(),
				"modTime": info.ModTime().Format(time.RFC3339Nano),
			}
			fileInfos = append(fileInfos, entryMap)
			return nil
		})
		// Check the error returned by WalkDir itself (e.g., if the callback returned an error)
		if walkErr != nil {
			errMsg := fmt.Sprintf("failed during recursive directory walk for '%s'", relPath)
			interpreter.Logger().Errorf("Tool: ListDirectory] %s: %v", errMsg, walkErr)
			// Determine specific ErrorCode if possible (e.g., from wrapped error)
			if errors.Is(walkErr, os.ErrPermission) { // Check if permission error halted the walk
				return nil, NewRuntimeError(ErrorCodePermissionDenied, errMsg, ErrPermissionDenied)
			}
			return nil, NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrIOFailed, walkErr))
		}
	} else { // Non-recursive
		interpreter.Logger().Debugf("Tool: ListDirectory] Reading directory non-recursively: %s...", absBasePath)
		entries, readErr := os.ReadDir(absBasePath)
		if readErr != nil {
			errMsg := fmt.Sprintf("failed reading directory '%s'", relPath)
			interpreter.Logger().Errorf("Tool: ListDirectory] %s (resolved: %s): %v", errMsg, absBasePath, readErr)
			if errors.Is(readErr, os.ErrPermission) {
				return nil, NewRuntimeError(ErrorCodePermissionDenied, errMsg, ErrPermissionDenied)
			}
			return nil, NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrIOFailed, readErr))
		}
		for _, entry := range entries {
			info, infoErr := entry.Info() // Best effort to get info
			var size int64 = 0
			var modTime time.Time
			if infoErr == nil && info != nil {
				size = info.Size()
				modTime = info.ModTime()
			} // Ignore infoErr for basic listing
			entryMap := map[string]interface{}{
				"name":    entry.Name(),
				"path":    filepath.ToSlash(entry.Name()),
				"isDir":   entry.IsDir(),
				"size":    size,
				"modTime": modTime.Format(time.RFC3339Nano),
			}
			fileInfos = append(fileInfos, entryMap)
		}
	}

	interpreter.Logger().Infof("Tool: ListDirectory] Listing successful for %q (Recursive: %t). Found %d entries.", relPath, recursive, len(fileInfos))
	return fileInfos, nil
}

// toolMkdir creates a directory (like mkdir -p).
func toolMkdir(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("Mkdir: expected 1 argument (path), got %d", len(args)), ErrArgumentMismatch)
	}
	relPath, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("Mkdir: path argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	if relPath == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "Mkdir: path argument cannot be empty", ErrInvalidArgument)
	}
	// Prevent creating "." or using ".."
	cleanRelPath := filepath.Clean(relPath)
	if cleanRelPath == "." || strings.HasPrefix(cleanRelPath, "..") {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("Mkdir: path '%s' is invalid or attempts to traverse upwards", relPath), ErrInvalidArgument)
	}

	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		interpreter.Logger().Error("Tool: Mkdir] Interpreter sandboxDir is empty, cannot proceed.")
		return nil, NewRuntimeError(ErrorCodeConfiguration, "Mkdir: interpreter sandbox directory is not set", ErrConfiguration)
	}

	absPathToCreate, secErr := SecureFilePath(relPath, sandboxRoot)
	if secErr != nil {
		interpreter.Logger().Infof("Tool: Mkdir] Path security error for %q: %v (Sandbox Root: %s)", relPath, secErr, sandboxRoot)
		return nil, secErr // SecureFilePath returns RuntimeError
	}

	interpreter.Logger().Infof("Tool: Mkdir] Validated path. Attempting to create directory: %s (Original Relative: %q)", absPathToCreate, relPath)

	info, statErr := os.Stat(absPathToCreate)
	if statErr == nil {
		if !info.IsDir() {
			// Path exists but is a file!
			errMsg := fmt.Sprintf("path '%s' already exists and is a file, not a directory", relPath)
			interpreter.Logger().Errorf("Tool: Mkdir] %s (resolved: %s)", errMsg, absPathToCreate)
			// Use ErrorCodePathExists and ErrPathNotDirectory sentinel
			return nil, NewRuntimeError(ErrorCodePathExists, errMsg, ErrPathNotDirectory)
		}
		// Path exists and is a directory - MkdirAll is idempotent, so this is fine.
		interpreter.Logger().Infof("Tool: Mkdir] Directory '%s' already exists.", relPath)
		return fmt.Sprintf("Directory '%s' already exists.", relPath), nil
	} else if !errors.Is(statErr, os.ErrNotExist) {
		// Error stating path other than "not found" (e.g., permission error)
		errMsg := fmt.Sprintf("failed to check path '%s'", relPath)
		interpreter.Logger().Errorf("Tool: Mkdir] Stat error: %s (resolved: %s): %v", errMsg, absPathToCreate, statErr)
		if errors.Is(statErr, os.ErrPermission) {
			return nil, NewRuntimeError(ErrorCodePermissionDenied, errMsg, ErrPermissionDenied)
		}
		return nil, NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrIOFailed, statErr))
	}

	// Path does not exist, proceed with MkdirAll
	err := os.MkdirAll(absPathToCreate, 0755)
	if err != nil {
		errMsg := fmt.Sprintf("failed to create directory '%s'", relPath)
		interpreter.Logger().Errorf("Tool: Mkdir] %s (resolved: %s): %v", errMsg, absPathToCreate, err)
		// Use ErrorCodeIOFailed and wrap ErrCannotCreateDir sentinel + OS error
		return nil, NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrCannotCreateDir, err))
	}

	successMsg := fmt.Sprintf("Successfully created directory: %s", relPath)
	interpreter.Logger().Infof("Tool: Mkdir] %s", successMsg)
	return successMsg, nil
}
