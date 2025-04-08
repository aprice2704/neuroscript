// pkg/neurodata/blocks/blocks_extractor.go

// Package blocks extracts fenced code blocks (```lang ... ```) from text content,
// identifying the language identifier and the raw content within the fences.
// It uses an ANTLR lexer for tokenization and Go code to process the token stream.
// Metadata extraction within the block content is handled separately.
package blocks

import (
	"fmt"
	"io"
	"log"
	"strings"

	// *** Ensure this path includes /v4 ***
	"github.com/antlr4-go/antlr/v4"
	// Import generated code (should also use /v4 internally if generated correctly)
	"github.com/aprice2704/neuroscript/pkg/neurodata/blocks/generated"
)

// FencedBlock holds information about an extracted fenced code block.
type FencedBlock struct {
	LanguageID string
	RawContent string
	StartLine  int
	EndLine    int
}

// ExtractAll extracts all fenced blocks using the ANTLR lexer and Go logic.
func ExtractAll(content string, logger *log.Logger) ([]FencedBlock, error) {
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}
	logger.Printf("[DEBUG BLOCKS Extractor] Starting ExtractAll")

	// Use antlr/v4 types
	inputStream := antlr.NewInputStream(content)
	lexer := generated.NewFencedBlockExtractorLexer(inputStream)
	lexerErrorListener := &BlockErrorListener{SourceName: "lexer"}
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(lexerErrorListener) // This uses the v4 interface

	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	tokenStream.Fill()

	if lexerErrorListener.HasErrors() {
		return nil, fmt.Errorf("lexer errors found:\n%s", lexerErrorListener.GetErrors())
	}

	tokens := tokenStream.GetAllTokens()
	logger.Printf("[DEBUG BLOCKS Extractor] Lexer produced %d tokens", len(tokens))

	var blocks []FencedBlock
	var currentContent strings.Builder
	inBlock := false
	nestingLevel := 0
	currentLangID := ""
	currentStartLine := -1
	lastClosedFenceLine := -1

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		// Ensure token conforms to v4 interface if needed (usually okay)
		_ = token.(antlr.Token)

		// Skip hidden channel tokens (WS might be hidden depending on final grammar used)
		// Check channel ID using GetChannel() which is part of v4 Token interface
		// CommonTokenStream typically places hidden tokens off the default channel (0)
		if token.GetChannel() != antlr.TokenDefaultChannel {
			// If WS was put back on default channel, handle it here instead
			if token.GetTokenType() == generated.FencedBlockExtractorLexerWS && inBlock {
				currentContent.WriteString(token.GetText())
			}
			continue // Skip other hidden tokens
		}

		tokenType := token.GetTokenType()
		tokenText := token.GetText()
		tokenLine := token.GetLine()

		// (Rest of the loop logic remains the same as the previous version...)
		if tokenType == generated.FencedBlockExtractorLexerFENCE_MARKER {
			// --- Ambiguity Check ---
			if !inBlock && lastClosedFenceLine != -1 && tokenLine == lastClosedFenceLine+1 {
				prevTokenIndex := i - 1
				for prevTokenIndex >= 0 && tokens[prevTokenIndex].GetChannel() != antlr.TokenDefaultChannel {
					prevTokenIndex--
				} // Find prev non-hidden
				if prevTokenIndex >= 0 && tokens[prevTokenIndex].GetTokenType() == generated.FencedBlockExtractorLexerNEWLINE && tokens[prevTokenIndex].GetLine() == lastClosedFenceLine {
					errMsg := fmt.Sprintf("ambiguous fence detected on line %d, immediately following block closed on line %d", tokenLine, lastClosedFenceLine)
					logger.Printf("[ERROR BLOCKS Extractor] %s", errMsg)
					return blocks, fmt.Errorf(errMsg)
				}
			}
			// --- Lookahead (simplified logic, ensure indices skip hidden) ---
			isPotentialCloseFence := false
			isPotentialOpenFence := false
			langID := ""
			skipTokens := 0
			nextTokenIndex := i + 1
			for nextTokenIndex < len(tokens) && tokens[nextTokenIndex].GetChannel() != antlr.TokenDefaultChannel {
				nextTokenIndex++
			}

			if nextTokenIndex < len(tokens) {
				nextToken := tokens[nextTokenIndex]
				nextTokenType := nextToken.GetTokenType()
				nextTokenLine := nextToken.GetLine()
				if nextTokenType == generated.FencedBlockExtractorLexerNEWLINE || nextTokenType == antlr.TokenEOF {
					isPotentialCloseFence = true
					skipTokens = nextTokenIndex - i
				} else if nextTokenType == generated.FencedBlockExtractorLexerLANG_ID && nextTokenLine == tokenLine {
					langID = nextToken.GetText()
					thirdTokenIndex := nextTokenIndex + 1
					for thirdTokenIndex < len(tokens) && tokens[thirdTokenIndex].GetChannel() != antlr.TokenDefaultChannel {
						thirdTokenIndex++
					}
					if thirdTokenIndex < len(tokens) {
						thirdToken := tokens[thirdTokenIndex]
						if thirdToken.GetTokenType() == generated.FencedBlockExtractorLexerNEWLINE || thirdToken.GetTokenType() == antlr.TokenEOF {
							isPotentialOpenFence = true
							skipTokens = thirdTokenIndex - i
						}
					} else {
						isPotentialOpenFence = true
						skipTokens = nextTokenIndex - i
					} // ```lang at EOF
				} else if nextTokenType == generated.FencedBlockExtractorLexerNEWLINE { // Check for ``` NEWLINE (no lang)
					isPotentialOpenFence = true
					langID = ""
					skipTokens = nextTokenIndex - i
				}
			} else {
				isPotentialCloseFence = true
			} // ``` at EOF

			// --- Process Fence ---
			if inBlock && isPotentialCloseFence { // Closing
				nestingLevel--
				logger.Printf("[DEBUG BLOCKS Extractor L%d] Potential closing fence. Nesting: %d", tokenLine, nestingLevel)
				if nestingLevel == 0 {
					endLine := tokenLine
					logger.Printf("[DEBUG BLOCKS Extractor L%d] Closing block. Start: %d, LangID: %q", endLine, currentStartLine, currentLangID)
					block := FencedBlock{LanguageID: currentLangID, RawContent: currentContent.String(), StartLine: currentStartLine, EndLine: endLine}
					blocks = append(blocks, block)
					inBlock = false
					lastClosedFenceLine = tokenLine
					i += skipTokens
					continue
				} else { // Nested close
					logger.Printf("[DEBUG BLOCKS Extractor L%d] Nested closing fence -> Content", tokenLine)
					currentContent.WriteString(tokenText)
					// Add newline if needed
					nlIdx := i + 1
					for nlIdx < len(tokens) && tokens[nlIdx].GetChannel() != antlr.TokenDefaultChannel {
						nlIdx++
					}
					if nlIdx < len(tokens) && tokens[nlIdx].GetTokenType() == generated.FencedBlockExtractorLexerNEWLINE {
						currentContent.WriteString(tokens[nlIdx].GetText())
					}
				}
			} else if !inBlock && isPotentialOpenFence { // Opening
				logger.Printf("[DEBUG BLOCKS Extractor L%d] Opening fence. LangID: %q", tokenLine, langID)
				inBlock = true
				nestingLevel = 1
				currentLangID = langID
				currentStartLine = tokenLine
				currentContent.Reset()
				lastClosedFenceLine = -1
				i += skipTokens
				continue
			} else { // Malformed or ``` as content
				if inBlock {
					logger.Printf("[DEBUG BLOCKS Extractor L%d] Non-fence '```' -> Content", tokenLine)
					currentContent.WriteString(tokenText)
				} else {
					logger.Printf("[DEBUG BLOCKS Extractor L%d] Ignoring non-opening '```'", tokenLine)
				}
			}
		} else if tokenType == antlr.TokenEOF {
			break
		} else { // --- Other Tokens ---
			if inBlock {
				currentContent.WriteString(tokenText)
			} // Append *all* non-hidden tokens
			if lastClosedFenceLine != -1 && tokenLine > lastClosedFenceLine {
				lastClosedFenceLine = -1
			}
		}
	} // End loop

	if inBlock {
		errMsg := fmt.Sprintf("malformed content: unclosed fenced block starting on line %d", currentStartLine)
		logger.Printf("[ERROR BLOCKS Extractor] %s", errMsg)
		return blocks, fmt.Errorf(errMsg)
	}

	logger.Printf("[DEBUG BLOCKS Extractor] Finished ExtractAll. Found %d blocks.", len(blocks))
	for i := range blocks {
		blocks[i].RawContent = strings.TrimSpace(blocks[i].RawContent)
	} // Final trim
	return blocks, nil
}

// --- Custom Error Listener (using antlr/v4) ---
type BlockErrorListener struct {
	*antlr.DefaultErrorListener // Embed v4 default listener
	Errors                      []string
	SourceName                  string
}

// Override v4 SyntaxError
func (l *BlockErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	errMsg := fmt.Sprintf("%s:%d:%d: %s", l.SourceName, line, column, msg)
	l.Errors = append(l.Errors, errMsg)
}
func (l *BlockErrorListener) HasErrors() bool   { return len(l.Errors) > 0 }
func (l *BlockErrorListener) GetErrors() string { return strings.Join(l.Errors, "\n") }

// Implement other methods required by antlr.ErrorListener if DefaultErrorListener doesn't cover them fully (ReportAmbiguity etc. usually are covered)
// func (l *BlockErrorListener) ReportAmbiguity(...) {} // Implement if needed
// func (l *BlockErrorListener) ReportAttemptingFullContext(...) {} // Implement if needed
// func (l *BlockErrorListener) ReportContextSensitivity(...) {} // Implement if needed
