package core

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

// Assume yySymType and token constants (like IDENTIFIER, STRING_LIT, KW_DEFINE, etc.)
// are defined in neuroscript.y.go, generated from the parser definition.

// Lexer state struct
type lexer struct {
	input       string
	pos         int
	line        int
	startPos    int         // Start position of the current token being scanned
	lastVal     yySymType   // Used internally by parser? Keep for now.
	result      []Procedure // Holds the final parsed procedures
	state       int         // Current lexing state (stDefault or stDocstring)
	docBuffer   bytes.Buffer
	returnedEOF bool // Flag to ensure NEWLINE before final EOF if needed
}

// State constants
const (
	stDefault   = 0
	stDocstring = 1
)

// NewLexer creates a new lexer instance.
func NewLexer(input string) *lexer {
	return &lexer{input: input, line: 1, state: stDefault}
}

// Error handles syntax errors reported by the parser.
func (l *lexer) Error(s string) {
	context := l.currentTokenText()
	if context == "" && l.pos < len(l.input) {
		context = string(l.input[l.pos])
	}
	fmt.Printf("Syntax error on line %d near '%s': %s\n", l.line, context, s)
}

// currentTokenText returns the text of the token currently being processed, truncated for logging.
func (l *lexer) currentTokenText() string {
	start := l.startPos
	end := l.pos
	// Adjust bounds safely
	if start >= len(l.input) {
		start = len(l.input)
	}
	if end > len(l.input) {
		end = len(l.input)
	}
	if start < 0 {
		start = 0
	}
	if start >= end {
		return ""
	}

	text := l.input[start:end]
	if len(text) > 30 {
		text = text[:27] + "..."
	}
	return text
}

// Lex is the main entry point called by the parser to get the next token.
func (l *lexer) Lex(lval *yySymType) int {
	if l.state == stDocstring {
		return l.lexDocstring(lval)
	}
	return l.lexDefault(lval)
}

// lexDocstring handles lexing within a COMMENT: block.
func (l *lexer) lexDocstring(lval *yySymType) int {
	if l.pos >= len(l.input) {
		l.Error("EOF reached while parsing docstring (missing END?)")
		return 0 // EOF
	}
	l.startPos = l.pos // Track start for error context
	startLine := l.line
	l.docBuffer.Reset()

	for l.pos < len(l.input) {
		lineEnd := l.pos
		for lineEnd < len(l.input) && l.input[lineEnd] != '\n' {
			lineEnd++
		}
		lineContent := l.input[l.pos:lineEnd]
		trimmedLine := strings.TrimSpace(lineContent)

		// Check for the END keyword specifically
		if trimmedLine == "END" {
			// Found the end of the docstring
			l.state = stDefault // Switch back to default state
			lval.str = l.docBuffer.String()

			// Consume the 'END' line itself, including the potential newline after it
			l.pos = lineEnd
			if l.pos < len(l.input) && l.input[l.pos] == '\n' {
				l.pos++
				l.line++
			}
			// Don't consume the newline *before* returning DOC_COMMENT_CONTENT,
			// let the main loop handle newlines after this token.
			return DOC_COMMENT_CONTENT
		}

		// Not the end line, append to buffer
		if l.docBuffer.Len() > 0 {
			l.docBuffer.WriteByte('\n')
		}
		l.docBuffer.WriteString(lineContent)

		// Move past the current line
		l.pos = lineEnd
		if l.pos < len(l.input) && l.input[l.pos] == '\n' {
			l.pos++
			l.line++
		} else if l.pos >= len(l.input) {
			// Reached EOF without finding END
			l.pos = l.startPos // Reset for error context
			l.line = startLine
			l.Error("EOF reached while parsing docstring (missing END?)")
			return 0 // EOF
		}
	}
	// Should not be reached if EOF handling is correct
	l.Error("Unexpected exit from docstring lexing loop")
	return 0
}

// lexDefault handles lexing in the default state (outside docstrings).
func (l *lexer) lexDefault(lval *yySymType) int {
	// 1. Skip insignificant characters (whitespace, comments, line continuations)
	if l.lexSkipInsignificant() {
		// If skipping encountered EOF, handle final NEWLINE injection if needed
		return l.handleEOF()
	}

	// 2. Check for EOF *after* skipping
	if l.pos >= len(l.input) {
		return l.handleEOF()
	}

	// 3. Set start position for the significant token
	l.startPos = l.pos
	char := rune(l.input[l.pos])

	// 4. Handle specific single characters or keywords first
	if char == '\n' {
		return l.lexNewline()
	}
	if strings.HasPrefix(l.input[l.pos:], "COMMENT:") {
		return l.lexCommentKeyword(lval)
	}

	// 5. Try matching different token types
	// Order can matter here (e.g., check operators before identifiers if symbols overlap)
	if token := l.lexOperator(lval); token > 0 {
		return token
	}
	if unicode.IsLetter(char) || char == '_' {
		return l.lexIdentifierOrKeyword(lval)
	}
	if char == '"' || char == '\'' {
		return l.lexStringLiteral(lval)
	}
	if unicode.IsDigit(char) {
		return l.lexNumericLiteral(lval)
	}

	// 6. If none matched, it's an unexpected character
	l.Error(fmt.Sprintf("unexpected character: %q", char))
	l.pos++ // Consume the invalid character to prevent infinite loops
	return INVALID
}

// lexSkipInsignificant consumes whitespace, comments, and line continuations.
// Returns true if EOF was reached during skipping, false otherwise.
func (l *lexer) lexSkipInsignificant() bool {
	for l.pos < len(l.input) {
		startPosBeforeSkip := l.pos
		skippedSomethingThisPass := false

		// Skip horizontal whitespace (' ' and '\t')
		for l.pos < len(l.input) {
			char := rune(l.input[l.pos])
			if char == ' ' || char == '\t' {
				l.pos++
				skippedSomethingThisPass = true
			} else {
				break
			}
		}

		// Skip Comments ( '#' or '--' to end of line, including newline)
		if l.pos < len(l.input) {
			char := rune(l.input[l.pos])
			isComment := false
			if char == '#' {
				isComment = true
			}
			if char == '-' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '-' {
				isComment = true
			}
			if isComment {
				for l.pos < len(l.input) && l.input[l.pos] != '\n' {
					l.pos++
				}
				if l.pos < len(l.input) { // Found newline
					l.pos++ // Consume newline
					l.line++
				}
				skippedSomethingThisPass = true
				continue // Restart skipping immediately after a comment line
			}
		}

		// Skip Line Continuations ('\' followed by optional space/tab then '\n')
		if l.pos < len(l.input) {
			if rune(l.input[l.pos]) == '\\' {
				peekPos := l.pos + 1
				for peekPos < len(l.input) {
					peekChar := rune(l.input[peekPos])
					if peekChar == ' ' || peekChar == '\t' {
						peekPos++
					} else {
						break
					}
				}
				if peekPos < len(l.input) && l.input[peekPos] == '\n' {
					l.pos = peekPos + 1 // Consume '\', whitespace, and '\n'
					l.line++
					skippedSomethingThisPass = true
					continue // Restart skipping immediately after a line continuation
				}
			}
		}

		// If we didn't skip anything significant in this pass, break the loop
		if !skippedSomethingThisPass && l.pos == startPosBeforeSkip {
			break
		}
	}
	// Return true if EOF was reached
	return l.pos >= len(l.input)
}

// handleEOF manages the end-of-file condition, injecting a final NEWLINE if necessary.
func (l *lexer) handleEOF() int {
	if !l.returnedEOF {
		l.returnedEOF = true
		// Inject NEWLINE if the last non-EOF character wasn't one
		needsNewlineInjection := (l.pos > 0 && l.input[l.pos-1] != '\n')
		if needsNewlineInjection {
			return NEWLINE
		}
	}
	// If EOF already handled or no NEWLINE needed, return 0 (EOF code for parser)
	return 0
}

// lexNewline handles the NEWLINE token.
func (l *lexer) lexNewline() int {
	l.pos++
	l.line++
	return NEWLINE
}

// lexCommentKeyword handles the COMMENT: keyword and transitions state.
func (l *lexer) lexCommentKeyword(lval *yySymType) int {
	const commentKeyword = "COMMENT:"
	l.pos += len(commentKeyword)
	// Consume the immediately following NEWLINE if present
	if l.pos < len(l.input) && l.input[l.pos] == '\n' {
		l.pos++
		l.line++
	}
	l.state = stDocstring
	l.docBuffer.Reset()
	return KW_COMMENT
}

// lexOperator tries to match and return 1 or 2 character operators.
// Returns the token type (e.g., EQ, PLUS) or 0 if no operator found.
func (l *lexer) lexOperator(lval *yySymType) int {
	// Check 2-character operators first
	if l.pos+1 < len(l.input) {
		twoChars := l.input[l.pos : l.pos+2]
		switch twoChars {
		case "==":
			l.pos += 2
			return EQ
		case "!=":
			l.pos += 2
			return NEQ
		case ">=":
			l.pos += 2
			return GTE
		case "<=":
			l.pos += 2
			return LTE
		case "{{":
			l.pos += 2
			return PLACEHOLDER_START
		case "}}":
			l.pos += 2
			return PLACEHOLDER_END
			// '--' comment handled in lexSkipInsignificant
		}
	}
	// Check 1-character operators
	char := rune(l.input[l.pos])
	switch char {
	case '=':
		l.pos++
		return ASSIGN
	case '+':
		l.pos++
		return PLUS
	case '(':
		l.pos++
		return LPAREN
	case ')':
		l.pos++
		return RPAREN
	case ',':
		l.pos++
		return COMMA
	case '[':
		l.pos++
		return LBRACK
	case ']':
		l.pos++
		return RBRACK
	case '{':
		l.pos++
		return LBRACE
	case '}':
		l.pos++
		return RBRACE
	case ':':
		l.pos++
		return COLON
	case '.':
		l.pos++
		return DOT
	case '>':
		l.pos++
		return GT
	case '<':
		l.pos++
		return LT
		// '#' comment handled in lexSkipInsignificant
		// '\n' handled separately
		// '\' line continuation handled in lexSkipInsignificant
	}
	// No operator matched at this position
	return 0
}

// lexIdentifierOrKeyword handles identifiers and keywords.
func (l *lexer) lexIdentifierOrKeyword(lval *yySymType) int {
	start := l.pos
	l.pos++ // Consume the first letter or underscore

	// Check for the specific __last_call_result identifier early
	const lastCall = "__last_call_result"
	if strings.HasPrefix(l.input[start:], lastCall) {
		boundary := start + len(lastCall)
		// Check if it's correctly bounded (EOF or non-identifier char)
		if boundary == len(l.input) || !isalnum_(rune(l.input[boundary])) {
			l.pos = boundary // Consume the whole identifier
			return KW_LAST_CALL_RESULT
		}
		// If not bounded correctly, reset pos and let general logic handle it
		// (though this specific case is unlikely to be part of a larger identifier)
		l.pos = start + 1
	}

	// Consume remaining identifier characters
	for l.pos < len(l.input) && isalnum_(rune(l.input[l.pos])) {
		l.pos++
	}
	segment := l.input[start:l.pos]

	// Check if it's a keyword (case-insensitive)
	upperSegment := strings.ToUpper(segment)
	switch upperSegment {
	case "DEFINE":
		return KW_DEFINE
	case "PROCEDURE":
		return KW_PROCEDURE
	case "END":
		return KW_END // Note: END also terminates COMMENT:, handled there too.
	case "SET":
		return KW_SET
	case "CALL":
		return KW_CALL
	case "RETURN":
		return KW_RETURN
	case "IF":
		return KW_IF
	case "THEN":
		return KW_THEN
	case "ELSE":
		return KW_ELSE
	case "WHILE":
		return KW_WHILE
	case "DO":
		return KW_DO
	case "FOR":
		return KW_FOR
	case "EACH":
		return KW_EACH
	case "IN":
		return KW_IN
	case "TOOL":
		return KW_TOOL
	case "LLM":
		return KW_LLM
	// Note: COMMENT handled via COMMENT: prefix
	// Note: TRUE/FALSE not keywords yet
	default:
		// Not a keyword, it's an identifier
		lval.str = segment // Store original case
		return IDENTIFIER
	}
}

// lexStringLiteral handles single or double quoted string literals.
func (l *lexer) lexStringLiteral(lval *yySymType) int {
	start := l.pos
	quote := rune(l.input[l.pos])
	l.pos++ // Consume opening quote

	escaped := false
	foundEndQuote := false
	for l.pos < len(l.input) {
		curr := rune(l.input[l.pos])
		if escaped {
			escaped = false // Consume character after escape
		} else if curr == '\\' {
			escaped = true // Mark next character as potentially escaped
		} else if curr == quote {
			l.pos++ // Consume closing quote
			// Store the literal *including* the quotes for now, parser/interpreter can unquote.
			lval.str = l.input[start:l.pos]
			foundEndQuote = true
			break
		} else if curr == '\n' {
			l.line++ // Allow multi-line strings
		}
		l.pos++
	}

	if foundEndQuote {
		return STRING_LIT
	}

	// If loop finished without finding end quote
	l.pos = start // Reset position for better error context
	l.Error(fmt.Sprintf("unclosed string literal starting with %c", quote))
	return INVALID
}

// lexNumericLiteral handles simple integer literals.
// TODO: Expand for floats, different bases?
func (l *lexer) lexNumericLiteral(lval *yySymType) int {
	start := l.pos
	l.pos++ // Consume first digit
	for l.pos < len(l.input) && unicode.IsDigit(rune(l.input[l.pos])) {
		l.pos++
	}
	// TODO: Add float support (check for '.', then more digits)
	lval.str = l.input[start:l.pos]
	return NUMBER_LIT
}

// isalnum_ checks if a rune is a letter, digit, or underscore. (Helper)
func isalnum_(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// SetResult allows the parser to store the final result back into the lexer.
func (l *lexer) SetResult(res []Procedure) {
	l.result = res
}
