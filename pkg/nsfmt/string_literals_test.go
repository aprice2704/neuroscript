// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Explicitly tests string literal formatting, including the whitespace-only exception.
// filename: pkg/nsfmt/string_literals_test.go
// nlines: 96

package nsfmt

import (
	"fmt"
	"strings"
	"testing"
)

// Helper to construct inputs without fighting Go's string escaping
func wrapInFunc(content string) string {
	return fmt.Sprintf("\nfunc main() means\n    set s = %s\nendfunc\n", content)
}

func TestFormat_Strings_Standard(t *testing.T) {
	// 1. Standard simple string
	input := wrapInFunc(`"hello world"`)
	expected := wrapInFunc(`"hello world"`)
	assertFormat(t, input, expected)
}

func TestFormat_Strings_EscapedChars(t *testing.T) {
	// 2. String with escaped characters (tabs, quotes) BUT NO NEWLINES
	// Input in NS: "tab\tquote\""
	input := wrapInFunc(`"tab\tquote\""`)
	expected := wrapInFunc(`"tab\tquote\""`)
	assertFormat(t, input, expected)
}

func TestFormat_Strings_ToTripleBacktick(t *testing.T) {
	// 3. String with internal newline -> Should become triple backtick
	// We must use "input" that the parser accepts.
	// The parser accepts standard strings with \n escapes.
	input := wrapInFunc(`"line1\nline2"`)

	// The formatter should see the newline in the value and upgrade it.
	expected := wrapInFunc("```line1\nline2```")
	assertFormat(t, input, expected)
}

func TestFormat_Strings_PureWhitespace_Exception(t *testing.T) {
	// 4. Pure whitespace string with newlines -> Should REMAIN double quotes
	// This prevents "\n" from exploding into a 3-line backtick block.
	input := wrapInFunc(`"\n"`)
	expected := wrapInFunc(`"\n"`)
	assertFormat(t, input, expected)

	input2 := wrapInFunc(`"\n\n"`)
	expected2 := wrapInFunc(`"\n\n"`)
	assertFormat(t, input2, expected2)
}

func TestFormat_Strings_AlreadyTripleBacktick(t *testing.T) {
	// 5. Input is already triple backtick.
	// We simulate this by constructing the input with backticks.
	input := wrapInFunc("```raw\nstring```")
	expected := wrapInFunc("```raw\nstring```")
	assertFormat(t, input, expected)
}

func TestFormat_Strings_BacktickInside(t *testing.T) {
	// 6. String contains a backtick, but not triple.
	// Should be safe to use triple backticks if it has newlines.

	// FIX: Removed invalid backslash escape before the backtick.
	// Input represents: "has ` backtick\nand newline"
	input := wrapInFunc(`"has ` + "`" + ` backtick\nand newline"`)

	expected := wrapInFunc("```has ` backtick\nand newline```")
	assertFormat(t, input, expected)
}

func TestFormat_Strings_TripleBacktickInside_Fallback(t *testing.T) {
	// 7. String contains "```" inside.
	// Even if it has newlines, we CANNOT use triple backtick wrapping.
	// It must fallback to double quotes with escaping.

	// Construct input: "has ``` triple and \n newline"
	input := wrapInFunc(`"has ` + "```" + ` triple and \n newline"`)

	// Expectation: It stays as double quotes, with \n escaped.
	// Note: The formatter (TestString) typically outputs %q.
	expected := wrapInFunc(`"has ` + "```" + ` triple and \n newline"`)

	assertFormat(t, input, expected)
}

func TestFormat_Strings_ComplexPrompt(t *testing.T) {
	// 8. A realistic "Prompt" scenario
	rawPrompt := `You are a helper.
Please assist.
`
	// Input uses \n escapes
	inputVal := strings.ReplaceAll(rawPrompt, "\n", "\\n")
	input := wrapInFunc(fmt.Sprintf(`"%s"`, inputVal))

	// Output uses raw backticks
	expected := wrapInFunc(fmt.Sprintf("```%s```", rawPrompt))

	assertFormat(t, input, expected)
}
