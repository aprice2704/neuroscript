// filename: pkg/core/tools_fs_dirs.go
package core

import (
	"errors" // Import errors for defining local errors
	"fmt"
	"os"
	// "path/filepath" // Not needed directly here
)

// --- Errors specific to directory operations ---
var (
	ErrCannotCreateDir = errors.New("cannot create directory")
	// Define ErrCannotDelete here later if needed
)

// --- Tool Implementations ---

// toolMkdir creates a directory, ensuring it's within the sandbox.
func toolMkdir(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation guarantees args[0] is a string
	relativePath := args[0].(string)

	cwd, errWd := os.Getwd() // Get current working directory (should be sandbox root in agent/secure mode)
	if errWd != nil {
		// This is an internal error, return it directly
		return nil, fmt.Errorf("TOOL.Mkdir failed to get working directory: %w", errWd)
	}

	// Validate the path using SecureFilePath first to ensure it's within bounds
	// SecureFilePath returns the *absolute* path if valid
	absPath, secErr := SecureFilePath(relativePath, cwd)
	if secErr != nil {
		// Path violation (absolute, outside CWD, etc.)
		errMsg := fmt.Sprintf("Mkdir path error for '%s': %s", relativePath, secErr.Error())
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL Mkdir] %s", errMsg)
		}
		// Return the error message string for NeuroScript, and the actual error for Go
		return errMsg, secErr // Return the specific path violation error (which wraps ErrPathViolation)
	}

	// Attempt to create the directory(ies) using the validated absolute path
	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL Mkdir] Attempting to create directory: %s (Original: %s)", absPath, relativePath)
	}
	mkdirErr := os.MkdirAll(absPath, 0755) // Use MkdirAll to create parent dirs if needed
	if mkdirErr != nil {
		errMsg := fmt.Sprintf("Mkdir failed for '%s': %s", relativePath, mkdirErr.Error())
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL Mkdir] %s", errMsg)
		}
		// Return error message string for NeuroScript, wrap internal error for Go
		// Wrap the OS error with our defined error type
		return errMsg, fmt.Errorf("%w: creating directory '%s': %w", ErrCannotCreateDir, relativePath, mkdirErr)
	}

	// Success
	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL Mkdir] Successfully created directory: %s", relativePath)
	}
	return "OK", nil
}

// --- Registration Function ---

// registerFsDirTools registers directory-specific filesystem tools.
func registerFsDirTools(registry *ToolRegistry) error {
	// --- Mkdir Registration ---
	err := registry.RegisterTool(ToolImplementation{
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

	// --- Register DeleteFile here later ---
	/*
		err = registry.RegisterTool(ToolImplementation{
			Spec: ToolSpec{ Name: "DeleteFile", ... },
			Func: toolDeleteFile,
		})
		if err != nil {
			return fmt.Errorf("failed to register FS tool DeleteFile: %w", err)
		}
	*/

	return nil // Success
}
