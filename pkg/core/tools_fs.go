// filename: pkg/core/tools_fs.go
package core

import "fmt" // Keep fmt

// registerFsTools registers all filesystem-related tools by calling specific registration functions.
func registerFsTools(registry *ToolRegistry) error {
	// Register file-specific tools (Read, Write)
	if err := registerFsFileTools(registry); err != nil {
		return fmt.Errorf("failed registering file tools: %w", err)
	}
	// Register directory-specific tools (List, Mkdir, Delete later)
	// if err := registerFsDirTools(registry); err != nil {
	// 	return fmt.Errorf("failed registering directory tools: %w", err)
	// }
	// Register utility tools (LineCountFile, SanitizeFilename)
	if err := registerFsUtilTools(registry); err != nil {
		return fmt.Errorf("failed registering FS utility tools: %w", err)
	}
	// +++ ADDED: Register hash tool +++
	if err := registerFsHashTools(registry); err != nil {
		return fmt.Errorf("failed registering FS hash tools: %w", err)
	}
	// +++ ADDED: Register move tool +++
	// if err := registerFsMoveTools(registry); err != nil {
	// 	return fmt.Errorf("failed registering FS move tools: %w", err)
	// }
	// +++ ADDED: Register delete tool +++
	// if err := registerFsDeleteTools(registry); err != nil { // Call added
	// 	return fmt.Errorf("failed registering FS delete tools: %w", err)
	// }

	return nil // Success
}

// --- Registration helpers for specific categories ---
// (These would contain the actual ToolImplementation structs and RegisterTool calls)

// registerFsFileTools registers ReadFile, WriteFile
// Defined in tools_fs_read.go and tools_fs_write.go
func registerFsFileTools(registry *ToolRegistry) error {
	// Implementations are in tools_fs_read.go and tools_fs_write.go
	// This function ensures they are called correctly.
	// For brevity, assuming implementations like toolReadFile, toolWriteFile exist.
	tools := []ToolImplementation{
		{Spec: ToolSpec{Name: "ReadFile", Description: "Reads the entire content of a specific file...", Args: []ArgSpec{{Name: "filepath", Type: ArgTypeString, Required: true, Description: "The relative path..."}}, ReturnType: ArgTypeString}, Func: toolReadFile},            // Assumes toolReadFile exists
		{Spec: ToolSpec{Name: "WriteFile", Description: "Writes content to a specific file...", Args: []ArgSpec{{Name: "filepath", Type: ArgTypeString, Required: true}, {Name: "content", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}, Func: toolWriteFile}, // Assumes toolWriteFile exists
	}
	for _, tool := range tools {
		if err := registry.RegisterTool(tool); err != nil {
			// Log or return error specific to registration phase
			return fmt.Errorf("error registering file tool %s: %w", tool.Spec.Name, err) // More specific error
		}
	}
	return nil
}

// Note: Implementations like toolReadFile, toolWriteFile, toolListDirectory,
// toolLineCountFile, toolSanitizeFilename would remain in their respective
// implementation files (e.g., tools_fs_read.go, tools_fs_write.go, etc.)
// The registerFsDirTools function is defined in the new tools_fs_dirs.go file.
// The registerFsUtilTools function is defined in tools_fs_utils.go.
// The registerFsHashTools function is defined in tools_fs_hash.go.
// The registerFsMoveTools function is defined in tools_fs_move.go.
// The registerFsDeleteTools function is defined in tools_fs_delete.go. // Added note
