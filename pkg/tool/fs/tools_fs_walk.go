// NeuroScript Version: 0.3.1
// File version: 0.0.6 // Changed INFO logs to DEBUG
// nlines: 117
// risk_rating: MEDIUM
// filename: pkg/tool/fs/tools_fs_walk.go
package fs

import (
	"errors"
	"fmt"
	"io/fs"	// Use io/fs for WalkDir and DirEntry
	"os"
	"path/filepath"
	"time"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// toolWalkDir recursively walks a directory, returning a list of maps,
// where each map describes a file or subdirectory found.
func toolWalkDir(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// --- Argument Validation ---
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("WalkDir: expected 1 argument (path string), got %d", len(args)), ErrArgumentMismatch)
	}
	relPath, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(ErrorCodeType, fmt.Sprintf("WalkDir: path argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	// --- ADDED: Explicit check for empty path BEFORE resolving ---
	if relPath == "" {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "WalkDir: path argument cannot be empty", ErrInvalidArgument)
	}
	// Allow "." - handled by ResolveAndSecurePath

	// --- Sandbox Check ---
	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		interpreter.Logger().Error("Tool: WalkDir] Interpreter sandboxDir is empty, cannot proceed.")
		return nil, lang.NewRuntimeError(ErrorCodeConfiguration, "WalkDir: interpreter sandbox directory is not set", ErrConfiguration)
	}

	// --- Path Security Validation ---
	// ResolveAndSecurePath handles validation (absolute, traversal, null bytes, empty)
	absBasePath, secErr := ResolveAndSecurePath(relPath, sandboxRoot)
	if secErr != nil {
		interpreter.Logger().Debug("Tool: WalkDir] Path validation failed", "error", secErr.Error(), "path", relPath)	// Changed from Info
		return nil, secErr												// Return the *RuntimeError directly
	}

	interpreter.Logger().Debugf("Tool: WalkDir] Validated base path: %s (Original Relative: %q, Sandbox: %q)", absBasePath, relPath, sandboxRoot)	// Changed from Infof

	// --- Check if Start Path is a Directory ---
	baseInfo, statErr := os.Stat(absBasePath)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			errMsg := fmt.Sprintf("WalkDir: start path not found '%s'", relPath)
			interpreter.Logger().Debug("Tool: WalkDir] %s", errMsg)	// Changed from Info
			return nil, lang.NewRuntimeError(ErrorCodeFileNotFound, errMsg, ErrFileNotFound)
		}
		if errors.Is(statErr, os.ErrPermission) {
			errMsg := fmt.Sprintf("WalkDir: permission denied for start path '%s'", relPath)
			return nil, lang.NewRuntimeError(ErrorCodePermissionDenied, errMsg, ErrPermissionDenied)
		}
		errMsg := fmt.Sprintf("WalkDir: failed to stat start path '%s'", relPath)
		return nil, lang.NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrIOFailed, statErr))
	}

	if !baseInfo.IsDir() {
		errMsg := fmt.Sprintf("WalkDir: start path '%s' is not a directory", relPath)
		interpreter.Logger().Debug("Tool: WalkDir] %s", errMsg)	// Changed from Info
		return nil, lang.NewRuntimeError(ErrorCodePathTypeMismatch, errMsg, ErrPathNotDirectory)
	}

	// --- Walk the Directory ---
	// --- CHANGED: Use specific slice type ---
	var fileInfos = make([]map[string]interface{}, 0)

	walkErr := filepath.WalkDir(absBasePath, func(currentPath string, d fs.DirEntry, walkPathErr error) error {
		if walkPathErr != nil {
			interpreter.Logger().Warnf("Tool: WalkDir] Error accessing %q during walk: %v. Skipping entry/subtree.", currentPath, walkPathErr)
			return nil
		}	// Skip entry on error
		if currentPath == absBasePath {
			return nil
		}	// Skip root

		info, infoErr := d.Info()
		if infoErr != nil {
			interpreter.Logger().Warnf("Tool: WalkDir] Error getting FileInfo for %q: %v. Skipping entry.", currentPath, infoErr)
			return nil
		}	// Skip entry

		entryRelPath, relErr := filepath.Rel(absBasePath, currentPath)
		if relErr != nil {
			interpreter.Logger().Errorf("Tool: WalkDir] Internal error calculating relative path for %q (base %q): %v", currentPath, absBasePath, relErr)
			return lang.NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("WalkDir: internal error calculating relative path for '%s'", currentPath), ErrInternal)
		}

		entryMap := map[string]interface{}{
			"name":			d.Name(),
			"path_relative":	filepath.ToSlash(entryRelPath),	// Use consistent slashes
			"is_dir":		d.IsDir(),
			"size_bytes":		info.Size(),
			"modified_unix":	info.ModTime().Unix(),
			"modified_rfc3339":	info.ModTime().Format(time.RFC3339Nano),
			"mode_string":		info.Mode().String(),
		}
		// --- CHANGED: Append directly to specific slice type ---
		fileInfos = append(fileInfos, entryMap)
		return nil
	})

	// --- Handle Final Error from WalkDir ---
	if walkErr != nil {
		var rtErr *RuntimeError
		if errors.As(walkErr, &rtErr) {
			interpreter.Logger().Errorf("Tool: WalkDir] Walk failed due to propagated error: %v", rtErr)
			return nil, rtErr
		}
		errMsg := fmt.Sprintf("WalkDir: failed walking directory '%s'", relPath)
		interpreter.Logger().Errorf("Tool: WalkDir] %s: %v", errMsg, walkErr)
		return nil, lang.NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrIOFailed, walkErr))
	}

	interpreter.Logger().Debugf("Tool: WalkDir] Walk successful", "path", relPath, "entries_found", len(fileInfos))	// Changed from Infof
	// Return the correctly typed slice (even if empty)
	return fileInfos, nil
}