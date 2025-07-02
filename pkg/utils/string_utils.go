// NeuroScript Version: 0.4.0
// File version: 0.1.3 // Added handling for line continuation '\' + newline/CR
// Purpose: Provides string un-escaping for NeuroScript literals and common string escaping utilities.
// filename: pkg/utils/string_utils.go
// nlines: 171 // Estimated
// risk_rating: MEDIUM

package utils

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"strconv"
	"strings"
	"unicode/utf16"
)

// UnescapeNeuroScriptString processes a raw string from a NeuroScript literal
// (the content between the quotes) and resolves escape sequences.
// It handles:
// - Line continuations: '\' followed by an actual newline or carriage return.
// - Standard escapes: \b, \t, \n, \f, \r, \v, \~, \`
// - Quote escapes: \", \', \\
// - Unicode escapes: \uXXXX
func UnescapeNeuroScriptString(rawString string) (string, error) {
	var sb strings.Builder
	reader := strings.NewReader(rawString)

	for reader.Len() > 0 {
		char, _, err := reader.ReadRune()
		if err != nil {
			return "", fmt.Errorf("error reading string for un-escaping: %w", err)
		}

		if char != '\\' {
			sb.WriteRune(char)
			continue
		}

		// At this point, 'char' is '\'. Read the next character.
		if reader.Len() == 0 {
			return "", fmt.Errorf("string ends with a bare backslash")
		}
		escChar, _, err := reader.ReadRune()
		if err != nil {
			return "", fmt.Errorf("error reading character after backslash: %w", err)
		}

		// Handle actual newline or carriage return characters following a backslash,
		// which signifies a line continuation from the lexer's CONTINUED_LINE fragment.
		if escChar == '\n' || escChar == '\r' {
			// This is a line continuation. The '\' and the newline/CR
			// (which is escChar) have been consumed from the reader.
			// We append nothing to the string builder, effectively joining the lines.
			// The loop will then continue with the character that was after the newline/CR.
			continue	// Continue the "for reader.Len() > 0" loop
		}

		// If it wasn't a line continuation, process as a standard escape sequence:
		switch escChar {
		case 'b':
			sb.WriteRune('\b')
		case 't':
			sb.WriteRune('\t')
		case 'n':	// This is for the escape sequence literal "\n"
			sb.WriteRune('\n')
		case 'f':
			sb.WriteRune('\f')
		case 'r':	// This is for the escape sequence literal "\r"
			sb.WriteRune('\r')
		case 'v':
			sb.WriteRune('\v')
		case '~':
			sb.WriteRune('~')
		case '`':
			sb.WriteRune('`')
		case '"':
			sb.WriteRune('"')
		case '\'':
			sb.WriteRune('\'')
		case '\\':
			sb.WriteRune('\\')
		case 'u':
			hexSequence := make([]rune, 4)
			for i := 0; i < 4; i++ {
				if reader.Len() == 0 {
					return "", fmt.Errorf("incomplete unicode escape sequence \\u%s", string(hexSequence[:i]))
				}
				r, _, readErr := reader.ReadRune()
				if readErr != nil {
					return "", fmt.Errorf("error reading unicode escape sequence: %w", readErr)
				}
				hexSequence[i] = r
			}
			val, parseErr := strconv.ParseUint(string(hexSequence), 16, 32)
			if parseErr != nil {
				return "", fmt.Errorf("invalid unicode escape sequence \\u%s: %w", string(hexSequence), parseErr)
			}

			r1 := rune(val)
			isR1HighSurrogate := (r1 >= 0xD800 && r1 <= 0xDBFF)

			if isR1HighSurrogate && reader.Len() >= 6 {
				var peekedRunes [6]rune
				var bytesReadForPeek int64 = 0
				var actualRunesReadForPeek int = 0
				var successfullyPeeked6Runes bool = true

				for i := 0; i < 6; i++ {
					if reader.Len() == 0 {
						successfullyPeeked6Runes = false
						break
					}
					r, size, readErr := reader.ReadRune()
					if readErr != nil {
						successfullyPeeked6Runes = false
						if bytesReadForPeek > 0 {
							_, seekErr := reader.Seek(-bytesReadForPeek, io.SeekCurrent)
							if seekErr != nil {
								return "", fmt.Errorf("failed to rollback peeked runes after read error: %w", seekErr)
							}
						}
						break
					}
					peekedRunes[i] = r
					bytesReadForPeek += int64(size)
					actualRunesReadForPeek++
				}

				if successfullyPeeked6Runes && actualRunesReadForPeek == 6 && peekedRunes[0] == '\\' && peekedRunes[1] == 'u' {
					lowSurrogateHex := string(peekedRunes[2:])
					val2, err2 := strconv.ParseUint(lowSurrogateHex, 16, 32)
					if err2 == nil {
						r2 := rune(val2)
						isR2LowSurrogate := (r2 >= 0xDC00 && r2 <= 0xDFFF)
						if isR2LowSurrogate {	// r1 is already known to be a high surrogate here
							combinedRune := utf16.DecodeRune(r1, r2)
							sb.WriteRune(combinedRune)
							continue	// Continue the main "for" loop
						}
					}
				}
				// If surrogate pair wasn't formed, roll back the peeked runes
				if bytesReadForPeek > 0 {
					_, seekErr := reader.Seek(-bytesReadForPeek, io.SeekCurrent)
					if seekErr != nil {
						return "", fmt.Errorf("failed to rollback peeked runes: %w", seekErr)
					}
				}
			}
			sb.WriteRune(r1)	// Write the original rune (high surrogate or regular BMP)
		default:
			return "", fmt.Errorf("unknown escape sequence: \\%c", escChar)
		}
	}
	return sb.String(), nil
}

// SafeEscapeHTML prepares a string for safe embedding in HTML by escaping special characters.
// It wraps Go's standard html.EscapeString.
func SafeEscapeHTML(s string) string {
	return html.EscapeString(s)
}

// SafeEscapeJavaScriptString prepares a string for safe embedding within a JavaScript string literal.
func SafeEscapeJavaScriptString(s string) (string, error) {
	marshaled, err := json.Marshal(s)
	if err != nil {
		return "", fmt.Errorf("failed to marshal string for JavaScript escaping: %w", err)
	}
	// json.Marshal returns a JSON string, which is double-quoted.
	// For direct embedding in JS code as a string literal, this is usually correct.
	return string(marshaled), nil
}

// StripJSONStringQuotes removes the outer quotes from a JSON-encoded string literal
// and unescapes its content. This is useful if you have a string that is already
// a valid JSON string literal (e.g., from `SafeEscapeJavaScriptString` or other JSON source)
// and you want the raw Go string value.
func StripJSONStringQuotes(jsonStr string) (string, error) {
	if len(jsonStr) < 2 || jsonStr[0] != '"' || jsonStr[len(jsonStr)-1] != '"' {
		return "", fmt.Errorf("string is not a valid JSON-encoded string literal: %s", jsonStr)
	}
	// strconv.Unquote handles JSON string unescaping.
	return strconv.Unquote(jsonStr)
}