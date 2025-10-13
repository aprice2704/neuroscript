// NeuroScript Version: 0.7.2
// File version: 8
// Purpose: Removes EquateEmpty workaround now that the parser and canon packages are consistent in producing non-nil empty collections.
// filename: pkg/parser/ast_canonicalization_roundtrip_test.go
// nlines: 73
// risk_rating: HIGH

package parser_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/canon"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// TestASTCanonicalizationRoundTrip performs a full serialization and deserialization
// of a complex AST to verify that the process is completely lossless. This test
// is a strong safeguard against bugs in the `canon` package where parts of the
// AST might be dropped or altered.
func TestASTCanonicalizationRoundTrip(t *testing.T) {
	script := `
:: title: Comprehensive Round-Trip Test
:: version: 1.0

func main(returns result) means
	set my_list = [1, "two", true, {"key": "value"}]
	set total = 0
	for each item in my_list
		if typeof(item) == "number"
			set total = total + item
		endif
	endfor
	call tool.Log(total)
	return total
endfunc

on event "system.startup" do
	emit "System is starting up!"
endon
`
	logger := logging.NewTestLogger(t)
	parserAPI := parser.NewParserAPI(logger)
	astBuilder := parser.NewASTBuilder(logger)

	// 1. Parse the source to get the original AST.
	antlrTree, _, pErr := parserAPI.ParseAndGetStream("roundtrip_test.ns", script)
	if pErr != nil {
		t.Fatalf("Parse() failed unexpectedly: %v", pErr)
	}
	originalProgram, _, bErr := astBuilder.Build(antlrTree)
	if bErr != nil {
		t.Fatalf("Build() failed unexpectedly: %v", bErr)
	}
	originalTree := &ast.Tree{Root: originalProgram}

	// 2. Canonicalize the original AST to its binary format.
	encoded, _, err := canon.CanonicaliseWithRegistry(originalTree)
	if err != nil {
		t.Fatalf("canon.CanonicaliseWithRegistry failed: %v", err)
	}

	// 3. Decode the binary format back into a new AST.
	decodedTree, err := canon.DecodeWithRegistry(encoded)
	if err != nil {
		t.Fatalf("canon.DecodeWithRegistry returned an error: %v", err)
	}

	// 4. Compare the original and decoded ASTs.
	cmpOpts := []cmp.Option{
		cmpopts.IgnoreFields(ast.BaseNode{}, "StartPos", "StopPos", "NodeKind"),
		cmp.AllowUnexported(ast.Procedure{}, ast.Step{}),
	}

	if diff := cmp.Diff(originalTree, decodedTree, cmpOpts...); diff != "" {
		t.Fatalf("AST mismatch after round-trip (-want +got):\n%s", diff)
	}
}
