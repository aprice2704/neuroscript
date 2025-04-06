// pkg/core/tools_composite_doc.go
package core

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
	// "log" // Use interpreter's logger
)

// registerCompositeDocTools registers composite document tools
func registerCompositeDocTools(registry *ToolRegistry) {
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "ExtractFencedBlock",
			Description: "Extracts the raw text content from within a specific fenced code block found within the provided string content. The block is identified by its unique ID (from `# id: ...` or `-- id: ...` metadata) or matches the first block if ID is empty.", // Modified Desc
			Args: []ArgSpec{
				{Name: "content", Type: ArgTypeString, Required: true, Description: "The string content to search within."},
				{Name: "block_id", Type: ArgTypeString, Required: true, Description: "The unique identifier of the block, or empty string to match the first block."}, // Modified Desc
				{Name: "block_type", Type: ArgTypeString, Required: false, Description: "Optional: Expected language tag (e.g., 'neuroscript')."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolExtractFencedBlock, // Points to the ID-matching version below
	})

	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "ParseChecklist",
			Description: "Parses a string formatted as a simple markdown checklist (lines starting with '- [ ]' or '- [x]') into a list of maps. Each map contains 'text' and 'status' ('pending' or 'done'). Ignores lines not matching the pattern.",
			Args: []ArgSpec{
				{Name: "content", Type: ArgTypeString, Required: true, Description: "The string containing the checklist."},
			},
			ReturnType: ArgTypeSliceAny,
		},
		Func: toolParseChecklist,
	})
}

// toolExtractFencedBlock (ID-matching version - v52 logic)
// This function is NOT being tested directly right now, but is kept here.
// It uses ID matching and the logic that passed empty/EOF tests but failed content tests.
func toolExtractFencedBlock(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Using v52 code logic here as requested (passed empty/eof, failed content)
	content := args[0].(string)
	targetBlockID := args[1].(string) // Can be empty string
	expectedBlockTypeOpt := ""
	if len(args) > 2 && args[2] != nil {
		if blockTypeStr, ok := args[2].(string); ok {
			expectedBlockTypeOpt = blockTypeStr
		} else {
			return nil, fmt.Errorf("internal error: optional block_type arg was not a string, got %T", args[2])
		}
	}

	logDebug := func(format string, v ...interface{}) {
		fmt.Printf("[EBF DBG] "+format+"\n", v...)
		if interpreter != nil && interpreter.Logger() != nil {
			interpreter.Logger().Printf("[TOOL.ExtractFencedBlock] "+format, v...)
		}
	}

	matchFirstBlock := (targetBlockID == "")
	if matchFirstBlock {
		logDebug("Start: targetID='' (Matching first block), expectedType='%s'", expectedBlockTypeOpt)
	} else {
		logDebug("Start: targetID='%s', expectedType='%s'", targetBlockID, expectedBlockTypeOpt)
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	var finalContentResult string
	foundTargetAndClosed := false
	targetBlockWasEntered := false
	var currentCapturedLines []string
	inAnyBlock := false
	isCurrentBlockTarget := false
	var currentBlockType string
	var currentBlockID string
	checkedForID := false
	fencePattern := regexp.MustCompile("^```([a-zA-Z0-9-_]*)")
	metadataPattern := regexp.MustCompile(`^(?:#|--)\s*id:\s*(\S+)`)
	skipMetadataPattern := regexp.MustCompile(`^\s*(#|--)\s*(version:|template:|lang_version:|rendering_hint:|canonical_format:|dependsOn:|howToUpdate:|status:)`)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		if inAnyBlock && trimmedLine == "```" { // --- Closing Fence ---
			logDebug("L%d: Potential Closing Fence found. isCurrentBlockTarget=%v", lineNum, isCurrentBlockTarget)
			blockWasTarget := isCurrentBlockTarget
			if blockWasTarget {
				foundTargetAndClosed = true
				logDebug("L%d: Closing fence matches target block (ID: '%s'). Setting foundTargetAndClosed=true.", lineNum, targetBlockID)
				// Post-capture cleanup from v52
				if currentCapturedLines != nil && len(currentCapturedLines) > 0 {
					if strings.TrimSpace(currentCapturedLines[0]) == "" {
						logDebug("L%d: Removing initial blank line from captured content.", lineNum)
						if len(currentCapturedLines) > 1 {
							currentCapturedLines = currentCapturedLines[1:]
						} else {
							currentCapturedLines = nil
						}
					}
				}
				if currentCapturedLines == nil {
					finalContentResult = ""
				} else {
					finalContentResult = strings.Join(currentCapturedLines, "\n")
				}
				logDebug("L%d: Processed final content len: %d", lineNum, len(finalContentResult))
			} else {
				logDebug("L%d: Closing fence found for NON-TARGET block (Type: %s, ID found: %s)", lineNum, currentBlockType, currentBlockID)
			}
			inAnyBlock = false
			isCurrentBlockTarget = false
			currentBlockType = ""
			currentBlockID = ""
			currentCapturedLines = nil
			checkedForID = false
			if matchFirstBlock && blockWasTarget {
				logDebug("L%d: Matched and closed the first block, stopping scan.", lineNum)
				break
			}
			continue
		}
		if !inAnyBlock { // --- Opening Fence ---
			matches := fencePattern.FindStringSubmatch(line)
			if len(matches) > 1 {
				if foundTargetAndClosed {
					logDebug("L%d: Ignoring subsequent opening fence, target already found and closed.", lineNum)
					continue
				}
				logDebug("L%d: Opening Fence. Type='%s'", lineNum, matches[1])
				inAnyBlock = true
				currentBlockType = matches[1]
				isCurrentBlockTarget = false
				currentBlockID = ""
				currentCapturedLines = nil
				checkedForID = false
				if matchFirstBlock && !targetBlockWasEntered {
					logDebug("L%d: Matching first block encountered.", lineNum)
					if expectedBlockTypeOpt != "" && !strings.EqualFold(currentBlockType, expectedBlockTypeOpt) {
						logDebug("L%d: Type mismatch ERROR for first block.", lineNum)
						return fmt.Sprintf("Error: First block found, but type mismatch: expected '%s', got '%s'", expectedBlockTypeOpt, currentBlockType), nil
					}
					isCurrentBlockTarget = true
					targetBlockWasEntered = true
				}
				continue
			}
			continue
		}
		if inAnyBlock { // --- Inside Block ---
			if !checkedForID { // --- Header Processing ---
				if trimmedLine == "" {
					logDebug("L%d: Skipping blank line during header check.", lineNum)
					continue
				}
				if skipMetadataPattern.MatchString(line) {
					logDebug("L%d: Skipping specific metadata line: %q", lineNum, line)
					continue
				}
				idMatches := metadataPattern.FindStringSubmatch(line)
				if idMatches != nil && len(idMatches) > 1 { // Found ID Line
					foundID := idMatches[1]
					currentBlockID = foundID
					logDebug("L%d: Found metadata ID line: id='%s'. Target is '%s'.", lineNum, currentBlockID, targetBlockID)
					if !matchFirstBlock {
						if currentBlockID == targetBlockID {
							if expectedBlockTypeOpt != "" && !strings.EqualFold(currentBlockType, expectedBlockTypeOpt) {
								logDebug("L%d: Type mismatch ERROR.", lineNum)
								return fmt.Sprintf("Error: Block ID '%s' found, but type mismatch: expected '%s', got '%s'", targetBlockID, expectedBlockTypeOpt, currentBlockType), nil
							}
							logDebug("L%d: ID/Type match. Setting isCurrentBlockTarget=true & targetBlockWasEntered=true.", lineNum)
							isCurrentBlockTarget = true
							targetBlockWasEntered = true
						} else {
							logDebug("L%d: ID does not match target.", lineNum)
							isCurrentBlockTarget = false
						}
					}
					checkedForID = true
					continue
				}
				logDebug("L%d: First significant line is not ID/Metadata. Ending ID check phase.", lineNum)
				checkedForID = true // Header phase over
				if !matchFirstBlock {
					logDebug("L%d: No ID found, and specific target ID was given. Block is not target.", lineNum)
					isCurrentBlockTarget = false
				}
				// Fall through to capture this line
			} // --- Content Capture ---
			if isCurrentBlockTarget {
				if currentCapturedLines == nil {
					logDebug("L%d: Initializing capture buffer for target block.", lineNum)
					currentCapturedLines = make([]string, 0)
				}
				currentCapturedLines = append(currentCapturedLines, line)
			}
		} // End if inAnyBlock
	} // End loop

	// --- EOF Handling (same as v52) ---
	if err := scanner.Err(); err != nil {
		logDebug("Scanner Error: %v", err)
		return nil, fmt.Errorf("error scanning content: %w", err)
	}
	logDebug("EOF Reached. Final state: inAnyBlock=%v, foundTargetAndClosed=%v, targetBlockWasEntered=%v", inAnyBlock, foundTargetAndClosed, targetBlockWasEntered)
	if inAnyBlock {
		blockInfo := fmt.Sprintf("type '%s'", currentBlockType)
		if currentBlockID != "" {
			blockInfo = fmt.Sprintf("ID '%s', type '%s'", currentBlockID, currentBlockType)
		}
		errMsg := fmt.Sprintf("Error: Malformed block structure: Block %s started but closing fence '```' not found before end of input", blockInfo)
		logDebug("EOF Final Check [1]: Condition 'inAnyBlock' is TRUE. Returning Missing Fence Error.")
		return errMsg, nil
	}
	if foundTargetAndClosed {
		logDebug("EOF Final Check [2]: Condition 'foundTargetAndClosed' is TRUE. Returning content (len %d).", len(finalContentResult))
		return finalContentResult, nil
	}
	if targetBlockID != "" && !targetBlockWasEntered {
		errMsg := fmt.Sprintf("Error: Block ID '%s' not found in content", targetBlockID)
		logDebug("EOF Final Check [4]: Condition 'targetBlockWasEntered' is FALSE. Returning ID Not Found Error.")
		return errMsg, nil
	}
	if matchFirstBlock && !targetBlockWasEntered {
		errMsg := "Error: No fenced code blocks found in content"
		logDebug("EOF Final Check [5]: Matching first block, but none found.")
		return errMsg, nil
	}
	logDebug("EOF Final Check [6]: Fallback Condition (Unexpected state). Returning Generic Error.")
	return fmt.Sprintf("Error: Could not extract block ID '%s' (unexpected state at EOF).", targetBlockID), nil
}

// toolParseChecklist remains unchanged
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
