// NeuroScript Version: 0.4.0
// File version: 0.1.4 // Updated error check to use errors.Is with strconv.ErrSyntax
// Purpose: Tests for string utility functions in string_utils.go.
// filename: pkg/utils/string_utils_test.go
// nlines: 223 // Approximate
// risk_rating: LOW

package utils

import (
	"errors"	// For errors.Is
	// For t.Errorf
	"strconv"	// For strconv.ErrSyntax
	"strings"	// For strings.Contains
	"testing"	// For testing framework
)

func TestUnescapeNeuroScriptString(t *testing.T) {
	tests := []struct {
		name		string
		input		string
		want		string
		wantErr		bool
		errContains	string
	}{
		// Basic escapes
		{"empty string", "", "", false, ""},
		{"no escapes", "hello world", "hello world", false, ""},
		{"newline", "hello\\nworld", "hello\nworld", false, ""},
		{"tab", "hello\\tworld", "hello\tworld", false, ""},
		{"carriage return", "hello\\rworld", "hello\rworld", false, ""},
		{"backslash", "hello\\\\world", "hello\\world", false, ""},
		{"double quote", "hello\\\"world\\\"", "hello\"world\"", false, ""},
		{"single quote", "hello\\'world\\'", "hello'world'", false, ""},
		{"backtick", "hello\\`world\\`", "hello`world`", false, ""},
		{"tilde", "hello\\~world\\~", "hello~world~", false, ""},
		{"form feed", "hello\\fworld", "hello\fworld", false, ""},
		{"vertical tab", "hello\\vworld", "hello\vworld", false, ""},
		{"backspace", "hello\\bworld", "hello\bworld", false, ""},
		{"multiple basic", "\\n\\t\\\\\\\"\\'\\`\\~\\f\\v\\b", "\n\t\\\"'`~\f\v\b", false, ""},

		// Unicode escapes
		{"unicode A", "\\u0041", "A", false, ""},
		{"unicode Omega", "\\u03A9", "Ω", false, ""},
		{"unicode Heart", "\\u2764", "❤", false, ""},
		{"unicode mixed", "Value: \\u0041-\\u03A9", "Value: A-Ω", false, ""},
		{"unicode emoji grinning face", "\\uD83D\\uDE00", "😀", false, ""},	// U+1F600
		{"unicode emoji laughing face", "\\uD83D\\uDE04", "😄", false, ""},	// U+1F604
		{"unicode high surrogate only", "\\uD83D", string(rune(0xD83D)), false, ""},
		{"unicode high surrogate then text", "\\uD83Dhello", string(rune(0xD83D)) + "hello", false, ""},
		{"unicode high surrogate then incomplete low surrogate 1", "\\uD83D\\u", string(rune(0xD83D)) + "\\u", true, "incomplete unicode"},
		{"unicode high surrogate then incomplete low surrogate 2", "\\uD83D\\uDE0", string(rune(0xD83D)) + "\\uDE0", true, "incomplete unicode"},
		{"unicode high surrogate then invalid low surrogate hex", "\\uD83D\\uXXXX", string(rune(0xD83D)) + "\\uXXXX", true, "invalid unicode"},
		{"unicode high surrogate not followed by \\u", "\\uD83DDE04", string(rune(0xD83D)) + "DE04", false, ""},

		// Error cases
		{"trailing backslash", "abc\\", "", true, "ends with a bare backslash"},
		{"incomplete unicode short", "abc\\u12", "", true, "incomplete unicode escape sequence"},
		{"incomplete unicode mid", "abc\\u004", "", true, "incomplete unicode escape sequence"},
		{"invalid unicode hex", "abc\\uXXYY", "", true, "invalid unicode escape sequence"},
		{"unknown escape", "abc\\z", "", true, "unknown escape sequence: \\z"},
		{"unknown escape at end", "\\z", "", true, "unknown escape sequence: \\z"},
		{"unicode incomplete in surrogate attempt", "hello\\uD83D\\u12", "", true, "incomplete unicode escape sequence \\u12"},
		{"unicode invalid hex in surrogate attempt", "hello\\uD83D\\uXXXX", "", true, "invalid unicode escape sequence \\uXXXX"},

		// Mixed cases
		{"text and escapes", "first\\nsecond\\tthird\\\\fourth", "first\nsecond\tthird\\fourth", false, ""},
		{"escapes and unicode", "\\t\\u0041\\n", "\tA\n", false, ""},
		{"complex emoji sequence", "Emoji: \\uD83E\\uDD26\\u200D\\u2642\\uFE0F (facepalm)", "Emoji: 🤦‍♂️ (facepalm)", false, ""},
		{"multiple emojis", "\\uD83D\\uDE00\\uD83D\\uDE04", "😀😄", false, ""},
		{"emoji then text", "\\uD83D\\uDE00 hello", "😀 hello", false, ""},
		{"text then emoji", "hello \\uD83D\\uDE00", "hello 😀", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnescapeNeuroScriptString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnescapeNeuroScriptString(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil {	// Should not happen if wantErr is true and previous check passed
					t.Errorf("UnescapeNeuroScriptString(%q) expected error but got nil (errContains: %q)", tt.input, tt.errContains)
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("UnescapeNeuroScriptString(%q) error = %q, want err containing %q", tt.input, err.Error(), tt.errContains)
				}
				// If tt.errContains is empty, any error is fine as long as err != nil.
				return
			}
			if got != tt.want {
				t.Errorf("UnescapeNeuroScriptString(%q) = %q (bytes: %v), want %q (bytes: %v)",
					tt.input, got, []byte(got), tt.want, []byte(tt.want))
			}
		})
	}
}

func TestSafeEscapeHTML(t *testing.T) {
	tests := []struct {
		name	string
		input	string
		want	string
	}{
		{"empty", "", ""},
		{"no special chars", "hello world", "hello world"},
		{"simple html", "<p>Hello</p>", "&lt;p&gt;Hello&lt;/p&gt;"},
		{"quotes and ampersand", "He said \"Hi & Welcome!\"", "He said &#34;Hi &amp; Welcome!&#34;"},
		{"single quote", "It's mine.", "It&#39;s mine."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SafeEscapeHTML(tt.input); got != tt.want {
				t.Errorf("SafeEscapeHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeEscapeJavaScriptString(t *testing.T) {
	tests := []struct {
		name	string
		input	string
		want	string
		wantErr	bool
	}{
		{"empty", "", "\"\"", false},
		{"simple string", "hello", "\"hello\"", false},
		{"with quotes", "hello \"world\"", "\"hello \\\"world\\\"\"", false},
		{"with newline", "hello\nworld", "\"hello\\nworld\"", false},
		{"with backslash", "hello\\world", "\"hello\\\\world\"", false},
		{"all together", "a\"b'c\\d\ne\rf\tg", "\"a\\\"b'c\\\\d\\ne\\rf\\tg\"", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeEscapeJavaScriptString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeEscapeJavaScriptString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SafeEscapeJavaScriptString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStripJSONStringQuotes(t *testing.T) {
	tests := []struct {
		name		string
		input		string
		want		string
		wantErr		bool
		errContains	string	// Used for message check, especially if not checking a specific sentinel
	}{
		{"valid simple", "\"hello\"", "hello", false, ""},
		{"valid with newline", "\"hello\\nworld\"", "hello\nworld", false, ""},
		{"valid with quote", "\"hello \\\"world\\\"\"", "hello \"world\"", false, ""},
		{"valid with backslash", "\"hello\\\\world\"", "hello\\world", false, ""},
		{"empty content", "\"\"", "", false, ""},
		{"error not a string", "hello", "", true, "not a valid JSON-encoded string literal"},
		{"error no leading quote", "hello\"", "", true, "not a valid JSON-encoded string literal"},
		{"error no trailing quote", "\"hello", "", true, "not a valid JSON-encoded string literal"},
		{"error only one quote", "\"", "", true, "not a valid JSON-encoded string literal"},
		// This case now expects strconv.ErrSyntax, whose message is "invalid syntax".
		{"error invalid internal escape", "\"hello\\xworld\"", "", true, "invalid syntax"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StripJSONStringQuotes(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("StripJSONStringQuotes(%q): error presence mismatch, got err %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {	// Should be caught by the check above
					t.Errorf("StripJSONStringQuotes(%q): expected error but got nil", tt.input)
				} else {
					// Specific check for the "error invalid internal escape" test case using errors.Is
					if tt.name == "error invalid internal escape" {
						if !errors.Is(err, strconv.ErrSyntax) {
							t.Errorf("StripJSONStringQuotes(%q): expected error to be strconv.ErrSyntax, but got %v (type %T)",
								tt.input, err, err)
						}
						// Optionally, also check the message if errContains is set and matches strconv.ErrSyntax.Error()
						// This is a secondary check; the errors.Is is primary for the sentinel.
						if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
							t.Errorf("StripJSONStringQuotes(%q): for strconv.ErrSyntax, error message %q does not contain %q",
								tt.input, err.Error(), tt.errContains)
						}
					} else if tt.errContains != "" {	// Fallback for other expected error cases (check message substring)
						if !strings.Contains(err.Error(), tt.errContains) {
							t.Errorf("StripJSONStringQuotes(%q): error message %q does not contain %q",
								tt.input, err.Error(), tt.errContains)
						}
					}
					// If tt.errContains is empty and not the special "error invalid internal escape" case,
					// any error is acceptable, and we've already confirmed err != nil.
				}
				return	// Done with error case
			}

			// If no error was wanted (tt.wantErr is false)
			if got != tt.want {
				t.Errorf("StripJSONStringQuotes(%q): got result %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}