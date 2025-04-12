// filename: pkg/core/tools_fs.go
package core

import "fmt" // Keep fmt

// registerFsTools registers all filesystem-related tools.
// *** MODIFIED: Returns error ***
func registerFsTools(registry *ToolRegistry) error {
	tools := []ToolImplementation{
		{Spec: ToolSpec{Name: "ReadFile", Description: "Reads the entire content of a specific file within the designated sandbox directory. Use this tool when asked to get the contents of, read, or show a local file specified by name.", Args: []ArgSpec{{Name: "filepath", Type: ArgTypeString, Required: true, Description: "The relative path (within the sandbox) of the file to read."}}, ReturnType: ArgTypeString}, Func: toolReadFile},
		{Spec: ToolSpec{Name: "WriteFile", Description: "Writes content to a specific file within the designated sandbox directory, creating directories if needed. Overwrites existing files.", Args: []ArgSpec{{Name: "filepath", Type: ArgTypeString, Required: true, Description: "The relative path (within the sandbox) of the file to write."}, {Name: "content", Type: ArgTypeString, Required: true, Description: "The content to write."}}, ReturnType: ArgTypeString}, Func: toolWriteFile},
		{Spec: ToolSpec{Name: "ListDirectory", Description: "Lists directory content within the sandbox. Returns a list of maps, each map containing {'name': string, 'is_dir': bool}.", Args: []ArgSpec{{Name: "path", Type: ArgTypeString, Required: true, Description: "The relative path (within the sandbox) of the directory to list."}}, ReturnType: ArgTypeSliceAny}, Func: toolListDirectory},
		{Spec: ToolSpec{Name: "LineCountFile", Description: "Counts lines in a specified file within the sandbox. Returns -1 on file path or read error.", Args: []ArgSpec{{Name: "filepath", Type: ArgTypeString, Required: true, Description: "Relative file path (within the sandbox) to count lines in."}}, ReturnType: ArgTypeInt}, Func: toolLineCountFile},
		{Spec: ToolSpec{Name: "SanitizeFilename", Description: "Cleans a string to make it suitable for use as part of a filename.", Args: []ArgSpec{{Name: "name", Type: ArgTypeString, Required: true, Description: "The string to sanitize."}}, ReturnType: ArgTypeString}, Func: toolSanitizeFilename},
	}
	for _, tool := range tools {
		// *** Check error from RegisterTool ***
		if err := registry.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register FS tool %s: %w", tool.Spec.Name, err)
		}
	}
	return nil // Success
}
