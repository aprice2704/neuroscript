// filename: pkg/core/tools_fs_list.go
package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// toolListDirectory lists the contents of a specified directory.
// Assumes path validation/sandboxing is handled by the SecurityLayer.
func toolListDirectory(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation guarantees args[0] is a string
	dirPath := args[0].(string)

	cwd, errWd := os.Getwd()
	if errWd != nil {
		errMsg := fmt.Sprintf("ListDirectory failed get CWD: %s", errWd.Error())
		return errMsg, fmt.Errorf(errMsg) // Return error string for script, Go error internally
	}

	var absPath string
	var secErr error
	// Resolve path using SecureFilePath, security checks happened earlier
	if filepath.Clean(dirPath) == "." {
		absPath = cwd // Special case for current directory needs careful sandbox handling
		// Check if CWD itself is within sandbox if '.' is used? SecurityLayer should handle this.
		// For now, assume if validation passed, '.' means the sandbox root.
		// Re-validate '.' against sandbox root?
		_, secErr = SecureFilePath(".", cwd) // Validate '.' against CWD itself
		if secErr != nil {
			errMsg := fmt.Sprintf("ListDirectory path error for '.': %s", secErr.Error())
			return errMsg, fmt.Errorf(errMsg)
		}

	} else {
		absPath, secErr = SecureFilePath(dirPath, cwd)
		if secErr != nil {
			errMsg := fmt.Sprintf("ListDirectory path error for '%s': %s", dirPath, secErr.Error())
			return errMsg, fmt.Errorf(errMsg)
		}
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL ListDirectory] Listing validated path: %s (Original Relative: %s)", absPath, dirPath)
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		errMsg := fmt.Sprintf("ListDirectory read error for '%s': %s", dirPath, err.Error())
		// Check if it's a "not a directory" error specifically?
		if os.IsNotExist(err) {
			errMsg = fmt.Sprintf("ListDirectory failed: Directory not found at path '%s'", dirPath)
		} else if strings.Contains(err.Error(), "not a directory") {
			errMsg = fmt.Sprintf("ListDirectory failed: Path '%s' is not a directory", dirPath)
		}
		return errMsg, fmt.Errorf(errMsg) // Return error string and Go error
	}

	// Prepare results as []interface{} containing map[string]interface{}
	resultsList := make([]interface{}, 0, len(entries))
	entryInfos := make([]map[string]interface{}, 0, len(entries))

	for _, entry := range entries {
		entryInfos = append(entryInfos, map[string]interface{}{
			"name":   entry.Name(),
			"is_dir": entry.IsDir(),
		})
	}

	// Sort for deterministic output
	sort.Slice(entryInfos, func(i, j int) bool { return entryInfos[i]["name"].(string) < entryInfos[j]["name"].(string) })

	// Convert sorted temp slice to final result type
	for _, info := range entryInfos {
		resultsList = append(resultsList, info)
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL ListDirectory] Found %d entries in %s.", len(resultsList), dirPath)
	}
	// Return the list of maps
	return resultsList, nil
}
