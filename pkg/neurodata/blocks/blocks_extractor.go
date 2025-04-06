// pkg/neurodata/blocks2/blocks_extractor.go
package blocks

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	// Assuming generated code is in a 'generated' sub-package
	"github.com/aprice2704/neuroscript/pkg/neurodata/blocks/generated"
)

// FencedBlock holds information about an extracted fenced code block.
type FencedBlock struct {
	LanguageID string // Language identifier from the opening fence (e.g., "go", "python")
	RawContent string // The raw string content between the fences (excluding fence lines)
	StartLine  int    // Line number where the block started (opening ``` line)
	EndLine    int    // Line number where the block ended (closing ``` line)
}

// ExtractAll extracts all fenced blocks from the input content using the ANTLR tokenizer.
// It handles nested fences and returns an error for ambiguous or unclosed fences.
func ExtractAll(content string) ([]FencedBlock, error) {
	fmt.Println("[DEBUG BLOCKS2 Extractor] Starting ExtractAll") // Debug

	// 1. Setup ANTLR Lexer
	inputStream := antlr.NewInputStream(content)
	lexer := generated.NewFencedBlockExtractorLexer(inputStream)

	// Optional: Add error listener for lexer errors if needed
	// lexerErrorListener := NewYourErrorListener() // Implement if necessary
	// lexer.RemoveErrorListeners()
	// lexer.AddErrorListener(lexerErrorListener)

	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	tokenStream.Fill() // Load all tokens
	tokens := tokenStream.GetAllTokens()

	fmt.Printf("[DEBUG BLOCKS2 Extractor] Lexer produced %d tokens\n", len(tokens)) // Debug

	// 2. Process Tokens
	var blocks []FencedBlock
	var currentContent strings.Builder
	inBlock := false
	nestingLevel := 0
	currentLangID := ""
	currentStartLine := -1
	var lastClosedFenceLine int = -1 // Track the line number of the last closing fence

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		tokenType := token.GetTokenType()
		tokenText := token.GetText()
		tokenLine := token.GetLine()

		// fmt.Printf("[DEBUG BLOCKS2 Token %d] Type: %s, Text: %q, Line: %d, State: InBlock=%t, Nest=%d\n",
		// 	i, generated.FencedBlockExtractorLexerSymbolicNames[tokenType], tokenText, tokenLine, inBlock, nestingLevel) // Detailed Debug

		// --- Check for Closing Fence ---
		// A closing fence is FENCE_MARKER followed by NEWLINE or EOF
		isPotentialCloseFence := false
		isEOF := false
		if tokenType == generated.FencedBlockExtractorLexerFENCE_MARKER {
			// Check next token (if exists)
			if i+1 < len(tokens) {
				nextToken := tokens[i+1]
				// Closing fence requires a NEWLINE immediately after
				// Allow EOF token type as well
				if nextToken.GetTokenType() == generated.FencedBlockExtractorLexerNEWLINE || nextToken.GetTokenType() == antlr.TokenEOF {
					isPotentialCloseFence = true
					if nextToken.GetTokenType() == antlr.TokenEOF {
						isEOF = true
					}
				}
			} else {
				// FENCE_MARKER at the very end of the token stream implies EOF follows
				isPotentialCloseFence = true
				isEOF = true
			}
		}

		if inBlock && isPotentialCloseFence {
			nestingLevel--
			fmt.Printf("[DEBUG BLOCKS2 Extractor L%d] Potential closing fence. New Nesting Level: %d\n", tokenLine, nestingLevel)

			if nestingLevel == 0 { // --- End of the current block ---
				endLine := tokenLine
				fmt.Printf("[DEBUG BLOCKS2 Extractor L%d] Closing block. StartLine: %d, LangID: %q\n", endLine, currentStartLine, currentLangID)

				// Post-process content: remove leading/trailing blank lines
				rawContent := currentContent.String()
				// Split into lines, remove empty prefix/suffix lines, rejoin
				contentLines := strings.Split(rawContent, "\n")
				startIdx := 0
				endIdx := len(contentLines)
				for startIdx < endIdx && strings.TrimSpace(contentLines[startIdx]) == "" {
					startIdx++
				}
				for endIdx > startIdx && strings.TrimSpace(contentLines[endIdx-1]) == "" {
					endIdx--
				}
				// Handle case where block was only whitespace/newlines
				if startIdx >= endIdx {
					rawContent = ""
				} else {
					rawContent = strings.Join(contentLines[startIdx:endIdx], "\n")
				}

				block := FencedBlock{
					LanguageID: currentLangID,
					RawContent: rawContent,
					StartLine:  currentStartLine,
					EndLine:    endLine,
				}
				blocks = append(blocks, block)

				// Reset state
				inBlock = false
				currentLangID = ""
				currentStartLine = -1
				currentContent.Reset()
				lastClosedFenceLine = tokenLine // Record line of closing fence

				// Consume the NEWLINE/EOF after the FENCE_MARKER
				// Only consume if it wasn't the EOF marker itself
				if !isEOF && i+1 < len(tokens) && tokens[i+1].GetTokenType() == generated.FencedBlockExtractorLexerNEWLINE {
					i++ // Skip the newline token
				}
				continue // Move to the token after the fence sequence
			} else {
				// Nested closing fence, treat ``` as content
				fmt.Printf("[DEBUG BLOCKS2 Extractor L%d] Nested closing fence marker encountered. Treating as content.\n", tokenLine)
				currentContent.WriteString(tokenText) // Add the ```
				// Also add the following newline if it exists and was part of the check
				if !isEOF && i+1 < len(tokens) && tokens[i+1].GetTokenType() == generated.FencedBlockExtractorLexerNEWLINE {
					currentContent.WriteString(tokens[i+1].GetText())
					i++ // Consume newline
				}
				continue // Continue loop after processing nested close
			}
		} else if tokenType == generated.FencedBlockExtractorLexerFENCE_MARKER { // --- Check for Opening Fence ---
			// Potential opening fence marker

			// Ambiguity Check: Fence marker found immediately after a block closed on the previous line.
			// We need to check the *next* non-skippable token's line if the fence marker itself is on the same line as the close.
			// A simpler check: if lastClosedFenceLine != -1 AND the current FENCE_MARKER's line is <= lastClosedFenceLine + 1 (allowing same line or next)
			// This might be too strict, let's stick to the "next line" rule.
			if lastClosedFenceLine != -1 && tokenLine == lastClosedFenceLine+1 {
				// More specific check: was the *previous* token the NEWLINE after the close?
				if i > 0 && tokens[i-1].GetTokenType() == generated.FencedBlockExtractorLexerNEWLINE && tokens[i-1].GetLine() == lastClosedFenceLine {
					fmt.Printf("[DEBUG BLOCKS2 Extractor L%d] Ambiguity detected! Fence follows immediately after closed block on line %d.\n", tokenLine, lastClosedFenceLine)
					return blocks, fmt.Errorf("ambiguous fence detected on line %d, immediately following block closed on line %d", tokenLine, lastClosedFenceLine)
				}
			}

			langID := ""
			hasNextToken := i+1 < len(tokens)
			hasLangID := false
			hasSecondToken := hasNextToken // Alias for readability
			hasThirdToken := i+2 < len(tokens)
			isStartOfBlock := false

			// Look ahead for LANG_ID and NEWLINE pattern or just NEWLINE
			if hasSecondToken && tokens[i+1].GetTokenType() == generated.FencedBlockExtractorLexerLANG_ID && tokens[i+1].GetLine() == tokenLine {
				// Potential ``` LANG_ID ...
				hasLangID = true
				if hasThirdToken && tokens[i+2].GetTokenType() == generated.FencedBlockExtractorLexerNEWLINE {
					// Pattern: ``` LANG_ID NEWLINE
					isStartOfBlock = true
					langID = tokens[i+1].GetText()
				}
			} else if hasSecondToken && tokens[i+1].GetTokenType() == generated.FencedBlockExtractorLexerNEWLINE {
				// Pattern: ``` NEWLINE
				isStartOfBlock = true
			}

			if isStartOfBlock {
				if !inBlock { // Starting a new top-level block
					fmt.Printf("[DEBUG BLOCKS2 Extractor L%d] Opening fence detected. LangID: %q\n", tokenLine, langID)
					inBlock = true
					nestingLevel = 1
					currentLangID = langID
					currentStartLine = tokenLine
					currentContent.Reset()
					lastClosedFenceLine = -1 // Reset ambiguity check flag

					// Consume LANG_ID (if present) and NEWLINE tokens
					if hasLangID {
						i += 2 // Skip FENCE_MARKER, LANG_ID, NEWLINE
					} else {
						i += 1 // Skip FENCE_MARKER, NEWLINE
					}
					continue // Move to next token after the opening sequence
				} else { // Nested opening fence
					fmt.Printf("[DEBUG BLOCKS2 Extractor L%d] Nested opening fence detected. Increasing nesting level.\n", tokenLine)
					nestingLevel++
					// Treat the opening fence sequence as content
					currentContent.WriteString(tokenText) // Add ```
					if hasLangID {
						currentContent.WriteString(tokens[i+1].GetText()) // Add LangID
						if hasThirdToken {
							currentContent.WriteString(tokens[i+2].GetText()) // Add Newline
						}
						i += 2 // Consume all 3
					} else {
						if hasSecondToken {
							currentContent.WriteString(tokens[i+1].GetText()) // Add Newline
						}
						i += 1 // Consume 2
					}
					continue
				}
			} else {
				// FENCE_MARKER not followed by LANG_ID?/NEWLINE - treat as regular text
				if inBlock {
					fmt.Printf("[DEBUG BLOCKS2 Extractor L%d] Non-fence '```' sequence found inside block. Treating as content.\n", tokenLine)
					currentContent.WriteString(tokenText)
				} else {
					// If outside a block, it might be an unterminated fence attempt or just text. Ignore for block extraction.
					fmt.Printf("[DEBUG BLOCKS2 Extractor L%d] Non-fence '```' sequence found outside block. Ignoring.\n", tokenLine)
				}
			}
		} else if tokenType == antlr.TokenEOF {
			// End of File token reached
			break // Exit loop cleanly
		} else { // --- Regular Content Token ---
			if inBlock {
				// Append text of any other token type to the content builder
				currentContent.WriteString(tokenText)
			}
			// Ignore tokens outside blocks that aren't fences
		}

		// Reset ambiguity flag if the current token means we are no longer immediately after a closed block
		if lastClosedFenceLine != -1 && tokenLine > lastClosedFenceLine {
			// Or if it's a different token type on the same line? Check if needed.
			// For now, simply moving to the next line resets it.
			lastClosedFenceLine = -1
		}

	} // End token loop

	// Check for unclosed block at EOF
	if inBlock {
		fmt.Printf("[DEBUG BLOCKS2 Extractor] EOF reached while still inside a block (started line %d).\n", currentStartLine)
		return blocks, fmt.Errorf("malformed content: unclosed fenced block starting on line %d", currentStartLine)
	}

	fmt.Printf("[DEBUG BLOCKS2 Extractor] Finished ExtractAll. Found %d blocks.\n", len(blocks))
	return blocks, nil
}
