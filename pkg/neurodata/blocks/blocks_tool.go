// pkg/neurodata/blocks/blocks_tool.go
package blocks

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core" // Depends on core
)

// RegisterBlockTools adds the fenced block extraction tool to the registry.
func RegisterBlockTools(registry *core.ToolRegistry) {
	registry.RegisterTool(core.ToolImplementation{
		Spec: core.ToolSpec{
			Name:        "ExtractFencedBlock",
			Description: "Extracts the raw text content from within a specific fenced code block (```) found within the provided string content. The block is identified by its unique ID (from `# id: ...` or `-- id: ...` metadata) or matches the first block if ID is empty string.",
			Args: []core.ArgSpec{
				{Name: "content", Type: core.ArgTypeString, Required: true, Description: "The string content to search within."},
				{Name: "block_id", Type: core.ArgTypeString, Required: true, Description: "The unique identifier of the block (e.g., 'my-block-1'), or empty string to match the first block found."},
				{Name: "block_type", Type: core.ArgTypeString, Required: false, Description: "Optional: Expected language tag (e.g., 'neuroscript', 'go'). If provided, block must match ID AND type."},
			},
			ReturnType: core.ArgTypeString,
		},
		Func: toolExtractFencedBlockByID,
	})
}

// toolExtractFencedBlockByID - RESTORED to original logic from initial context,
// plus fix to avoid capturing opening fence when matching first block.
func toolExtractFencedBlockByID(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	content := args[0].(string)
	targetBlockID := args[1].(string)
	expectedBlockTypeOpt := ""
	if len(args) > 2 && args[2] != nil {
		expectedBlockTypeOpt = args[2].(string)
	}

	logger := interpreter.Logger()
	logDebug := func(format string, v ...interface{}) {
		logMsg := fmt.Sprintf("[TOOL.ExtractFencedBlock] "+format, v...)
		if logger != nil {
			logger.Output(2, logMsg)
		}
		fmt.Println("[DEBUG EFB Tool] " + fmt.Sprintf(format, v...)) // Keep console debug
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
	// Define skip pattern based on original successful logic (might need adjustment later)
	skipMetadataPattern := regexp.MustCompile(`^\s*(#|--)\s*(version:|template:|lang_version:|rendering_hint:|canonical_format:|dependsOn:|howToUpdate:|status:)`)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// --- Closing Fence Handling ---
		if inAnyBlock && trimmedLine == "```" {
			logDebug("L%d: Potential Closing Fence found. isCurrentBlockTarget=%t", lineNum, isCurrentBlockTarget)
			blockWasTarget := isCurrentBlockTarget

			// Reset state for the next potential block
			inAnyBlock = false
			isCurrentBlockTarget = false
			currentBlockType = ""
			currentBlockID = ""
			checkedForID = false

			if blockWasTarget {
				foundTargetAndClosed = true
				logDebug("L%d: Closing fence matches target block (ID: '%s' or first block). Setting foundTargetAndClosed=true.", lineNum, targetBlockID)
				// Assign captured content (handle nil buffer)
				if currentCapturedLines == nil {
					finalContentResult = ""
				} else {
					finalContentResult = strings.Join(currentCapturedLines, "\n")
				}
				logDebug("L%d: Assigned final content (len %d)", lineNum, len(finalContentResult))
				// Original logic: Stop scan once the target block is found and closed.
				// Keep this break as it seemed to work for simple ID cases initially.
				// We might need to revisit this for the 'unclosed first block' test later.
				logDebug("L%d: Target block found and closed, stopping scan.", lineNum)
				break
			} else {
				logDebug("L%d: Closing fence found for NON-TARGET block.", lineNum)
			}
			currentCapturedLines = nil // Clear buffer after processing fence
			continue
		} // --- End Closing Fence Handling ---

		// --- Opening Fence Handling ---
		if !inAnyBlock {
			matches := fencePattern.FindStringSubmatch(line)
			if len(matches) > 1 { // Found opening fence
				if foundTargetAndClosed { // Should only happen if break wasn't hit? Defensive check.
					logDebug("L%d: Ignoring subsequent opening fence, target already processed.", lineNum)
					continue
				}
				logDebug("L%d: Opening Fence found. Type='%s'", lineNum, matches[1])
				inAnyBlock = true
				currentBlockType = matches[1]
				isCurrentBlockTarget = false // Assume not target until ID/First check passes
				currentBlockID = ""
				currentCapturedLines = nil
				checkedForID = false

				// Check if this *newly opened* block is our target (First Block case)
				if matchFirstBlock && !targetBlockWasEntered {
					logDebug("L%d: Matching FIRST block encountered.", lineNum)
					if expectedBlockTypeOpt != "" && !strings.EqualFold(currentBlockType, expectedBlockTypeOpt) {
						logDebug("L%d: Type mismatch ERROR for first block.", lineNum)
						return fmt.Sprintf("Error: First block found, but type mismatch: expected '%s', got '%s'", expectedBlockTypeOpt, currentBlockType), nil
					}
					logDebug("L%d: First block TYPE matches (or no type expected). Setting as target.", lineNum)
					isCurrentBlockTarget = true
					targetBlockWasEntered = true
				}
				// *** IMPORTANT: Do NOT capture the opening fence line itself ***
				continue // Skip processing the fence line as content
			} else {
				// Not in a block, and not an opening fence line, ignore.
				continue
			}
		} // --- End Opening Fence Handling ---

		// --- Inside Block Handling ---
		if inAnyBlock {
			// --- ID / Metadata Check Phase (only if matching by ID) ---
			if !checkedForID && !matchFirstBlock {
				// Skip standard metadata lines (version, etc.) before looking for ID
				if skipMetadataPattern.MatchString(line) {
					logDebug("L%d: Skipping specific metadata line: %q", lineNum, line)
					continue
				}
				idMatches := metadataPattern.FindStringSubmatch(line)
				if idMatches != nil && len(idMatches) > 1 { // Found ID Line
					foundID := idMatches[1]
					currentBlockID = foundID
					logDebug("L%d: Found metadata ID line: id='%s'. Target is '%s'.", lineNum, currentBlockID, targetBlockID)
					if currentBlockID == targetBlockID {
						// Check type if ID matches
						if expectedBlockTypeOpt != "" && !strings.EqualFold(currentBlockType, expectedBlockTypeOpt) {
							logDebug("L%d: ID matches, but type mismatch ERROR.", lineNum)
							return fmt.Sprintf("Error: Block ID '%s' found, but type mismatch: expected '%s', got '%s'", targetBlockID, expectedBlockTypeOpt, currentBlockType), nil
						}
						logDebug("L%d: ID and Type match (or no type expected). Setting as target.", lineNum)
						isCurrentBlockTarget = true
						targetBlockWasEntered = true // Mark that the specific target block was found
					} else {
						logDebug("L%d: ID '%s' does not match target '%s'.", lineNum, currentBlockID, targetBlockID)
						isCurrentBlockTarget = false
					}
					checkedForID = true // ID check phase is complete for this block
					// Capture the ID line itself if this IS the target block
					if isCurrentBlockTarget {
						if currentCapturedLines == nil {
							currentCapturedLines = make([]string, 0)
						}
						currentCapturedLines = append(currentCapturedLines, line)
					}
					continue // Move to next line after processing ID line
				} else if trimmedLine != "" { // First non-blank, non-skipped-metadata line is NOT ID
					logDebug("L%d: First significant line is not ID/Metadata. Ending ID check phase.", lineNum)
					checkedForID = true // Header phase over for this block
					// Since ID wasn't found, this block is not the target (we are in !matchFirstBlock case)
					isCurrentBlockTarget = false
					// Fall through to potentially capture this line if isCurrentBlockTarget was somehow true (shouldn't happen here)
				} else {
					// It's a blank line within the potential metadata section, capture if target
					if isCurrentBlockTarget { // Only capture if we already know it's the target (e.g., matchFirstBlock)
						if currentCapturedLines == nil {
							currentCapturedLines = make([]string, 0)
						}
						currentCapturedLines = append(currentCapturedLines, line)
					}
					continue // Continue scanning metadata lines
				}
			} // --- End ID / Metadata Check Phase ---

			// --- Content Capture Phase ---
			// Capture the line if this block is the target
			if isCurrentBlockTarget {
				if currentCapturedLines == nil {
					logDebug("L%d: Initializing capture buffer for target block.", lineNum)
					currentCapturedLines = make([]string, 0)
				}
				currentCapturedLines = append(currentCapturedLines, line) // Capture raw line
			}
		} // --- End Inside Block Handling ---
	} // End loop over lines

	// --- EOF Handling --- (Copied from reverted version)
	if err := scanner.Err(); err != nil {
		logDebug("Scanner Error: %v", err)
		return fmt.Sprintf("Error scanning content: %s", err.Error()), nil
	}
	logDebug("EOF Reached. Final state: inAnyBlock=%t, foundTargetAndClosed=%t, targetBlockWasEntered=%t", inAnyBlock, foundTargetAndClosed, targetBlockWasEntered)

	if foundTargetAndClosed {
		logDebug("EOF Final Check [1]: Condition 'foundTargetAndClosed' is TRUE. Returning content (len %d).", len(finalContentResult))
		return finalContentResult, nil
	}
	if inAnyBlock {
		blockInfo := fmt.Sprintf("type '%s'", currentBlockType)
		if currentBlockID != "" {
			blockInfo = fmt.Sprintf("ID '%s', type '%s'", currentBlockID, currentBlockType)
		} else if matchFirstBlock && targetBlockWasEntered {
			blockInfo = "first block"
		} else {
			blockInfo = fmt.Sprintf("last encountered block (type '%s')", currentBlockType)
		}
		errMsg := fmt.Sprintf("Error: Malformed block structure: Block %s started but closing fence '```' not found before end of input", blockInfo)
		logDebug("EOF Final Check [2]: Condition 'inAnyBlock' is TRUE. Returning Missing Fence Error.")
		return errMsg, nil
	}
	if targetBlockID != "" && !targetBlockWasEntered {
		errMsg := fmt.Sprintf("Error: Block ID '%s' not found in content", targetBlockID)
		logDebug("EOF Final Check [3]: Target ID specified but 'targetBlockWasEntered' is FALSE. Returning ID Not Found Error.")
		return errMsg, nil
	}
	if matchFirstBlock && !targetBlockWasEntered {
		errMsg := "Error: No fenced code blocks found in content"
		logDebug("EOF Final Check [4]: Matching first block, but 'targetBlockWasEntered' is FALSE. Returning No Blocks Found Error.")
		return errMsg, nil
	}
	logDebug("EOF Final Check [5]: Fallback Condition (Unexpected state). Returning Generic Error.")
	errMsg := fmt.Sprintf("Error: Could not extract block ID '%s'", targetBlockID)
	if matchFirstBlock {
		errMsg = "Error: Could not extract first block"
	}
	return errMsg + " (unexpected state at EOF).", nil
}
