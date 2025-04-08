// lexer.go
package checklist

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"
)

// Assumes types CheckItem, MetadataLines AND constants ITEM, NEWLINE_TOK, METADATA_STRING
// are available via the generated checklist_parser.go within this package.

type ChecklistLexer struct {
	scanner       *bufio.Scanner
	Result        []CheckItem   // Stores parsed checklist items
	MetadataLines MetadataLines // Stores raw metadata lines
	lastToken     int           // Tracks last token for NEWLINE logic
	lineNum       int
}

func NewChecklistLexer(r io.Reader) *ChecklistLexer {
	return &ChecklistLexer{
		scanner:       bufio.NewScanner(r),
		Result:        []CheckItem{},
		MetadataLines: make(MetadataLines, 0), // Initialize slice
		lineNum:       0,
	}
}

func (l *ChecklistLexer) Lex(lval *yySymType) int {
	itemRegex := regexp.MustCompile(`^\s*-\s*\[\s*([xX ])\s*\](?:\s*(.*))?\s*$`)
	// Regex for ':: key: value' (requires space after ::)
	metaRegex := regexp.MustCompile(`^\s*::\s+([a-zA-Z0-9_.-]+)\s*:\s*(.*)\s*$`)

	// Handle NEWLINE_TOK after ITEM
	if l.lastToken == ITEM {
		l.lastToken = 0
		return NEWLINE_TOK
	}

	for l.scanner.Scan() {
		l.lineNum++
		line := l.scanner.Text() // Raw line
		trimmedLine := strings.TrimSpace(line)

		// 1. Check for Metadata line FIRST (':: key: value')
		metaMatches := metaRegex.FindStringSubmatch(line)
		if len(metaMatches) == 3 {
			// Store the raw line
			if l.MetadataLines == nil {
				l.MetadataLines = make(MetadataLines, 0)
			}
			l.MetadataLines = append(l.MetadataLines, line)
			// Set lval string field (assuming field 'str' in yySymType for METADATA_STRING)
			lval.str = line
			l.lastToken = METADATA_STRING
			return METADATA_STRING
		}

		// 2. Check for generic Comment line ('#') and SKIP
		if strings.HasPrefix(trimmedLine, "#") {
			l.lastToken = 0 // Reset last token, comment doesn't count for newline logic
			continue        // Skip to next line
		}

		// 3. Skip Blank lines
		if trimmedLine == "" {
			l.lastToken = 0 // Reset last token
			continue        // Skip to next line
		}

		// 4. Try to match Checklist Item ('- [ ] ...')
		itemMatches := itemRegex.FindStringSubmatch(line)
		if len(itemMatches) == 3 {
			// Calculate indentation
			indentationLevel := 0
			for _, char := range line {
				if unicode.IsSpace(char) {
					indentationLevel++
				} else {
					break
				}
			}
			mark := itemMatches[1]
			text := strings.TrimSpace(itemMatches[2])
			status := "pending"
			if strings.ToLower(mark) == "x" {
				status = "done"
			}
			// Populate lval.item (field 'item' based on %token <item> ITEM)
			lval.item = CheckItem{Text: text, Status: status, Indent: indentationLevel}
			l.lastToken = ITEM // Remember ITEM for newline handling
			return ITEM
		}

		// 5. If none of the above: Syntax Error
		l.Error(fmt.Sprintf("line %d: invalid format or unknown line type: %q", l.lineNum, line))
		l.lastToken = 0
		return 0 // Treat as EOF on error
	}

	if err := l.scanner.Err(); err != nil {
		l.Error(fmt.Sprintf("scanner error: %v", err))
	}

	// EOF handling: Emit final NEWLINE_TOK if needed
	if l.lastToken == ITEM {
		l.lastToken = 0
		return NEWLINE_TOK
	}
	return 0 // Standard EOF
}

func (l *ChecklistLexer) Error(s string) {
	fmt.Printf("Syntax Error: %s (at approx line %d)\n", s, l.lineNum)
}
