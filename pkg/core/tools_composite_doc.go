// pkg/core/tools_composite_doc.go
package core

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

// registerCompositeDocTools registration remains the same...
func registerCompositeDocTools(registry *ToolRegistry) {
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "ExtractFencedBlock",
			Description: "Extracts the raw text content from within a specific fenced code block found within the provided string content. The block is identified by its unique ID (from `# id: ...` or `-- id: ...` metadata) and optionally verified against its language tag.",
			Args: []ArgSpec{
				{Name: "content", Type: ArgTypeString, Required: true, Description: "The string content to search within."},
				{Name: "block_id", Type: ArgTypeString, Required: true, Description: "The unique identifier of the block."},
				{Name: "block_type", Type: ArgTypeString, Required: false, Description: "Optional: Expected language tag (e.g., 'neuroscript')."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolExtractFencedBlock,
	})

	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "ParseChecklist",
			Description: "Parses a string formatted as a simple markdown checklist (lines starting with '- [ ]' or '- [x]') into a list of maps. Each map contains 'text' and 'status' ('pending' or 'done'). Ignores lines not matching the pattern.",
			Args: []ArgSpec{
				{Name: "content", Type: ArgTypeString, Required: true, Description: "The string containing the checklist."},
			},
			ReturnType: ArgTypeSliceAny, // Returns a list of maps
		},
		Func: toolParseChecklist,
	})
}

// toolExtractFencedBlock implementation (EOF Fix v6 + Logging)
func toolExtractFencedBlock(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	content := args[0].(string)
	targetBlockID := args[1].(string)
	expectedBlockTypeOpt := ""
	if len(args) > 2 && args[2] != nil {
		if blockTypeStr, ok := args[2].(string); ok {
			expectedBlockTypeOpt = blockTypeStr
		} else {
			return nil, fmt.Errorf("internal error: optional block_type arg was not a string, got %T", args[2])
		}
	}

	// Use fmt.Printf for logging to ensure visibility during testing
	logger := func(format string, v ...interface{}) {
		fmt.Printf("[DEBUG EXTRACT LOG] "+format+"\n", v...)
		// Use actual logger if interpreter is available and has one
		// if interpreter != nil && interpreter.logger != nil {
		// 	interpreter.logger.Printf(format, v...)
		// }
	}

	logger("Start: targetID='%s', expectedType='%s'", targetBlockID, expectedBlockTypeOpt)

	scanner := bufio.NewScanner(strings.NewReader(content))
	var capturedLines []string

	foundTargetHeader := false // Found the target ID line within the *current* block
	inTargetBlock := false     // Are we currently inside the block matching the target ID?
	inAnyBlock := false        // Are we currently inside *any* fenced block?
	targetIdEverFound := false // Did we ever find the target ID in *any* block?
	var currentBlockType string

	fencePattern := regexp.MustCompile("^```([a-zA-Z0-9-_]*)")
	metadataPattern := regexp.MustCompile(`^(?:#|--)\s*id:\s*(\S+)`)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if inAnyBlock {
			if trimmedLine == "```" {
				wasInTargetBlock := inTargetBlock
				logger("L%d: Closing Fence '```' detected. Was in target block: %v", lineNum, wasInTargetBlock)
				inAnyBlock = false
				inTargetBlock = false
				currentBlockType = ""
				foundTargetHeader = false

				if wasInTargetBlock {
					logger("L%d: Closed target block '%s'. Processing captured lines.", lineNum, targetBlockID)
					firstRealContentIndex := -1
					for i, l := range capturedLines {
						trimmedL := strings.TrimSpace(l)
						if trimmedL != "" && !strings.HasPrefix(trimmedL, "#") && !strings.HasPrefix(trimmedL, "--") {
							firstRealContentIndex = i
							break
						}
					}
					finalContent := ""
					if firstRealContentIndex != -1 {
						finalContent = strings.Join(capturedLines[firstRealContentIndex:], "\n")
					} else {
						finalContent = strings.Join(capturedLines, "\n")
					}
					logger("L%d: Success - Returning content (len %d).", lineNum, len(finalContent))
					return finalContent, nil // Successful return
				} else {
					logger("L%d: Closed non-target block.", lineNum)
				}
				continue
			}

			if !foundTargetHeader {
				metadataMatches := metadataPattern.FindStringSubmatch(line)
				if len(metadataMatches) > 1 {
					blockID := metadataMatches[1]
					logger("L%d: Found metadata ID line: id='%s'. Target is '%s'.", lineNum, blockID, targetBlockID)
					if blockID == targetBlockID {
						targetIdEverFound = true
						if expectedBlockTypeOpt != "" && !strings.EqualFold(currentBlockType, expectedBlockTypeOpt) {
							logger("L%d: Type mismatch ERROR.", lineNum)
							return fmt.Sprintf("Error: Block ID '%s' found, but type mismatch: expected '%s', got '%s'", targetBlockID, expectedBlockTypeOpt, currentBlockType), nil
						}
						logger("L%d: ID/Type match. Setting foundTargetHeader=true, inTargetBlock=true.", lineNum)
						foundTargetHeader = true
						inTargetBlock = true
						capturedLines = []string{}
					}
					continue // Skip metadata line
				}
				// Inside block, not fence, not metadata, header not found yet -> skip
				continue
			}

			// Inside target block, capture line
			if inTargetBlock {
				capturedLines = append(capturedLines, line)
			}
			continue

		} else { // Not in any block
			matches := fencePattern.FindStringSubmatch(line)
			if len(matches) > 1 {
				logger("L%d: Found opening fence. Type='%s'", lineNum, matches[1])
				inAnyBlock = true
				inTargetBlock = false
				foundTargetHeader = false
				currentBlockType = matches[1]
				capturedLines = make([]string, 0)
			}
			continue
		}
	} // End loop

	// --- After the loop (End of Input reached) ---
	if err := scanner.Err(); err != nil {
		logger("Scanner Error: %v", err)
		return nil, fmt.Errorf("error scanning content: %w", err)
	}

	// ** Add detailed logging before final checks **
	logger("EOF Reached. Final state: inAnyBlock=%v, inTargetBlock=%v, targetIdEverFound=%v", inAnyBlock, inTargetBlock, targetIdEverFound)

	// ** Final Checks - Prioritize missing fence **
	if inTargetBlock {
		// If loop finished while still marked as being inside the target block.
		errMsg := fmt.Sprintf("Error: Malformed block structure in content: Block ID '%s' started but closing fence '```' not found", targetBlockID)
		logger("EOF Final Check: Condition 'inTargetBlock' is TRUE. Returning Missing Fence Error: %s", errMsg)
		return errMsg, nil
	}

	if !targetIdEverFound {
		// If the target ID was never set to true during the scan.
		errMsg := fmt.Sprintf("Error: Block ID '%s' not found in content", targetBlockID)
		logger("EOF Final Check: Condition '!targetIdEverFound' is TRUE. Returning ID Not Found Error: %s", errMsg)
		return errMsg, nil
	}

	// Fallback: ID was found, but not in target block at EOF (implies closed or error earlier).
	errMsg := fmt.Sprintf("Error: Block ID '%s' not found in content (or processing failed after finding)", targetBlockID)
	logger("EOF Final Check: Fallback Condition. Returning Error: %s", errMsg)
	return errMsg, nil

}

// toolParseChecklist implementation (Regex unchanged)
func toolParseChecklist(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	content := args[0].(string)
	result := make([]interface{}, 0)
	itemPattern := regexp.MustCompile(`^-\s+\[([ xX])\]\s+(.*)`)
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		matches := itemPattern.FindStringSubmatch(line)
		if len(matches) == 3 {
			statusChar := strings.ToLower(matches[1])
			taskText := strings.TrimSpace(matches[2])
			status := "pending"
			if statusChar == "x" {
				status = "done"
			}
			itemMap := map[string]interface{}{"text": taskText, "status": status}
			result = append(result, itemMap)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning checklist content: %w", err)
	}
	return result, nil
}
