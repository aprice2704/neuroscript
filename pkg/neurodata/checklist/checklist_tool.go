// pkg/neurodata/checklist/checklist_tool.go
package checklist

import (
	"errors" // Needed for errors.Is
	"fmt"
	"log"

	"github.com/aprice2704/neuroscript/pkg/core" // Import core
)

// RegisterChecklistTools adds the checklist parsing tool.
func RegisterChecklistTools(registry *core.ToolRegistry) error {
	err := registry.RegisterTool(core.ToolImplementation{
		Spec: core.ToolSpec{
			Name: "ParseChecklistFromString",
			Description: "Parses text content formatted as a NeuroData Checklist. " +
				"Extracts file-level ':: key: value' metadata and checklist items. " +
				"Returns a map containing 'metadata' (map[string]string) and 'items' (a list of maps). " +
				"Each item map contains: 'text' (string), 'status' (string: pending, done, partial, special), " +
				"'symbol' (string), 'indent' (int), 'is_automatic' (bool), 'line_number' (int). " +
				"Returns an error message string on parsing failure (e.g., malformed item, no content).",
			Args: []core.ArgSpec{
				{Name: "content", Type: core.ArgTypeString, Required: true, Description: "The string content containing the checklist."},
			},
			ReturnType: core.ArgTypeAny, // Returns map[string]interface{} or error string
		},
		Func: toolParseChecklistFromString,
	})
	if err != nil {
		// Log or handle registration error if needed immediately
		log.Printf("[ERROR] Failed to register tool 'ParseChecklistFromString': %v", err)
		return fmt.Errorf("failed to register checklist tool: %w", err)
	}
	return nil
}

// toolParseChecklistFromString is the Go function implementing the tool.
func toolParseChecklistFromString(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures args[0] is a string
	content := args[0].(string)
	logger := interpreter.Logger() // Use interpreter's logger

	if logger != nil {
		logSnippet := content
		if len(logSnippet) > 100 {
			logSnippet = logSnippet[:100] + "..."
		}
		logger.Info("[TOOL ParseChecklistFromString] Parsing content (snippet): %q", logSnippet)
	}

	// Call the actual parser function (V12 or later recommended)
	parsedData, parseErr := ParseChecklist(content, logger)

	// Handle specific parse errors by returning error strings
	if parseErr != nil {
		errMsg := ""
		if errors.Is(parseErr, ErrNoContent) {
			errMsg = "ParseChecklistFromString Error: Input contains no valid checklist content or metadata."
		} else if errors.Is(parseErr, ErrMalformedItem) {
			// Include the original error message detail for malformed items
			errMsg = fmt.Sprintf("ParseChecklistFromString Error: %s", parseErr.Error())
		} else if errors.Is(parseErr, ErrScannerFailed) {
			errMsg = fmt.Sprintf("ParseChecklistFromString Error: Failed to scan input - %s", parseErr.Error())
		} else {
			// Generic error
			errMsg = fmt.Sprintf("ParseChecklistFromString Error: An unexpected parsing error occurred: %s", parseErr.Error())
		}
		if logger != nil {
			logger.Error("[ERROR TOOL ParseChecklistFromString] %s", errMsg)
		}
		return errMsg, nil // Return error message as string result
	}

	// --- Convert ParsedChecklist struct to map[string]interface{} ---

	// Convert metadata map[string]string to map[string]interface{}
	metadataMapInterface := make(map[string]interface{}, len(parsedData.Metadata))
	for k, v := range parsedData.Metadata {
		metadataMapInterface[k] = v
	}

	// Convert []ChecklistItem to []interface{} of maps
	itemsListInterface := make([]interface{}, len(parsedData.Items))
	for i, item := range parsedData.Items {
		itemMap := map[string]interface{}{
			"text":         item.Text,
			"status":       item.Status,
			"symbol":       string(item.Symbol), // Convert rune to string
			"indent":       int64(item.Indent),  // Convert int to int64 for NeuroScript numbers
			"is_automatic": item.IsAutomatic,
			"line_number":  int64(item.LineNumber), // Convert int to int64
		}
		itemsListInterface[i] = itemMap
	}

	// Create the final result map
	resultMap := map[string]interface{}{
		"metadata": metadataMapInterface,
		"items":    itemsListInterface,
	}

	if logger != nil {
		logger.Info("[TOOL ParseChecklistFromString] Successfully parsed. Returning map with %d metadata keys and %d items.", len(metadataMapInterface), len(itemsListInterface))
	}

	return resultMap, nil // Return the map on success
}
