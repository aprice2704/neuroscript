// filename: pkg/core/tools_fs.go
package core

import "fmt" // Keep fmt

// registerFsTools registers all filesystem-related tools by calling specific registration functions.
func registerFsTools(registry *ToolRegistry) error {
	// Register file-specific tools (Read, Write, Utils)
	if err := registerFsFileTools(registry); err != nil {
		return fmt.Errorf("failed registering file tools: %w", err)
	}
	// Register directory-specific tools (List, Mkdir, Delete later)
	if err := registerFsDirTools(registry); err != nil {
		return fmt.Errorf("failed registering directory tools: %w", err)
	}
	// Register utility tools (LineCountFile, SanitizeFilename)
	if err := registerFsUtilTools(registry); err != nil {
		return fmt.Errorf("failed registering FS utility tools: %w", err)
	}

	return nil // Success
}

// --- Registration helpers for specific categories ---
// (These would contain the actual ToolImplementation structs and RegisterTool calls)

// registerFsFileTools registers ReadFile, WriteFile
func registerFsFileTools(registry *ToolRegistry) error {
	tools := []ToolImplementation{
		{Spec: ToolSpec{Name: "ReadFile", Description: "Reads the entire content of a specific file...", Args: []ArgSpec{{Name: "filepath", Type: ArgTypeString, Required: true, Description: "The relative path..."}}, ReturnType: ArgTypeString}, Func: toolReadFile},
		{Spec: ToolSpec{Name: "WriteFile", Description: "Writes content to a specific file...", Args: []ArgSpec{{Name: "filepath", Type: ArgTypeString, Required: true}, {Name: "content", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}, Func: toolWriteFile},
	}
	for _, tool := range tools {
		if err := registry.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register file tool %s: %w", tool.Spec.Name, err)
		}
	}
	return nil
}

// registerFsUtilTools registers LineCountFile, SanitizeFilename
func registerFsUtilTools(registry *ToolRegistry) error {
	tools := []ToolImplementation{
		{Spec: ToolSpec{Name: "LineCountFile", Description: "Counts lines in a specified file...", Args: []ArgSpec{{Name: "filepath", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeInt}, Func: toolLineCountFile},
		{Spec: ToolSpec{Name: "SanitizeFilename", Description: "Cleans a string to make it suitable for use as part of a filename.", Args: []ArgSpec{{Name: "name", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}, Func: toolSanitizeFilename},
		{Spec: ToolSpec{Name: "ListDirectory", Description: "Lists directory content within the sandbox...", Args: []ArgSpec{{Name: "path", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeSliceAny}, Func: toolListDirectory}, // Moved ListDirectory here conceptually
	}
	for _, tool := range tools {
		if err := registry.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register FS util tool %s: %w", tool.Spec.Name, err)
		}
	}
	return nil
}

// Note: Implementations like toolReadFile, toolWriteFile, toolListDirectory,
// toolLineCountFile, toolSanitizeFilename would remain in their respective
// implementation files (e.g., tools_fs_read.go, tools_fs_write.go, etc.)
// The registerFsDirTools function is defined in the new tools_fs_dirs.go file.
