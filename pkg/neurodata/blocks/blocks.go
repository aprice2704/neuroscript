// pkg/neurodata/blocks/blocks.go
package blocks

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	// Import core for tool types
	"github.com/aprice2704/neuroscript/pkg/core"
)

// --- Block Extraction Logic (Moved from core/embed_block_tools.go) ---

// ExtractAllFencedBlocks finds all fenced code blocks (```) in content
// and returns their raw content as a slice of strings. STRICTLY RAW CAPTURE.
// Captures all lines *between* the opening and closing fences.
// Returns an error if an ambiguous fence pattern is detected.
func ExtractAllFencedBlocks(content string) ([]string, error) {
	fmt.Println("[DEBUG Block Extractor] Starting ExtractAllFencedBlocks") // Debug Start
	var allBlocksContent []string
	scanner := bufio.NewScanner(strings.NewReader(content))

	var currentCapturedLines []string = nil
	inBlock := false
	justClosed := false // Flag: Was the *previous* line processed a closing fence?

	fencePattern := regexp.MustCompile("^\\s*```") // Pattern matches start of line

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		isFenceLine := fencePattern.MatchString(line) // Checks if line STARTS with ```
		isExactClosingFence := trimmedLine == "```"   // Checks if line IS EXACTLY ```

		wasJustClosed := justClosed
		justClosed = false

		if inBlock {
			if isExactClosingFence { // --- Closing Fence Found ---
				var blockContent string
				if currentCapturedLines != nil {
					blockContent = strings.Join(currentCapturedLines, "\n")
				} else {
					blockContent = ""
				}
				allBlocksContent = append(allBlocksContent, blockContent)
				inBlock = false
				justClosed = true // Set flag
				currentCapturedLines = nil
			} else { // --- Still Inside Block ---
				if currentCapturedLines == nil {
					currentCapturedLines = make([]string, 0)
				}
				currentCapturedLines = append(currentCapturedLines, line)
			}
		} else { // Not inBlock
			if wasJustClosed && isFenceLine { // --- Ambiguous Fence ---
				fmt.Printf("[DEBUG Block Extractor L%d] **** AMBIGUITY DETECTED ****\n", lineNum)
				return allBlocksContent, fmt.Errorf("ambiguous fence pattern: line %d starts with '```' immediately after a previous block closed at line %d", lineNum, lineNum-1)
			} else if isFenceLine { // --- Opening Fence Found ---
				inBlock = true
				currentCapturedLines = nil
			}
			// --- Outside Block, Not a Fence (Ignore) ---
		}
	} // End loop
	fmt.Println("[DEBUG Block Extractor] --- Loop Finished ---")

	if err := scanner.Err(); err != nil {
		return allBlocksContent, fmt.Errorf("scanner error: %w", err)
	}

	var returnError error = nil
	if inBlock { // --- Handling for Unclosed Block at EOF ---
		fmt.Printf("[DEBUG Block Extractor] EOF Check: InBlock is true (unclosed block).\n")
		var blockContent string
		if currentCapturedLines != nil {
			blockContent = strings.Join(currentCapturedLines, "\n")
		} else {
			blockContent = ""
		}
		allBlocksContent = append(allBlocksContent, blockContent)
		returnError = fmt.Errorf("malformed content: unclosed fence found at EOF")
	}

	fmt.Printf("[DEBUG Block Extractor] --- Returning. Error: %v ---\n", returnError)
	return allBlocksContent, returnError
}

// --- Tool Definition & Implementation (Moved/Adapted from core/tools_composite_doc.go) ---

// RegisterBlockTools adds block manipulation tools to the core registry.
// Called from gonsi/main.go
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
			ReturnType: core.ArgTypeString, // Returns content or error message string
		},
		Func: toolExtractFencedBlockByID, // Use the ID-matching implementation
	})
	// Add other block-related tools here in the future (e.g., ExtractAllFencedBlocksTool)
}

// toolExtractFencedBlockByID is the implementation for TOOL.ExtractFencedBlock.
// Extracts a single block based on ID (or first block if ID is empty).
// This reuses the complex logic previously developed in core/tools_composite_doc.go
func toolExtractFencedBlockByID(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	// Validation handled by core.ValidateAndConvertArgs
	content := args[0].(string)
	targetBlockID := args[1].(string)
	expectedBlockTypeOpt := ""
	if len(args) > 2 && args[2] != nil {
		expectedBlockTypeOpt = args[2].(string) // Validation ensures it's string if present
	}

	logger := interpreter.Logger() // Get logger safely
	logDebug := func(format string, v ...interface{}) {
		if logger != nil {
			logger.Printf("[TOOL.ExtractFencedBlock] "+format, v...)
		}
		// Also print to console for easier debugging during runs
		fmt.Printf("[DEBUG EFB Tool] "+format+"\n", v...)
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
	targetBlockWasEntered := false // Track if we ever entered the target block
	var currentCapturedLines []string
	inAnyBlock := false
	isCurrentBlockTarget := false
	var currentBlockType string
	var currentBlockID string
	checkedForID := false // Have we checked the metadata lines for the current block?

	// Regex patterns (same as before)
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
				foundTargetAndClosed = true // Mark we found AND closed the target
				logDebug("L%d: Closing fence matches target block (ID: '%s' or first block). Setting foundTargetAndClosed=true.", lineNum, targetBlockID)
				// Assign captured content (handle nil slice)
				if currentCapturedLines == nil {
					finalContentResult = ""
				} else {
					finalContentResult = strings.Join(currentCapturedLines, "\n")
				}
				logDebug("L%d: Assigned final content (len %d)", lineNum, len(finalContentResult))
			} else {
				logDebug("L%d: Closing fence found for NON-TARGET block (Type: %s, ID found: %s)", lineNum, currentBlockType, currentBlockID)
			}
			// Reset state for potential next block
			inAnyBlock = false
			isCurrentBlockTarget = false
			currentBlockType = ""
			currentBlockID = ""
			currentCapturedLines = nil
			checkedForID = false
			// If we found and closed the target (either specific ID or the first block), stop.
			if foundTargetAndClosed {
				logDebug("L%d: Target block found and closed, stopping scan.", lineNum)
				break
			}
			continue // Move to next line
		}

		if !inAnyBlock { // --- Looking for Opening Fence ---
			matches := fencePattern.FindStringSubmatch(line)
			if len(matches) > 1 { // Found opening fence ```optional_lang
				// Ignore subsequent blocks if we already found and closed our target
				if foundTargetAndClosed {
					logDebug("L%d: Ignoring subsequent opening fence, target already found and closed.", lineNum)
					continue
				}
				logDebug("L%d: Opening Fence found. Type='%s'", lineNum, matches[1])
				inAnyBlock = true
				currentBlockType = matches[1]
				isCurrentBlockTarget = false // Assume not target until ID/type checked
				currentBlockID = ""
				currentCapturedLines = nil
				checkedForID = false

				// Handle "match first block" scenario
				if matchFirstBlock && !targetBlockWasEntered {
					logDebug("L%d: Matching FIRST block encountered.", lineNum)
					// Check type if specified
					if expectedBlockTypeOpt != "" && !strings.EqualFold(currentBlockType, expectedBlockTypeOpt) {
						logDebug("L%d: Type mismatch ERROR for first block.", lineNum)
						return fmt.Sprintf("Error: First block found, but type mismatch: expected '%s', got '%s'", expectedBlockTypeOpt, currentBlockType), nil
					}
					logDebug("L%d: First block TYPE matches (or no type expected). Setting as target.", lineNum)
					isCurrentBlockTarget = true
					targetBlockWasEntered = true // Mark that we entered it
				}
				continue // Move to next line (process metadata/content)
			}
			// Not in block, and not an opening fence, ignore line
			continue
		}

		// --- Inside a Block ---
		if !checkedForID { // --- Metadata Header Processing Phase ---
			// Skip blank lines at the start of the block content
			if trimmedLine == "" {
				logDebug("L%d: Skipping blank line during metadata check phase.", lineNum)
				continue
			}
			// Skip known metadata lines that aren't the ID
			if skipMetadataPattern.MatchString(line) {
				logDebug("L%d: Skipping specific known metadata line: %q", lineNum, line)
				continue
			}
			// Look for the ID line specifically
			idMatches := metadataPattern.FindStringSubmatch(line)
			if idMatches != nil && len(idMatches) > 1 { // Found ID Line
				foundID := idMatches[1]
				currentBlockID = foundID
				logDebug("L%d: Found metadata ID line: id='%s'. Target is '%s'.", lineNum, currentBlockID, targetBlockID)
				// If we are NOT matching the first block, check ID/type now
				if !matchFirstBlock {
					if currentBlockID == targetBlockID {
						// ID matches, now check type if specified
						if expectedBlockTypeOpt != "" && !strings.EqualFold(currentBlockType, expectedBlockTypeOpt) {
							logDebug("L%d: ID matches, but type mismatch ERROR.", lineNum)
							return fmt.Sprintf("Error: Block ID '%s' found, but type mismatch: expected '%s', got '%s'", targetBlockID, expectedBlockTypeOpt, currentBlockType), nil
						}
						logDebug("L%d: ID and Type match (or no type expected). Setting as target.", lineNum)
						isCurrentBlockTarget = true
						targetBlockWasEntered = true // Mark that we entered it
					} else {
						logDebug("L%d: ID '%s' does not match target '%s'.", foundID, targetBlockID)
						isCurrentBlockTarget = false // Ensure it's false
					}
				}
				// Whether it matched or not, continue to next line (still in metadata phase)
				continue
			}
			// If we reach here, the line is not blank, not known metadata, and not the ID line.
			// This means the metadata header phase is over.
			logDebug("L%d: First non-metadata/non-ID line encountered. Ending metadata check phase.", lineNum)
			checkedForID = true
			// If we were looking for a specific ID and didn't find it in the header,
			// then this block cannot be the target.
			if !matchFirstBlock && !targetBlockWasEntered {
				logDebug("L%d: Specific ID '%s' was required but not found in header. Block is not target.", targetBlockID)
				isCurrentBlockTarget = false
			}
			// Fall through to capture this line as content *if* this block is the target
		} // End metadata check phase

		// --- Content Capture Phase ---
		if isCurrentBlockTarget {
			if currentCapturedLines == nil {
				logDebug("L%d: Initializing capture buffer for target block.", lineNum)
				currentCapturedLines = make([]string, 0)
			}
			// Capture the current line (which might be the first content line)
			currentCapturedLines = append(currentCapturedLines, line)
			// logDebug("L%d: Captured line: %q", lineNum, line) // Can be very verbose
		}
	} // End loop over lines

	// --- EOF Handling (Error checking after loop) ---
	if err := scanner.Err(); err != nil {
		logDebug("Scanner Error: %v", err)
		return nil, fmt.Errorf("error scanning content: %w", err) // Return Go error for internal issues
	}
	logDebug("EOF Reached. Final state: inAnyBlock=%v, foundTargetAndClosed=%v, targetBlockWasEntered=%v", inAnyBlock, foundTargetAndClosed, targetBlockWasEntered)

	// Case 1: Still inside a block at EOF (means unclosed fence)
	if inAnyBlock {
		blockInfo := fmt.Sprintf("type '%s'", currentBlockType)
		if currentBlockID != "" {
			blockInfo = fmt.Sprintf("ID '%s', type '%s'", currentBlockID, currentBlockType)
		} else if matchFirstBlock && targetBlockWasEntered {
			blockInfo = "first block"
		}
		errMsg := fmt.Sprintf("Error: Malformed block structure: Block %s started but closing fence '```' not found before end of input", blockInfo)
		logDebug("EOF Final Check [1]: Condition 'inAnyBlock' is TRUE. Returning Missing Fence Error.")
		return errMsg, nil // Return error string
	}

	// Case 2: We found and successfully closed the target block
	if foundTargetAndClosed {
		logDebug("EOF Final Check [2]: Condition 'foundTargetAndClosed' is TRUE. Returning content (len %d).", len(finalContentResult))
		return finalContentResult, nil // Return success string
	}

	// Case 3: We were looking for a specific ID, but never found/entered it
	if targetBlockID != "" && !targetBlockWasEntered {
		errMsg := fmt.Sprintf("Error: Block ID '%s' not found in content", targetBlockID)
		logDebug("EOF Final Check [3]: Target ID specified but 'targetBlockWasEntered' is FALSE. Returning ID Not Found Error.")
		return errMsg, nil // Return error string
	}

	// Case 4: We were looking for the first block, but didn't find any opening fence
	if matchFirstBlock && !targetBlockWasEntered {
		errMsg := "Error: No fenced code blocks found in content"
		logDebug("EOF Final Check [4]: Matching first block, but 'targetBlockWasEntered' is FALSE. Returning No Blocks Found Error.")
		return errMsg, nil // Return error string
	}

	// Case 5: Should be unreachable if logic is sound, but acts as a fallback.
	// This might happen if we entered a block that wasn't the target, and no target was ever found.
	logDebug("EOF Final Check [5]: Fallback Condition (Unexpected state). Returning Generic Error.")
	errMsg := fmt.Sprintf("Error: Could not extract block ID '%s'", targetBlockID)
	if matchFirstBlock {
		errMsg = "Error: Could not extract first block"
	}
	return errMsg + " (unexpected state at EOF).", nil // Return error string
}
