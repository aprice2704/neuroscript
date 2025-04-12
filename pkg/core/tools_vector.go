// pkg/core/tools_vector.go
package core

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// registerVectorTools adds Vector DB related tools to the registry.
// *** MODIFIED: Returns error ***
func registerVectorTools(registry *ToolRegistry) error {
	tools := []ToolImplementation{
		{
			Spec: ToolSpec{
				Name:        "SearchSkills",
				Description: "Searches the (mock) vector index for skills matching a query.",
				Args: []ArgSpec{
					{Name: "query", Type: ArgTypeString, Required: true, Description: "Natural language query."},
				},
				ReturnType: ArgTypeString, // Returns JSON string of results
			},
			Func: toolSearchSkills,
		},
		{
			Spec: ToolSpec{
				Name:        "VectorUpdate",
				Description: "Updates the (mock) vector index for a given file.",
				Args: []ArgSpec{
					{Name: "filepath", Type: ArgTypeString, Required: true, Description: "Relative path to the skill file to index."},
				},
				ReturnType: ArgTypeString, // Returns "OK" or error message
			},
			Func: toolVectorUpdate,
		},
	}
	for _, tool := range tools {
		// *** Check error from RegisterTool ***
		if err := registry.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register Vector tool %s: %w", tool.Spec.Name, err)
		}
	}
	return nil // Success
}

// toolSearchSkills performs a mock similarity search.
func toolSearchSkills(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation handled by ValidateAndConvertArgs
	query := args[0].(string)

	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.SearchSkills (Mock) for query: %q", query)
	}

	if interpreter.vectorIndex == nil {
		interpreter.vectorIndex = make(map[string][]float32) // Initialize if nil
		interpreter.logger.Printf("[INFO] Vector index was nil, initialized.")
	}

	// 1. Generate embedding for the query
	queryEmb, embErr := interpreter.GenerateEmbedding(query)
	if embErr != nil {
		return fmt.Sprintf("SearchSkills embedding generation failed: %s", embErr.Error()), nil
	}

	// 2. Define result structure and search
	type SearchResult struct {
		Path  string  `json:"path"` // Use JSON tags for output consistency
		Score float64 `json:"score"`
	}
	results := []SearchResult{}
	threshold := 0.5 // Example similarity threshold

	for path, storedEmb := range interpreter.vectorIndex {
		score, simErr := cosineSimilarity(queryEmb, storedEmb)
		// Handle potential similarity calculation errors gracefully
		if simErr != nil {
			interpreter.logger.Printf("[WARN] Could not calculate similarity for '%s': %v", path, simErr)
			continue // Skip this entry if similarity fails
		}
		if score >= threshold {
			// Store the relative path if possible, or the key as stored
			// For simplicity, using the map key directly for now.
			results = append(results, SearchResult{Path: path, Score: score})
		}
	}

	// 3. Sort results by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 4. Marshal results to JSON string
	resultBytes, jsonErr := json.Marshal(results)
	if jsonErr != nil {
		// This is an internal error
		return nil, fmt.Errorf("SearchSkills failed to marshal results to JSON: %w", jsonErr)
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      SearchSkills found %d results.", len(results))
	}

	return string(resultBytes), nil
}

// toolVectorUpdate adds or updates a file's mock embedding in the index.
func toolVectorUpdate(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation handled by ValidateAndConvertArgs
	filePath := args[0].(string)

	// Use SecureFilePath to ensure path is safe relative to CWD
	cwd, errWd := os.Getwd()
	if errWd != nil {
		return nil, fmt.Errorf("VectorUpdate failed to get working directory: %w", errWd)
	}
	absPath, secErr := SecureFilePath(filePath, cwd)
	if secErr != nil {
		return fmt.Sprintf("VectorUpdate path error: %s", secErr.Error()), nil
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.VectorUpdate (Mock) for %s (Resolved: %s)", filePath, absPath)
	}

	// 1. Read file content
	contentBytes, readErr := os.ReadFile(absPath)
	if readErr != nil {
		return fmt.Sprintf("VectorUpdate read error for '%s': %s", filePath, readErr.Error()), nil
	}

	// 2. Generate embedding
	embedding, embErr := interpreter.GenerateEmbedding(string(contentBytes))
	if embErr != nil {
		return fmt.Sprintf("VectorUpdate embedding generation failed: %s", embErr.Error()), nil
	}

	// 3. Update index (using absolute path as key for consistency)
	if interpreter.vectorIndex == nil {
		interpreter.vectorIndex = make(map[string][]float32) // Initialize if nil
	}
	interpreter.vectorIndex[absPath] = embedding // Store with absolute path key

	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      VectorUpdate successful for %s", filePath)
	}
	return "OK", nil
}
