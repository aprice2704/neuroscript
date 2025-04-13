// pkg/core/tools_vector.go
package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath" // *** ADDED filepath import ***
	"sort"
)

// registerVectorTools adds Vector DB related tools to the registry.
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
					{Name: "filepath", Type: ArgTypeString, Required: true, Description: "Relative path to the skill file to index (within the sandbox)."},
				},
				ReturnType: ArgTypeString, // Returns "OK" or error message
			},
			Func: toolVectorUpdate,
		},
	}
	for _, tool := range tools {
		if err := registry.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register Vector tool %s: %w", tool.Spec.Name, err)
		}
	}
	return nil
}

// toolSearchSkills performs a mock similarity search.
// (Implementation remains the same)
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
		if simErr != nil {
			interpreter.logger.Printf("[WARN] Could not calculate similarity for '%s': %v", path, simErr)
			continue // Skip this entry if similarity fails
		}
		if score >= threshold {
			// Convert absolute path from index back to relative path for result consistency
			relativePath := path // Default if conversion fails
			if interpreter.sandboxDir != "" && filepath.IsAbs(path) {
				rel, err := filepath.Rel(interpreter.sandboxDir, path)
				if err == nil {
					relativePath = rel
				} else {
					interpreter.logger.Printf("[WARN SearchSkills] Could not make path relative to sandbox '%s': %s (%v)", interpreter.sandboxDir, path, err)
				}
			}
			results = append(results, SearchResult{Path: relativePath, Score: score})
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
// *** MODIFIED: Use interpreter.sandboxDir instead of os.Getwd() ***
func toolVectorUpdate(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation handled by ValidateAndConvertArgs
	filePathRel := args[0].(string)

	// *** Get sandbox root directly from the interpreter ***
	sandboxRoot := interpreter.sandboxDir // Use the field name you added
	if sandboxRoot == "" {
		if interpreter.logger != nil {
			interpreter.logger.Printf("[WARN TOOL VectorUpdate] Interpreter sandboxDir is empty, using default relative path validation.")
		}
		sandboxRoot = "." // Ensure it's at least relative to CWD if empty
	}

	// Use SecureFilePath to ensure path is safe relative to sandboxDir
	absPath, secErr := SecureFilePath(filePathRel, sandboxRoot) // *** Use sandboxRoot ***
	if secErr != nil {
		// Path validation failed
		errMsg := fmt.Sprintf("VectorUpdate path error: %s", secErr.Error())
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL VectorUpdate] %s (Sandbox Root: %s)", errMsg, sandboxRoot)
		}
		return errMsg, secErr // Return error message and actual error
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.VectorUpdate (Mock) for %s (Resolved: %s, Sandbox: %s)", filePathRel, absPath, sandboxRoot)
	}

	// 1. Read file content using the absolute path
	contentBytes, readErr := os.ReadFile(absPath)
	if readErr != nil {
		errMsg := fmt.Sprintf("VectorUpdate read error for '%s': %s", filePathRel, readErr.Error())
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL VectorUpdate] %s", errMsg)
		}
		return errMsg, fmt.Errorf("%w: reading file '%s': %w", ErrInternalTool, filePathRel, readErr) // Return error message and wrapped Go error
	}

	// 2. Generate embedding
	embedding, embErr := interpreter.GenerateEmbedding(string(contentBytes))
	if embErr != nil {
		errMsg := fmt.Sprintf("VectorUpdate embedding generation failed: %s", embErr.Error())
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL VectorUpdate] %s", errMsg)
		}
		// Decide if embedding error is internal or should be reported as string
		return errMsg, fmt.Errorf("%w: generating embedding for '%s': %w", ErrInternalTool, filePathRel, embErr) // Return error message and wrapped Go error
	}

	// 3. Update index (using absolute path as key for consistency internally)
	if interpreter.vectorIndex == nil {
		interpreter.vectorIndex = make(map[string][]float32) // Initialize if nil
	}
	interpreter.vectorIndex[absPath] = embedding // Store with absolute path key

	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      VectorUpdate successful for %s", filePathRel)
	}
	return "OK", nil
}
