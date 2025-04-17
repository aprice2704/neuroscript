// filename: pkg/core/tools_fs_stat.go
package core

import (
	"fmt"
	"os"   // For os.Stat and os.IsNotExist
	"time" // For formatting ModTime
)

// toolStat gets information about a file or directory within the sandbox.
func toolStat(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation guarantees args[0] is a string
	relPath := args[0].(string)
	sandboxRoot := interpreter.sandboxDir
	if sandboxRoot == "" {
		if interpreter.logger != nil {
			interpreter.logger.Printf("[WARN TOOL Stat] Interpreter sandboxDir is empty, using default relative path validation.")
		}
		sandboxRoot = "."
	}

	// Validate path is within sandbox and get absolute path
	absPath, secErr := SecureFilePath(relPath, sandboxRoot)
	if secErr != nil {
		errMsg := fmt.Sprintf("Stat path error for '%s': %s", relPath, secErr.Error())
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL Stat] %s (Sandbox Root: %s)", errMsg, sandboxRoot)
		}
		// Return nil for non-existent or invalid paths, consistent with returning map or nil
		// Return the security error for Go context
		return nil, secErr
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL Stat] Attempting to stat validated path: %s (Original Relative: %s, Sandbox: %s)", absPath, relPath, sandboxRoot)
	}

	// Get file info using the validated absolute path
	info, statErr := os.Stat(absPath)
	if statErr != nil {
		// Handle errors, specifically file not found
		if os.IsNotExist(statErr) {
			errMsg := fmt.Sprintf("Stat: Path not found '%s'", relPath)
			if interpreter.logger != nil {
				interpreter.logger.Printf("[TOOL Stat] %s", errMsg)
			}
			return nil, nil // Return nil value and nil error for script if not found
		}
		// Other errors (permissions, etc.)
		errMsg := fmt.Sprintf("Stat failed for '%s': %s", relPath, statErr.Error())
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL Stat] %s", errMsg)
		}
		// Return error message for script, wrap OS error for Go context
		return nil, fmt.Errorf("%w: stating file '%s': %w", ErrInternalTool, relPath, statErr)
	}

	// Success, create result map
	resultMap := map[string]interface{}{
		"name":     info.Name(), // Base name of the file/dir
		"path":     relPath,     // The original relative path requested
		"size":     info.Size(),
		"is_dir":   info.IsDir(),
		"mod_time": info.ModTime().Format(time.RFC3339), // Standard time format
		// Add other fields if needed, e.g., permissions info.Mode().String()
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL Stat] Stat successful for '%s'", relPath)
	}

	// Return the map and nil error
	return resultMap, nil
}
