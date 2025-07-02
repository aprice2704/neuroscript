// pkg/neurodata/checklist/scanner_parser.go
package checklist

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/neurodata/metadata"
)

// ChecklistItem definition... (no changes)
type ChecklistItem struct {
	Text        string
	Status      string
	Symbol      rune
	Indent      int
	LineNumber  int
	IsAutomatic bool
}

// ParsedChecklist definition... (no changes)
type ParsedChecklist struct {
	Items    []ChecklistItem
	Metadata map[string]string
}

var (
	// Regexes for skipping lines (unchanged)
	manualItemRegex = regexp.MustCompile(`^\s*-\s\[(.*?)\](?:\s+(.*))?$`)
	autoItemRegex   = regexp.MustCompile(`^\s*-\s\|(.*?)\|(?:\s+(.*))?$`)
	headingRegex    = regexp.MustCompile(`^\s*#+\s+.*`)
	commentRegex    = regexp.MustCompile(`^\s*(#|--).*`)
	blankLineRegex  = regexp.MustCompile(`^\s*$`)
)

// ParseChecklist scans content line by line using string manipulation.
func ParseChecklist(content string, logger interfaces.Logger) (*ParsedChecklist, error) {
	logger.Debug("[DEBUG ChecklistParser V12] Starting ParseChecklist (String Manipulation)") // Version bump

	meta, metaErr := metadata.Extract(content)
	if metaErr != nil {
		logger.Debug("[WARN ChecklistParser V12] Error extracting metadata: %v", metaErr)
	}
	if len(meta) > 0 {
		logger.Debug("[DEBUG ChecklistParser V12] Extracted %d metadata pairs.", len(meta))
	}

	var items []ChecklistItem
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNumber := 0
	itemsSeen := false
	contentFound := false // Track if we found *any* non-skippable line

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// 1. Skip blank/comment/heading
		if blankLineRegex.MatchString(line) || commentRegex.MatchString(line) || headingRegex.MatchString(line) {
			// logger.Debug("[DEBUG ChecklistParser V12] L%d: Skipping blank/comment/heading: %q", lineNumber, line)
			continue
		}

		// 2. Skip metadata before items
		if !itemsSeen && metadata.IsMetadataLine(line) {
			// logger.Debug("[DEBUG ChecklistParser V12] L%d: Skipping pre-item metadata line: %q", lineNumber, line)
			// Even if skipped, it counts as content if metadata was actually extracted
			if len(meta) > 0 {
				contentFound = true
			}
			continue
		}

		contentFound = true // Found a line that wasn't skipped initially
		logger.Debug("[DEBUG ChecklistParser V12] L%d: Processing line: %q", lineNumber, line)

		// --- V11 String Manipulation Logic ---
		trimmedLine := strings.TrimSpace(line)

		if !strings.HasPrefix(trimmedLine, "- ") {
			logger.Debug("[DEBUG ChecklistParser V12] L%d: Non-checklist line, stopping parse: %q", lineNumber, line)
			break
		}

		contentPart := strings.TrimSpace(trimmedLine[2:])

		var delimiterContent string
		var description string
		var isAutomatic bool
		var itemFound bool

		if strings.HasPrefix(contentPart, "[") {
			endBracketPos := strings.Index(contentPart, "]")
			if endBracketPos > 0 {
				delimiterContent = contentPart[1:endBracketPos]
				description = strings.TrimSpace(contentPart[endBracketPos+1:])
				isAutomatic = false
				itemFound = true
				// logger.Debug("[DEBUG ChecklistParser V12] L%d: Found potential Manual Item. Delimiter: %q, Desc: %q", lineNumber, delimiterContent, description)
			}
		}

		if !itemFound && strings.HasPrefix(contentPart, "|") {
			endPipePos := strings.Index(contentPart, "|")
			if endPipePos == 0 {
				endPipePos = strings.Index(contentPart[1:], "|")
				if endPipePos != -1 {
					endPipePos += 1
				}
			}
			if endPipePos > 0 {
				delimiterContent = contentPart[1:endPipePos]
				description = strings.TrimSpace(contentPart[endPipePos+1:])
				isAutomatic = true
				itemFound = true
				// logger.Debug("[DEBUG ChecklistParser V12] L%d: Found potential Automatic Item. Delimiter: %q, Desc: %q", lineNumber, delimiterContent, description)
			}
		}

		if !itemFound {
			logger.Debug("[DEBUG ChecklistParser V12] L%d: Non-checklist line (bad format after '- '), stopping parse: %q", lineNumber, line)
			break
		}

		// --- Process the found item ---
		itemsSeen = true

		indentationLevel := 0
		dashIndex := strings.Index(line, "-")
		if dashIndex >= 0 {
			indentStr := line[:dashIndex]
			indentationLevel = utf8.RuneCountInString(indentStr)
		}
		// logger.Debug("[DEBUG ChecklistParser V12] L%d: Calculated Indent: %d", lineNumber, indentationLevel)

		item := ChecklistItem{
			Text:        description,
			Indent:      indentationLevel,
			LineNumber:  lineNumber,
			IsAutomatic: isAutomatic,
		}

		// Determine Status/Symbol
		switch delimiterContent {
		case " ", "":
			item.Status, item.Symbol = "pending", ' '
		case "x", "X":
			item.Status, item.Symbol = "done", 'x'
		case "-":
			item.Status, item.Symbol = "partial", '-'
		default:
			if utf8.RuneCountInString(delimiterContent) == 1 {
				item.Status = "special"
				runeValue, _ := utf8.DecodeRuneInString(delimiterContent)
				item.Symbol = runeValue
			} else {
				err := fmt.Errorf("line %d: %w: invalid content %q inside %s delimiters",
					lineNumber, ErrMalformedItem, delimiterContent, map[bool]string{true: "| |", false: "[ ]"}[isAutomatic])
				logger.Debug("[ERROR ChecklistParser V12] %v", err)
				return nil, err
			}
		}
		// logger.Debug("[DEBUG ChecklistParser V12] L%d: Parsed Item Status: %q, Symbol: '%c', Auto: %t", lineNumber, item.Status, item.Symbol, item.IsAutomatic)
		items = append(items, item)

	} // End scanner loop

	if err := scanner.Err(); err != nil {
		logger.Debug("[ERROR ChecklistParser V12] Scanner error: %v", err)
		return nil, fmt.Errorf("error scanning checklist content: %w: %w", ErrScannerFailed, err)
	}

	// --- V12 Add check for no content ---
	if !contentFound && len(items) == 0 && len(meta) == 0 {
		logger.Debug("[DEBUG ChecklistParser V12] Finished ParseChecklist. Input had no metadata or items.")
		return nil, ErrNoContent // Return specific error for no content
	}
	// --- End V12 check ---

	logger.Debug("[DEBUG ChecklistParser V12] Finished ParseChecklist. Found %d items.", len(items))
	return &ParsedChecklist{
		Items:    items,
		Metadata: meta,
	}, nil
}
