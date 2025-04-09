// filename: pkg/core/tools_fs_write.go
package core

import (
	"fmt"
	"os"
	"path/filepath"
)

// toolWriteFile writes content to a specified file.
// Assumes path validation/sandboxing is handled by the SecurityLayer.
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
		return fmt.Sprintf("WriteFile path error for '%s': %s", filePath, secErr.Error()), nil
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL WriteFile] Writing to validated path: %s (Original Relative: %s)", absPath, filePath)
	}

	// Ensure directory exists before writing
	dirPath := filepath.Dir(absPath)
	if dirErr := os.MkdirAll(dirPath, 0755); dirErr != nil {
		return fmt.Sprintf("WriteFile mkdir failed for dir '%s': %s", dirPath, dirErr.Error()), nil
	}

	// Write the file
	writeErr := os.WriteFile(absPath, []byte(content), 0644)
	if writeErr != nil {
		return fmt.Sprintf("WriteFile failed for '%s': %s", filePath, writeErr.Error()), nil
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL WriteFile] Wrote %d bytes successfully to %s", len(content), filePath)
	}
	// Return "OK" on success
	return "OK", nil
}
