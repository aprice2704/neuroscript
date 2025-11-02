// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Changed fileInfos slice type from []map[string]interface{} to []interface{} to fix lang.Wrap error.
// filename: pkg/tool/fs/tools_fs_walk.go
// nlines: 91
package fs

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/security"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// toolWalkDir implements the FS.Walk tool.
func toolWalkDir(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Walk: expected 1 argument (path)", lang.ErrArgumentMismatch)
	}

	relPath, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("Walk: path argument must be a string, got %T", args[0]), lang.ErrInvalidArgument)
	}
	if relPath == "" {
		relPath = "."
	}

	sandboxRoot := interpreter.SandboxDir()
	absBasePath, secErr := security.ResolveAndSecurePath(relPath, sandboxRoot)
	if secErr != nil {
		return nil, secErr
	}

	startInfo, statErr := os.Stat(absBasePath)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			return nil, lang.NewRuntimeError(lang.ErrorCodeFileNotFound, fmt.Sprintf("Walk: path not found '%s'", relPath), lang.ErrFileNotFound)
		}
		return nil, lang.NewRuntimeError(lang.ErrorCodeIOFailed, fmt.Sprintf("Walk: failed to stat path '%s'", relPath), errors.Join(lang.ErrIOFailed, statErr))
	}
	if !startInfo.IsDir() {
		return nil, lang.NewRuntimeError(lang.ErrorCodePathTypeMismatch, fmt.Sprintf("Walk: path '%s' is not a directory", relPath), lang.ErrPathNotDirectory)
	}

	var fileInfos []interface{} // <--- MODIFIED

	walkErr := filepath.WalkDir(absBasePath, func(currentPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip the root directory itself in the output list.
		if currentPath == absBasePath {
			return nil
		}

		info, infoErr := d.Info()
		if infoErr != nil {
			// Log or handle error if we can't get info for an entry
			interpreter.GetLogger().Warnf("Walk: could not get FileInfo for '%s': %v", currentPath, infoErr)
			return nil // Continue walking
		}

		// Make the path relative to the starting directory for consistent output.
		entryRelPath, relErr := filepath.Rel(absBasePath, currentPath)
		if relErr != nil {
			return lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("Walk: could not make path relative for '%s'", currentPath), relErr)
		}

		fileInfos = append(fileInfos, map[string]interface{}{
			"name":             d.Name(),
			"path_relative":    filepath.ToSlash(entryRelPath), // Use forward slashes for consistency.
			"is_dir":           d.IsDir(),
			"size_bytes":       info.Size(),
			"modified_unix":    info.ModTime().Unix(),
			"modified_rfc3339": info.ModTime().Format(time.RFC3339Nano),
			"mode_string":      info.Mode().String(),
		})
		return nil
	})

	if walkErr != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeIOFailed, fmt.Sprintf("failed directory walk for '%s'", relPath), errors.Join(lang.ErrIOFailed, walkErr))
	}

	return fileInfos, nil
}
