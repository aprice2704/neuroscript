// pkg/neurodata/blocks/blocks_extractor.go
package blocks

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
	// No core dependency needed here
)

// ExtractAllFencedBlocks extracts all fenced code blocks (content between ```)
// from a string. It performs raw capture, including metadata lines within the fences.
// It returns an error if ambiguous fences (``` immediately after a closing ```)
// or unclosed fences at EOF are detected.
func ExtractAllFencedBlocks(content string) ([]string, error) {
	fmt.Println("[DEBUG Block Extractor] Starting ExtractAllFencedBlocks")
	var allBlocksContent []string
	scanner := bufio.NewScanner(strings.NewReader(content))
	var currentCapturedLines []string = nil
	inBlock := false
	justClosed := false
	fencePattern := regexp.MustCompile("^\\s*```") // Matches opening ``` potentially with lang id
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		isFenceLine := fencePattern.MatchString(line)
		isExactClosingFence := trimmedLine == "```" // Specific check for closing fence
		wasJustClosed := justClosed
		justClosed = false // Reset flag for the current line

		if inBlock {
			if isExactClosingFence { // Closing Fence
				blockContent := ""
				if currentCapturedLines != nil {
					blockContent = strings.Join(currentCapturedLines, "\n")
				}
				allBlocksContent = append(allBlocksContent, blockContent)
				inBlock = false
				justClosed = true // Mark that we just closed a block
				currentCapturedLines = nil
			} else { // Still Inside: Capture raw line
				if currentCapturedLines == nil {
					currentCapturedLines = make([]string, 0)
				}
				currentCapturedLines = append(currentCapturedLines, line)
			}
		} else { // Not inBlock
			if wasJustClosed && isFenceLine { // Ambiguous Fence (fence immediately after closing fence)
				fmt.Printf("[DEBUG Block Extractor L%d] **** AMBIGUITY DETECTED ****\n", lineNum)
				// Return blocks found *before* the ambiguity
				return allBlocksContent, fmt.Errorf("ambiguous fence pattern: line %d starts with '```' immediately after a previous block closed at line %d", lineNum, lineNum-1)
			} else if isFenceLine { // Opening Fence
				inBlock = true
				currentCapturedLines = nil // Reset capture buffer for the new block
			} // Outside Block, Not a Fence: Ignore
		}
	} // End loop

	fmt.Println("[DEBUG Block Extractor] --- Loop Finished ---")
	if err := scanner.Err(); err != nil {
		// Return blocks found before the scanner error
		return allBlocksContent, fmt.Errorf("scanner error: %w", err)
	}

	var returnError error = nil
	if inBlock { // Unclosed Block at EOF
		fmt.Printf("[DEBUG Block Extractor] EOF Check: InBlock is true (unclosed block).\n")
		blockContent := ""
		if currentCapturedLines != nil {
			blockContent = strings.Join(currentCapturedLines, "\n")
		}
		allBlocksContent = append(allBlocksContent, blockContent) // Add the content found so far
		returnError = fmt.Errorf("malformed content: unclosed fence found at EOF")
	}
	fmt.Printf("[DEBUG Block Extractor] --- Returning. Error: %v ---\n", returnError)
	return allBlocksContent, returnError // Return blocks found AND the error status
}
