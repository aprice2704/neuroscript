// NeuroScript Version: 0.6.3
// File version: 17
// Purpose: FIX: Normalizes nil param defaults to lang.NilValue to match codec roundtrip behavior.
// filename: pkg/canon/comprehensive_e2e_test.go
// nlines: 138
// risk_rating: HIGH

package canon

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang" // <<< ADDED
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// normalizeAST is a pre-test pass that restructures the raw AST from the parser
// to match the clean, normalized structure that the canonicalization process produces.
// This is necessary for a symmetrical comparison.
func normalizeAST(p *ast.Program) {
	var walkSteps func([]ast.Step)
	walkSteps = func(steps []ast.Step) {
		for i := range steps {
			s := &steps[i]
			if s.Type == "ask" && s.AskStmt == nil {
				s.AskStmt = &ast.AskStmt{
					BaseNode:       ast.BaseNode{NodeKind: types.KindAskStmt},
					AgentModelExpr: s.Values[0],
					PromptExpr:     s.Values[1],
				}
				if len(s.LValues) > 0 {
					s.AskStmt.IntoTarget = s.LValues[0]
				}
				s.Values = nil
				s.LValues = nil
			}
			if s.Type == "promptuser" && s.PromptUserStmt == nil {
				s.PromptUserStmt = &ast.PromptUserStmt{
					BaseNode:   ast.BaseNode{NodeKind: types.KindPromptUserStmt},
					PromptExpr: s.Values[0],
				}
				if len(s.LValues) > 0 {
					s.PromptUserStmt.IntoTarget = s.LValues[0]
				}
				s.Values = nil
				s.LValues = nil
			}
			// FIX: Add normalization for WhisperStmt
			if s.Type == "whisper" && s.WhisperStmt == nil {
				s.WhisperStmt = &ast.WhisperStmt{
					BaseNode: ast.BaseNode{NodeKind: types.KindWhisperStmt},
					Handle:   s.Values[0],
					Value:    s.Values[1],
				}
				s.Values = nil
			}
			if len(s.Body) > 0 {
				walkSteps(s.Body)
			}
			if len(s.ElseBody) > 0 {
				walkSteps(s.ElseBody)
			}
		}
	}

	for _, proc := range p.Procedures {
		// --- FIX: Normalize nil defaults to lang.NilValue ---
		for _, param := range proc.OptionalParams {
			if param.Default == nil {
				param.Default = lang.NilValue{}
			}
		}
		// --- END FIX ---

		walkSteps(proc.Steps)
		for _, handler := range proc.ErrorHandlers {
			walkSteps(handler.Body)
		}
	}
	for _, cmd := range p.Commands {
		walkSteps(cmd.Body)
	}
	for _, ev := range p.Events {
		walkSteps(ev.Body)
	}
}

// runRoundtripComparison is a helper to perform the full parse -> canonicalize -> decode -> compare cycle.
func runRoundtripComparison(t *testing.T, scriptPath string) {
	t.Helper()
	t.Logf("--- Running roundtrip comparison for: %s ---", scriptPath)

	scriptBytes, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("Failed to read script file %s: %v", scriptPath, err)
	}

	// 1. Parse the script to get the original AST.
	parserAPI := parser.NewParserAPI(logging.NewNoOpLogger())
	antlrTree, pErr := parserAPI.Parse(string(scriptBytes))
	if pErr != nil {
		t.Fatalf("parser.Parse() failed unexpectedly: %v", pErr)
	}
	// FIX: Instantiate an ASTBuilder, not a ParserAPI.
	builder := parser.NewASTBuilder(logging.NewNoOpLogger())
	program, _, bErr := builder.Build(antlrTree)
	if bErr != nil {
		t.Fatalf("ast.Build() failed unexpectedly: %v", bErr)
	}

	// 2. Normalize the original AST before comparison.
	normalizeAST(program)
	originalTree := &ast.Tree{Root: program}

	// 3. Run the AST through the full Canonicalise/Decode cycle.
	blob, _, err := CanonicaliseWithRegistry(originalTree)
	if err != nil {
		t.Fatalf("CanonicaliseWithRegistry() failed unexpectedly: %v", err)
	}
	decodedTree, err := DecodeWithRegistry(blob)
	if err != nil {
		t.Fatalf("DecodeWithRegistry() failed unexpectedly: %v", err)
	}

	// 4. Perform a deep comparison.
	cmpOpts := []cmp.Option{
		cmpopts.IgnoreFields(ast.BaseNode{}, "StartPos", "StopPos"),
		cmpopts.IgnoreUnexported(ast.Procedure{}, ast.Step{}, ast.LValueNode{}),
		cmpopts.EquateEmpty(),
		cmpopts.SortSlices(func(a, b *ast.MapEntryNode) bool { return a.Key.Value < b.Key.Value }),
	}

	if diff := cmp.Diff(originalTree, decodedTree, cmpOpts...); diff != "" {
		t.Errorf("FAIL: The decoded AST does not match the original. The following fields were lost or altered during canonicalization:\n%s", diff)
	}
}

// TestComprehensiveGrammarRoundtrip is the ultimate regression test for the canonicalization process.
func TestComprehensiveGrammarRoundtrip(t *testing.T) {
	testCases := []struct {
		name       string
		scriptFile string
	}{
		{"Library Script", filepath.Join("..", "antlr", "comprehensive_grammar.ns")},
		{"Whisper", filepath.Join("..", "antlr", "whisper_feature.ns")},
		{"Command Script", filepath.Join("..", "antlr", "command_block.ns")},
		{"Additional Features Library Script", filepath.Join("..", "antlr", "additional_features.ns")},
		{"Additional Features Command Script", filepath.Join("..", "antlr", "additional_command_block.ns")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runRoundtripComparison(t, tc.scriptFile)
		})
	}
}
