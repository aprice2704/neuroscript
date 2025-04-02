// pkg/core/tools_fs.go
package core

import (
	"fmt"
	"os"
	"path/filepath" // Needed for Join
	"sort"
)

// toolListDirectory lists the contents of a directory.
func toolListDirectory(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("TOOL.ListDirectory internal error: expected 1 arg (path), got %d", len(args))
	}
	dirPath, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.ListDirectory internal error: path must be a string, got %T", args[0])
	}

	cwd, errWd := os.Getwd()
	if errWd != nil {
		// Return error as string result for NeuroScript
		return fmt.Sprintf("ListDirectory failed to get working directory: %s", errWd.Error()), nil
	}

	// Secure the path - allow listing the CWD itself if path is "."
	var absPath string
	var secErr error
	if filepath.Clean(dirPath) == "." {
		absPath = cwd // Use cwd directly if input is "."
	} else {
		absPath, secErr = secureFilePath(dirPath, cwd)
		if secErr != nil {
			// Return security error as string result
			return fmt.Sprintf("ListDirectory path error: %s", secErr.Error()), nil
		}
	}

	// Read directory entries
	entries, err := os.ReadDir(absPath)
	if err != nil {
		// Return read error as string result
		return fmt.Sprintf("ListDirectory read error for '%s': %s", dirPath, err.Error()), nil
	}

	// Extract names and sort them
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

// Placeholder for LineCount - implement next
func toolLineCount(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// TODO: Implement LineCount logic
	return nil, fmt.Errorf("TOOL.LineCount not implemented yet")
}
