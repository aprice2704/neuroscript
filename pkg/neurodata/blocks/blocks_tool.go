// pkg/neurodata/blocks/blocks_tool.go
package blocks

import (
	"fmt"

	// Import core for tool types and Interpreter interface
	"github.com/aprice2704/neuroscript/pkg/core"
	// NOTE: We are in the 'blocks' package now, so no need to import it.
	// We directly call ExtractAll and LookForMetadata from this package.
)

// RegisterBlockTools adds the new ANTLR-based block tools to the registry.
func RegisterBlockTools(registry *core.ToolRegistry) {
	registry.RegisterTool(core.ToolImplementation{
		Spec: core.ToolSpec{
			Name: "BlocksExtractAll",
			Description: "Extracts all fenced code blocks (handling nesting and errors) from input content using ANTLR. " +
				"Returns a list of maps, where each map represents a block and contains keys: " +
				"'language_id' (string), 'raw_content' (string), 'start_line' (int), 'end_line' (int), 'metadata' (map[string]string).",
			Args: []core.ArgSpec{
				{Name: "content", Type: core.ArgTypeString, Required: true, Description: "The string content to search within."},
			},
			ReturnType: core.ArgTypeSliceAny, // Returns []map[string]interface{}
		},
		Func: toolBlocksExtractAll,
	})

	registry.RegisterTool(core.ToolImplementation{
		Spec: core.ToolSpec{
			Name:        "BlockGetMetadata",
			Description: "Parses the raw content string of a single code block to find metadata lines (e.g., '# id: ...'). Returns a map[string]string of found key-value pairs.",
			Args: []core.ArgSpec{
				{Name: "raw_content", Type: core.ArgTypeString, Required: true, Description: "The raw content string of the block."},
			},
			ReturnType: core.ArgTypeAny, // Returns map[string]string (as interface{})
		},
		Func: toolBlockGetMetadata,
	})
}

// toolBlocksExtractAll wraps the blocks.ExtractAll and blocks.LookForMetadata functions.
func toolBlocksExtractAll(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	content := args[0].(string)
	logger := interpreter.Logger() // Get logger safely

	if logger != nil {
		logSnippet := content
		if len(logSnippet) > 50 {
			logSnippet = logSnippet[:50] + "..."
		}
		logger.Printf("[DEBUG TOOL] Calling TOOL.BlocksExtractAll on content (snippet): %q", logSnippet)
	}

	// Call the main extraction function from this package
	extractedBlocks, extractErr := ExtractAll(content)

	// Handle errors from ExtractAll (e.g., ambiguity, unclosed fence)
	// Return the specific error message as a string result for NeuroScript
	if extractErr != nil {
		errMsg := fmt.Sprintf("Error during block extraction: %s", extractErr.Error())
		if logger != nil {
			logger.Printf("[ERROR TOOL] TOOL.BlocksExtractAll failed: %s", extractErr.Error())
		}
		return errMsg, nil // Return error message as string
	}

	// Process successful results
	resultsList := make([]interface{}, 0, len(extractedBlocks))
	for _, block := range extractedBlocks {
		// For each extracted block, get its metadata
		metadataMap, metaErr := LookForMetadata(block.RawContent)
		if metaErr != nil {
			// Log the metadata error but maybe continue? Or return an error?
			// Let's log and continue, returning potentially partial metadata.
			if logger != nil {
				logger.Printf("[WARN TOOL] TOOL.BlocksExtractAll: Failed to get metadata for block at line %d: %v", block.StartLine, metaErr)
			}
			// Use an empty map if metadata parsing failed
			metadataMap = make(map[string]string)
		}

		// Convert metadataMap (map[string]string) to map[string]interface{} for the result structure
		metadataInterfaceMap := make(map[string]interface{}, len(metadataMap))
		for k, v := range metadataMap {
			metadataInterfaceMap[k] = v
		}

		// Create the result map for this block
		blockMap := map[string]interface{}{
			"language_id": block.LanguageID,
			"raw_content": block.RawContent,
			"start_line":  int64(block.StartLine), // Ensure int64 for NeuroScript compatibility
			"end_line":    int64(block.EndLine),   // Ensure int64
			"metadata":    metadataInterfaceMap,
		}
		resultsList = append(resultsList, blockMap)
	}

	if logger != nil {
		logger.Printf("[DEBUG TOOL] TOOL.BlocksExtractAll successful. Found %d blocks.", len(resultsList))
	}

	// Return the list of maps
	return resultsList, nil
}

// toolBlockGetMetadata wraps the blocks.LookForMetadata function.
func toolBlockGetMetadata(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	rawContent := args[0].(string)
	logger := interpreter.Logger()

	if logger != nil {
		logSnippet := rawContent
		if len(logSnippet) > 50 {
			logSnippet = logSnippet[:50] + "..."
		}
		logger.Printf("[DEBUG TOOL] Calling TOOL.BlockGetMetadata on content (snippet): %q", logSnippet)
	}

	// Call the metadata function from this package
	metadataMap, metaErr := LookForMetadata(rawContent)

	if metaErr != nil {
		errMsg := fmt.Sprintf("Error getting block metadata: %s", metaErr.Error())
		if logger != nil {
			logger.Printf("[ERROR TOOL] TOOL.BlockGetMetadata failed: %s", metaErr.Error())
		}
		return errMsg, nil // Return error message as string
	}

	// Convert map[string]string to map[string]interface{} for return
	metadataInterfaceMap := make(map[string]interface{}, len(metadataMap))
	for k, v := range metadataMap {
		metadataInterfaceMap[k] = v
	}

	if logger != nil {
		logger.Printf("[DEBUG TOOL] TOOL.BlockGetMetadata successful. Found metadata: %v", metadataInterfaceMap)
	}

	return metadataInterfaceMap, nil
}
