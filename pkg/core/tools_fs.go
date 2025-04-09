// pkg/core/tools_fs.go
package core

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// registerFsTools registration needs to update ListDirectory's ReturnType
func registerFsTools(registry *ToolRegistry) {
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "ReadFile", Description: "Reads file content.", Args: []ArgSpec{{Name: "filepath", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}, Func: toolReadFile})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "WriteFile", Description: "Writes content to file.", Args: []ArgSpec{{Name: "filepath", Type: ArgTypeString, Required: true}, {Name: "content", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}, Func: toolWriteFile})
	// *** MODIFIED ListDirectory Spec ***
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{
		Name:        "ListDirectory",
		Description: "Lists directory content. Returns a list of maps, each map containing {'name': string, 'is_dir': bool}.",
		Args:        []ArgSpec{{Name: "path", Type: ArgTypeString, Required: true}},
		ReturnType:  ArgTypeSliceAny, // Changed to SliceAny because it returns []map[string]interface{}
	}, Func: toolListDirectory})
	// *** END MODIFICATION ***
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "LineCount", Description: "Counts lines in file or string. Returns -1 on file error.", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeInt}, Func: toolLineCount})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "SanitizeFilename", Description: "Cleans string for filename.", Args: []ArgSpec{{Name: "name", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}, Func: toolSanitizeFilename})
}

// toolReadFile, toolWriteFile, toolSanitizeFilename (remain the same)
func toolReadFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	filePath := args[0].(string)
	cwd, errWd := os.Getwd()
	if errWd != nil {
		return nil, fmt.Errorf("ReadFile failed get CWD: %w", errWd)
	}
	absPath, secErr := SecureFilePath(filePath, cwd) // Use exported SecureFilePath
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
	absPath, secErr := SecureFilePath(filePath, cwd) // Use exported SecureFilePath
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
func toolSanitizeFilename(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	name := args[0].(string)
	sanitized := sanitizeFilename(name) // Assumes internal sanitizeFilename exists
	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      SanitizeFilename: %q -> %q", name, sanitized)
	}
	return sanitized, nil
}

// *** MODIFIED toolListDirectory IMPLEMENTATION ***
func toolListDirectory(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	dirPath := args[0].(string)
	cwd, errWd := os.Getwd()
	if errWd != nil {
		// Return error string for script, but Go error for internal handling
		errMsg := fmt.Sprintf("ListDirectory failed get CWD: %s", errWd.Error())
		return errMsg, fmt.Errorf(errMsg) // Return error string for script side
	}
	var absPath string
	var secErr error
	if filepath.Clean(dirPath) == "." {
		absPath = cwd
	} else {
		absPath, secErr = SecureFilePath(dirPath, cwd) // Use exported SecureFilePath
		if secErr != nil {
			errMsg := fmt.Sprintf("ListDirectory path error: %s", secErr.Error())
			return errMsg, fmt.Errorf(errMsg) // Return error string for script side
		}
	}
	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.ListDirectory for %s (Resolved: %s)", dirPath, absPath)
	}
	entries, err := os.ReadDir(absPath)
	if err != nil {
		errMsg := fmt.Sprintf("ListDirectory read error for '%s': %s", dirPath, err.Error())
		return errMsg, fmt.Errorf(errMsg) // Return error string for script side
	}

	// Create list of maps
	resultsList := make([]interface{}, 0, len(entries))
	entryInfos := make([]map[string]interface{}, 0, len(entries)) // Temp slice for sorting

	for _, entry := range entries {
		entryInfos = append(entryInfos, map[string]interface{}{
			"name":   entry.Name(), // Return just the name, no trailing slash
			"is_dir": entry.IsDir(),
		})
	}

	// Sort by name for deterministic order
	sort.Slice(entryInfos, func(i, j int) bool {
		return entryInfos[i]["name"].(string) < entryInfos[j]["name"].(string)
	})

	// Convert sorted temp slice to []interface{}
	for _, info := range entryInfos {
		resultsList = append(resultsList, info)
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      ListDirectory successful for %s. Found %d entries.", dirPath, len(resultsList))
	}
	// Return the list of maps (as []interface{}) and nil error
	return resultsList, nil
}

// *** END MODIFICATION ***

// toolLineCount (remains the same)
func toolLineCount(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	input := args[0].(string)
	content := ""
	isPath := false
	foundContent := false

	if strings.ContainsAny(input, "/\\") || strings.HasSuffix(input, ".go") || strings.HasSuffix(input, ".txt") || strings.HasSuffix(input, ".md") {
		cwd, errWd := os.Getwd()
		if errWd != nil {
			if interpreter.logger != nil {
				interpreter.logger.Printf("[ERROR] LineCount failed get CWD: %v", errWd)
			}
			return int64(-1), nil
		}
		absPath, secErr := SecureFilePath(input, cwd) // Use exported SecureFilePath
		if secErr == nil {
			isPath = true
			if interpreter.logger != nil {
				interpreter.logger.Printf("[DEBUG-INTERP]      LineCount treating input as path: %s (Resolved: %s)", input, absPath)
			}
			contentBytes, readErr := os.ReadFile(absPath)
			if readErr == nil {
				content = string(contentBytes)
				foundContent = true
			} else {
				if interpreter.logger != nil {
					interpreter.logger.Printf("[WARN] LineCount failed to read secure path '%s': %v.", input, readErr)
				}
				return int64(-1), nil
			}
		} else {
			if interpreter.logger != nil {
				interpreter.logger.Printf("[WARN] LineCount input '%s' failed security check: %v.", input, secErr)
			}
			return int64(-1), nil
		}
	}

	if !isPath && !foundContent {
		content = input
		if interpreter.logger != nil {
			interpreter.logger.Printf("[DEBUG-INTERP]      LineCount treating input as raw string.")
		}
		foundContent = true
	}

	if foundContent {
		if len(content) == 0 {
			if interpreter.logger != nil {
				interpreter.logger.Printf("[DEBUG-INTERP]      LineCount result: 0 (empty content)")
			}
			return int64(0), nil
		}
		lineCount := int64(strings.Count(content, "\n"))
		if !strings.HasSuffix(content, "\n") {
			lineCount++
		}
		if content == "\n" {
			lineCount = 1
		}

		if interpreter.logger != nil {
			interpreter.logger.Printf("[DEBUG-INTERP]      LineCount result: %d", lineCount)
		}
		return lineCount, nil
	}
	if interpreter.logger != nil {
		interpreter.logger.Printf("[ERROR] LineCount reached unexpected state.")
	}
	return int64(-1), nil
}

// Internal helper, assumed to exist
func SanitizeFilename(name string) string {
	// Placeholder - copy implementation from utils.go if needed here,
	// but better practice is to keep it in one place (e.g., tools_helpers.go)
	// and ensure tools_fs.go can access it or call it via an interface/registry.
	// For now, assume it's accessible or reimplemented if strictly necessary.
	name = strings.ReplaceAll(name, " ", "_")
	name = regexp.MustCompile(`[^a-zA-Z0-9._-]`).ReplaceAllString(name, "")
	name = strings.Trim(name, "._-")
	if name == "" {
		return "default_sanitized_name"
	}
	return name
}

var (
// Regex moved here temporarily if sanitizeFilename is copied
// removeCharsPattern = regexp.MustCompile(`[^a-zA-Z0-9._-]`)
)
