// filename: pkg/core/tools_fs_dirs.go
package core

import (
	"errors" // Import errors for defining local errors and checking os errors
	"fmt"
	"os"
	// "path/filepath" // Not needed directly here
)

// --- Errors specific to directory/file operations ---
// Moved ErrCannotCreateDir to errors.go

// --- Tool Implementations ---

// toolMkdir creates a directory, ensuring it's within the sandbox.
// (Implementation uses interpreter.sandboxDir as updated previously)
func toolMkdir(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation guarantees args[0] is a string
	relativePath := args[0].(string)
	sandboxRoot := interpreter.sandboxDir
	if sandboxRoot == "" {
		if interpreter.logger != nil {
			interpreter.logger.Printf("[WARN TOOL Mkdir] Interpreter sandboxDir is empty, using default relative path validation.")
		}
		sandboxRoot = "."
	}

	absPath, secErr := SecureFilePath(relativePath, sandboxRoot)
	if secErr != nil {
		errMsg := fmt.Sprintf("Mkdir path error for '%s': %s", relativePath, secErr.Error())
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL Mkdir] %s (Sandbox Root: %s)", errMsg, sandboxRoot)
		}
		return errMsg, secErr // Return error message and original error
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL Mkdir] Attempting to create directory: %s (Original: %s, Sandbox: %s)", absPath, relativePath, sandboxRoot)
	}
	mkdirErr := os.MkdirAll(absPath, 0755)
	if mkdirErr != nil {
		errMsg := fmt.Sprintf("Mkdir failed for '%s': %s", relativePath, mkdirErr.Error())
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL Mkdir] %s", errMsg)
		}
		// Check if the error is because a file exists at the path
		if _, pathErr := os.Stat(absPath); pathErr == nil {
			// A file/dir exists, check if it's a directory
			if info, statErr := os.Stat(absPath); statErr == nil && !info.IsDir() {
				// It exists and it's a file, wrap ErrCannotCreateDir
				return errMsg, fmt.Errorf("%w: path '%s' exists and is not a directory: %w", ErrCannotCreateDir, relativePath, mkdirErr)
			}
		}
		// Return generic creation error otherwise
		return errMsg, fmt.Errorf("%w: creating directory '%s': %w", ErrCannotCreateDir, relativePath, mkdirErr)
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL Mkdir] Successfully created directory: %s", relativePath)
	}
	return "OK", nil
}

// toolDeleteFile deletes a file or an empty directory within the sandbox.
func toolDeleteFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation guarantees args[0] is a string
	relativePath := args[0].(string)
	sandboxRoot := interpreter.sandboxDir
	if sandboxRoot == "" {
		if interpreter.logger != nil {
			interpreter.logger.Printf("[WARN TOOL DeleteFile] Interpreter sandboxDir is empty, using default relative path validation.")
		}
		sandboxRoot = "."
	}

	absPath, secErr := SecureFilePath(relativePath, sandboxRoot)
	if secErr != nil {
		errMsg := fmt.Sprintf("DeleteFile path error for '%s': %s", relativePath, secErr.Error())
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL DeleteFile] %s (Sandbox Root: %s)", errMsg, sandboxRoot)
		}
		return errMsg, secErr // Return error message and original error
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL DeleteFile] Attempting to delete: %s (Original: %s, Sandbox: %s)", absPath, relativePath, sandboxRoot)
	}

	// Attempt to remove the file/directory
	removeErr := os.Remove(absPath)

	if removeErr != nil {
		// Check if the error is 'file not found' - this is considered success for delete
		if errors.Is(removeErr, os.ErrNotExist) {
			if interpreter.logger != nil {
				interpreter.logger.Printf("[TOOL DeleteFile] Path '%s' not found, considered deleted successfully.", relativePath)
			}
			return "OK", nil // File already doesn't exist, that's fine
		}

		// Handle other errors (permissions, non-empty directory, etc.)
		errMsg := fmt.Sprintf("DeleteFile failed for '%s': %s", relativePath, removeErr.Error())
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL DeleteFile] %s", errMsg)
		}
		// Wrap the specific os error with our defined error
		return errMsg, fmt.Errorf("%w: deleting '%s': %w", ErrCannotDelete, relativePath, removeErr)
	}

	// Success
	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL DeleteFile] Successfully deleted: %s", relativePath)
	}
	return "OK", nil
}

// --- Registration Function ---

// registerFsDirTools registers directory/file operations like Mkdir, DeleteFile, ListDirectory.
func registerFsDirTools(registry *ToolRegistry) error {
	var err error // Declare err variable

	// --- Mkdir Registration ---
	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "Mkdir",
			Description: "Creates a directory (including any necessary parent directories) at the specified relative path within the sandbox.",
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "The relative path (within the sandbox) of the directory to create."},
			},
			ReturnType: ArgTypeString, // Returns "OK" or error message string
		},
		Func: toolMkdir,
	})
	if err != nil {
		return fmt.Errorf("failed to register FS tool Mkdir: %w", err)
	}

	// --- ListDirectory Registration ---
	// Note: Implementation toolListDirectory is in tools_fs_list.go
	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "ListDirectory",
			Description: "Lists directory content within the sandbox, returning a list of maps, each with 'name', 'is_dir', and 'size'.",
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "The relative path (within the sandbox) of the directory to list."},
			},
			ReturnType: ArgTypeSliceAny, // Returns slice of maps
		},
		Func: toolListDirectory, // Function defined in tools_fs_list.go
	})
	if err != nil {
		return fmt.Errorf("failed to register FS tool ListDirectory: %w", err)
	}

	// --- DeleteFile Registration ---
	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "DeleteFile",
			Description: "Deletes a file or an empty directory at the specified relative path within the sandbox. Returns 'OK' even if the file doesn't exist.",
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "The relative path (within the sandbox) of the file or empty directory to delete."},
			},
			ReturnType: ArgTypeString, // Returns "OK" or error message string
		},
		Func: toolDeleteFile,
	})
	if err != nil {
		return fmt.Errorf("failed to register FS tool DeleteFile: %w", err)
	}

	return nil // Success
}
