// NeuroScript Version: 0.3.1
// File version: 0.1.2 // Correctly use FileAPI.ResolvePath and os.ReadFile.
// nlines: 123 // Approximate
// risk_rating: LOW // Mock implementation
// filename: pkg/core/tools_vector.go

package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os" // Import os package for ReadFile
	"path/filepath"
	"sort"
)

// toolSearchSkills performs a mock similarity search.
// Corresponds to ToolSpec "SearchSkills".
func toolSearchSkills(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "SearchSkills"

	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: expected 1 argument (query), got %d", toolName, len(args)), ErrArgumentMismatch)
	}
	query, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: query argument must be a string, got %T", toolName, args[0]), ErrInvalidArgument)
	}

	interpreter.Logger().Debug(fmt.Sprintf("[%s] (Mock) searching for query", toolName), "query", query)

	if interpreter.vectorIndex == nil {
		interpreter.vectorIndex = make(map[string][]float32)
		interpreter.Logger().Debug(fmt.Sprintf("[%s] Vector index was nil, initialized.", toolName))
	}

	queryEmb, embErr := interpreter.GenerateEmbedding(query)
	if embErr != nil {
		errMsg := fmt.Sprintf("%s: embedding generation failed", toolName)
		interpreter.Logger().Error(errMsg, "error", embErr)
		return nil, NewRuntimeError(ErrorCodeInternal, errMsg, errors.Join(ErrInternalTool, embErr))
	}

	type SearchResult struct {
		Path  string  `json:"path"`
		Score float64 `json:"score"`
	}
	results := []SearchResult{}
	threshold := 0.5

	for pathKeyAbs, storedEmb := range interpreter.vectorIndex {
		score, simErr := cosineSimilarity(queryEmb, storedEmb)
		if simErr != nil {
			interpreter.Logger().Warn(fmt.Sprintf("[%s] Could not calculate similarity", toolName), "path", pathKeyAbs, "error", simErr)
			continue
		}
		if score >= threshold {
			relativePath := pathKeyAbs
			sandboxRoot := interpreter.SandboxDir()
			if sandboxRoot != "" && filepath.IsAbs(pathKeyAbs) {
				rel, err := filepath.Rel(sandboxRoot, pathKeyAbs)
				if err == nil {
					relativePath = rel
				} else {
					interpreter.Logger().Warn(fmt.Sprintf("[%s] Could not make path relative to sandbox", toolName), "sandbox", sandboxRoot, "absPath", pathKeyAbs, "error", err)
				}
			}
			results = append(results, SearchResult{Path: relativePath, Score: score})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	resultBytes, jsonErr := json.Marshal(results)
	if jsonErr != nil {
		errMsg := fmt.Sprintf("%s: failed to marshal results to JSON", toolName)
		interpreter.Logger().Error(errMsg, "error", jsonErr)
		return nil, NewRuntimeError(ErrorCodeInternal, errMsg, errors.Join(ErrInternalTool, jsonErr))
	}

	interpreter.Logger().Debug(fmt.Sprintf("[%s] Search complete", toolName), "results_count", len(results))
	return string(resultBytes), nil
}

// toolVectorUpdate adds or updates a file's mock embedding in the index.
// Corresponds to ToolSpec "VectorUpdate".
func toolVectorUpdate(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "VectorUpdate"

	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: expected 1 argument (filepath), got %d", toolName, len(args)), ErrArgumentMismatch)
	}
	filePathRel, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: filepath argument must be a string, got %T", toolName, args[0]), ErrInvalidArgument)
	}
	if filePathRel == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: filepath cannot be empty", toolName), ErrInvalidArgument)
	}

	fileAPI := interpreter.FileAPI()
	if fileAPI == nil {
		return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("%s: FileAPI not initialized in interpreter", toolName), ErrInternal)
	}

	interpreter.Logger().Debug(fmt.Sprintf("[%s] (Mock) updating index for", toolName), "relative_path", filePathRel)

	// Step 1: Resolve the relative path to a safe, absolute path using FileAPI
	absPath, pathErr := fileAPI.ResolvePath(filePathRel)
	if pathErr != nil {
		interpreter.Logger().Error(fmt.Sprintf("%s: failed to resolve path", toolName), "relative_path", filePathRel, "error", pathErr)
		// ResolvePath already returns a RuntimeError, so just propagate it
		return nil, pathErr
	}

	// Step 2: Read the file content using the resolved absolute path
	contentBytes, readErr := os.ReadFile(absPath)
	if readErr != nil {
		interpreter.Logger().Error(fmt.Sprintf("%s: failed to read file", toolName), "absolute_path", absPath, "error", readErr)
		return nil, NewRuntimeError(ErrorCodeIOFailed, fmt.Sprintf("failed to read file '%s'", filePathRel), readErr)
	}

	// Generate embedding
	embedding, embErr := interpreter.GenerateEmbedding(string(contentBytes))
	if embErr != nil {
		errMsg := fmt.Sprintf("%s: embedding generation failed for %q", toolName, filePathRel)
		interpreter.Logger().Error(errMsg, "error", embErr)
		return nil, NewRuntimeError(ErrorCodeInternal, errMsg, errors.Join(ErrInternalTool, embErr))
	}

	// The absPath from ResolvePath is already the correct key for the vector index.
	if interpreter.vectorIndex == nil {
		interpreter.vectorIndex = make(map[string][]float32)
	}
	interpreter.vectorIndex[absPath] = embedding // Store with absolute path key

	interpreter.Logger().Debug(fmt.Sprintf("[%s] Update successful", toolName), "relative_path", filePathRel, "absolute_path", absPath)
	return "OK", nil
}

// Note: cosineSimilarity function assumed to exist elsewhere or be defined.
