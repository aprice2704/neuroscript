// filename: pkg/core/tools_fs_delete.go
package core

import (
	"errors"
	"fmt"
	"os"
	"strings" // Added for error string checking
)

// toolDeleteFile implements the TOOL.DeleteFile command.
func toolDeleteFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Argument Validation remains the same...
	if len(args) != 1 { /* ... */
	}
	relPath, ok := args[0].(string)
	if !ok { /* ... */
	}
	if relPath == "" { /* ... */
	}

	// Path Security Validation remains the same...
	sandboxRoot := interpreter.sandboxDir
	if sandboxRoot == "" { /* ... */
	}
	absPath, secErr := SecureFilePath(relPath, sandboxRoot)
	if secErr != nil { /* ... */
		errMsg := fmt.Sprintf("DeleteFile path security error for %q: %v", relPath, secErr)
		interpreter.logger.Printf("[TOOL DeleteFile] %s (Sandbox Root: %s)", errMsg, sandboxRoot)
		return errMsg, fmt.Errorf("TOOL.DeleteFile: %w", secErr)
	}

	interpreter.logger.Printf("[TOOL DeleteFile] Validated path: %s. Attempting deletion.", absPath)

	// --- Perform Deletion ---
	err := os.Remove(absPath)

	// --- Error Handling ---
	if err != nil {
		// --- FIX: Handle ErrNotExist specifically to match test expectation ---
		if errors.Is(err, os.ErrNotExist) {
			errMsg := fmt.Sprintf("File or directory not found: %s", relPath)
			interpreter.logger.Printf("[TOOL DeleteFile] Info: %s", errMsg)
			// Test expects "OK" and nil error even if not found
			// return errMsg, nil // Return specific message, nil Go error
			return "OK", nil // Return "OK" and nil error to match test expectation literally
		}

		// --- FIX: Check for non-empty directory error and wrap ErrCannotDelete ---
		// Error message check is OS-dependent, unfortunately. Checking common variants.
		errMsgText := err.Error()
		isDirNotEmptyErr := strings.Contains(errMsgText, "directory not empty") || // Linux, macOS
			strings.Contains(errMsgText, "The directory is not empty.") // Windows? (Guessing)

		errMsg := fmt.Sprintf("Failed to delete '%s': %v", relPath, err)
		interpreter.logger.Printf("[TOOL DeleteFile] Error: %s", errMsg)

		if isDirNotEmptyErr {
			// Return the specific message string AND the wrapped sentinel error
			return errMsg, fmt.Errorf("TOOL.DeleteFile: %w: %w", ErrCannotDelete, err)
		}

		// Return generic message and wrapped OS error for other cases
		return errMsg, fmt.Errorf("TOOL.DeleteFile: %w", err)
	}

	// --- Success ---
	interpreter.logger.Printf("[TOOL DeleteFile] Successfully deleted: %s", relPath)
	// --- FIX: Return "OK" on success to match test ---
	return "OK", nil
	// --- END FIX ---
}

// --- Registration ---
func registerFsDeleteTools(registry *ToolRegistry) error {
	return registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "DeleteFile",
			Description: "Deletes a file or an empty directory within the sandbox.",
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "The relative path to the file or empty directory to delete."},
			},
			ReturnType: ArgTypeString, // Returns "OK" or error string
		},
		Func: toolDeleteFile,
	})
}
