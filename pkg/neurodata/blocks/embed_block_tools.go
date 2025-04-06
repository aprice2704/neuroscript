// pkg/core/extract_blocks.go
package blocks

import (
	"bufio"
	"fmt" // Added for debugging
	"regexp"
	"strings"
)

// ExtractAllFencedBlocks finds all fenced code blocks (```) in content
// and returns their raw content as a slice of strings. STRICTLY RAW CAPTURE.
// Captures all lines *between* the opening and closing fences.
// Returns an error if an ambiguous fence pattern is detected.
func ExtractAllFencedBlocks(content string) ([]string, error) {
	fmt.Println("--- Starting ExtractAllFencedBlocks ---") // Debug Start
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

		// --- Debug Print Start of Iteration ---
		// fmt.Printf("[DEBUG L%d] Start: line=%q, inBlock=%t, justClosed=%t\n", lineNum, line, inBlock, justClosed)

		wasJustClosed := justClosed // Store flag's value from the previous line
		justClosed = false          // Reset the flag for the current line by default

		// --- Debug Print After Flag Update ---
		// fmt.Printf("[DEBUG L%d] Flags: wasJustClosed=%t, justClosed=%t (reset), isFenceLine=%t, isExactClosingFence=%t\n",
		//	lineNum, wasJustClosed, justClosed, isFenceLine, isExactClosingFence)

		if inBlock {
			// fmt.Printf("[DEBUG L%d] Path: Inside Block\n", lineNum)
			if isExactClosingFence { // We are inside, and this line IS EXACTLY ```
				// fmt.Printf("[DEBUG L%d] Action: Closing Fence found.\n", lineNum)
				// --- Closing Fence Found ---
				var blockContent string
				if currentCapturedLines != nil {
					blockContent = strings.Join(currentCapturedLines, "\n")
					// fmt.Printf("[DEBUG L%d] Captured %d lines for closing block.\n", lineNum, len(currentCapturedLines))
				} else {
					blockContent = ""
					// fmt.Printf("[DEBUG L%d] Captured 0 lines for closing block (currentCapturedLines was nil).\n", lineNum)
				}
				allBlocksContent = append(allBlocksContent, blockContent)
				inBlock = false   // Go OUTSIDE
				justClosed = true // Set flag that we JUST closed a block
				currentCapturedLines = nil
				// fmt.Printf("[DEBUG L%d] State Change: inBlock=false, justClosed=true\n", lineNum)
			} else {
				// --- Still Inside Block (and line is not exact closer) ---
				// fmt.Printf("[DEBUG L%d] Action: Capturing line inside block.\n", lineNum)
				if currentCapturedLines == nil {
					currentCapturedLines = make([]string, 0)
				}
				currentCapturedLines = append(currentCapturedLines, line)
				// inBlock remains true, justClosed remains false
			}
		} else { // Not inBlock
			// fmt.Printf("[DEBUG L%d] Path: Outside Block\n", lineNum)
			// Check if the *previous* line closed a block AND the current line STARTS with ```.
			// fmt.Printf("[DEBUG L%d] Checking Ambiguity: wasJustClosed=%t && isFenceLine=%t\n", lineNum, wasJustClosed, isFenceLine)
			if wasJustClosed && isFenceLine {
				// --- Ambiguous Fence ---
				// fmt.Printf("[DEBUG L%d] **** AMBIGUITY DETECTED - RETURNING ERROR ****\n", lineNum)
				return allBlocksContent, fmt.Errorf("ambiguous fence pattern: line %d starts with '```' immediately after a previous block closed at line %d", lineNum, lineNum-1)
			} else if isFenceLine { // Current line starts with ```, and previous didn't close a block
				// --- Opening Fence Found ---
				// fmt.Printf("[DEBUG L%d] Action: Opening Fence found.\n", lineNum)
				inBlock = true // Go INSIDE for the next line
				currentCapturedLines = nil
				// justClosed remains false
				// fmt.Printf("[DEBUG L%d] State Change: inBlock=true\n", lineNum)
			} else {
				// --- Outside Block, Not a Fence ---
				// fmt.Printf("[DEBUG L%d] Action: Ignoring line outside block.\n", lineNum)
				// inBlock remains false, justClosed remains false
			}
		}
		// fmt.Printf("[DEBUG L%d] End:   inBlock=%t, justClosed=%t\n", lineNum, inBlock, justClosed)
	} // End loop
	fmt.Println("--- Loop Finished ---")

	if err := scanner.Err(); err != nil {
		// fmt.Printf("[DEBUG] Scanner Error: %v\n", err)
		return allBlocksContent, fmt.Errorf("scanner error: %w", err)
	}

	// --- Handling for Unclosed Block at EOF ---
	var returnError error = nil
	if inBlock {
		// fmt.Printf("[DEBUG] EOF Check: InBlock is true (unclosed block).\n")
		var blockContent string
		if currentCapturedLines != nil {
			blockContent = strings.Join(currentCapturedLines, "\n")
			// fmt.Printf("[DEBUG] EOF Captured %d lines for unclosed block.\n", len(currentCapturedLines))
		} else {
			blockContent = ""
			// fmt.Printf("[DEBUG] EOF Captured 0 lines for unclosed block (currentCapturedLines was nil).\n")
		}
		allBlocksContent = append(allBlocksContent, blockContent)
		returnError = fmt.Errorf("malformed content: unclosed fence found at EOF")
	} else {
		// fmt.Printf("[DEBUG] EOF Check: InBlock is false.\n")
	}

	// fmt.Printf("--- Returning from ExtractAllFencedBlocks. Error: %v ---\n", returnError)
	return allBlocksContent, returnError
}
