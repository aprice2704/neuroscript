// pkg/neurodata/blocks/blocks_tool.go
package blocks

import (
	"fmt"
	"log" // Keep if used

	// Import core
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/toolsets" // <<< ADDED import for init()
)

// --- ADDED: init() function for self-registration ---
func init() {
	fmt.Println("Blocks package init() running...") // Debug output
	// Register the main registration function with the toolsets package.
	toolsets.AddToolsetRegistration("Blocks", RegisterBlockTools)
}

// RegisterBlockTools adds the updated block extraction tool.
// This function is now called via the init() mechanism.
func RegisterBlockTools(registry tool.ToolRegistrar) error { // Use interface
	// --- TOOL.BlocksExtractAll registration ---
	err := registry.RegisterTool(tool.ToolImplementation{
		Spec: tool.ToolSpec{
			Name: "BlocksExtractAll",
			Description: "Extracts all fenced code blocks (handling nesting) from input content using ANTLR Listener. " +
				"Recognizes ':: key: value' metadata lines immediately preceding the opening fence. " +
				"Returns a list of maps, where each map represents a block and contains keys: " +
				"'language_id' (string), 'raw_content' (string), 'start_line' (int), 'end_line' (int), 'metadata' (map[string]string). Silently ignores unclosed blocks.",
			Args: []tool.ArgSpec{
				{Name: "content", Type: parser.ArgTypeString, Required: true, Description: "The string content to search within."},
			},
			ReturnType: parser.ArgTypeSliceAny, // Returns slice of maps
		},
		Func: toolBlocksExtractAll,
	})

	if err != nil {
		// Log or wrap the error appropriately
		log.Printf("[ERROR] Failed to register tool 'BlocksExtractAll': %v", err)
		return fmt.Errorf("failed to register blocks tool 'BlocksExtractAll': %w", err)
	}

	// Add registration for other block-related tools here if needed

	fmt.Println("Blocks tools registered via RegisterBlockTools.") // Debug
	return nil                                                     // Return nil on successful registration
}

// --- toolBlocksExtractAll implementation ---
func toolBlocksExtractAll(interpreter *neurogo.Interpreter, args []interface{}) (interface{}, error) {
	// Argument Validation (should ideally happen before calling the func, but good practice)
	if len(args) != 1 {
		return nil, fmt.Errorf("%w: %s expected 1 argument (content), got %d", lang.ErrValidationArgCount, "BlocksExtractAll", len(args))
	}
	content, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[0] 'content', got %T", lang.ErrValidationTypeMismatch, "BlocksExtractAll", args[0])
	}

	logger := interpreter.Logger()

	// Log snippet for debugging
	logSnippet := content
	if len(logSnippet) > 100 {
		logSnippet = logSnippet[:100] + "..."
	}
	logger.Debug("Calling TOOL.BlocksExtractAll (Listener Based)", "snippet", logSnippet)

	// Call the listener-based extractor
	extractedBlocks, extractErr := ExtractAll(content, logger) // Pass logger

	// Handle potential errors from the parser/lexer
	if extractErr != nil {
		logger.Error("TOOL.BlocksExtractAll failed during extraction", "error", extractErr)
		// Return a user-friendly error, masking internal details unless necessary
		// Consider if this should be ErrInternalTool or ErrInvalidArgument depending on cause
		return nil, fmt.Errorf("%w: error during block extraction: %w", lang.ErrInternalTool, extractErr)
	}

	// Convert FencedBlock structs to []interface{} of maps for NeuroScript
	resultsList := make([]interface{}, 0, len(extractedBlocks))
	for _, block := range extractedBlocks {
		// Convert metadata map[string]string to map[string]interface{}
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

	logger.Info("TOOL.BlocksExtractAll successful.", "blocks_found", len(resultsList))
	return resultsList, nil // Return the list of maps
}
