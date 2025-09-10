// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Corrected the invalid base64 padding test case.
// filename: pkg/tool/strtools/tools_string_codecs_test.go
// nlines: 95
// risk_rating: LOW

package strtools

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestToolStringCodecs(t *testing.T) {
	interp := interpreter.NewInterpreter()
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		// Base64
		{name: "ToBase64 Simple", toolName: "ToBase64", args: MakeArgs("hello world"), wantResult: "aGVsbG8gd29ybGQ="},
		{name: "ToBase64 Empty", toolName: "ToBase64", args: MakeArgs(""), wantResult: ""},
		{name: "FromBase64 Simple", toolName: "FromBase64", args: MakeArgs("aGVsbG8gd29ybGQ="), wantResult: "hello world"},
		{name: "FromBase64 Empty", toolName: "FromBase64", args: MakeArgs(""), wantResult: ""},
		{name: "FromBase64 Invalid Chars", toolName: "FromBase64", args: MakeArgs("aGVsbG8gd29ybGQ"), wantErrIs: lang.ErrInvalidArgument},
		{name: "FromBase64 Missing Padding", toolName: "FromBase64", args: MakeArgs("aGVsbG8"), wantErrIs: lang.ErrInvalidArgument},
		{name: "ToBase64 Wrong Type", toolName: "ToBase64", args: MakeArgs(123), wantErrIs: lang.ErrArgumentMismatch},

		// Hex
		{name: "ToHex Simple", toolName: "ToHex", args: MakeArgs("hello"), wantResult: "68656c6c6f"},
		{name: "ToHex Empty", toolName: "ToHex", args: MakeArgs(""), wantResult: ""},
		{name: "FromHex Simple", toolName: "FromHex", args: MakeArgs("68656c6c6f"), wantResult: "hello"},
		{name: "FromHex Empty", toolName: "FromHex", args: MakeArgs(""), wantResult: ""},
		{name: "FromHex Invalid Chars", toolName: "FromHex", args: MakeArgs("68656c6c6g"), wantErrIs: lang.ErrInvalidArgument},
		{name: "FromHex Invalid Length", toolName: "FromHex", args: MakeArgs("68656c6c6"), wantErrIs: lang.ErrInvalidArgument},
		{name: "ToHex Wrong Type", toolName: "ToHex", args: MakeArgs(true), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, tt)
	}
}

func TestToolStringCompression(t *testing.T) {
	interp := interpreter.NewInterpreter()

	// The exact compressed output can vary, so we compress and then decompress to test for integrity.
	originalText := "this is some text that will be compressed and then decompressed. it has repetition. it has repetition."

	// Manually run compression to get a valid input for decompress test
	compressedResult, err := toolStringCompress(interp, MakeArgs(originalText))
	if err != nil {
		t.Fatalf("Compression failed during test setup: %v", err)
	}
	compressedString, ok := compressedResult.(string)
	if !ok {
		t.Fatalf("Expected compressed result to be a string, but got %T", compressedResult)
	}

	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		// Compress & Decompress Cycle
		{
			name:       "Compress then Decompress",
			toolName:   "Decompress",
			args:       MakeArgs(compressedString),
			wantResult: originalText,
		},
		{
			name:      "Decompress Empty String (invalid)",
			toolName:  "Decompress",
			args:      MakeArgs(""),
			wantErrIs: lang.ErrInvalidArgument, // Empty string is not valid base64 or gzip
		},
		{
			name:       "Compress Empty String",
			toolName:   "Compress",
			args:       MakeArgs(""),
			wantResult: "H4sIAAAAAAAA/wEAAP//AAAAAAAAAAA=", // Predictable output for empty string
		},
		{
			name:      "Decompress Invalid Base64",
			toolName:  "Decompress",
			args:      MakeArgs("not_base64_string?"),
			wantErrIs: lang.ErrInvalidArgument,
		},
		{
			name:      "Decompress Valid Base64 but Invalid Gzip",
			toolName:  "Decompress",
			args:      MakeArgs("aGVsbG8="), // "hello" in base64, not valid gzip
			wantErrIs: lang.ErrInvalidArgument,
		},
		{
			name:      "Compress Wrong Type",
			toolName:  "Compress",
			args:      MakeArgs(12345),
			wantErrIs: lang.ErrArgumentMismatch,
		},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, tt)
	}
}
