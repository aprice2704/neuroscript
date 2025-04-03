// pkg/core/tools_fs.go
package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// registerFsTools registration (remains the same)
func registerFsTools(registry *ToolRegistry) {
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "ReadFile", Description: "Reads file content.", Args: []ArgSpec{{Name: "filepath", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}, Func: toolReadFile})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "WriteFile", Description: "Writes content to file.", Args: []ArgSpec{{Name: "filepath", Type: ArgTypeString, Required: true}, {Name: "content", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}, Func: toolWriteFile})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "ListDirectory", Description: "Lists directory content.", Args: []ArgSpec{{Name: "path", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeSliceString}, Func: toolListDirectory})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "LineCount", Description: "Counts lines in file or string. Returns -1 on file error.", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeInt}, Func: toolLineCount}) // ReturnType is Int
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "SanitizeFilename", Description: "Cleans string for filename.", Args: []ArgSpec{{Name: "name", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}, Func: toolSanitizeFilename})
}

// toolReadFile, toolWriteFile, toolListDirectory, toolSanitizeFilename (remain the same)
func toolReadFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	filePath := args[0].(string)
	cwd, errWd := os.Getwd()
	if errWd != nil {
		return nil, fmt.Errorf("ReadFile failed get CWD: %w", errWd)
	}
	absPath, secErr := secureFilePath(filePath, cwd)
	if secErr != nil {
		return fmt.Sprintf("ReadFile failed for '%s': %s", filePath, secErr.Error()), nil
	}
	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.ReadFile for %s (Resolved: %s)", filePath, absPath)
	}
	contentBytes, readErr := os.ReadFile(absPath)
	if readErr != nil {
		return fmt.Sprintf("ReadFile failed for '%s': %s", filePath, readErr.Error()), nil
	}
	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      ReadFile successful for %s", filePath)
	}
	return string(contentBytes), nil
}
func toolWriteFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	filePath := args[0].(string)
	content := args[1].(string)
	cwd, errWd := os.Getwd()
	if errWd != nil {
		return nil, fmt.Errorf("WriteFile failed get CWD: %w", errWd)
	}
	absPath, secErr := secureFilePath(filePath, cwd)
	if secErr != nil {
		return fmt.Sprintf("WriteFile path error: %s", secErr.Error()), nil
	}
	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.WriteFile for %s (Resolved: %s)", filePath, absPath)
	}
	dirPath := filepath.Dir(absPath)
	if dirErr := os.MkdirAll(dirPath, 0755); dirErr != nil {
		return fmt.Sprintf("WriteFile mkdir failed for dir '%s': %s", dirPath, dirErr.Error()), nil
	}
	writeErr := os.WriteFile(absPath, []byte(content), 0644)
	if writeErr != nil {
		return fmt.Sprintf("WriteFile failed for '%s': %s", filePath, writeErr.Error()), nil
	}
	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      WriteFile successful for %s", filePath)
	}
	return "OK", nil
}
func toolListDirectory(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	dirPath := args[0].(string)
	cwd, errWd := os.Getwd()
	if errWd != nil {
		return fmt.Sprintf("ListDirectory failed get CWD: %s", errWd.Error()), nil
	}
	var absPath string
	var secErr error
	if filepath.Clean(dirPath) == "." {
		absPath = cwd
	} else {
		absPath, secErr = secureFilePath(dirPath, cwd)
		if secErr != nil {
			return fmt.Sprintf("ListDirectory path error: %s", secErr.Error()), nil
		}
	}
	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.ListDirectory for %s (Resolved: %s)", dirPath, absPath)
	}
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return fmt.Sprintf("ListDirectory read error for '%s': %s", dirPath, err.Error()), nil
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		names = append(names, name)
	}
	sort.Strings(names)
	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      ListDirectory successful for %s. Found %d entries.", dirPath, len(names))
	}
	return names, nil
}
func toolSanitizeFilename(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	name := args[0].(string)
	sanitized := sanitizeFilename(name)
	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      SanitizeFilename: %q -> %q", name, sanitized)
	}
	return sanitized, nil
}

// *** UPDATED toolLineCount IMPLEMENTATION ***
func toolLineCount(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	input := args[0].(string)
	content := ""
	isPath := false
	foundContent := false

	// Check if input might be a path
	if strings.ContainsAny(input, "/\\") || strings.HasSuffix(input, ".go") || strings.HasSuffix(input, ".txt") || strings.HasSuffix(input, ".md") {
		cwd, errWd := os.Getwd()
		if errWd != nil {
			if interpreter.logger != nil {
				interpreter.logger.Printf("[ERROR] LineCount failed get CWD: %v", errWd)
			}
			return int64(-1), nil // Return -1 on internal error
		}
		absPath, secErr := secureFilePath(input, cwd)
		if secErr == nil { // Path is secure, try reading
			isPath = true
			if interpreter.logger != nil {
				interpreter.logger.Printf("[DEBUG-INTERP]      LineCount treating input as path: %s (Resolved: %s)", input, absPath)
			}
			contentBytes, readErr := os.ReadFile(absPath)
			if readErr == nil {
				content = string(contentBytes)
				foundContent = true
			} else {
				// Secure path, but read failed (e.g., not found) -> Return error
				if interpreter.logger != nil {
					interpreter.logger.Printf("[WARN] LineCount failed to read secure path '%s': %v.", input, readErr)
				}
				return int64(-1), nil // Return -1 on read error
			}
		} else {
			// Path was insecure -> Return error
			if interpreter.logger != nil {
				interpreter.logger.Printf("[WARN] LineCount input '%s' failed security check: %v.", input, secErr)
			}
			return int64(-1), nil // Return -1 on security error
		}
	}

	// If not treated as a path or reading failed unexpectedly (shouldn't happen after checks), treat as raw string
	if !isPath && !foundContent {
		content = input
		if interpreter.logger != nil {
			interpreter.logger.Printf("[DEBUG-INTERP]      LineCount treating input as raw string.")
		}
		foundContent = true // Mark that we have content to count
	}

	// Count lines if content was found
	if foundContent {
		if len(content) == 0 {
			if interpreter.logger != nil {
				interpreter.logger.Printf("[DEBUG-INTERP]      LineCount result: 0 (empty content)")
			}
			return int64(0), nil
		}
		// Use scanner for potentially more robust line counting? Or stick with strings.Count
		// strings.Count is simpler. Add 1 unless the only content is ""
		lineCount := int64(strings.Count(content, "\n"))
		// If content doesn't end with newline, the last line still counts
		if !strings.HasSuffix(content, "\n") {
			lineCount++
		}
		// Handle case where content is just "\n" -> should be 1 line
		if content == "\n" {
			lineCount = 1
		}

		if interpreter.logger != nil {
			interpreter.logger.Printf("[DEBUG-INTERP]      LineCount result: %d", lineCount)
		}
		return lineCount, nil
	}

	// Should not be reached, but return error code just in case
	if interpreter.logger != nil {
		interpreter.logger.Printf("[ERROR] LineCount reached unexpected state.")
	}
	return int64(-1), nil
}

// *** END UPDATED IMPLEMENTATION ***
