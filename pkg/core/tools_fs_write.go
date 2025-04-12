// filename: pkg/core/tools_fs_write.go
package core

import (
	"fmt"
	"os"
	"path/filepath"
)

// toolWriteFile writes content to a specified file.
// Assumes path validation/sandboxing is handled by the SecurityLayer.
// *** MODIFIED: Propagate error from SecureFilePath ***
func toolWriteFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation guarantees args[0] and args[1] are strings
	filePath := args[0].(string)
	content := args[1].(string)

	cwd, errWd := os.Getwd()
	if errWd != nil {
		return nil, fmt.Errorf("TOOL WriteFile failed get CWD: %w", errWd) // Internal error
	}

	// Use SecureFilePath for consistency and absolute path resolution.
	// Security check against sandbox happened before tool call.
	absPath, secErr := SecureFilePath(filePath, cwd)
	if secErr != nil {
		// *** Propagate the path violation error ***
		errMsg := fmt.Sprintf("WriteFile path error for '%s': %s", filePath, secErr.Error()) // Log unwrapped
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL WriteFile] %s", errMsg)
		}
		// Return error message string for script, but the error itself for Go context
		return errMsg, secErr
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL WriteFile] Writing to validated path: %s (Original Relative: %s)", absPath, filePath)
	}

	// Ensure directory exists before writing
	dirPath := filepath.Dir(absPath)
	if dirErr := os.MkdirAll(dirPath, 0755); dirErr != nil {
		errMsg := fmt.Sprintf("WriteFile mkdir failed for dir '%s': %s", dirPath, dirErr.Error())
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL WriteFile] %s", errMsg)
		}
		return errMsg, fmt.Errorf("%w: creating directory '%s': %w", ErrInternalTool, dirPath, dirErr)
	}

	// Write the file
	writeErr := os.WriteFile(absPath, []byte(content), 0644)
	if writeErr != nil {
		errMsg := fmt.Sprintf("WriteFile failed for '%s': %s", filePath, writeErr.Error())
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL WriteFile] %s", errMsg)
		}
		return errMsg, fmt.Errorf("%w: writing file '%s': %w", ErrInternalTool, filePath, writeErr)
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL WriteFile] Wrote %d bytes successfully to %s", len(content), filePath)
	}
	// Return "OK" on success
	return "OK", nil
}
