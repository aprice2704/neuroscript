// Package blocks extracts fenced code blocks (```lang ... ```) from text content.
// This file defines the NeuroScript TOOL functions that wrap the block
// extraction and metadata parsing logic for use within NeuroScript procedures.
package blocks

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/core"
)

// RegisterBlockTools adds the block extraction/metadata tools to the registry.
func RegisterBlockTools(registry *core.ToolRegistry) {
	// --- TOOL.BlocksExtractAll registration ---
	registry.RegisterTool(core.ToolImplementation{
		Spec: core.ToolSpec{
			Name: "BlocksExtractAll",
			Description: "Extracts all fenced code blocks (handling nesting and errors) from input content using ANTLR. " +
				"Returns a list of maps, where each map represents a block and contains keys: " +
				"'language_id' (string), 'raw_content' (string), 'start_line' (int), 'end_line' (int), 'metadata' (map[string]string). Returns error string on failure.",
			Args: []core.ArgSpec{
				{Name: "content", Type: core.ArgTypeString, Required: true, Description: "The string content to search within."},
			},
			ReturnType: core.ArgTypeSliceAny,
		},
		Func: toolBlocksExtractAll,
	})

	// --- TOOL.BlockGetMetadata registration ---
	registry.RegisterTool(core.ToolImplementation{
		Spec: core.ToolSpec{
			Name:        "BlockGetMetadata",
			Description: "Parses the raw content string of a single code block to find metadata lines (':: key: value'). Returns a map[string]string of found key-value pairs or error string on failure.",
			Args: []core.ArgSpec{
				{Name: "raw_content", Type: core.ArgTypeString, Required: true, Description: "The raw content string of the block."},
			},
			ReturnType: core.ArgTypeAny,
		},
		Func: toolBlockGetMetadata,
	})
}

// (toolBlocksExtractAll and toolBlockGetMetadata functions remain unchanged from previous version)
// --- toolBlocksExtractAll implementation ---
func toolBlocksExtractAll(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	content := args[0].(string)
	logger := interpreter.Logger()

	if logger != nil {
		logSnippet := content
		if len(logSnippet) > 50 {
			logSnippet = logSnippet[:50] + "..."
		}
		logger.Printf("[DEBUG TOOL] Calling TOOL.BlocksExtractAll on content (snippet): %q", logSnippet)
	}

	extractedBlocks, extractErr := ExtractAll(content, logger)

	if extractErr != nil {
		errMsg := fmt.Sprintf("Error during block extraction: %s", extractErr.Error())
		if logger != nil {
			logger.Printf("[ERROR TOOL] TOOL.BlocksExtractAll failed: %s", extractErr.Error())
		}
		return errMsg, nil
	}

	resultsList := make([]interface{}, 0, len(extractedBlocks))
	for _, block := range extractedBlocks {
		metadataMap, metaErr := LookForMetadata(block.RawContent)
		if metaErr != nil {
			if logger != nil {
				logger.Printf("[WARN TOOL] TOOL.BlocksExtractAll: Failed to get metadata for block at line %d: %v", block.StartLine, metaErr)
			}
			metadataMap = make(map[string]string)
		}

		metadataInterfaceMap := make(map[string]interface{}, len(metadataMap))
		for k, v := range metadataMap {
			metadataInterfaceMap[k] = v
		}

		blockMap := map[string]interface{}{
			"language_id": block.LanguageID,
			"raw_content": block.RawContent,
			"start_line":  int64(block.StartLine),
			"end_line":    int64(block.EndLine),
			"metadata":    metadataInterfaceMap,
		}
		resultsList = append(resultsList, blockMap)
	}

	if logger != nil {
		logger.Printf("[DEBUG TOOL] TOOL.BlocksExtractAll successful. Found %d blocks.", len(resultsList))
	}
	return resultsList, nil
}

// --- toolBlockGetMetadata implementation ---
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

	metadataMap, metaErr := LookForMetadata(rawContent)

	if metaErr != nil {
		errMsg := fmt.Sprintf("Error getting block metadata: %s", metaErr.Error())
		if logger != nil {
			logger.Printf("[ERROR TOOL] TOOL.BlockGetMetadata failed: %s", metaErr.Error())
		}
		return errMsg, nil
	}

	metadataInterfaceMap := make(map[string]interface{}, len(metadataMap))
	for k, v := range metadataMap {
		metadataInterfaceMap[k] = v
	}

	if logger != nil {
		logger.Printf("[DEBUG TOOL] TOOL.BlockGetMetadata successful. Found metadata: %v", metadataInterfaceMap)
	}
	return metadataInterfaceMap, nil
}
