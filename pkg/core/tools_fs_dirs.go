// filename: pkg/core/tools_fs_dirs.go
package core

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings" // Added for Mkdir path check fix
	"time"    // Needed again for ModTime
)

// registerFsDirTools registers directory-related filesystem tools.
func registerFsDirTools(registry *ToolRegistry) error {
	// Register ListDirectory
	err := registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name: "ListDirectory",
			// --- Standardized Description ---
			Description: "Lists the contents (files and subdirectories) of a specified directory. " +
				"Returns a list of maps, each containing 'name', 'path', 'isDir', 'size', 'modTime'.", // Added modTime back
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "The relative path to the directory to list."},
				{Name: "recursive", Type: ArgTypeBool, Required: false, Description: "If true, list contents recursively. Defaults to false."},
			},
			ReturnType: ArgTypeAny, // List of maps -> Any
		},
		Func: toolListDirectory, // Points to the function below in this file
	})
	if err != nil {
		return fmt.Errorf("failed to register tool ListDirectory: %w", err)
	}

	// Register Mkdir
	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "Mkdir",
			Description: "Creates a new directory (including any necessary parents) within the sandbox.",
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "The relative path of the directory to create."},
			},
			ReturnType: ArgTypeString, // Returns success message or error string
		},
		Func: toolMkdir, // Points to the function below in this file
	})
	if err != nil {
		return fmt.Errorf("failed to register tool Mkdir: %w", err)
	}
	return nil // Success
}

// --- Implementations ---

// toolListDirectory lists contents of a directory, now with recursion.
// Returns []map[string]interface{} with keys: name, path, isDir, size, modTime
func toolListDirectory(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// --- Argument Validation ---
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("TOOL.ListDirectory: expected 1 or 2 arguments (path, [recursive]), got %d", len(args))
	}
	relPath, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.ListDirectory: expected argument 1 (path) to be a string, got %T", args[0])
	}
	if relPath == "" {
		return nil, fmt.Errorf("TOOL.ListDirectory: path argument cannot be empty: %w", ErrInvalidArgument)
	}

	recursive := false // Default value
	if len(args) == 2 {
		recursiveVal, okBool := args[1].(bool)
		if !okBool {
			if args[1] != nil { // Only error if it's non-nil and not a bool
				return nil, fmt.Errorf("TOOL.ListDirectory: expected argument 2 (recursive) to be a boolean or null, got %T", args[1])
			}
		} else {
			recursive = recursiveVal
		}
	}

	// --- Path Security Validation ---
	sandboxRoot := interpreter.sandboxDir
	if sandboxRoot == "" {
		interpreter.logger.Warn("TOOL ListDirectory] Interpreter sandboxDir is empty, using default relative path validation from current directory.")
		sandboxRoot = "."
	}
	absBasePath, secErr := SecureFilePath(relPath, sandboxRoot)
	if secErr != nil {
		errMsg := fmt.Sprintf("ListDirectory path security error for %q: %v", relPath, secErr)
		interpreter.logger.Info("Tool: ListDirectory] %s (Sandbox Root: %s)", errMsg, sandboxRoot)
		// Propagate security error directly
		return nil, fmt.Errorf("TOOL.ListDirectory: %w", secErr)
	}
	interpreter.logger.Info("Tool: ListDirectory] Validated base path: %s (Original Relative: %q, Sandbox: %q, Recursive: %t)", absBasePath, relPath, sandboxRoot, recursive)

	// --- Check if Path is a Directory ---
	baseInfo, statErr := os.Stat(absBasePath)
	if statErr != nil {
		// Return specific errors based on stat failure
		if errors.Is(statErr, os.ErrNotExist) {
			interpreter.logger.Info("Tool: ListDirectory] Path not found %q", relPath)
			// Tests expect ErrInternalTool here, maintain consistency for now
			return nil, fmt.Errorf("TOOL.ListDirectory: %w: %w", ErrInternalTool, statErr)
		}
		interpreter.logger.Info("Tool: ListDirectory] Failed to stat path %q: %v", relPath, statErr)
		// Tests expect ErrInternalTool here
		return nil, fmt.Errorf("TOOL.ListDirectory: %w: %w", ErrInternalTool, statErr)
	}

	if !baseInfo.IsDir() {
		errMsg := fmt.Sprintf("ListDirectory: Path %q is not a directory", relPath)
		interpreter.logger.Info("Tool: ListDirectory] %s", errMsg)
		// Tests expect ErrInternalTool here, wrapping ErrInvalidArgument
		return nil, fmt.Errorf("TOOL.ListDirectory: %w: %w", ErrInternalTool, ErrInvalidArgument)
	}

	// --- List Directory Contents ---
	var fileInfos []map[string]interface{} // Return correct type
	var listErr error

	if recursive {
		interpreter.logger.Info("Tool: ListDirectory] Walking recursively...")
		walkErr := filepath.WalkDir(absBasePath, func(currentPath string, d fs.DirEntry, err error) error {
			if err != nil {
				interpreter.logger.Info("Tool: ListDirectory Walk] Error accessing %q during walk: %v", currentPath, err)
				return fmt.Errorf("error accessing %q: %w", currentPath, err)
			}
			if currentPath == absBasePath { // Skip root
				return nil
			}

			info, infoErr := d.Info()
			if infoErr != nil {
				interpreter.logger.Info("Tool: ListDirectory Walk] Error getting FileInfo for %q: %v", currentPath, infoErr)
				return nil // Skip entry
			}

			entryRelPath, relErr := filepath.Rel(absBasePath, currentPath)
			if relErr != nil {
				interpreter.logger.Info("Tool: ListDirectory Walk] Error calculating relative path for %q (base %q): %v", currentPath, absBasePath, relErr)
				return nil // Skip entry
			}

			// --- Standardized Map ---
			entryMap := make(map[string]interface{})
			entryMap["name"] = d.Name()
			entryMap["path"] = filepath.ToSlash(entryRelPath)
			entryMap["isDir"] = d.IsDir() // Use camelCase
			entryMap["size"] = info.Size()
			entryMap["modTime"] = info.ModTime().Format(time.RFC3339) // Add ModTime back

			fileInfos = append(fileInfos, entryMap)
			return nil
		})
		listErr = walkErr

	} else {
		interpreter.logger.Info("Tool: ListDirectory] Reading directory non-recursively...")
		entries, readErr := os.ReadDir(absBasePath)
		if readErr != nil {
			listErr = fmt.Errorf("failed reading directory %q: %w", relPath, readErr)
		} else {
			for _, entry := range entries {
				info, infoErr := entry.Info()
				var size int64 = 0
				var modTime time.Time = time.Time{} // Zero time
				if infoErr == nil && info != nil {
					size = info.Size()
					modTime = info.ModTime()
				} else {
					interpreter.logger.Info("Tool: ListDirectory] Error getting FileInfo for %q: %v", entry.Name(), infoErr)
				}

				// --- Standardized Map ---
				entryMap := make(map[string]interface{})
				entryMap["name"] = entry.Name()
				entryMap["path"] = filepath.ToSlash(entry.Name()) // Path is just name
				entryMap["isDir"] = entry.IsDir()                 // Use camelCase
				entryMap["size"] = size
				entryMap["modTime"] = modTime.Format(time.RFC3339) // Add ModTime back

				fileInfos = append(fileInfos, entryMap)
			}
		}
	}

	// --- Handle Errors from Listing ---
	if listErr != nil {
		errMsg := fmt.Sprintf("ListDirectory: Failed listing directory %q (Recursive: %t): %v", relPath, recursive, listErr)
		interpreter.logger.Info("Tool: ListDirectory] %s", errMsg)
		// Tests expect ErrInternalTool here
		return nil, fmt.Errorf("TOOL.ListDirectory: %w: %w", ErrInternalTool, listErr)
	}

	interpreter.logger.Info("Tool: ListDirectory] Listing successful for %q (Recursive: %t). Found %d entries.", relPath, recursive, len(fileInfos))

	// --- REMOVED conversion to []interface{} ---
	return fileInfos, nil // Return []map[string]interface{} directly
}

// toolMkdir implements TOOL.Mkdir
func toolMkdir(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Argument Validation
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

	// Path Security Validation
	sandboxRoot := interpreter.sandboxDir
	if sandboxRoot == "" {
		interpreter.logger.Warn("TOOL Mkdir] Interpreter sandboxDir is empty, using default relative path validation from current directory.")
		sandboxRoot = "."
	}
	parentDir := filepath.Dir(relPath)
	// Fix for parent being root (".")
	if parentDir == "." && relPath != "." && !strings.Contains(relPath, string(filepath.Separator)) {
		parentDir = "." // Treat as relative to current sandbox if creating top-level dir
	} else if parentDir == "" { // Should not happen if relPath is not empty, but safer
		parentDir = "."
	}

	absBasePath, secErr := SecureFilePath(parentDir, sandboxRoot)
	if secErr != nil {
		errMsg := fmt.Sprintf("Mkdir path security error for parent of %q: %v", relPath, secErr)
		interpreter.logger.Info("Tool: Mkdir] %s (Sandbox Root: %s)", errMsg, sandboxRoot)
		return errMsg, fmt.Errorf("TOOL.Mkdir: %w", secErr)
	}

	// Construct the full absolute path to create
	absPathToCreate := filepath.Join(absBasePath, filepath.Base(relPath))

	// Security double-check using Clean and HasPrefix
	cleanAbsPathToCreate := filepath.Clean(absPathToCreate)
	cleanAbsBasePath := filepath.Clean(absBasePath)

	// Check if the cleaned path is still prefixed by the cleaned base path
	// Handle the case where the base path is the root (".") separately
	isOutside := false
	if cleanAbsBasePath == "." {
		// If sandbox is current dir, ensure path doesn't become absolute or go up
		if filepath.IsAbs(cleanAbsPathToCreate) || strings.HasPrefix(cleanAbsPathToCreate, ".."+string(filepath.Separator)) {
			isOutside = true
		}
	} else {
		// Check if it starts with the base path + separator, or is exactly the base path
		if !strings.HasPrefix(cleanAbsPathToCreate, cleanAbsBasePath+string(filepath.Separator)) && cleanAbsPathToCreate != cleanAbsBasePath {
			isOutside = true
		}
	}

	if isOutside {
		secErr = fmt.Errorf("%w: resultant path '%s' escapes validated base '%s'", ErrPathViolation, relPath, parentDir)
		errMsg := fmt.Sprintf("Mkdir path security error for %q: %v", relPath, secErr)
		interpreter.logger.Info("Tool: Mkdir] Error: %s", errMsg)
		return errMsg, fmt.Errorf("TOOL.Mkdir: %w", secErr)
	}

	interpreter.logger.Info("Tool: Mkdir] Validated base path: %s. Attempting to create: %s", cleanAbsBasePath, cleanAbsPathToCreate)

	// Create Directory
	err := os.MkdirAll(cleanAbsPathToCreate, 0755)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to create directory '%s': %v", relPath, err)
		interpreter.logger.Info("Tool: Mkdir] %s", errMsg)
		return errMsg, fmt.Errorf("TOOL.Mkdir: %w", err)
	}

	// Success
	successMsg := fmt.Sprintf("Successfully created directory: %s", relPath)
	interpreter.logger.Info("Tool: Mkdir] %s", successMsg)
	return successMsg, nil
}
