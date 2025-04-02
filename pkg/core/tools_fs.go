// pkg/core/tools_fs.go
package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings" // Needed for LineCount placeholder
)

// registerFsTools adds File System related tools to the registry.
func registerFsTools(registry *ToolRegistry) {
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "ReadFile",
			Description: "Reads the content of a file.",
			Args: []ArgSpec{
				{Name: "filepath", Type: ArgTypeString, Required: true, Description: "Relative path to the file."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolReadFile,
	})

	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "WriteFile",
			Description: "Writes content to a file, creating directories if needed.",
			Args: []ArgSpec{
				{Name: "filepath", Type: ArgTypeString, Required: true, Description: "Relative path to the file."},
				{Name: "content", Type: ArgTypeString, Required: true, Description: "The content to write."},
			},
			ReturnType: ArgTypeString, // Returns "OK" or error message
		},
		Func: toolWriteFile,
	})

	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "ListDirectory",
			Description: "Lists the files and subdirectories within a given directory path, adding '/' suffix to directories.",
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "Relative path to the directory."},
			},
			ReturnType: ArgTypeSliceString, // Returns []string of names or error message string
		},
		Func: toolListDirectory,
	})

	// Placeholder - Keep registration but function is not implemented
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "LineCount",
			Description: "Counts the lines in a file or string. [NOT IMPLEMENTED]",
			Args: []ArgSpec{
				{Name: "input", Type: ArgTypeString, Required: true, Description: "File path or string content."},
			},
			ReturnType: ArgTypeInt,
		},
		Func: toolLineCount,
	})

	// Keep SanitizeFilename here as it's FS path related
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "SanitizeFilename",
			Description: "Cleans a string to make it suitable for use as a filename.",
			Args: []ArgSpec{
				{Name: "name", Type: ArgTypeString, Required: true, Description: "The input name string."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolSanitizeFilename,
	})
}

// toolReadFile reads content from a secured file path.
func toolReadFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation handled by ValidateAndConvertArgs
	filePath := args[0].(string)

	cwd, errWd := os.Getwd()
	if errWd != nil {
		return nil, fmt.Errorf("ReadFile failed to get working directory: %w", errWd)
	}
	absPath, secErr := secureFilePath(filePath, cwd) // Use helper
	if secErr != nil {
		// Return error message as string result for NeuroScript
		return fmt.Sprintf("ReadFile failed for '%s': %s", filePath, secErr.Error()), nil
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.ReadFile for %s (Resolved: %s)", filePath, absPath)
	}

	contentBytes, readErr := os.ReadFile(absPath)
	if readErr != nil {
		// Return read error message as string result
		return fmt.Sprintf("ReadFile failed for '%s': %s", filePath, readErr.Error()), nil
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      ReadFile successful for %s", filePath)
	}
	return string(contentBytes), nil
}

// toolWriteFile writes content to a secured file path, creating directories.
func toolWriteFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation handled by ValidateAndConvertArgs
	filePath := args[0].(string)
	content := args[1].(string)

	cwd, errWd := os.Getwd()
	if errWd != nil {
		return nil, fmt.Errorf("WriteFile failed to get working directory: %w", errWd)
	}
	absPath, secErr := secureFilePath(filePath, cwd) // Use helper
	if secErr != nil {
		return fmt.Sprintf("WriteFile path error: %s", secErr.Error()), nil
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.WriteFile for %s (Resolved: %s)", filePath, absPath)
	}

	// Ensure the directory exists
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
		interpreter.logger.Printf("[DEBUG-INTERP]      WriteFile successful for %s", filePath)
	}
	return "OK", nil
}

// toolListDirectory lists the contents of a secured directory path.
func toolListDirectory(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation handled by ValidateAndConvertArgs
	dirPath := args[0].(string)

	cwd, errWd := os.Getwd()
	if errWd != nil {
		return fmt.Sprintf("ListDirectory failed to get working directory: %s", errWd.Error()), nil
	}

	// Secure the path - allow listing the CWD itself if path is "."
	var absPath string
	var secErr error
	if filepath.Clean(dirPath) == "." {
		absPath = cwd // Use cwd directly if input is "."
	} else {
		absPath, secErr = secureFilePath(dirPath, cwd) // Use helper
		if secErr != nil {
			return fmt.Sprintf("ListDirectory path error: %s", secErr.Error()), nil
		}
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.ListDirectory for %s (Resolved: %s)", dirPath, absPath)
	}

	// Read directory entries
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return fmt.Sprintf("ListDirectory read error for '%s': %s", dirPath, err.Error()), nil
	}

	// Extract names, add '/' to dirs, and sort
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/" // Add trailing slash to indicate directory
		}
		names = append(names, name)
	}
	sort.Strings(names) // Ensure deterministic order

	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      ListDirectory successful for %s. Found %d entries.", dirPath, len(names))
	}

	// Return the list of names (which is []string, compatible with ArgTypeSliceString)
	return names, nil
}

// toolLineCount counts lines in a file or string content. [Placeholder]
func toolLineCount(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation handled by ValidateAndConvertArgs
	input := args[0].(string)

	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.LineCount for input (first 50 chars): %q...", input[:min(len(input), 50)])
	}

	// TODO: Implement LineCount logic.
	// - Check if input is a likely file path (needs robust check) or string content.
	// - If path, use secureFilePath and read the file.
	// - Count newlines ('\n') in the content.
	// - Return int64 count or error string.

	// Placeholder implementation
	lineCount := int64(strings.Count(input, "\n") + 1) // Basic count for now
	if len(input) == 0 {
		lineCount = 0
	}
	interpreter.logger.Printf("[WARN] TOOL.LineCount is not fully implemented. Using basic newline count.")

	return lineCount, nil // Placeholder return
}

// toolSanitizeFilename cleans a string for use as a filename.
func toolSanitizeFilename(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation handled by ValidateAndConvertArgs
	name := args[0].(string)
	sanitized := sanitizeFilename(name) // Use helper from utils.go (or move here/tools_helpers.go later)
	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      SanitizeFilename: %q -> %q", name, sanitized)
	}
	return sanitized, nil
}
