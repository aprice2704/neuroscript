// filename: pkg/core/tools_fs_utils.go
package core

import (
	"fmt"
	"os"
	"strings"
	// No extra import needed for ArgTypeSliceAny as it's defined in tools_types.go
)

// --- Tool Implementations ---

// toolLineCountFile counts lines in a specified file.
func toolLineCountFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// ... (implementation remains unchanged) ...
	filePath := args[0].(string)
	sandboxRoot := interpreter.sandboxDir
	if sandboxRoot == "" {
		if interpreter.logger != nil {
			interpreter.logger.Warn("TOOL LineCountFile] Interpreter sandboxDir is empty, using default relative path validation.")
		}
		sandboxRoot = "."
	}
	absPath, secErr := SecureFilePath(filePath, sandboxRoot)
	if secErr != nil {
		if interpreter.logger != nil {
			interpreter.logger.Warn("TOOL LineCountFile] Path validation failed for '%s': %v. (Sandbox Root: %s)", filePath, secErr, sandboxRoot)
		}
		// Return consistent type on error path if possible, or handle differently
		return int64(-1), fmt.Errorf("LineCountFile path error: %w", secErr) // Return error for consistency?
	}
	if interpreter.logger != nil {
		interpreter.logger.Info("Tool: LineCountFile] Attempting to read validated path: %s (Original: %s, Sandbox: %s)", absPath, filePath, sandboxRoot)
	}
	contentBytes, readErr := os.ReadFile(absPath)
	if readErr != nil {
		if interpreter.logger != nil {
			interpreter.logger.Warn("TOOL LineCountFile] Read error for path '%s': %v.", filePath, readErr)
		}
		return int64(-1), fmt.Errorf("LineCountFile read error: %w", readErr) // Return error for consistency?
	}
	content := string(contentBytes)
	if len(content) == 0 {
		if interpreter.logger != nil {
			interpreter.logger.Info("Tool: LineCountFile] Counted 0 lines (empty file '%s').", filePath)
		}
		return int64(0), nil
	}
	lineCount := int64(strings.Count(content, "\n"))
	// Handle files that don't end with a newline
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		lineCount++
	}
	// Handle edge case of single newline file (strings.Count returns 1, which is correct)
	// if content == "\n" { // This check might be redundant now
	// 	lineCount = 1
	// }
	if interpreter.logger != nil {
		interpreter.logger.Info("Tool: LineCountFile] Counted %d lines in file '%s'.", lineCount, filePath)
	}
	return lineCount, nil
}

// toolSanitizeFilename calls the exported helper function.
func toolSanitizeFilename(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// ... (implementation remains unchanged) ...
	name := args[0].(string)
	sanitized := SanitizeFilename(name)
	if interpreter.logger != nil {
		interpreter.logger.Info("Tool: SanitizeFilename] Input: %q -> Output: %q", name, sanitized)
	}
	return sanitized, nil
}

// --- Tool Specifications ---

var lineCountFileSpec = ToolSpec{
	Name:        "LineCountFile",
	Description: "Counts lines in a specified file within the sandbox. Returns line count or error.", // Updated description slightly
	Args: []ArgSpec{
		{Name: "filepath", Type: ArgTypeString, Required: true, Description: "Relative path to the file."},
	},
	ReturnType: ArgTypeInt,
}

var sanitizeFilenameSpec = ToolSpec{
	Name:        "SanitizeFilename",
	Description: "Cleans a string to make it suitable for use as part of a filename.",
	Args: []ArgSpec{
		{Name: "name", Type: ArgTypeString, Required: true, Description: "The string to sanitize."},
	},
	ReturnType: ArgTypeString,
}

var walkDirSpec = ToolSpec{
	// *** CORRECTED Name to use Base Name ***
	Name:        "WalkDir",
	Description: "Recursively walks a directory, returning a list of maps describing files/subdirectories found.",
	Args: []ArgSpec{
		{Name: "path", Type: ArgTypeString, Required: true, Description: "Relative path to the directory to walk."},
	},
	ReturnType: ArgTypeSliceAny, // Expects a list of maps (represented as []interface{} internally)
}

// --- Tool Implementations Slice (for potential registration) ---

// Assumes toolWalkDir implementation exists in tools_fs_walk.go
var fsUtilTools = []ToolImplementation{
	{Spec: lineCountFileSpec, Func: toolLineCountFile},
	{Spec: sanitizeFilenameSpec, Func: toolSanitizeFilename},
	{Spec: walkDirSpec, Func: toolWalkDir}, // Uses the corrected walkDirSpec
}

// registerFsUtilTools registers the utility filesystem tools.
func registerFsUtilTools(registry *ToolRegistry) error {
	for _, tool := range fsUtilTools {
		if err := registry.RegisterTool(tool); err != nil {
			// Make error message more specific
			return fmt.Errorf("registering FS util tool '%s': %w", tool.Spec.Name, err)
		}
	}
	return nil
}
