// filename: pkg/core/tools_fs_list.go
package core

import (
	"fmt"
	"os" // Keep os for ReadDir and Stat
	// Keep filepath for Clean
	"sort"
	// "time" // Keep if adding modified time later
)

// toolListDirectory lists files and directories at a given path within the sandbox.
// *** MODIFIED: Use interpreter.sandboxDir instead of os.Getwd() ***
func toolListDirectory(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures args[0] is string
	pathRel := args[0].(string)

	// *** Get sandbox root directly from the interpreter ***
	sandboxRoot := interpreter.sandboxDir // Use the field name you added
	if sandboxRoot == "" {
		if interpreter.logger != nil {
			interpreter.logger.Printf("[WARN TOOL ListDirectory] Interpreter sandboxDir is empty, using default relative path validation.")
		}
		sandboxRoot = "." // Ensure it's at least relative to CWD if empty
	}

	// Ensure path is safe and within the sandbox
	absPath, secErr := SecureFilePath(pathRel, sandboxRoot) // *** Use sandboxRoot ***
	if secErr != nil {
		// Path validation failed
		errMsg := fmt.Sprintf("ListDirectory path error for '%s': %s", pathRel, secErr.Error()) // Log unwrapped
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL ListDirectory] %s (Sandbox Root: %s)", errMsg, sandboxRoot)
		}
		return nil, secErr // Return the original error
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL ListDirectory] Listing directory: %s (resolved: %s, Sandbox: %s)", pathRel, absPath, sandboxRoot)
	}

	// Read directory contents using the validated absolute path
	entries, readErr := os.ReadDir(absPath)
	if readErr != nil {
		// Error reading directory (e.g., doesn't exist, not a directory, permissions)
		errMsg := fmt.Sprintf("ListDirectory failed: %s", readErr.Error()) // Log the unwrapped error
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL ListDirectory] ReadDir error: %s", errMsg)
		}
		return nil, fmt.Errorf("%w: reading directory '%s': %w", ErrInternalTool, pathRel, readErr) // Wrap underlying OS error
	}

	// Prepare result slice
	result := make([]interface{}, 0, len(entries))
	for _, entry := range entries {
		info, infoErr := entry.Info()
		if infoErr != nil {
			if interpreter.logger != nil {
				interpreter.logger.Printf("[WARN TOOL ListDirectory] Could not get info for entry '%s': %v", entry.Name(), infoErr)
			}
			continue // Skip entries we can't get info for
		}
		// Construct map for each entry
		entryMap := map[string]interface{}{
			"name":   entry.Name(),
			"is_dir": entry.IsDir(),
			"size":   info.Size(), // Include size
			// "modified": info.ModTime().Format(time.RFC3339), // Optional
		}
		result = append(result, entryMap)
	}

	// Sort results by name for predictable order
	sort.SliceStable(result, func(i, j int) bool {
		iMap, iOk := result[i].(map[string]interface{})
		jMap, jOk := result[j].(map[string]interface{})
		if !iOk || !jOk {
			return false
		}
		iName, iNameOk := iMap["name"].(string)
		jName, jNameOk := jMap["name"].(string)
		if !iNameOk || !jNameOk {
			return false
		}
		return iName < jName
	})

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL ListDirectory] Successfully listed %d entries for %s", len(result), pathRel)
	}

	return result, nil // Return the slice of maps
}
