// pkg/neurodata/blocks/blocks_parser.go
package blocks

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
	// Import the checklist package to use its exported parser
	"github.com/aprice2704/neuroscript/pkg/neurodata/checklist"
)

// toolBlockParseContent is the implementation for TOOL.BlockParseContent.
// It parses raw block content based on the provided language ID.
func toolBlockParseContent(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	rawContent := args[0].(string)
	languageID := args[1].(string)
	logger := interpreter.Logger()

	if logger != nil {
		logSnippet := rawContent
		if len(logSnippet) > 50 {
			logSnippet = logSnippet[:50] + "..."
		}
		logger.Printf("[DEBUG TOOL] Calling TOOL.BlockParseContent for lang '%s', content (snippet): %q", languageID, logSnippet)
	}

	switch strings.ToLower(languageID) { // Use lowercase for case-insensitive matching
	case "neuroscript":
		// Use the core NeuroScript parser
		parseOptions := core.ParseOptions{
			DebugAST: false,  // Disable detailed AST logging from within the tool by default
			Logger:   logger, // Pass the logger for potential parser warnings/errors
		}
		stringReader := strings.NewReader(rawContent)
		// Source name can be generic here as we're parsing a fragment
		procedures, fileVersion, parseErr := core.ParseNeuroScript(stringReader, fmt.Sprintf("block(lang=%s)", languageID), parseOptions)

		if parseErr != nil {
			errMsg := fmt.Sprintf("Error parsing neuroscript block: %s", parseErr.Error())
			if logger != nil {
				logger.Printf("[ERROR TOOL] TOOL.BlockParseContent neuroscript parse failed: %v", parseErr)
			}
			return errMsg, nil // Return error string
		}

		// Successfully parsed. Return a summary map.
		// Returning the full AST might be too complex for NeuroScript interface.
		procSummaries := make([]map[string]interface{}, len(procedures))
		for i, proc := range procedures {
			procSummaries[i] = map[string]interface{}{
				"name":        proc.Name,
				"param_count": int64(len(proc.Params)), // Use int64
				"purpose":     proc.Docstring.Purpose,  // Example metadata
			}
		}
		resultMap := map[string]interface{}{
			"parse_success": true,
			"file_version":  fileVersion, // Include file version if found
			"procedures":    procSummaries,
		}
		if logger != nil {
			logger.Printf("[DEBUG TOOL] TOOL.BlockParseContent neuroscript parse successful. Procedures: %d", len(procedures))
		}
		return resultMap, nil

	case "neurodata-checklist":
		// Use the exported checklist parser
		parsedItems, parseErr := checklist.ParseChecklistContent(rawContent)
		if parseErr != nil {
			errMsg := fmt.Sprintf("Error parsing neurodata-checklist block: %s", parseErr.Error())
			if logger != nil {
				logger.Printf("[ERROR TOOL] TOOL.BlockParseContent checklist parse failed: %v", parseErr)
			}
			return errMsg, nil // Return error string
		}

		// Return the list of maps directly (as []interface{})
		// Convert []map[string]interface{} to []interface{}
		resultSlice := make([]interface{}, len(parsedItems))
		for i, item := range parsedItems {
			resultSlice[i] = item
		}

		if logger != nil {
			logger.Printf("[DEBUG TOOL] TOOL.BlockParseContent checklist parse successful. Items: %d", len(resultSlice))
		}
		return resultSlice, nil

	default:
		// Language not supported for parsing by this tool
		errMsg := fmt.Sprintf("Parsing not currently supported for language ID: '%s'", languageID)
		if logger != nil {
			logger.Printf("[WARN TOOL] TOOL.BlockParseContent: %s", errMsg)
		}
		// Return map indicating lack of support, not necessarily an "error" string
		resultMap := map[string]interface{}{
			"parse_success": false,
			"error":         errMsg,
		}
		return resultMap, nil
	}
}
