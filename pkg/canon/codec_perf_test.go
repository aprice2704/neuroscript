// NeuroScript Version: 0.6.3
// File version: 1
// Purpose: Adds benchmark tests for the canonicalization and decoding processes.
// filename: pkg/canon/codec_perf_test.go
// nlines: 50
// risk_rating: LOW

package canon

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// BenchmarkCanonicalise measures the performance of the full encoding process.
func BenchmarkCanonicalise(b *testing.B) {
	scriptPath := filepath.Join("..", "antlr", "comprehensive_grammar.ns")
	scriptBytes, err := os.ReadFile(scriptPath)
	if err != nil {
		b.Fatalf("Failed to read script file: %v", err)
	}

	parserAPI := parser.NewParserAPI(logging.NewNoOpLogger())
	antlrTree, _ := parserAPI.Parse(string(scriptBytes))
	builder := parser.NewASTBuilder(logging.NewNoOpLogger())
	program, _, _ := builder.Build(antlrTree)
	tree := &ast.Tree{Root: program}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, err := CanonicaliseWithRegistry(tree)
		if err != nil {
			b.Fatalf("CanonicaliseWithRegistry failed: %v", err)
		}
	}
}

// BenchmarkDecode measures the performance of the full decoding process.
func BenchmarkDecode(b *testing.B) {
	scriptPath := filepath.Join("..", "antlr", "comprehensive_grammar.ns")
	scriptBytes, err := os.ReadFile(scriptPath)
	if err != nil {
		b.Fatalf("Failed to read script file: %v", err)
	}

	parserAPI := parser.NewParserAPI(logging.NewNoOpLogger())
	antlrTree, _ := parserAPI.Parse(string(scriptBytes))
	builder := parser.NewASTBuilder(logging.NewNoOpLogger())
	program, _, _ := builder.Build(antlrTree)
	tree := &ast.Tree{Root: program}
	blob, _, _ := CanonicaliseWithRegistry(tree)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := DecodeWithRegistry(blob)
		if err != nil {
			b.Fatalf("DecodeWithRegistry failed: %v", err)
		}
	}
}
