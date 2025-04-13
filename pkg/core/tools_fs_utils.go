// filename: pkg/core/tools_fs_utils.go
package core

import (
	"fmt" // *** ADDED fmt import ***
	"os"  // Keep for os.ReadFile
	"strings"
)

// toolLineCountFile counts lines in a specified file.
// Returns -1 on any path validation or file read error.
// *** MODIFIED: Use interpreter.sandboxDir instead of os.Getwd() ***
func toolLineCountFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures args[0] is a string
	filePath := args[0].(string)
	// *** Get sandbox root directly from the interpreter ***
	sandboxRoot := interpreter.sandboxDir // Use the field name you added
	if sandboxRoot == "" {
		if interpreter.logger != nil {
			interpreter.logger.Printf("[WARN TOOL LineCountFile] Interpreter sandboxDir is empty, using default relative path validation.")
		}
		sandboxRoot = "." // Ensure it's at least relative to CWD if empty
	}

	// Validate path relative to sandboxDir using SecureFilePath
	absPath, secErr := SecureFilePath(filePath, sandboxRoot) // *** Use sandboxRoot ***
	if secErr != nil {
		// Path validation failed (absolute, outside sandboxDir, etc.)
		if interpreter.logger != nil {
			interpreter.logger.Printf("[WARN TOOL LineCountFile] Path validation failed for '%s': %v. (Sandbox Root: %s)", filePath, secErr, sandboxRoot)
		}
		// Return specific error code and nil Go error for script level
		return int64(-1), nil // Indicate failure to the script
	}

	// Path is valid, attempt to read the file using the absolute path
	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL LineCountFile] Attempting to read validated path: %s (Original: %s, Sandbox: %s)", absPath, filePath, sandboxRoot)
	}
	contentBytes, readErr := os.ReadFile(absPath)
	if readErr != nil {
		// Read error (not found, permissions, etc.)
		if interpreter.logger != nil {
			interpreter.logger.Printf("[WARN TOOL LineCountFile] Read error for path '%s': %v.", filePath, readErr)
		}
		// Return specific error code and nil Go error for script level
		return int64(-1), nil // Indicate failure to the script
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

// registerFsUtilTools needs to be defined elsewhere or incorporated into registerFsTools
// Assuming it's called correctly by registerFsTools
func registerFsUtilTools(registry *ToolRegistry) error {
	tools := []ToolImplementation{
		{Spec: ToolSpec{Name: "LineCountFile", Description: "Counts lines in a specified file...", Args: []ArgSpec{{Name: "filepath", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeInt}, Func: toolLineCountFile},
		{Spec: ToolSpec{Name: "SanitizeFilename", Description: "Cleans a string to make it suitable for use as part of a filename.", Args: []ArgSpec{{Name: "name", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}, Func: toolSanitizeFilename},
		// ListDirectory registration moved to tools_fs_list.go or registerFsDirTools
	}
	for _, tool := range tools {
		if err := registry.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register FS util tool %s: %w", tool.Spec.Name, err)
		}
	}
	return nil
}
