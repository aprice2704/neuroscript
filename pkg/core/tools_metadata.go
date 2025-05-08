// NeuroScript Version: 0.3.1
// File version: 0.1.1 // Removed local registerMetadataTools function.
// nlines: 45 // Approximate
// risk_rating: LOW
// filename: pkg/core/tools_metadata.go

package core

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/neurodata/metadata" // Import the actual metadata package
	// "errors" - Not needed currently
	// "log" - Not needed currently
)

// toolExtractMetadataFromString extracts metadata from a string.
// Corresponds to ToolSpec "ExtractMetadata".
func toolExtractMetadataFromString(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "ExtractMetadata" // Assuming this is the public name

	// Argument validation (Count=1, Type=string) expected from validation layer
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: expected 1 argument (content), got %d", toolName, len(args)), ErrArgumentMismatch)
	}
	content, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: content argument must be a string, got %T", toolName, args[0]), ErrInvalidArgument)
	}

	// Log snippet before extraction
	logSnippet := content
	maxLen := 100
	if len(logSnippet) > maxLen {
		logSnippet = logSnippet[:maxLen] + "..."
	}
	interpreter.Logger().Debug(fmt.Sprintf("%s: Extracting from content", toolName), "snippet", logSnippet)

	// Call the actual extraction function
	metadataMapString, err := metadata.Extract(content)
	if err != nil {
		// If metadata.Extract can return errors, wrap them. Assuming it doesn't for now.
		// If it could fail (e.g., malformed input?), use appropriate ErrorCode/Sentinel.
		interpreter.Logger().Error(fmt.Sprintf("%s: Error from metadata.Extract", toolName), "error", err)
		// Depending on the error type, choose ErrorCode. Using ErrorCodeInternal as placeholder.
		return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("%s: failed during metadata extraction: %v", toolName, err), err)
	}

	// Convert map[string]string to map[string]interface{} for NeuroScript compatibility
	metadataMapInterface := make(map[string]interface{}, len(metadataMapString))
	for k, v := range metadataMapString {
		metadataMapInterface[k] = v
	}

	interpreter.Logger().Debug(fmt.Sprintf("%s: Extraction complete", toolName), "pairs_found", len(metadataMapInterface))
	return metadataMapInterface, nil
}

// Removed registerMetadataTools function - registration handled centrally.
