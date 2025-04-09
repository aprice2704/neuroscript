// filename: pkg/core/tools_fs_utils.go
package core

import (
	"os" // Keep for SecureFilePath call
	"strings"
)

// toolLineCountFile counts lines in a specified file.
// Returns -1 on any path validation or file read error.
func toolLineCountFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures args[0] is a string
	filePath := args[0].(string)
	cwd, errWd := os.Getwd()
	if errWd != nil {
		if interpreter.logger != nil {
			interpreter.logger.Printf("[ERROR TOOL LineCountFile] failed get CWD: %v", errWd)
		}
		return int64(-1), nil // Return error code
	}

	// Validate path relative to CWD using SecureFilePath
	absPath, secErr := SecureFilePath(filePath, cwd)
	if secErr != nil {
		// Path validation failed (absolute, outside CWD, etc.)
		if interpreter.logger != nil {
			interpreter.logger.Printf("[WARN TOOL LineCountFile] Path validation failed for '%s': %v.", filePath, secErr)
		}
		return int64(-1), nil // Return error code
	}

	// Path is valid, attempt to read the file
	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL LineCountFile] Attempting to read validated path: %s (Original: %s)", absPath, filePath)
	}
	contentBytes, readErr := os.ReadFile(absPath)
	if readErr != nil {
		// Read error (not found, permissions, etc.)
		if interpreter.logger != nil {
			interpreter.logger.Printf("[WARN TOOL LineCountFile] Read error for path '%s': %v.", filePath, readErr)
		}
		return int64(-1), nil // Return error code
	}

	// Successfully read file, now count lines
	content := string(contentBytes)
	if len(content) == 0 {
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL LineCountFile] Counted 0 lines (empty file '%s').", filePath)
		}
		return int64(0), nil
	}
	lineCount := int64(strings.Count(content, "\n"))
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		lineCount++
	}
	if content == "\n" {
		lineCount = 1
	} // Handle single newline case

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL LineCountFile] Counted %d lines in file '%s'.", lineCount, filePath)
	}
	return lineCount, nil
}

// toolSanitizeFilename calls the exported helper function.
// (Implementation unchanged)
func toolSanitizeFilename(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	name := args[0].(string)
	sanitized := SanitizeFilename(name) // Calls exported helper
	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL SanitizeFilename] Input: %q -> Output: %q", name, sanitized)
	}
	return sanitized, nil
}
