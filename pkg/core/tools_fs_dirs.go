// NeuroScript Version: 0.3.1
// File version: 0.1.1 // Removed redundant fmt.Printf in toolMkdir.
// nlines: 156 // Approximate
// risk_rating: MEDIUM
// filename: pkg/core/tools_fs_dirs.go
package core

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// --- toolListDirectory unchanged ---
func toolListDirectory(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("ListDirectory: expected 1 or 2 arguments (path, [recursive]), got %d", len(args)), ErrArgumentMismatch)
	}
	relPath, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("ListDirectory: path argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	if relPath == "" {
		relPath = "."
	}

	recursive := false
	if len(args) == 2 {
		if args[1] != nil {
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

	absBasePath, secErr := ResolveAndSecurePath(relPath, sandboxRoot)
	if secErr != nil {
		interpreter.Logger().Infof("Tool: ListDirectory] Path security error for %q: %v (Sandbox Root: %s)", relPath, secErr.Error(), sandboxRoot)
		return nil, secErr
	}

	interpreter.Logger().Infof("Tool: ListDirectory] Validated base path: %s (Original Relative: %q, Sandbox: %q, Recursive: %t)", absBasePath, relPath, sandboxRoot, recursive)
	baseInfo, statErr := os.Stat(absBasePath)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			interpreter.Logger().Infof("Tool: ListDirectory] Path not found %q (resolved: %s)", relPath, absBasePath)
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
		return nil, NewRuntimeError(ErrorCodePathTypeMismatch, errMsg, ErrPathNotDirectory)
	}

	var fileInfos = make([]map[string]interface{}, 0)
	if recursive {
		interpreter.Logger().Debugf("Tool: ListDirectory] Walking recursively from %s...", absBasePath)
		walkErr := filepath.WalkDir(absBasePath, func(currentPath string, d fs.DirEntry, err error) error {
			if err != nil {
				if errors.Is(err, fs.ErrPermission) {
					interpreter.Logger().Warnf("Tool: ListDirectory Walk] Permission error accessing %q: %v. Skipping entry/subtree.", currentPath, err)
				} else {
					interpreter.Logger().Warnf("Tool: ListDirectory Walk] Error accessing %q during walk: %v. Skipping entry/subtree.", currentPath, err)
				}
				if d != nil && d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			if currentPath == absBasePath {
				return nil
			}
			info, infoErr := d.Info()
			if infoErr != nil {
				interpreter.Logger().Warnf("Tool: ListDirectory Walk] Error getting FileInfo for %q: %v. Skipping entry.", currentPath, infoErr)
				return nil
			}
			entryRelPath, relErr := filepath.Rel(absBasePath, currentPath)
			if relErr != nil {
				interpreter.Logger().Warnf("Tool: ListDirectory Walk] Error calculating relative path for %q (base %q): %v. Skipping entry.", currentPath, absBasePath, relErr)
				return nil
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
		if walkErr != nil {
			errMsg := fmt.Sprintf("failed during recursive directory walk for '%s'", relPath)
			interpreter.Logger().Errorf("Tool: ListDirectory] %s: %v", errMsg, walkErr)
			if errors.Is(walkErr, os.ErrPermission) {
				return nil, NewRuntimeError(ErrorCodePermissionDenied, errMsg, ErrPermissionDenied)
			}
			return nil, NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrIOFailed, walkErr))
		}
	} else {
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
			info, infoErr := entry.Info()
			var size int64 = 0
			var modTime time.Time
			if infoErr == nil && info != nil {
				size = info.Size()
				modTime = info.ModTime()
			} else if infoErr != nil {
				interpreter.Logger().Warnf("Tool: ListDirectory ReadDir] Error getting FileInfo for %q: %v. Size/ModTime omitted.", entry.Name(), infoErr)
			}
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
	cleanRelPath := filepath.Clean(relPath)
	if cleanRelPath == "." {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "Mkdir: path '.' is invalid for creating a directory", ErrInvalidArgument)
	}

	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		interpreter.Logger().Error("Tool: Mkdir] Interpreter sandboxDir is empty, cannot proceed.")
		return nil, NewRuntimeError(ErrorCodeConfiguration, "Mkdir: interpreter sandbox directory is not set", ErrConfiguration)
	}

	// Resolve and secure the path
	interpreter.Logger().Debugf("Tool: Mkdir] PRE ResolveAndSecurePath for %q", relPath)
	absPathToCreate, secErr := ResolveAndSecurePath(relPath, sandboxRoot)
	interpreter.Logger().Debugf("Tool: Mkdir] POST ResolveAndSecurePath for %q -> err: <%v> (type: %T)", relPath, secErr, secErr)

	if secErr != nil {
		interpreter.Logger().Infof("Tool: Mkdir] Path validation failed for %q: %v (Sandbox Root: %s)", relPath, secErr, sandboxRoot)
		return nil, secErr
	}

	interpreter.Logger().Debugf("Tool: Mkdir] Validated path. Checking state for: %s (Original Relative: %q)", absPathToCreate, relPath)
	info, statErr := os.Stat(absPathToCreate)
	if statErr == nil {
		if !info.IsDir() {
			errMsg := fmt.Sprintf("path '%s' already exists and is a file, not a directory", relPath)
			interpreter.Logger().Infof("Tool: Mkdir] %s (resolved: %s)", errMsg, absPathToCreate)
			return nil, NewRuntimeError(ErrorCodePathTypeMismatch, errMsg, ErrPathNotDirectory)
		}
		errMsg := fmt.Sprintf("directory '%s' already exists", relPath)
		interpreter.Logger().Infof("Tool: Mkdir] %s (resolved: %s)", errMsg, absPathToCreate)
		return nil, NewRuntimeError(ErrorCodePathExists, errMsg, ErrPathExists)

	} else if !errors.Is(statErr, os.ErrNotExist) {
		errMsg := fmt.Sprintf("failed to check path status for '%s'", relPath)
		interpreter.Logger().Errorf("Tool: Mkdir] Stat error: %s (resolved: %s): %v", errMsg, absPathToCreate, statErr)
		if errors.Is(statErr, os.ErrPermission) {
			return nil, NewRuntimeError(ErrorCodePermissionDenied, errMsg, ErrPermissionDenied)
		}
		return nil, NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrIOFailed, statErr))
	}

	interpreter.Logger().Infof("Tool: Mkdir] Path does not exist, attempting to create directory: %s", absPathToCreate)
	err := os.MkdirAll(absPathToCreate, 0755)
	if err != nil {
		errMsg := fmt.Sprintf("failed to create directory '%s'", relPath)
		interpreter.Logger().Errorf("Tool: Mkdir] %s (resolved: %s): %v", errMsg, absPathToCreate, err)
		if errors.Is(err, os.ErrPermission) {
			return nil, NewRuntimeError(ErrorCodePermissionDenied, errMsg, ErrPermissionDenied)
		}
		return nil, NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrCannotCreateDir, err))
	}

	successMsg := fmt.Sprintf("Successfully created directory: %s", relPath)
	interpreter.Logger().Infof("Tool: Mkdir] %s", successMsg)
	resultMap := map[string]interface{}{
		"status":  "success",
		"message": successMsg,
		"path":    relPath,
	}
	return resultMap, nil
}
