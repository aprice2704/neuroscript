package core

import (
	"fmt"
	// Keep log
	"github.com/aprice2704/neuroscript/pkg/neurodata/metadata" // Import the actual metadata package
)

// registerMetadataTools adds metadata extraction tool.
// *** MODIFIED: Returns error ***
func registerMetadataTools(registry *ToolRegistry) error {
	err := registry.RegisterTool(ToolImplementation{ // Capture potential error
		Spec: ToolSpec{
			Name:        "ExtractMetadata", // Use base name for registry key
			Description: "Extracts ':: key: value' metadata from the beginning of the provided string content. Skips comments/blank lines before the first non-metadata line.",
			Args: []ArgSpec{
				{Name: "content", Type: ArgTypeString, Required: true, Description: "The string content from which to extract metadata."},
			},
			ReturnType: ArgTypeAny, // Returns a map[string]interface{}
		},
		Func: toolExtractMetadataFromString,
	})
	// *** Check error from RegisterTool ***
	if err != nil {
		return fmt.Errorf("failed to register Metadata tool ExtractMetadata: %w", err)
	}
	return nil // Success
}

// toolExtractMetadataFromString extracts metadata from a string.
func toolExtractMetadataFromString(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures args[0] is a string
	content := args[0].(string)

	if interpreter.logger != nil {
		logSnippet := content
		maxLen := 100
		if len(logSnippet) > maxLen {
			logSnippet = logSnippet[:maxLen] + "..."
		}
		interpreter.logger.Info("Tool: ExtractMetadata] Extracting from content (snippet): %q", logSnippet)
	}

	// Call the actual extraction function from the metadata package
	metadataMapString, err := metadata.Extract(content)
	if err != nil {
		// Should Extract return an error? Currently it doesn't seem to based on tests.
		// If it could, we'd handle it here.
		errMsg := fmt.Sprintf("ExtractMetadata failed: %s", err.Error())
		if interpreter.logger != nil {
			interpreter.logger.Error("TOOL ExtractMetadata] %s", errMsg)
		}
		return errMsg, nil // Return error message as string
	}

	// Convert map[string]string to map[string]interface{} for NeuroScript compatibility
	metadataMapInterface := make(map[string]interface{}, len(metadataMapString))
	for k, v := range metadataMapString {
		metadataMapInterface[k] = v
	}

	if interpreter.logger != nil {
		interpreter.logger.Info("Tool: ExtractMetadata] Extracted %d metadata pairs.", len(metadataMapInterface))
	}

	return metadataMapInterface, nil // Return the map
}
