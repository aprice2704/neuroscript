// pkg/neurodata/blocks/blocks_extractor.go
package blocks

import (
	"bufio" // Keep errors import
	"fmt"
	"io"
	"log"
	"regexp"
	"sort" // Import sort for stable metadata output
	"strings"
	// No need for core package import here unless tools need more than logger
)

// FencedBlock structure - Matches the structure expected by existing tests.
type FencedBlock struct {
	LanguageID string            // Language identifier from the opening fence (e.g., "go", "python")
	RawContent string            // The raw content within the fences
	StartLine  int               // Line number of the opening ```
	EndLine    int               // Line number of the closing ```
	Metadata   map[string]string // Metadata accumulated before the block (:: key: value)
}

var (
	// --- UPDATED Regex ---
	// Regex to capture the language ID (any non-whitespace characters) from an opening fence.
	// Allows empty language ID (just ```).
	openingFenceRegex = regexp.MustCompile("^```(\\S*)$")
	// --- END UPDATED Regex ---

	// Regex for metadata lines (local copy for simplicity, or import your metadata package)
	metadataRegex = regexp.MustCompile(`^\s*::\s+([a-zA-Z0-9_.-]+)\s*:\s*(.*)`)
)

// ExtractAll scans through the input string (file content) line by line
// and extracts blocks based on the specified fence and metadata logic.
func ExtractAll(content string, logger *log.Logger) ([]FencedBlock, error) {
	if logger == nil {
		logger = log.New(io.Discard, "", 0) // Ensure logger is never nil
	}
	logger.Printf("[DEBUG BLOCKS Extractor - Line Scanner v4] Starting ExtractAll") // Version bump for clarity

	var blocks []FencedBlock
	var metadataAccumulator = make(map[string]string)
	var blockAccumulator []string
	fenceLevel := 0
	currentLangID := ""   // Store language ID when fence level becomes 1
	currentStartLine := 0 // Store line number of opening fence

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line) // Trim whitespace for fence checks

		// logger.Printf("[LINE SCAN DEBUG] L%d | Level: %d | Line: %q", lineNumber, fenceLevel, line)

		if fenceLevel == 0 {
			// --- Handling Outside Fences (Level 0) ---

			// Check for Metadata (:: key: value)
			if metadataMatch := metadataRegex.FindStringSubmatch(line); len(metadataMatch) == 3 {
				key := strings.TrimSpace(metadataMatch[1])
				value := strings.TrimSpace(metadataMatch[2])
				// logger.Printf("[LINE SCAN DEBUG] L%d | Level: %d | Action: Found Metadata '%s' = '%s'", lineNumber, fenceLevel, key, value)
				if _, exists := metadataAccumulator[key]; !exists {
					metadataAccumulator[key] = value
				}
				continue // Metadata line processed
			}

			// Check for Opening Fence (```<token> or ```) - Use trimmed line and NEW REGEX
			if openingMatch := openingFenceRegex.FindStringSubmatch(trimmedLine); openingMatch != nil && strings.HasPrefix(trimmedLine, "```") {
				langID := ""
				if len(openingMatch) > 1 {
					langID = openingMatch[1] // Group 1 contains the language ID (or is empty)
				}
				// logger.Printf("[LINE SCAN DEBUG] L%d | Level: %d | Action: Found Opening Fence (Lang: %q)", lineNumber, fenceLevel, langID)
				fenceLevel++
				if fenceLevel == 1 {
					currentLangID = langID
					currentStartLine = lineNumber
					blockAccumulator = []string{}
					// logger.Printf("[LINE SCAN DEBUG] L%d | Level: %d | Action: ---> Entering Level 1 Block (Lang: %q)", lineNumber, fenceLevel, langID)
				} else {
					// Nested opening fence
					// logger.Printf("[LINE SCAN DEBUG] L%d | Level: %d | Action: ---> Entering Nested Level %d", lineNumber, fenceLevel, fenceLevel)
					if blockAccumulator != nil {
						blockAccumulator = append(blockAccumulator, line)
					}
				}
				continue // Opening fence processed
			}

			// Check for *exact* Closing Fence (```) - Error if encountered at level 0
			if trimmedLine == "```" {
				err := fmt.Errorf("line %d: closing fence '```' encountered while not inside a block (level 0)", lineNumber)
				logger.Printf("[ERROR BLOCKS Extractor] %v", err)
				return blocks, err
			}

			// Ignore other lines & clear metadata if needed
			if len(metadataAccumulator) > 0 && trimmedLine != "" {
				// logger.Printf("[LINE SCAN DEBUG] L%d | Level: %d | Action: Clearing metadata due to non-meta/fence line", lineNumber, fenceLevel)
				metadataAccumulator = make(map[string]string)
			}

		} else {
			// --- Handling Inside Fences (Level > 0) ---

			// Check for *exact* Closing Fence (```) - Use trimmed line
			if trimmedLine == "```" {
				// logger.Printf("[LINE SCAN DEBUG] L%d | Level: %d | Action: Found Closing Fence", lineNumber, fenceLevel)
				currentEndLine := lineNumber
				fenceLevel--

				if fenceLevel < 0 {
					err := fmt.Errorf("line %d: fence level decreased below zero, potential mismatch", lineNumber)
					logger.Printf("[ERROR BLOCKS Extractor] %v", err)
					return blocks, err
				}

				if fenceLevel == 0 {
					// Emit the completed block
					finalMetadata := make(map[string]string)
					for k, v := range metadataAccumulator {
						finalMetadata[k] = v
					}
					newBlock := FencedBlock{
						LanguageID: currentLangID,
						Metadata:   finalMetadata,
						RawContent: strings.Join(blockAccumulator, "\n"),
						StartLine:  currentStartLine,
						EndLine:    currentEndLine,
					}
					blocks = append(blocks, newBlock)
					// logger.Printf("[LINE SCAN DEBUG] L%d | Level: %d | Action: <--- Exiting Level 1 Block (Emit)", lineNumber, fenceLevel)

					// Clear accumulators
					metadataAccumulator = make(map[string]string)
					blockAccumulator = nil
					currentLangID = ""
					currentStartLine = 0
				} else {
					// Closing a nested fence
					// logger.Printf("[LINE SCAN DEBUG] L%d | Level: %d | Action: <--- Exiting Nested Level %d", lineNumber, fenceLevel, fenceLevel+1)
					if blockAccumulator != nil {
						blockAccumulator = append(blockAccumulator, line)
					}
				}
				continue // Closing fence processed
			}

			// Check for Opening Fence (```<token> or ```) - handles nested fences (using NEW REGEX)
			if openingMatch := openingFenceRegex.FindStringSubmatch(trimmedLine); openingMatch != nil && strings.HasPrefix(trimmedLine, "```") {
				// logger.Printf("[LINE SCAN DEBUG] L%d | Level: %d | Action: Found Nested Opening Fence", lineNumber, fenceLevel)
				fenceLevel++
				// logger.Printf("[LINE SCAN DEBUG] L%d | Level: %d | Action: ---> Entering Nested Level %d", lineNumber, fenceLevel, fenceLevel)
				if blockAccumulator != nil {
					blockAccumulator = append(blockAccumulator, line)
				}
				continue // Nested opening fence processed
			}

			// If still inside a fence (level > 0), add the line to the current block accumulator.
			if blockAccumulator != nil {
				// logger.Printf("[LINE SCAN DEBUG] L%d | Level: %d | Action: Accumulating line content", lineNumber, fenceLevel)
				blockAccumulator = append(blockAccumulator, line)
			} else {
				err := fmt.Errorf("line %d: internal error: blockAccumulator is nil while fenceLevel > 0", lineNumber)
				logger.Printf("[ERROR BLOCKS Extractor] %v", err)
				return blocks, err
			}
		}
	} // End scanner loop

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		logger.Printf("[ERROR BLOCKS Extractor] Scanner error: %v", err)
		return blocks, fmt.Errorf("error scanning input: %w", err)
	}

	// Check for unclosed fences at EOF
	if fenceLevel != 0 {
		logger.Printf("[WARN BLOCKS Extractor] Reached end of input with unclosed fences (final level: %d). Returning completed blocks.", fenceLevel)
	} else {
		logger.Printf("[DEBUG BLOCKS Extractor] Finished ExtractAll successfully. Found %d blocks.", len(blocks))
	}

	return blocks, nil // Success or partial success with warning logged
}

// --- Formatting Function (As provided by user) ---

// FormatBlocks takes a slice of FencedBlock structs and returns a
// human-readable string representation.
func FormatBlocks(blocks []FencedBlock) string {
	var builder strings.Builder
	separator := "\n" // strings.Repeat("-", 30) + "\n" // Consistent separator

	if len(blocks) == 0 {
		builder.WriteString("No blocks found.\n")
		return builder.String()
	}

	for i, block := range blocks {
		builder.WriteString(fmt.Sprintf("# %d is %q\n", i+1, block.LanguageID))

		// Format Metadata
		if len(block.Metadata) > 0 {
			builder.WriteString("Metadata: ")
			// Sort keys for consistent output order
			keys := make([]string, 0, len(block.Metadata))
			for k := range block.Metadata {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				builder.WriteString(fmt.Sprintf("  %s: %s  ", k, block.Metadata[k]))
			}
			builder.WriteString("\n")
		} else {
			builder.WriteString("Metadata: None\n")
		}

		// Format Line Numbers (Commented out as per user's version)
		// builder.WriteString(fmt.Sprintf("Lines: %d-%d\n", block.StartLine, block.EndLine))

		// Format Content
		// builder.WriteString("Content:\n") // Commented out as per user's version
		builder.WriteString(">" + block.RawContent + "<")
		// Add newline before closing fence if content doesn't end with one
		if !strings.HasSuffix(block.RawContent, "\n") {
			builder.WriteString("\n")
		}
		// builder.WriteString("\n") // Remove extra newline after content marker

		// Add separator between blocks
		if i < len(blocks)-1 {
			builder.WriteString(separator)
		}
	}

	// Ensure final output ends with a newline if there were blocks
	if len(blocks) > 0 && !strings.HasSuffix(builder.String(), "\n") {
		builder.WriteString("\n")
	}

	return builder.String()
}

// --- End Formatting Function ---
