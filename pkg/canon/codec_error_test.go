// NeuroScript Version: 0.6.3
// File version: 3
// Purpose: Tests the decoder's resilience against malformed and corrupted input data using sentinel errors.
// filename: pkg/canon/codec_error_test.go
// nlines: 70
// risk_rating: LOW

package canon

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

func TestDecodeWithRegistry_ErrorHandling(t *testing.T) {
	// 1. Get a valid blob to mutate.
	script := `func main() means; set x = 1; endfunc`
	parserAPI := parser.NewParserAPI(logging.NewNoOpLogger())
	antlrTree, _ := parserAPI.Parse(script)
	builder := parser.NewASTBuilder(logging.NewNoOpLogger())
	program, _, _ := builder.Build(antlrTree)
	validBlob, _, _ := CanonicaliseWithRegistry(&ast.Tree{Root: program})

	testCases := []struct {
		name          string
		input         []byte
		expectedError error
	}{
		{
			name:          "Nil blob",
			input:         nil,
			expectedError: ErrInvalidMagic,
		},
		{
			name:          "Empty blob",
			input:         []byte{},
			expectedError: ErrInvalidMagic,
		},
		{
			name:          "Invalid magic number",
			input:         []byte{'B', 'A', 'D', ' '},
			expectedError: ErrInvalidMagic,
		},
		{
			name:          "Truncated blob after magic number",
			input:         magicNumber,
			expectedError: ErrTruncatedData,
		},
		{
			name: "Truncated blob mid-stream",
			// Truncate after the first few bytes of the program node
			input:         validBlob[:len(validBlob)-10],
			expectedError: ErrTruncatedData,
		},
		{
			name: "Corrupted kind value",
			// Overwrite a valid kind with an invalid one
			input: func() []byte {
				badBlob := append([]byte(nil), validBlob...)
				// This replaces the KindProgram varint with an invalid one (254).
				badBlob[len(magicNumber)] = 0xFE
				badBlob[len(magicNumber)+1] = 0x01
				return badBlob
			}(),
			expectedError: ErrUnknownCodec,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := DecodeWithRegistry(tc.input)
			if err == nil {
				t.Fatal("expected an error but got nil")
			}
			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expected error to be %v, but got: %v", tc.expectedError, err)
			}
		})
	}
}

// Add this to the existing codec_error_test.go file

func FuzzDecodeWithRegistry(f *testing.F) {
	// 1. Add a few valid seed inputs.
	scriptPath := filepath.Join("..", "antlr", "comprehensive_grammar.ns")
	scriptBytes, err := os.ReadFile(scriptPath)
	if err != nil {
		f.Fatalf("Failed to read script file: %v", err)
	}
	parserAPI := parser.NewParserAPI(logging.NewNoOpLogger())
	antlrTree, _ := parserAPI.Parse(string(scriptBytes))
	builder := parser.NewASTBuilder(logging.NewNoOpLogger())
	program, _, _ := builder.Build(antlrTree)
	validBlob, _, _ := CanonicaliseWithRegistry(&ast.Tree{Root: program})
	f.Add(validBlob)

	f.Add(magicNumber) // A known bad case
	f.Add([]byte{})    // Another bad case

	// 2. The fuzzer will now mutate these inputs.
	f.Fuzz(func(t *testing.T, data []byte) {
		// The test simply decodes the data. The fuzzer's goal is to
		// find an input `data` that causes a panic. If it panics, the
		// test will fail and report the problematic input.
		// We don't need to check the error, as we already have unit
		// tests for specific error conditions.
		_, _ = DecodeWithRegistry(data)
	})
}
