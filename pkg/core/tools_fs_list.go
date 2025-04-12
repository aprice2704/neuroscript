// filename: pkg/core/tools_fs_list.go
package core

import (
	"fmt"
	"os" // Keep filepath
	"sort"
	// "time" // Keep if adding modified time later
)

// toolListDirectory lists files and directories at a given path within the sandbox.
// *** MODIFIED: Return error from SecureFilePath directly ***
func toolListDirectory(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures args[0] is string
	pathRel := args[0].(string)

	// Ensure path is safe and within the sandbox
	cwd, errWd := os.Getwd()
	if errWd != nil {
		// Internal error, return actual error
		return nil, fmt.Errorf("ListDirectory failed to get working directory: %w", errWd)
	}
	absPath, secErr := SecureFilePath(pathRel, cwd)
	if secErr != nil {
		// *** Return the security error directly ***
		// secErr should be or wrap ErrPathViolation
		errMsg := fmt.Sprintf("ListDirectory path error: %s", secErr.Error()) // Log the unwrapped error
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL ListDirectory] %s", errMsg)
		}
		return nil, secErr // Return the original error (which should wrap ErrPathViolation)
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL ListDirectory] Listing directory: %s (resolved: %s)", pathRel, absPath)
	}

	// Read directory contents
	entries, readErr := os.ReadDir(absPath)
	if readErr != nil {
		// *** Return wrapped internal tool error ***
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
			// Add modified time? Requires formatting.
			// "modified": info.ModTime().Format(time.RFC3339),
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
