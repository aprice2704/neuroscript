// filename: pkg/core/tools_string_utils_test.go
package core

import ( // Keep errors
	// Keep filepath
	// "reflect" // Remove reflect
	// Keep strings for error check
	"testing"
)

// Remove the duplicate testFsToolHelper function definition here
/*
func testFsToolHelper(...) {
    ... // REMOVED
}
*/

// Assuming testStringToolHelper exists (defined in tools_string_basic_test.go or testing_helpers_test.go)

func TestToolLineCountString(t *testing.T) {
	interp, _ := newDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		// *** FIXED toolName prefix ***
		{name: "Empty String", toolName: "LineCountString", args: makeArgs(""), wantResult: int64(0)},
		{name: "Single Line No NL", toolName: "LineCountString", args: makeArgs("hello"), wantResult: int64(1)},
		{name: "Single Line With NL", toolName: "LineCountString", args: makeArgs("hello\n"), wantResult: int64(1)},
		{name: "Two Lines No Trailing NL", toolName: "LineCountString", args: makeArgs("hello\nworld"), wantResult: int64(2)},
		{name: "Two Lines With Trailing NL", toolName: "LineCountString", args: makeArgs("hello\nworld\n"), wantResult: int64(2)},
		{name: "Multiple Blank Lines", toolName: "LineCountString", args: makeArgs("\n\n\n"), wantResult: int64(3)}, // Changed expectation, blank lines count
		{name: "Mixed Content and Blank", toolName: "LineCountString", args: makeArgs("line1\n\nline3\n"), wantResult: int64(3)},
		{name: "CRLF Line Endings", toolName: "LineCountString", args: makeArgs("line1\r\nline2\r\n"), wantResult: int64(2)}, // Should handle CRLF
		{name: "Validation Wrong Arg Type", toolName: "LineCountString", args: makeArgs(123), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Validation Nil Arg", toolName: "LineCountString", args: makeArgs(nil), valWantErrIs: ErrValidationRequiredArgNil},
	}
	for _, tt := range tests {
		// *** Use the correct helper ***
		// Assuming testStringToolHelper is defined elsewhere (e.g., tools_string_basic_test.go)
		// and handles the necessary validation and execution checks.
		testStringToolHelper(t, interp, tt)
	}
}

// Add tests for other utils if they exist, e.g., HasPrefix, HasSuffix, Contains (these might be in basic_test?)
// Assuming only LineCountString is in this file for now based on previous error output.
