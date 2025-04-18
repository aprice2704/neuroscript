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

// toolWalkDir recursively walks a directory within the sandbox and returns a list of maps
// containing file/directory information, using standard Go types.
func toolWalkDir(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// --- Argument Validation ---
	if len(args) != 1 {
		return nil, fmt.Errorf("expected 1 argument (path string), got %d", len(args))
	}
	relPath, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("expected argument 1 (path) to be a string, got %T", args[0])
	}
	if relPath == "" {
		return nil, fmt.Errorf("path argument cannot be empty: %w", ErrInvalidArgument)
	}

	// --- Path Security Validation ---
	sandboxRoot := interpreter.sandboxDir
	if sandboxRoot == "" {
		if interpreter.logger != nil {
			interpreter.logger.Printf("[WARN TOOL WalkDir] Interpreter sandboxDir is empty, using default relative path validation from current directory.")
		}
		sandboxRoot = "."
	}

	absBasePath, secErr := SecureFilePath(relPath, sandboxRoot)
	if secErr != nil {
		errMsg := fmt.Sprintf("WalkDir path security error for %q: %v", relPath, secErr)
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL WalkDir] %s (Sandbox Root: %s)", errMsg, sandboxRoot)
		}
		return nil, secErr
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL WalkDir] Validated base path: %s (Original Relative: %q, Sandbox: %q)", absBasePath, relPath, sandboxRoot)
	}

	// --- Check if Path is a Directory ---
	baseInfo, statErr := os.Stat(absBasePath)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			errMsg := fmt.Sprintf("WalkDir: Start path not found %q", relPath)
			if interpreter.logger != nil {
				interpreter.logger.Printf("[TOOL WalkDir] %s", errMsg)
			}
			return nil, nil // Return nil result, nil error if start path doesn't exist
		}
		errMsg := fmt.Sprintf("WalkDir: Failed to stat start path %q: %v", relPath, statErr)
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL WalkDir] %s", errMsg)
		}
		return nil, fmt.Errorf("failed getting info for path %q: %w", relPath, statErr)
	}

	if !baseInfo.IsDir() {
		errMsg := fmt.Sprintf("WalkDir: Start path %q is not a directory", relPath)
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL WalkDir] %s", errMsg)
		}
		return nil, fmt.Errorf("%s: %w", errMsg, ErrInvalidArgument)
	}

	// --- Walk the Directory ---
	// *** FIXED: Use standard Go slice of maps ***
	var fileInfos []map[string]interface{}

	walkErr := filepath.WalkDir(absBasePath, func(currentPath string, d fs.DirEntry, err error) error {
		if err != nil {
			if interpreter.logger != nil {
				interpreter.logger.Printf("[TOOL WalkDir] Error accessing %q during walk: %v", currentPath, err)
			}
			if errors.Is(err, fs.ErrPermission) {
				return fmt.Errorf("permission error accessing %q: %w", currentPath, err)
			}
			return fmt.Errorf("error accessing %q: %w", currentPath, err)
		}

		if currentPath == absBasePath {
			return nil // Skip root
		}

		info, infoErr := d.Info()
		if infoErr != nil {
			if interpreter.logger != nil {
				interpreter.logger.Printf("[TOOL WalkDir] Error getting FileInfo for %q: %v", currentPath, infoErr)
			}
			return fmt.Errorf("failed getting FileInfo for %q: %w", currentPath, infoErr)
		}

		entryRelPath, relErr := filepath.Rel(absBasePath, currentPath)
		if relErr != nil {
			if interpreter.logger != nil {
				interpreter.logger.Printf("[TOOL WalkDir] Error calculating relative path for %q (base %q): %v", currentPath, absBasePath, relErr)
			}
			return fmt.Errorf("internal error calculating relative path for %q: %w", currentPath, relErr)
		}

		// *** FIXED: Create standard map[string]interface{} ***
		entryMap := make(map[string]interface{})
		entryMap["name"] = d.Name()                               // string
		entryMap["path"] = filepath.ToSlash(entryRelPath)         // string
		entryMap["isDir"] = d.IsDir()                             // bool
		entryMap["size"] = info.Size()                            // int64
		entryMap["modTime"] = info.ModTime().Format(time.RFC3339) // string

		// Append the standard map to the standard slice
		fileInfos = append(fileInfos, entryMap)

		return nil // Continue walking
	})

	// --- Handle Errors from WalkDir ---
	if walkErr != nil {
		errMsg := fmt.Sprintf("WalkDir: Failed walking directory %q: %v", relPath, walkErr)
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL WalkDir] %s", errMsg)
		}
		return nil, fmt.Errorf("failed walking directory %q: %w", relPath, walkErr)
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL WalkDir] Walk successful for %q. Found %d entries.", relPath, len(fileInfos))
	}

	// --- Return Result ---
	// *** FIXED: Return the standard Go slice directly ***
	return fileInfos, nil
}
