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

// toolExtractFencedBlock implementation (EOF Fix v37 - Restore v31 logic + fix compile errors)
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

	// Use a local logger function for clarity
	logDebug := func(format string, v ...interface{}) {
		// *** FORCE output to stdout for testing ***
		fmt.Printf("[EBF DBG] "+format+"\n", v...)
		// Also try the interpreter logger if it exists
		if interpreter != nil && interpreter.Logger() != nil {
			interpreter.Logger().Printf("[TOOL.ExtractFencedBlock] "+format, v...)
		}
	}

	logDebug("Start: targetID='%s', expectedType='%s'", targetBlockID, expectedBlockTypeOpt)

	scanner := bufio.NewScanner(strings.NewReader(content))

	// Overall result tracking
	foundTargetAndClosed := false
	finalContentResult := ""
	targetBlockWasEntered := false // <<< RESTORED FLAG

	// Current block state
	var currentCapturedLines []string
	inAnyBlock := false
	isCurrentBlockTarget := false
	idLineFoundInCurrentBlock := false
	currentBlockType := ""
	currentBlockID := ""
	lastUnclosedBlockInfo := ""

	fencePattern := regexp.MustCompile("^```([a-zA-Z0-9-_]*)")
	metadataPattern := regexp.MustCompile(`^(?:#|--)\s*id:\s*(\S+)`)
	commentOrMetaPattern := regexp.MustCompile(`^\s*(#|--)\s*(version:|template:|lang_version:|rendering_hint:|canonical_format:|dependsOn:|howToUpdate:|status:|id:)`)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		// logDebug("L%d: Line: %q", lineNum, line) // Minimal logging unless needed

		// --- Handle Closing Fence ---
		if inAnyBlock && trimmedLine == "```" {
			// logDebug("L%d: Closing Fence. isCurrentBlockTarget=%v", lineNum, isCurrentBlockTarget) // Minimal logging
			if isCurrentBlockTarget {
				// Set success flag FIRST, even if content is empty
				foundTargetAndClosed = true
				logDebug("L%d: Setting foundTargetAndClosed=true for target block '%s'.", lineNum, targetBlockID)
				// Process captured lines
				if currentCapturedLines == nil {
					finalContentResult = ""
				} else {
					finalContentResult = strings.TrimSpace(strings.Join(currentCapturedLines, "\n"))
				}
				logDebug("L%d: Processed content len: %d", lineNum, len(finalContentResult))
			}
			// Reset state for exiting *any* block
			inAnyBlock = false
			isCurrentBlockTarget = false
			idLineFoundInCurrentBlock = false
			currentBlockID = ""
			currentBlockType = ""
			currentCapturedLines = nil
			lastUnclosedBlockInfo = "" // Clear this when *any* block closes
			continue
		}

		// --- Handle Opening Fence ---
		if !inAnyBlock {
			matches := fencePattern.FindStringSubmatch(line)
			if len(matches) > 1 {
				logDebug("L%d: Opening Fence. Type='%s'", lineNum, matches[1])
				inAnyBlock = true
				currentBlockType = matches[1]
				isCurrentBlockTarget = false
				idLineFoundInCurrentBlock = false
				currentBlockID = ""
				lastUnclosedBlockInfo = fmt.Sprintf("type '%s' starting near line %d", currentBlockType, lineNum)
				currentCapturedLines = nil
				continue
			}
		}

		// --- Process Lines Inside a Block ---
		if inAnyBlock {
			wasIdLine := false

			// Step 1: Try find the ID line if not already found for this block.
			if !idLineFoundInCurrentBlock {
				metadataMatches := metadataPattern.FindStringSubmatch(line)
				if len(metadataMatches) > 1 {
					foundID := metadataMatches[1]
					currentBlockID = foundID
					idLineFoundInCurrentBlock = true
					lastUnclosedBlockInfo = fmt.Sprintf("ID '%s'", currentBlockID)
					wasIdLine = true

					logDebug("L%d: Found metadata ID line: id='%s'. Target is '%s'.", lineNum, currentBlockID, targetBlockID)

					if currentBlockID == targetBlockID {
						if expectedBlockTypeOpt != "" && !strings.EqualFold(currentBlockType, expectedBlockTypeOpt) {
							logDebug("L%d: Type mismatch ERROR.", lineNum)
							return fmt.Sprintf("Error: Block ID '%s' found, but type mismatch: expected '%s', got '%s'", targetBlockID, expectedBlockTypeOpt, currentBlockType), nil
						}
						logDebug("L%d: ID/Type match. Setting isCurrentBlockTarget=true & targetBlockWasEntered=true.", lineNum)
						isCurrentBlockTarget = true
						targetBlockWasEntered = true             // <<< SET FLAG
						currentCapturedLines = make([]string, 0) // Initialize buffer
					} else {
						isCurrentBlockTarget = false
						currentCapturedLines = nil
					}
				}
			} // End ID check block

			// Step 2: Capture line if appropriate
			if isCurrentBlockTarget && idLineFoundInCurrentBlock && !wasIdLine {
				if !commentOrMetaPattern.MatchString(line) {
					if currentCapturedLines == nil {
						currentCapturedLines = make([]string, 0)
					}
					// logDebug("L%d: Capturing line: %q", lineNum, line) // Keep logging minimal
					currentCapturedLines = append(currentCapturedLines, line)
				} else {
					logDebug("L%d: Skipping other metadata/comment line: %q", lineNum, line)
				}
			}
			continue
		}
	} // End loop

	// --- After the loop (End of Input reached) ---
	if err := scanner.Err(); err != nil {
		logDebug("Scanner Error: %v", err)
		return nil, fmt.Errorf("error scanning content: %w", err)
	}

	logDebug("EOF Reached. Final state: inAnyBlock=%v, foundTargetAndClosed=%v, targetBlockWasEntered=%v", inAnyBlock, foundTargetAndClosed, targetBlockWasEntered)

	// Check 1: If we were still inside *any* block, it's a missing fence error.
	if inAnyBlock {
		errMsg := fmt.Sprintf("Error: Malformed block structure in content: Block %s started but closing fence '```' not found", lastUnclosedBlockInfo)
		logDebug("EOF Final Check [1]: Condition 'inAnyBlock' is TRUE. Returning Missing Fence Error.")
		return errMsg, nil
	}

	// Check 2: If we successfully found and closed the target block.
	if foundTargetAndClosed {
		logDebug("EOF Final Check [2]: Condition 'foundTargetAndClosed' is TRUE. Returning content (len %d).", len(finalContentResult))
		return finalContentResult, nil
	}

	// Check 3: If target block was entered but not closed successfully.
	if targetBlockWasEntered {
		// Use targetBlockID here because lastUnclosedBlockInfo might have been cleared if another block closed later
		errMsg := fmt.Sprintf("Error: Malformed block structure in content: Block ID '%s' was found but closing fence '```' not found", targetBlockID)
		logDebug("EOF Final Check [3]: Condition 'targetBlockWasEntered' is TRUE (and previous checks false). Returning Missing Target Fence Error.")
		return errMsg, nil
	}

	// Check 4: Otherwise, the target block was never found.
	errMsg := fmt.Sprintf("Error: Block ID '%s' not found in content", targetBlockID)
	logDebug("EOF Final Check [4]: Fallback Condition. Returning ID Not Found Error.")
	return errMsg, nil
}

// toolParseChecklist remains unchanged...
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
