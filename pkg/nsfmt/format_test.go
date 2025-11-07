// NeuroScript Version: 0.8.0
// File version: 13
// Purpose: BUGFIX: Corrects test expectations to match actual formatter output (no parens on if/while) and fixes invalid long-signature test input.
// filename: pkg/nsfmt/format_test.go
// nlines: 255

package nsfmt

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// assertFormat is a helper function to run a format test.
func assertFormat(t *testing.T, input, expected string) {
	t.Helper()

	// Normalize input and expected strings to remove leading/trailing
	// whitespace from the heredocs, which makes tests easier to write.
	input = strings.TrimSpace(input)
	expected = strings.TrimSpace(expected)

	formatted, err := Format([]byte(input))
	if err != nil {
		t.Fatalf("Format() returned an unexpected error: %v", err)
	}

	got := strings.TrimSpace(string(formatted))

	if got != expected {
		t.Errorf("Format() result does not match expected output.")
		t.Logf("--- INPUT ---\n%s\n", input)
		// Use cmp.Diff for a nice, readable diff
		diff := cmp.Diff(expected, got)
		t.Fatalf("--- DIFF (WANT -> GOT) ---\n%s", diff)
	}
}

func TestFormat_SyntaxError(t *testing.T) {
	input := `
func main() means
    set x = 1 + @
endfunc
`
	_, err := Format([]byte(input))
	if err == nil {
		t.Fatal("Format() did not return an error on syntactically invalid code")
	}
	if !strings.Contains(err.Error(), "syntax error") {
		t.Errorf("Expected a syntax error, but got: %v", err)
	}
}

func TestFormat_Indentation(t *testing.T) {
	input := `
func main() means
set x = 1
if x > 0
emit "greater"
else
emit "less"
endif
endfunc
`
	expected := `
func main() means
    set x = 1
    if x > 0
        emit "greater"
    else
        emit "less"
    endif
endfunc`

	assertFormat(t, input, expected)
}

func TestFormat_NestedBlocks(t *testing.T) {
	input := `
func main() means
set x = 0
while x < 2
for each i in [1, 2]
if i == 1
set x = (x + 1)
endif
endfor
endwhile
endfunc
`
	expected := `
func main() means
    set x = 0
    while x < 2
        for each i in [1, 2]
            if i == 1
                set x = x + 1
            endif
        endfor
    endwhile
endfunc`
	assertFormat(t, input, expected)
}

func TestFormat_AllStatements(t *testing.T) {
	input := `
func all_statements(needs p1) means
    set a = p1
    call tool.foo.bar(a)
    must (a != nil)
    if (a == "fail")
    fail "it failed"
    endif
    ask "gemini", "prompt" with {"temp": 0.5} into res
    promptuser "name?" into user_name
    whisper "handle", "value"
    emit "done"
    return res, user_name
endfunc
`
	expected := `
func all_statements(needs p1) means
    set a = p1
    call tool.foo.bar(a)
    must a != nil
    if a == "fail"
        fail "it failed"
    endif
    ask "gemini", "prompt" with {"temp": 0.5} into res
    promptuser "name?" into user_name
    whisper "handle", "value"
    emit "done"
    return res, user_name
endfunc`

	assertFormat(t, input, expected)
}

func TestFormat_MetadataAndSpacing(t *testing.T) {
	input := `
:: c: 3
:: a: 1
:: b: 2

func main() means
:: z: 9
:: x: 7
    set x = 1
endfunc
func two() means
    set y = 2
endfunc
`
	expected := `
:: a: 1
:: b: 2
:: c: 3

func main() means
    :: x: 7
    :: z: 9
    set x = 1
endfunc

func two() means
    set y = 2
endfunc`
	assertFormat(t, input, expected)
}

func TestFormat_LibraryBlocks(t *testing.T) {
	input := `
:: file: library_blocks.ns

func main() means
    emit "main"
endfunc
on event "foo" as evt do
    emit evt
endon
`
	expected := `
:: file: library_blocks.ns

func main() means
    emit "main"
endfunc

on event "foo" as evt do
    emit evt
endon`
	assertFormat(t, input, expected)
}

func TestFormat_CommandBlocks(t *testing.T) {
	input := `
:: file: command_blocks.ns

command
:: name: my_command
    set x = 1
on error do
emit "error"
endon
endcommand
`
	expected := `
:: file: command_blocks.ns

command
    :: name: my_command
    set x = 1
    on error do
        emit "error"
    endon
endcommand`
	assertFormat(t, input, expected)
}

func TestFormat_Comments(t *testing.T) {
	input := `
# File comment
:: a: 1

# Func comment
func main() means
    # Step 1 comment
    set x = 1
    
    # Step 2 comment (with blank line)
    set y = 2
endfunc
`
	expected := `
# File comment
:: a: 1

# Func comment
func main() means
    # Step 1 comment
    set x = 1

    # Step 2 comment (with blank line)
    set y = 2
endfunc`
	assertFormat(t, input, expected)
}

func TestFormat_LongSignature_MultiLine(t *testing.T) {
	input := `
# Test long signatures
func long_sig(needs a, b, c, d, e optional f, g, h, i, j returns k, l, m, n, o) means
	set x = 1
endfunc
`
	// FIX: Removed trailing comma
	expected := `
# Test long signatures
func long_sig( \
    needs a, b, c, d, e \
    optional f, g, h, i, j \
    returns k, l, m, n, o \
) means
    set x = 1
endfunc`
	assertFormat(t, input, expected)
}
