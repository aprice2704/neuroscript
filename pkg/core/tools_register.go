package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	// Needed for tool implementations here
	// "strconv" // Needed if doing number conversions here
)

// registerCoreTools defines the specs for built-in tools and registers them.
// Assumes tool functions now match ToolFunc signature:
// func(interpreter *Interpreter, args []interface{}) (interface{}, error)
func registerCoreTools(registry *ToolRegistry) {

	// --- File I/O Tools ---
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "ReadFile", Description: "Reads the content of a file.", Args: []ArgSpec{{Name: "filepath", Type: ArgTypeString}}, ReturnType: ArgTypeString},
		Func: toolReadFile, // Assumes toolReadFile matches ToolFunc signature
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "WriteFile", Description: "Writes content to a file.", Args: []ArgSpec{{Name: "filepath", Type: ArgTypeString}, {Name: "content", Type: ArgTypeString}}, ReturnType: ArgTypeString},
		Func: toolWriteFile, // Assumes toolWriteFile matches ToolFunc signature
	})

	// --- Vector DB / Search Tools ---
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "SearchSkills", Description: "Searches the skill registry (mock).", Args: []ArgSpec{{Name: "query", Type: ArgTypeString}}, ReturnType: ArgTypeString},
		Func: toolSearchSkills, // Assumes toolSearchSkills matches ToolFunc signature
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "VectorUpdate", Description: "Updates vector index (mock).", Args: []ArgSpec{{Name: "filepath", Type: ArgTypeString}}, ReturnType: ArgTypeString},
		Func: toolVectorUpdate, // Assumes toolVectorUpdate matches ToolFunc signature
	})

	// --- Git Tools ---
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "GitAdd", Description: "Stages changes in Git.", Args: []ArgSpec{{Name: "filepath", Type: ArgTypeString}}, ReturnType: ArgTypeString},
		Func: toolGitAdd, // Assumes toolGitAdd matches ToolFunc signature
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "GitCommit", Description: "Commits staged changes in Git.", Args: []ArgSpec{{Name: "message", Type: ArgTypeString}}, ReturnType: ArgTypeString},
		Func: toolGitCommit, // Assumes toolGitCommit matches ToolFunc signature
	})

	// --- Filename Utility ---
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "SanitizeFilename", Description: "Sanitizes a string for use as a filename.", Args: []ArgSpec{{Name: "name", Type: ArgTypeString}}, ReturnType: ArgTypeString},
		Func: toolSanitizeFilename, // Assumes toolSanitizeFilename matches ToolFunc signature
	})

	// --- String Manipulation Tools ---
	// References the rewritten functions from tools_string.go directly
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{ // <-- START of StringLength registration
			Name:        "StringLength",
			Description: "Returns the number of runes (characters) in a string.",
			Args:        []ArgSpec{{Name: "input", Type: ArgTypeString}},
			ReturnType:  ArgTypeInt, // <-- CORRECTED ReturnType
		}, // <-- END of StringLength registration
		Func: toolStringLength,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "Substring", Description: "Returns a substring (0-based, start inclusive, end exclusive).", Args: []ArgSpec{{Name: "input", Type: ArgTypeString}, {Name: "start", Type: ArgTypeInt}, {Name: "end", Type: ArgTypeInt}}, ReturnType: ArgTypeString},
		Func: toolSubstring,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "ToUpper", Description: "Converts a string to uppercase.", Args: []ArgSpec{{Name: "input", Type: ArgTypeString}}, ReturnType: ArgTypeString},
		Func: toolToUpper,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "ToLower", Description: "Converts a string to lowercase.", Args: []ArgSpec{{Name: "input", Type: ArgTypeString}}, ReturnType: ArgTypeString},
		Func: toolToLower,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "TrimSpace", Description: "Removes leading/trailing whitespace.", Args: []ArgSpec{{Name: "input", Type: ArgTypeString}}, ReturnType: ArgTypeString},
		Func: toolTrimSpace,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "SplitString", Description: "Splits a string by a delimiter.", Args: []ArgSpec{{Name: "input", Type: ArgTypeString}, {Name: "delimiter", Type: ArgTypeString}}, ReturnType: ArgTypeSliceString},
		Func: toolSplitString,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "SplitWords", Description: "Splits a string into words based on whitespace.", Args: []ArgSpec{{Name: "input", Type: ArgTypeString}}, ReturnType: ArgTypeSliceString},
		Func: toolSplitWords,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "JoinStrings", Description: "Joins a list of strings with a separator.", Args: []ArgSpec{{Name: "input_slice", Type: ArgTypeSliceString}, {Name: "separator", Type: ArgTypeString}}, ReturnType: ArgTypeString},
		Func: toolJoinStrings, // Directly uses the rewritten function
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "ReplaceAll", Description: "Replaces all occurrences of a substring.", Args: []ArgSpec{{Name: "input", Type: ArgTypeString}, {Name: "old", Type: ArgTypeString}, {Name: "new", Type: ArgTypeString}}, ReturnType: ArgTypeString},
		Func: toolReplaceAll,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "Contains", Description: "Checks if a string contains a substring.", Args: []ArgSpec{{Name: "input", Type: ArgTypeString}, {Name: "substring", Type: ArgTypeString}}, ReturnType: ArgTypeBool},
		Func: toolContains,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "HasPrefix", Description: "Checks if a string starts with a prefix.", Args: []ArgSpec{{Name: "input", Type: ArgTypeString}, {Name: "prefix", Type: ArgTypeString}}, ReturnType: ArgTypeBool},
		Func: toolHasPrefix,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "HasSuffix", Description: "Checks if a string ends with a suffix.", Args: []ArgSpec{{Name: "input", Type: ArgTypeString}, {Name: "suffix", Type: ArgTypeString}}, ReturnType: ArgTypeBool},
		Func: toolHasSuffix,
	})

}

// --- Tool Function Implementations Stubs ---
// Ensure functions like toolReadFile, toolWriteFile, toolSearchSkills, etc.
// exist and match the ToolFunc signature:
// func(interpreter *Interpreter, args []interface{}) (interface{}, error)
// (Actual implementations were provided previously or assumed to exist)
// --- Rewritten Tool Function Implementations ---
// (Moved from interpreter.go / interpreter_c.go, matching ToolFunc signature)

func toolReadFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Assumes ValidateAndConvertArgs ensures args[0] is string
	filePath := args[0].(string)
	cwd, _ := os.Getwd()
	absPath, secErr := secureFilePath(filePath, cwd) // secureFilePath from interpreter_c.go
	if secErr != nil {
		return nil, fmt.Errorf("ReadFile path error: %w", secErr)
	}
	contentBytes, readErr := os.ReadFile(absPath)
	if readErr != nil {
		return nil, fmt.Errorf("ReadFile failed for '%s': %w", absPath, readErr)
	}
	return string(contentBytes), nil
}

func toolWriteFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Assumes ValidateAndConvertArgs ensures args[0] and args[1] are strings
	filePath := args[0].(string)
	content := args[1].(string)
	cwd, _ := os.Getwd()
	absPath, secErr := secureFilePath(filePath, cwd) // secureFilePath from interpreter_c.go
	if secErr != nil {
		return nil, fmt.Errorf("WriteFile path error: %w", secErr)
	}
	dirPath := filepath.Dir(absPath)
	if dirErr := os.MkdirAll(dirPath, 0755); dirErr != nil {
		return nil, fmt.Errorf("WriteFile mkdir fail: %w", dirErr)
	}
	contentBytes := []byte(content)
	writeErr := os.WriteFile(absPath, contentBytes, 0644)
	if writeErr != nil {
		return nil, fmt.Errorf("WriteFile failed for '%s': %w", absPath, writeErr)
	}
	return "OK", nil
}

func toolSearchSkills(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Assumes ValidateAndConvertArgs ensures args[0] is string
	query := args[0].(string)
	// Mock logic using interpreter's vectorIndex
	fmt.Printf("      [Tool] Mock Searching skills for: %q\n", query)
	queryEmb, embErr := interpreter.GenerateEmbedding(query) // GenerateEmbedding from interpreter_c.go
	if embErr != nil {
		return nil, fmt.Errorf("embed fail: %w", embErr)
	}
	type SearchResult struct {
		Path  string
		Score float64
	}
	results := []SearchResult{}
	threshold := 0.5
	for path, storedEmb := range interpreter.vectorIndex {
		score, simErr := cosineSimilarity(queryEmb, storedEmb) // cosineSimilarity from interpreter_c.go
		if simErr == nil && score >= threshold {
			results = append(results, SearchResult{Path: path, Score: score})
		}
	}
	resultBytes, jsonErr := json.Marshal(results)
	if jsonErr != nil {
		return nil, fmt.Errorf("marshal results fail: %w", jsonErr)
	}
	return string(resultBytes), nil
}

func toolVectorUpdate(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Assumes ValidateAndConvertArgs ensures args[0] is string
	filePath := args[0].(string)
	cwd, _ := os.Getwd()
	absPath, secErr := secureFilePath(filePath, cwd) // secureFilePath from interpreter_c.go
	if secErr != nil {
		return nil, fmt.Errorf("VectorUpdate path error: %w", secErr)
	}
	fmt.Printf("      [Tool] Mock Updating vector index for: %s\n", filepath.Base(absPath))
	contentBytes, readErr := os.ReadFile(absPath)
	if readErr != nil {
		return nil, fmt.Errorf("read fail for VectorUpdate: %w", readErr)
	}
	textToEmbed := string(contentBytes)
	embedding, embErr := interpreter.GenerateEmbedding(textToEmbed) // GenerateEmbedding from interpreter_c.go
	if embErr != nil {
		return nil, fmt.Errorf("embed fail for VectorUpdate: %w", embErr)
	}
	interpreter.vectorIndex[absPath] = embedding // Assumes interpreter state is passed
	return "OK", nil
}

func toolGitAdd(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Assumes ValidateAndConvertArgs ensures args[0] is string
	filePath := args[0].(string)
	cwd, _ := os.Getwd()
	absPath, secErr := secureFilePath(filePath, cwd) // secureFilePath from interpreter_c.go
	if secErr != nil {
		return nil, fmt.Errorf("GitAdd path error: %w", secErr)
	}
	err := runGitCommand("add", absPath) // runGitCommand from interpreter_c.go
	if err != nil {
		return nil, err
	}
	return "OK", nil
}

func toolGitCommit(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Assumes ValidateAndConvertArgs ensures args[0] is string
	message := args[0].(string)
	err := runGitCommand("commit", "-m", message) // runGitCommand from interpreter_c.go
	if err != nil {
		return nil, err
	}
	return "OK", nil
}

func toolSanitizeFilename(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Assumes ValidateAndConvertArgs ensures args[0] is string
	name := args[0].(string)
	return sanitizeFilename(name), nil // sanitizeFilename from interpreter_c.go
}
