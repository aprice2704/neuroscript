// pkg/neurodata/blocks/blocks_tool.go
package blocks

import (
	"fmt"
	"log" // Import log

	"github.com/aprice2704/neuroscript/pkg/core" // Import core
)

// RegisterBlockTools adds the updated block extraction tool.
// --- UPDATED: Now returns an error ---
func RegisterBlockTools(registry *core.ToolRegistry) error {
	// --- TOOL.BlocksExtractAll registration (Updated) ---
	err := registry.RegisterTool(core.ToolImplementation{ // Capture potential error
		Spec: core.ToolSpec{
			Name: "BlocksExtractAll",
			Description: "Extracts all fenced code blocks (handling nesting) from input content using ANTLR Listener. " +
				"Recognizes ':: key: value' metadata lines immediately preceding the opening fence. " +
				"Returns a list of maps, where each map represents a block and contains keys: " +
				"'language_id' (string), 'raw_content' (string), 'start_line' (int), 'end_line' (int), 'metadata' (map[string]string). Silently ignores unclosed blocks.",
			Args: []core.ArgSpec{
				{Name: "content", Type: core.ArgTypeString, Required: true, Description: "The string content to search within."},
			},
			ReturnType: core.ArgTypeSliceAny, // Returns slice of maps
		},
		Func: toolBlocksExtractAll,
	})

	// --- Handle and return error if registration failed ---
	if err != nil {
		log.Printf("[ERROR] Failed to register tool 'BlocksExtractAll': %v", err) // Log the error
		return fmt.Errorf("failed to register blocks tool 'BlocksExtractAll': %w", err)
	}

	// Add registration for other block-related tools here if needed

	return nil // Return nil on successful registration
	// --- END ERROR HANDLING ---

	// --- TOOL.BlockGetMetadata removed ---
}

// --- toolBlocksExtractAll implementation (Updated) ---
func toolBlocksExtractAll(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	content := args[0].(string) // Assumes validation already done
	logger := interpreter.Logger()

	if logger != nil {
		logSnippet := content
		if len(logSnippet) > 100 {
			logSnippet = logSnippet[:100] + "..."
		}
		logger.Debug("[DEBUG TOOL] Calling TOOL.BlocksExtractAll (Listener Based) on content (snippet): %q", logSnippet)
	}

	// Call the new listener-based extractor
	extractedBlocks, extractErr := ExtractAll(content, logger)

	// ExtractAll now handles ignoring unclosed blocks and returns nil error for that case.
	// It only returns errors for lexer/parser issues.
	if extractErr != nil {
		errMsg := fmt.Sprintf("Error during block extraction: %s", extractErr.Error())
		logger.Error("[ERROR TOOL] TOOL.BlocksExtractAll failed: %s", extractErr.Error())

		// Return the error message as a string, consistent with other tools
		return errMsg, nil
	}

	// Convert FencedBlock structs (which now include metadata) to []interface{} of maps
	resultsList := make([]interface{}, 0, len(extractedBlocks))
	for _, block := range extractedBlocks {
		// Convert metadata map[string]string to map[string]interface{} for NeuroScript compatibility
		metadataInterfaceMap := make(map[string]interface{}, len(block.Metadata))
		for k, v := range block.Metadata {
			metadataInterfaceMap[k] = v
		}

		blockMap := map[string]interface{}{
			"language_id": block.LanguageID,
			"raw_content": block.RawContent,       // Content no longer includes metadata lines
			"start_line":  int64(block.StartLine), // Line of opening fence ```
			"end_line":    int64(block.EndLine),   // Line of closing fence ```
			"metadata":    metadataInterfaceMap,   // Include parsed metadata directly
		}
		resultsList = append(resultsList, blockMap)
	}

	logger.Info("[DEBUG TOOL] TOOL.BlocksExtractAll successful. Found %d blocks.", len(resultsList))

	return resultsList, nil // Return the list of maps
}

// --- toolBlockGetMetadata function removed ---
