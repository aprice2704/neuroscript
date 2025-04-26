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
	// ... (implementation as before) ...
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
		return int64(-1), nil
	}
	if interpreter.logger != nil {
		interpreter.logger.Info("Tool: LineCountFile] Attempting to read validated path: %s (Original: %s, Sandbox: %s)", absPath, filePath, sandboxRoot)
	}
	contentBytes, readErr := os.ReadFile(absPath)
	if readErr != nil {
		if interpreter.logger != nil {
			interpreter.logger.Warn("TOOL LineCountFile] Read error for path '%s': %v.", filePath, readErr)
		}
		return int64(-1), nil
	}
	content := string(contentBytes)
	if len(content) == 0 {
		if interpreter.logger != nil {
			interpreter.logger.Info("Tool: LineCountFile] Counted 0 lines (empty file '%s').", filePath)
		}
		return int64(0), nil
	}
	lineCount := int64(strings.Count(content, "\n"))
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		lineCount++
	}
	if content == "\n" {
		lineCount = 1
	}
	if interpreter.logger != nil {
		interpreter.logger.Info("Tool: LineCountFile] Counted %d lines in file '%s'.", lineCount, filePath)
	}
	return lineCount, nil
}

// toolSanitizeFilename calls the exported helper function.
func toolSanitizeFilename(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// ... (implementation as before) ...
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
	Description: "Counts lines in a specified file within the sandbox. Returns line count or -1 on error.",
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
	Name:        "TOOL.WalkDir",
	Description: "Recursively walks a directory, returning a list of maps describing files/subdirectories found.",
	Args: []ArgSpec{
		{Name: "path", Type: ArgTypeString, Required: true, Description: "Relative path to the directory to walk."},
	},
	// *** FIXED: Use the correct ArgType constant ***
	ReturnType: ArgTypeSliceAny, // Expects a list of maps (represented as []interface{} internally)
}

// --- Tool Implementations Slice (for potential registration) ---

var fsUtilTools = []ToolImplementation{
	{Spec: lineCountFileSpec, Func: toolLineCountFile},
	{Spec: sanitizeFilenameSpec, Func: toolSanitizeFilename},
	{Spec: walkDirSpec, Func: toolWalkDir}, // Assumes toolWalkDir is defined in tools_fs_walk.go
}

// registerFsUtilTools registers the utility filesystem tools.
// Note: This registration pattern might vary across the project.
// Ensure TOOL.WalkDir is actually registered where appropriate (e.g., in tools_register.go).
func registerFsUtilTools(registry *ToolRegistry) error {
	for _, tool := range fsUtilTools {
		if err := registry.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register FS util tool %s: %w", tool.Spec.Name, err)
		}
	}
	return nil
}
