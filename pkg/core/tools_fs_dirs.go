// NeuroScript Version: 0.3.0
// File version: 0.1.1 // Added init-based registration, corrected ErrorCodes & return types
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
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("ListDirectory expects 1 or 2 arguments, got %d", len(args)), ErrInvalidArgument)
	}
	relPath, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "ListDirectory expects argument 1 (path) to be a string", ErrInvalidArgument)
	}
	if relPath == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "ListDirectory path argument cannot be empty", ErrInvalidArgument)
	}

	recursive := false
	if len(args) == 2 {
		recursiveVal, okBool := args[1].(bool)
		if !okBool {
			if args[1] != nil {
				return nil, NewRuntimeError(ErrorCodeArgMismatch, "ListDirectory expects argument 2 (recursive) to be a boolean or null", ErrInvalidArgument)
			}
		} else {
			recursive = recursiveVal
		}
	}

	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		interpreter.Logger().Error("Tool: ListDirectory] Interpreter sandboxDir is empty, cannot proceed.")
		return nil, NewRuntimeError(ErrorCodeConfiguration, "interpreter sandbox directory is not set", ErrConfiguration)
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
			return nil, NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("directory not found at path '%s'", relPath), statErr)
		}
		interpreter.Logger().Errorf("Tool: ListDirectory] Failed to stat path %q (resolved: %s): %v", relPath, absBasePath, statErr)
		return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("failed to stat directory '%s'", relPath), statErr)
	}

	if !baseInfo.IsDir() {
		errMsg := fmt.Sprintf("path '%s' is not a directory", relPath)
		interpreter.Logger().Infof("Tool: ListDirectory] %s (resolved: %s)", errMsg, absBasePath)
		return nil, NewRuntimeError(ErrorCodeArgMismatch, errMsg, ErrInvalidArgument) // Path is not a dir is an arg mismatch
	}

	var fileInfos = make([]map[string]interface{}, 0) // Initialize directly as the target type

	if recursive {
		interpreter.Logger().Debugf("Tool: ListDirectory] Walking recursively from %s...", absBasePath)
		walkErr := filepath.WalkDir(absBasePath, func(currentPath string, d fs.DirEntry, err error) error {
			if err != nil {
				// Log and decide whether to skip or halt. Skipping is often better for WalkDir.
				interpreter.Logger().Warnf("Tool: ListDirectory Walk] Error accessing %q during walk: %v. Skipping entry/subtree.", currentPath, err)
				if currentPath == absBasePath { // If error on root, probably should halt
					return fmt.Errorf("error accessing root path %q: %w", currentPath, err)
				}
				return nil // Skip this entry/subtree but continue walking
			}
			// Skip the root directory itself from the list, but process its children.
			if currentPath == absBasePath && d.IsDir() {
				return nil
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
				"path":    filepath.ToSlash(entryRelPath), // Ensure consistent slash for paths
				"isDir":   d.IsDir(),
				"size":    info.Size(),
				"modTime": info.ModTime().Format(time.RFC3339Nano),
			}
			fileInfos = append(fileInfos, entryMap)
			return nil
		})
		if walkErr != nil {
			errMsg := fmt.Sprintf("failed during recursive directory walk for '%s'", relPath)
			interpreter.Logger().Errorf("Tool: ListDirectory] %s: %v", errMsg, walkErr)
			return nil, NewRuntimeError(ErrorCodeInternal, errMsg, walkErr)
		}
	} else {
		interpreter.Logger().Debugf("Tool: ListDirectory] Reading directory non-recursively: %s...", absBasePath)
		entries, readErr := os.ReadDir(absBasePath)
		if readErr != nil {
			errMsg := fmt.Sprintf("failed reading directory '%s'", relPath)
			interpreter.Logger().Errorf("Tool: ListDirectory] %s (resolved: %s): %v", errMsg, absBasePath, readErr)
			return nil, NewRuntimeError(ErrorCodeInternal, errMsg, readErr)
		}
		for _, entry := range entries {
			info, infoErr := entry.Info()
			var size int64 = 0
			var modTime time.Time
			if infoErr == nil && info != nil {
				size = info.Size()
				modTime = info.ModTime()
			} else {
				interpreter.Logger().Warnf("Tool: ListDirectory] Error getting FileInfo for entry '%s' in '%s': %v. Using zero values.", entry.Name(), relPath, infoErr)
			}
			entryMap := map[string]interface{}{
				"name":    entry.Name(),
				"path":    filepath.ToSlash(entry.Name()), // Path is just name relative to the listed dir
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

// toolMkdir implements TOOL.Mkdir
func toolMkdir(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("Mkdir expects 1 argument, got %d", len(args)), ErrInvalidArgument)
	}
	relPath, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "Mkdir expects argument 1 (path) to be a string", ErrInvalidArgument)
	}
	if relPath == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "Mkdir path argument cannot be empty", ErrInvalidArgument)
	}
	// Disallow "." and ".." as paths for Mkdir to prevent ambiguity or unintended behavior.
	if relPath == "." || relPath == ".." || strings.HasPrefix(relPath, "../") || strings.HasSuffix(relPath, "/..") || strings.Contains(relPath, "/../") {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("Mkdir path '%s' is invalid or attempts to traverse upwards", relPath), ErrInvalidArgument)
	}

	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		interpreter.Logger().Error("Tool: Mkdir] Interpreter sandboxDir is empty, cannot proceed.")
		return nil, NewRuntimeError(ErrorCodeConfiguration, "interpreter sandbox directory is not set", ErrConfiguration)
	}

	// SecureFilePath will ensure the final path is within the sandbox.
	// For MkdirAll, we want to ensure the *target directory itself* is valid.
	absPathToCreate, secErr := SecureFilePath(relPath, sandboxRoot)
	if secErr != nil {
		interpreter.Logger().Infof("Tool: Mkdir] Path security error for %q: %v (Sandbox Root: %s)", relPath, secErr, sandboxRoot)
		return nil, secErr // SecureFilePath returns RuntimeError
	}

	interpreter.Logger().Infof("Tool: Mkdir] Validated path. Attempting to create directory: %s (Original Relative: %q)", absPathToCreate, relPath)

	// Check if path already exists and is a file
	info, statErr := os.Stat(absPathToCreate)
	if statErr == nil && !info.IsDir() {
		errMsg := fmt.Sprintf("path '%s' already exists and is a file, not a directory", relPath)
		interpreter.Logger().Errorf("Tool: Mkdir] %s (resolved: %s)", errMsg, absPathToCreate)
		return nil, NewRuntimeError(ErrorCodePreconditionFailed, errMsg, ErrCannotCreateDir) // Use a more specific error
	}
	// If it exists and is a directory, MkdirAll will do nothing, which is fine.

	err := os.MkdirAll(absPathToCreate, 0755) // 0755 are standard directory permissions
	if err != nil {
		errMsg := fmt.Sprintf("failed to create directory '%s'", relPath)
		interpreter.Logger().Errorf("Tool: Mkdir] %s (resolved: %s): %v", errMsg, absPathToCreate, err)
		return nil, NewRuntimeError(ErrorCodeInternal, errMsg, err) // Use ErrorCodeInternal for OS errors
	}

	successMsg := fmt.Sprintf("Successfully created directory (or ensured it exists): %s", relPath)
	interpreter.Logger().Infof("Tool: Mkdir] %s", successMsg)
	return successMsg, nil
}
