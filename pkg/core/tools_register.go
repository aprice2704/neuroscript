// pkg/core/tools_register.go
package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// registerCoreTools defines the specs for built-in tools and registers them.
func registerCoreTools(registry *ToolRegistry) {

	// --- File I/O, Vector, Git, Filename, Strings ---
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "ReadFile" /*...*/}, Func: toolReadFile})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "WriteFile" /*...*/}, Func: toolWriteFile})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "SearchSkills" /*...*/}, Func: toolSearchSkills})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "VectorUpdate" /*...*/}, Func: toolVectorUpdate})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "GitAdd" /*...*/}, Func: toolGitAdd})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "GitCommit" /*...*/}, Func: toolGitCommit})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "SanitizeFilename" /*...*/}, Func: toolSanitizeFilename})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "StringLength" /*...*/, Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeInt}, Func: toolStringLength})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "Substring" /*...*/, Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "start", Type: ArgTypeInt, Required: true}, {Name: "end", Type: ArgTypeInt, Required: true}}, ReturnType: ArgTypeString}, Func: toolSubstring})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "ToUpper" /*...*/, Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}, Func: toolToUpper})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "ToLower" /*...*/, Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}, Func: toolToLower})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "TrimSpace" /*...*/, Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}, Func: toolTrimSpace})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "SplitString" /*...*/, Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "delimiter", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeSliceString}, Func: toolSplitString})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "SplitWords" /*...*/, Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeSliceString}, Func: toolSplitWords})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "JoinStrings" /*...*/, Args: []ArgSpec{{Name: "input_slice", Type: ArgTypeSliceString, Required: true}, {Name: "separator", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}, Func: toolJoinStrings})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "ReplaceAll" /*...*/, Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "old", Type: ArgTypeString, Required: true}, {Name: "new", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}, Func: toolReplaceAll})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "Contains" /*...*/, Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "substring", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeBool}, Func: toolContains})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "HasPrefix" /*...*/, Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "prefix", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeBool}, Func: toolHasPrefix})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "HasSuffix" /*...*/, Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "suffix", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeBool}, Func: toolHasSuffix})

	// --- Shell Execution & Go Tools ---
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "ExecuteCommand" /*...*/}, Func: toolExecuteCommand})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "GoBuild" /*...*/}, Func: toolGoBuild})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "GoTest" /*...*/}, Func: toolGoTest})
	// registry.RegisterTool(ToolImplementation{ Spec: ToolSpec{ Name: "ApplyPatch", /*...*/ }, Func: toolApplyPatch }) // REMOVED
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "GoFmt" /*...*/}, Func: toolGoFmt})
	registry.RegisterTool(ToolImplementation{Spec: ToolSpec{Name: "GoModTidy" /*...*/}, Func: toolGoModTidy})

}

// --- Existing Tool Implementations ---
func toolReadFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	filePath := args[0].(string)
	cwd, _ := os.Getwd()
	absPath, secErr := secureFilePath(filePath, cwd)
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
	filePath := args[0].(string)
	content := args[1].(string)
	cwd, _ := os.Getwd()
	absPath, secErr := secureFilePath(filePath, cwd)
	if secErr != nil {
		return nil, fmt.Errorf("WriteFile path error: %w", secErr)
	}
	dirPath := filepath.Dir(absPath)
	if dirErr := os.MkdirAll(dirPath, 0755); dirErr != nil {
		return nil, fmt.Errorf("WriteFile mkdir fail: %w", dirErr)
	}
	writeErr := os.WriteFile(absPath, []byte(content), 0644)
	if writeErr != nil {
		return nil, fmt.Errorf("WriteFile failed for '%s': %w", absPath, writeErr)
	}
	return "OK", nil
}
func toolSearchSkills(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	query := args[0].(string)
	if interpreter.logger != nil {
		interpreter.logger.Printf("      [Tool] Mock Searching skills for: %q\n", query)
	}
	queryEmb, embErr := interpreter.GenerateEmbedding(query)
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
		score, simErr := cosineSimilarity(queryEmb, storedEmb)
		if simErr == nil && score >= threshold {
			results = append(results, SearchResult{Path: path, Score: score})
		}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })
	resultBytes, jsonErr := json.Marshal(results)
	if jsonErr != nil {
		return nil, fmt.Errorf("marshal results fail: %w", jsonErr)
	}
	return string(resultBytes), nil
}
func toolVectorUpdate(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	filePath := args[0].(string)
	cwd, _ := os.Getwd()
	absPath, secErr := secureFilePath(filePath, cwd)
	if secErr != nil {
		return nil, fmt.Errorf("VectorUpdate path error: %w", secErr)
	}
	if interpreter.logger != nil {
		interpreter.logger.Printf("      [Tool] Mock Updating vector index for: %s\n", filepath.Base(absPath))
	}
	contentBytes, readErr := os.ReadFile(absPath)
	if readErr != nil {
		return nil, fmt.Errorf("read fail for VectorUpdate: %w", readErr)
	}
	embedding, embErr := interpreter.GenerateEmbedding(string(contentBytes))
	if embErr != nil {
		return nil, fmt.Errorf("embed fail for VectorUpdate: %w", embErr)
	}
	interpreter.vectorIndex[absPath] = embedding
	return "OK", nil
}
func toolGitAdd(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	filePath := args[0].(string)
	cwd, _ := os.Getwd()
	absPath, secErr := secureFilePath(filePath, cwd)
	if secErr != nil {
		return nil, fmt.Errorf("GitAdd path error: %w", secErr)
	}
	err := runGitCommand("add", absPath)
	if err != nil {
		return nil, err
	}
	return "OK", nil
}
func toolGitCommit(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	message := args[0].(string)
	err := runGitCommand("commit", "-m", message)
	if err != nil {
		return nil, err
	}
	return "OK", nil
}
func toolSanitizeFilename(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	name := args[0].(string)
	return sanitizeFilename(name), nil
}
