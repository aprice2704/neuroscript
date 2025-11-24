// NeuroScript Version: 0.8.0
// File version: 8
// Purpose: BUGFIX: Updates TestFormat_String_MultiLine to expect valid triple-backtick syntax for multi-line output.
// filename: pkg/nsfmt/format_expr_test.go
// nlines: 187

package nsfmt

import (
	"testing"
)

func TestFormat_Expressions_Precedence(t *testing.T) {
	input := `
func main() means
    set x = 1 + 2 * 3
    set y = (1 + 2) * 3
    set z = "a" + "b"
    set w = a or b and c
    set chaining = a + b + c
    set grouping = a - (b - c)
endfunc
`
	expected := `
func main() means
    set x = 1 + 2 * 3
    set y = (1 + 2) * 3
    set z = "a" + "b"
    set w = a or b and c
    set chaining = a + b + c
    set grouping = a - (b - c)
endfunc`
	assertFormat(t, input, expected)
}

func TestFormat_String_MultiLine(t *testing.T) {
	// The parser currently requires valid input (no raw newlines in strings).
	// We verify that nsfmt converts the valid escaped input into
	// TRIPLE BACKTICK strings for readability.
	input := `
func main() means
    set prompt = "You are a helper.\nPlease assist."
endfunc
`
	expected := `
func main() means
    set prompt = ` + "```" + `You are a helper.
Please assist.` + "```" + `
endfunc`
	assertFormat(t, input, expected)
}

func TestFormat_MultiLineMap_Smart(t *testing.T) {
	// This map is long and should be broken into multiple lines.
	input := `
func main() means
    set my_map = { "key1": "this is a very long value that should definitely exceed the line limit", "key2": 123, "key3": [1, 2, 3] }
endfunc
`
	// FIX: Removed trailing comma
	expected := `
func main() means
    set my_map = { \
        "key1": "this is a very long value that should definitely exceed the line limit", \
        "key2": 123, \
        "key3": [1, 2, 3] \
    }
endfunc`
	assertFormat(t, input, expected)
}

func TestFormat_ShortList_SingleLine(t *testing.T) {
	// This list is short and should stay on a single line.
	input := `
func main() means
    set my_list = [1, "two", 3]
endfunc
`
	expected := `
func main() means
    set my_list = [1, "two", 3]
endfunc`
	assertFormat(t, input, expected)
}

func TestFormat_LongList_MultiLine(t *testing.T) {
	// This list is long and should be broken up.
	input := `
func main() means
    set my_list = [1, "two", 3, "four", "five", "six", "seven", "eight", "nine", "ten"]
endfunc
`
	// FIX: Removed trailing comma
	expected := `
func main() means
    set my_list = [ \
        1, \
        "two", \
        3, \
        "four", \
        "five", \
        "six", \
        "seven", \
        "eight", \
        "nine", \
        "ten" \
    ]
endfunc`
	assertFormat(t, input, expected)
}

func TestFormat_LongCall_MultiLine(t *testing.T) {
	// This call is long and should be broken up.
	input := `
func main() means
    call my_long_function("arg1", "arg2", "this is a very long argument", "arg4", "arg5")
endfunc
`
	// Note: No trailing comma
	expected := `
func main() means
    call my_long_function( \
        "arg1", \
        "arg2", \
        "this is a very long argument", \
        "arg4", \
        "arg5" \
    )
endfunc`
	assertFormat(t, input, expected)
}

func TestFormat_UnaryNotOperator(t *testing.T) {
	// This tests the 'not' operator bug
	input := `
func main() means
    if not my_func(true)
        emit "was false"
    endif
    set x = -1
endfunc
`
	expected := `
func main() means
    if not my_func(true)
        emit "was false"
    endif
    set x = -1
endfunc`
	assertFormat(t, input, expected)
}

func TestFormat_UnaryWordOperators(t *testing.T) {
	// This tests the 'not', 'some', and 'no' operator bug
	input := `
func main() means
    if not my_func(true)
        emit "was false"
    endif
    if some script_capsule and some charter_capsule
        emit "both present"
    endif
    if no (a or b)
        emit "neither"
    endif
endfunc
`
	expected := `
func main() means
    if not my_func(true)
        emit "was false"
    endif
    if some script_capsule and some charter_capsule
        emit "both present"
    endif
    if no (a or b)
        emit "neither"
    endif
endfunc`
	assertFormat(t, input, expected)
}
