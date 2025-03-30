package core

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

// Assume yySymType and token constants are defined in neuroscript.y.go

// Lexer state struct
type lexer struct {
	input       string
	pos         int
	line        int
	startPos    int
	lastVal     yySymType
	result      []Procedure
	state       int
	docBuffer   bytes.Buffer
	returnedEOF bool // Flag to ensure NEWLINE before final EOF
}

// State constants
const (
	stDefault   = 0
	stDocstring = 1
)

// NewLexer
func NewLexer(input string) *lexer {
	return &lexer{input: input, line: 1, state: stDefault}
}

// Error
func (l *lexer) Error(s string) {
	context := l.currentTokenText()
	if context == "" && l.pos < len(l.input) {
		context = string(l.input[l.pos])
	}
	fmt.Printf("Syntax error on line %d near '%s': %s\n", l.line, context, s)
}

// currentTokenText
func (l *lexer) currentTokenText() string {
	start := l.startPos
	end := l.pos
	if start >= end && start < len(l.input) {
		end = start + 1
	}
	if end > len(l.input) {
		end = len(l.input)
	}
	if start >= len(l.input) {
		start = len(l.input) - 1
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

// Lex
func (l *lexer) Lex(lval *yySymType) int {
	if l.state == stDocstring {
		fmt.Printf("Lexer: Calling lexDocstring (pos %d)\n", l.pos) // DEBUG
		return l.lexDocstring(lval)
	}
	fmt.Printf("Lexer: Calling lexDefault (pos %d)\n", l.pos) // DEBUG
	return l.lexDefault(lval)                                 // Ensure this calls the v20 version below
}

// lexDocstring (Unchanged from v18 debug version)
func (l *lexer) lexDocstring(lval *yySymType) int {
	fmt.Printf("  lexDocstring: Entered (pos %d, line %d)\n", l.pos, l.line) // DEBUG
	if l.pos >= len(l.input) {
		l.Error("EOF reached while parsing docstring (missing END?)")
		fmt.Printf("  lexDocstring: Returning 0 (EOF)\n") // DEBUG
		return 0                                          // EOF
	}
	l.startPos = l.pos
	startLine := l.line
	l.docBuffer.Reset()

	for l.pos < len(l.input) {
		lineEnd := l.pos
		for lineEnd < len(l.input) && l.input[lineEnd] != '\n' {
			lineEnd++
		}
		lineContent := l.input[l.pos:lineEnd]
		trimmedLine := strings.TrimSpace(lineContent)
		fmt.Printf("  lexDocstring: Read line %d: %q (trimmed: %q)\n", l.line, lineContent, trimmedLine) // DEBUG

		isEndLine := (trimmedLine == "END")

		if isEndLine {
			fmt.Printf("  lexDocstring: Found END line.\n") // DEBUG
			l.state = stDefault
			lval.str = l.docBuffer.String()
			fmt.Printf("  lexDocstring: Returning DOC_COMMENT_CONTENT (length %d), content: %q\n", len(l.docBuffer.String()), l.docBuffer.String()) // DEBUG
			fmt.Printf("  lexDocstring: Switching state to stDefault, pos left at %d\n", l.pos)                                                     // DEBUG
			return DOC_COMMENT_CONTENT
		} else {
			if l.docBuffer.Len() > 0 {
				l.docBuffer.WriteByte('\n')
			}
			l.docBuffer.WriteString(lineContent)
			l.pos = lineEnd
			if l.pos < len(l.input) && l.input[l.pos] == '\n' {
				l.pos++
				l.line++
				fmt.Printf("  lexDocstring: Consumed newline, moving to line %d, pos %d\n", l.line, l.pos) // DEBUG
			} else if l.pos >= len(l.input) {
				l.pos = l.startPos // Use startPos for error context
				l.line = startLine
				l.Error("EOF reached while parsing docstring (missing END?)")
				fmt.Printf("  lexDocstring: Returning 0 (EOF in loop)\n") // DEBUG
				return 0
			}
		}
	}
	fmt.Printf("  lexDocstring: Returning 0 (Fell out of loop?)\n") // DEBUG
	return 0
}

// isalnum_ (Unchanged)
func isalnum_(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// SetResult (Unchanged)
func (l *lexer) SetResult(res []Procedure) {
	l.result = res
}

// =======================================================================
// --- lexDefault Function (v21 - Restructured Loop for Skipping) ---
// =======================================================================

func (l *lexer) lexDefault(lval *yySymType) int {
	for { // Outer loop restarts only after *returning* a token or hitting error

		// *** Phase 1: Skip all insignificant stuff ***
		skippedSomething := true // Assume we might skip something
		for skippedSomething {
			skippedSomething = false // Reset flag for this pass

			// Check EOF within skipper
			if l.pos >= len(l.input) {
				// Handle EOF logic (same as v18)
				if !l.returnedEOF {
					needsNewlineInjection := (l.pos > 0 && l.input[l.pos-1] != '\n')
					l.returnedEOF = true // Set flag immediately
					if needsNewlineInjection {
						fmt.Printf("  lexDefault-Skip: Injecting final NEWLINE before EOF\n") // DEBUG
						return NEWLINE
					} else {
						fmt.Printf("  lexDefault-Skip: Returning 0 (EOF - no injection needed)\n") // DEBUG
						return 0
					}
				} else {
					fmt.Printf("  lexDefault-Skip: Returning 0 (EOF - already handled)\n") // DEBUG
					return 0
				}
			}

			startPosBeforeSkip := l.pos

			// Skip horizontal whitespace
			posBeforeSpace := l.pos
			for l.pos < len(l.input) {
				char := rune(l.input[l.pos])
				if char == ' ' || char == '\t' {
					l.pos++
				} else {
					break
				}
			}
			if l.pos > posBeforeSpace {
				skippedSomething = true
			}

			// Skip Comments (whole line including newline)
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
					commentStartPos := l.pos
					for l.pos < len(l.input) && l.input[l.pos] != '\n' {
						l.pos++
					}
					if l.pos < len(l.input) {
						l.pos++
						l.line++
					} // Consume newline
					fmt.Printf("  lexDefault-Skip: Skipped comment line (%q)\n", l.input[commentStartPos:l.pos]) // DEBUG
					skippedSomething = true
					continue // Restart skipping loop *immediately* after skipping comment line
				}
			}

			// Skip Line Continuations (\ + optional space/tab + \n)
			if l.pos < len(l.input) {
				char := rune(l.input[l.pos])
				if char == '\\' {
					peekPos := l.pos + 1
					// Skip optional spaces/tabs AFTER the backslash
					for peekPos < len(l.input) {
						peekChar := rune(l.input[peekPos])
						if peekChar == ' ' || peekChar == '\t' {
							peekPos++
						} else {
							break
						}
					}
					// Check if the character AFTER the backslash (and any spaces) is a newline
					if peekPos < len(l.input) && l.input[peekPos] == '\n' {
						// Success: Consume everything from the backslash up to and including the newline
						l.pos = peekPos + 1
						l.line++
						fmt.Printf("  lexDefault-Skip: Handled line continuation, now at line %d, pos %d\n", l.line, l.pos) // DEBUG
						skippedSomething = true
						continue // Restart skipping loop *immediately* after skipping continuation
					} else {
						// Invalid continuation - treat backslash as unexpected char later
						// Don't set skippedSomething = true here, let it fall through
					}
				}
			}
			// If we skipped only whitespace, loop again to check for comments/continuations
			if l.pos > startPosBeforeSkip && skippedSomething {
				continue
			}

		} // End of skipping loop

		// *** Phase 2: Identify and return the next significant token ***

		// Re-check EOF after skipping everything
		if l.pos >= len(l.input) {
			if !l.returnedEOF {
				needsNewlineInjection := (l.pos > 0 && l.input[l.pos-1] != '\n')
				l.returnedEOF = true // Set flag immediately
				if needsNewlineInjection {
					fmt.Printf("  lexDefault-Token: Injecting final NEWLINE before EOF\n") // DEBUG
					return NEWLINE
				} else {
					fmt.Printf("  lexDefault-Token: Returning 0 (EOF - no injection needed)\n") // DEBUG
					return 0
				}
			} else {
				fmt.Printf("  lexDefault-Token: Returning 0 (EOF - already handled)\n") // DEBUG
				return 0
			}
		}

		l.startPos = l.pos // Set definitive token start position
		char := rune(l.input[l.pos])

		// Handle NEWLINE itself as a token (if not skipped above)
		if char == '\n' {
			l.pos++
			l.line++
			fmt.Printf("  lexDefault-Token: Returning NEWLINE (line %d)\n", l.line) // DEBUG
			return NEWLINE
		}

		// Handle COMMENT: keyword
		const commentKeyword = "COMMENT:"
		if strings.HasPrefix(l.input[l.pos:], commentKeyword) {
			l.pos += len(commentKeyword)
			// Consume the immediately following NEWLINE
			if l.pos < len(l.input) && l.input[l.pos] == '\n' {
				l.pos++
				l.line++
				fmt.Printf("  lexDefault-Token: Consumed newline after COMMENT:\n") // DEBUG
			} else {
				fmt.Printf("  lexDefault-Token: WARNING - No newline found immediately after COMMENT:\n") // DEBUG
			}
			l.state = stDocstring
			l.docBuffer.Reset()
			fmt.Printf("  lexDefault-Token: Returning KW_COMMENT, switching state to stDocstring\n") // DEBUG
			return KW_COMMENT
		}

		// Handle other tokens (Operators, Literals, Identifiers/Keywords)
		// [ Omitted identical code from v18/v19 for brevity - MAKE SURE IT IS PRESENT IN YOUR FILE ]
		// Operators / Delimiters
		if l.pos+1 < len(l.input) { // 2-char
			twoChars := l.input[l.pos : l.pos+2]
			switch twoChars {
			case "==":
				l.pos += 2
				fmt.Printf("  lexDefault-Token: Returning EQ\n")
				return EQ
			case "!=":
				l.pos += 2
				fmt.Printf("  lexDefault-Token: Returning NEQ\n")
				return NEQ
			case "{{":
				l.pos += 2
				fmt.Printf("  lexDefault-Token: Returning PLACEHOLDER_START\n")
				return PLACEHOLDER_START
			case "}}":
				l.pos += 2
				fmt.Printf("  lexDefault-Token: Returning PLACEHOLDER_END\n")
				return PLACEHOLDER_END
			}
		}
		switch char { // 1-char
		case '=':
			l.pos++
			fmt.Printf("  lexDefault-Token: Returning ASSIGN\n")
			return ASSIGN
		case '+':
			l.pos++
			fmt.Printf("  lexDefault-Token: Returning PLUS\n")
			return PLUS
		case '(':
			l.pos++
			fmt.Printf("  lexDefault-Token: Returning LPAREN\n")
			return LPAREN
		case ')':
			l.pos++
			fmt.Printf("  lexDefault-Token: Returning RPAREN\n")
			return RPAREN
		case ',':
			l.pos++
			fmt.Printf("  lexDefault-Token: Returning COMMA\n")
			return COMMA
		case '[':
			l.pos++
			fmt.Printf("  lexDefault-Token: Returning LBRACK\n")
			return LBRACK
		case ']':
			l.pos++
			fmt.Printf("  lexDefault-Token: Returning RBRACK\n")
			return RBRACK
		case '{':
			l.pos++
			fmt.Printf("  lexDefault-Token: Returning LBRACE\n")
			return LBRACE
		case '}':
			l.pos++
			fmt.Printf("  lexDefault-Token: Returning RBRACE\n")
			return RBRACE
		case ':':
			l.pos++
			fmt.Printf("  lexDefault-Token: Returning COLON\n")
			return COLON // General colon
		case '.':
			l.pos++
			fmt.Printf("  lexDefault-Token: Returning DOT\n")
			return DOT
		}

		// Identifiers / Keywords
		if unicode.IsLetter(char) || char == '_' {
			start := l.pos
			l.pos++
			const lastCall = "__last_call_result"
			if strings.HasPrefix(l.input[start:], lastCall) {
				boundary := start + len(lastCall)
				if boundary == len(l.input) || !isalnum_(rune(l.input[boundary])) {
					l.pos = boundary
					fmt.Printf("  lexDefault-Token: Returning KW_LAST_CALL_RESULT\n")
					return KW_LAST_CALL_RESULT
				}
			}
			l.pos = start + 1
			for l.pos < len(l.input) && isalnum_(rune(l.input[l.pos])) {
				l.pos++
			}
			segment := l.input[start:l.pos]
			switch segment {
			case "DEFINE":
				fmt.Printf("  lexDefault-Token: Returning KW_DEFINE\n")
				return KW_DEFINE
			case "PROCEDURE":
				fmt.Printf("  lexDefault-Token: Returning KW_PROCEDURE\n")
				return KW_PROCEDURE
			case "END":
				fmt.Printf("  lexDefault-Token: Returning KW_END\n")
				return KW_END
			case "SET":
				fmt.Printf("  lexDefault-Token: Returning KW_SET\n")
				return KW_SET
			case "CALL":
				fmt.Printf("  lexDefault-Token: Returning KW_CALL\n")
				return KW_CALL
			case "RETURN":
				fmt.Printf("  lexDefault-Token: Returning KW_RETURN\n")
				return KW_RETURN
			case "IF":
				fmt.Printf("  lexDefault-Token: Returning KW_IF\n")
				return KW_IF
			case "THEN":
				fmt.Printf("  lexDefault-Token: Returning KW_THEN\n")
				return KW_THEN
			case "WHILE":
				fmt.Printf("  lexDefault-Token: Returning KW_WHILE\n")
				return KW_WHILE
			case "DO":
				fmt.Printf("  lexDefault-Token: Returning KW_DO\n")
				return KW_DO
			case "FOR":
				fmt.Printf("  lexDefault-Token: Returning KW_FOR\n")
				return KW_FOR
			case "EACH":
				fmt.Printf("  lexDefault-Token: Returning KW_EACH\n")
				return KW_EACH
			case "IN":
				fmt.Printf("  lexDefault-Token: Returning KW_IN\n")
				return KW_IN
			case "TOOL":
				fmt.Printf("  lexDefault-Token: Returning KW_TOOL\n")
				return KW_TOOL
			case "LLM":
				fmt.Printf("  lexDefault-Token: Returning KW_LLM\n")
				return KW_LLM
			case "ELSE":
				fmt.Printf("  lexDefault-Token: Returning KW_ELSE\n")
				return KW_ELSE
			default:
				lval.str = segment
				fmt.Printf("  lexDefault-Token: Returning IDENTIFIER (%s)\n", segment)
				return IDENTIFIER
			}
		}

		// String Literals
		if char == '"' || char == '\'' {
			start := l.pos
			quote := char
			l.pos++
			escaped := false
			foundEndQuote := false
			for l.pos < len(l.input) {
				curr := rune(l.input[l.pos])
				if escaped {
					escaped = false
				} else if curr == '\\' {
					escaped = true
				} else if curr == quote {
					l.pos++
					lval.str = l.input[start:l.pos]
					foundEndQuote = true
					break
				} else if curr == '\n' {
					l.line++
				} // Allow multiline strings
				l.pos++
			}
			if foundEndQuote {
				fmt.Printf("  lexDefault-Token: Returning STRING_LIT (%s)\n", lval.str)
				return STRING_LIT
			}
			l.pos = start
			l.Error(fmt.Sprintf("unclosed string literal starting with %c", quote))
			return INVALID
		}

		// Unexpected Character
		l.Error(fmt.Sprintf("unexpected character: %q", char))
		l.pos++                                               // Consume to avoid infinite loop
		fmt.Printf("  lexDefault-Token: Returning INVALID\n") // DEBUG
		return INVALID

	} // End outer loop
} // End lexDefault
