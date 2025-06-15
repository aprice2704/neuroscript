// NeuroScript Version: 0.4.0
// File version: 2
// Purpose: Corrected toolMkdir to be idempotent, returning success if the directory already exists.
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
	absBasePath, secErr := ResolveAndSecurePath(relPath, sandboxRoot)
	if secErr != nil {
		return nil, secErr
	}

	baseInfo, statErr := os.Stat(absBasePath)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			return nil, NewRuntimeError(ErrorCodeFileNotFound, fmt.Sprintf("ListDirectory: path not found '%s'", relPath), ErrFileNotFound)
		}
		return nil, NewRuntimeError(ErrorCodeIOFailed, fmt.Sprintf("ListDirectory: failed to stat path '%s': %v", relPath, statErr), errors.Join(ErrIOFailed, statErr))
	}

	if !baseInfo.IsDir() {
		return nil, NewRuntimeError(ErrorCodePathTypeMismatch, fmt.Sprintf("path '%s' is not a directory", relPath), ErrPathNotDirectory)
	}

	var fileInfos = make([]map[string]interface{}, 0)
	if recursive {
		walkErr := filepath.WalkDir(absBasePath, func(currentPath string, d fs.DirEntry, err error) error {
			if err != nil || currentPath == absBasePath {
				return nil // Skip errors and the root itself
			}
			info, _ := d.Info()
			entryRelPath, _ := filepath.Rel(absBasePath, currentPath)
			fileInfos = append(fileInfos, map[string]interface{}{
				"name":    d.Name(),
				"path":    filepath.ToSlash(entryRelPath),
				"isDir":   d.IsDir(),
				"size":    info.Size(),
				"modTime": info.ModTime().Format(time.RFC3339Nano),
			})
			return nil
		})
		if walkErr != nil {
			return nil, NewRuntimeError(ErrorCodeIOFailed, fmt.Sprintf("failed directory walk for '%s'", relPath), errors.Join(ErrIOFailed, walkErr))
		}
	} else {
		entries, readErr := os.ReadDir(absBasePath)
		if readErr != nil {
			return nil, NewRuntimeError(ErrorCodeIOFailed, fmt.Sprintf("failed reading directory '%s'", relPath), errors.Join(ErrIOFailed, readErr))
		}
		for _, entry := range entries {
			info, _ := entry.Info()
			fileInfos = append(fileInfos, map[string]interface{}{
				"name":    entry.Name(),
				"path":    filepath.ToSlash(entry.Name()),
				"isDir":   entry.IsDir(),
				"size":    info.Size(),
				"modTime": info.ModTime().Format(time.RFC3339Nano),
			})
		}
	}
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
	if relPath == "" || filepath.Clean(relPath) == "." {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "Mkdir: path argument cannot be empty or '.'", ErrInvalidArgument)
	}

	absPathToCreate, secErr := ResolveAndSecurePath(relPath, interpreter.SandboxDir())
	if secErr != nil {
		return nil, secErr
	}

	info, statErr := os.Stat(absPathToCreate)
	if statErr == nil {
		if info.IsDir() {
			// Path exists and is a directory. This is a success case (idempotent).
			return map[string]interface{}{"status": "success", "message": "Directory already exists", "path": relPath}, nil
		}
		// Path exists but is a file. This is an error.
		return nil, NewRuntimeError(ErrorCodePathTypeMismatch, fmt.Sprintf("path '%s' already exists and is a file", relPath), ErrPathNotDirectory)
	}

	if !errors.Is(statErr, os.ErrNotExist) {
		// An unexpected error occurred during Stat.
		return nil, NewRuntimeError(ErrorCodeIOFailed, fmt.Sprintf("failed to check path status for '%s'", relPath), errors.Join(ErrIOFailed, statErr))
	}

	// Path does not exist, so we create it.
	if err := os.MkdirAll(absPathToCreate, 0755); err != nil {
		return nil, NewRuntimeError(ErrorCodeIOFailed, fmt.Sprintf("failed to create directory '%s'", relPath), errors.Join(ErrCannotCreateDir, err))
	}

	return map[string]interface{}{"status": "success", "message": "Successfully created directory", "path": relPath}, nil
}
