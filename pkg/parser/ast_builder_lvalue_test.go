// NeuroScript Version: 0.3.1
// File version: 0.0.3
// Purpose: Updated test helper to use the new `LValues` field on the ast.Step struct.
// filename: pkg/core/ast_builder_lValue_test.go
// nlines: 253
// risk_rating: MEDIUM

package parser

import (
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

// parseScriptToLValueNode is a helper function to parse a script snippet
// and return the ast.LValueNode from the first "set" statement encountered.
func parseScriptToLValueNode(t *testing.T, scriptContent string) *ast.LValueNode {
	t.Helper()

	input := antlr.NewInputStream(scriptContent)
	lexer := gen.NewNeuroScriptLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := gen.NewNeuroScriptParser(stream)

	errorListener := NewErrorListener(nil) // Using the existing error listener
	parser.RemoveErrorListeners()
	parser.AddErrorListener(errorListener)

	tree := parser.Program()

	if len(errorListener.RawErrors) > 0 {
		t.Fatalf("Syntax errors found during parsing script for LValue test:\n%s\nErrors: %v", scriptContent, errorListener.RawErrors)
		return nil
	}

	nopLogger := adapters.NewNoOpLogger()
	astBuilder := NewASTBuilder(nopLogger)
	programAST, _, err := astBuilder.Build(tree)
	if err != nil {
		t.Fatalf("AST build failed for script:\n%s\nError: %v", scriptContent, err)
		return nil
	}
	if programAST == nil {
		t.Fatalf("AST build resulted in nil program for script:\n%s", scriptContent)
		return nil
	}

	for _, proc := range programAST.Procedures {
		for _, step := range proc.Steps {
			if step.Type == "set" && len(step.LValues) > 0 {
				return step.LValues[0]
			}
		}
	}

	t.Fatalf("No 'set' statement with LValues found in script:\n%s", scriptContent)
	return nil
}

func TestLValueParsing(t *testing.T) {
	testCases := []struct {
		name                      string
		script                    string
		expectedIdentifier        string
		expectedAccessorCount     int
		expectedAccessors         []ast.AccessorNode // For detailed checks; only key fields usually
		expectParseError          bool
		skipDetailedAccessorCheck bool
	}{
		{
			name:                  "simple identifier",
			script:                "func t() means\nset myVar = 1\nendfunc",
			expectedIdentifier:    "myVar",
			expectedAccessorCount: 0,
			expectedAccessors:     []ast.AccessorNode{},
		},
		{
			name:                  "single dot access",
			script:                "func t() means\nset myMap.key = 1\nendfunc",
			expectedIdentifier:    "myMap",
			expectedAccessorCount: 1,
			expectedAccessors:     []ast.AccessorNode{{Type: ast.DotAccess, Key: &ast.StringLiteralNode{Value: "key"}}},
		},
		{
			name:                  "multiple dot access",
			script:                "func t() means\nset myMap.key1.key2 = 1\nendfunc",
			expectedIdentifier:    "myMap",
			expectedAccessorCount: 2,
			expectedAccessors:     []ast.AccessorNode{{Type: ast.DotAccess, Key: &ast.StringLiteralNode{Value: "key1"}}, {Type: ast.DotAccess, Key: &ast.StringLiteralNode{Value: "key2"}}},
		},
		{
			name:                      "single bracket access string key",
			script:                    "func t() means\nset myMap[\"a_key\"] = 1\nendfunc",
			expectedIdentifier:        "myMap",
			expectedAccessorCount:     1,
			expectedAccessors:         []ast.AccessorNode{{Type: ast.BracketAccess}},
			skipDetailedAccessorCheck: true,
		},
		{
			name:                      "single bracket access numeric index",
			script:                    "func t() means\nset myList[0] = 1\nendfunc",
			expectedIdentifier:        "myList",
			expectedAccessorCount:     1,
			expectedAccessors:         []ast.AccessorNode{{Type: ast.BracketAccess}},
			skipDetailedAccessorCheck: true,
		},
		{
			name:                      "multiple bracket access string keys",
			script:                    "func t() means\nset myMap[\"keyA\"][\"keyB\"] = 1\nendfunc",
			expectedIdentifier:        "myMap",
			expectedAccessorCount:     2,
			expectedAccessors:         []ast.AccessorNode{{Type: ast.BracketAccess}, {Type: ast.BracketAccess}},
			skipDetailedAccessorCheck: true,
		},
		{
			name:                      "mixed dot and bracket",
			script:                    "func t() means\nset data.items[0].name = 1\nendfunc",
			expectedIdentifier:        "data",
			expectedAccessorCount:     3,
			expectedAccessors:         []ast.AccessorNode{{Type: ast.DotAccess, Key: &ast.StringLiteralNode{Value: "items"}}, {Type: ast.BracketAccess}, {Type: ast.DotAccess, Key: &ast.StringLiteralNode{Value: "name"}}},
			skipDetailedAccessorCheck: true,
		},
		{
			name:                      "mixed bracket and dot",
			script:                    "func t() means\nset config[\"settings\"].port = 1\nendfunc",
			expectedIdentifier:        "config",
			expectedAccessorCount:     2,
			expectedAccessors:         []ast.AccessorNode{{Type: ast.BracketAccess}, {Type: ast.DotAccess, Key: &ast.StringLiteralNode{Value: "port"}}},
			skipDetailedAccessorCheck: true,
		},
		{
			name:                  "complex nested access",
			script:                "func t() means\nset obj.array[1][\"inner\"].field.anotherArray[0] = 1\nendfunc",
			expectedIdentifier:    "obj",
			expectedAccessorCount: 6,
			expectedAccessors: []ast.AccessorNode{
				{Type: ast.DotAccess, Key: &ast.StringLiteralNode{Value: "array"}},
				{Type: ast.BracketAccess},
				{Type: ast.BracketAccess},
				{Type: ast.DotAccess, Key: &ast.StringLiteralNode{Value: "field"}},
				{Type: ast.DotAccess, Key: &ast.StringLiteralNode{Value: "anotherArray"}},
				{Type: ast.BracketAccess},
			},
			skipDetailedAccessorCheck: true,
		},
		{
			name:                      "bracket access with simple expression (variable)",
			script:                    "func t() means\nset myList[x] = 1\nendfunc",
			expectedIdentifier:        "myList",
			expectedAccessorCount:     1,
			expectedAccessors:         []ast.AccessorNode{{Type: ast.BracketAccess}},
			skipDetailedAccessorCheck: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lvalNode := parseScriptToLValueNode(t, tc.script)
			if lvalNode == nil {
				return
			}

			if lvalNode.Identifier != tc.expectedIdentifier {
				t.Errorf("Expected LValue Identifier to be %q, got %q", tc.expectedIdentifier, lvalNode.Identifier)
			}

			if len(lvalNode.Accessors) != tc.expectedAccessorCount {
				t.Errorf("Expected %d accessors, got %d. Actual Accessors:", tc.expectedAccessorCount, len(lvalNode.Accessors))
				for i, acc := range lvalNode.Accessors {
					t.Logf("  Accessor %d: Type=%v, Key=%#v", i, acc.Type, acc.Key)
				}
			}

			if !tc.skipDetailedAccessorCheck && len(lvalNode.Accessors) == tc.expectedAccessorCount && tc.expectedAccessorCount > 0 {
				for i, expectedAcc := range tc.expectedAccessors {
					if i >= len(lvalNode.Accessors) {
						t.Fatalf("Error in test logic: trying to access index %d when only %d accessors exist", i, len(lvalNode.Accessors))
						break
					}
					actualAcc := lvalNode.Accessors[i]
					if actualAcc.Type != expectedAcc.Type {
						t.Errorf("Accessor %d: Expected Type %v, got %v", i, expectedAcc.Type, actualAcc.Type)
					}
					if expectedAcc.Type == ast.DotAccess {
						expectedKey, ok := expectedAcc.Key.(*ast.StringLiteralNode)
						if !ok {
							t.Fatalf("Test logic error: expected accessor key is not a string literal")
						}
						actualKey, ok := actualAcc.Key.(*ast.StringLiteralNode)
						if !ok {
							t.Errorf("Accessor %d: Expected Key to be a string literal, but got %T", i, actualAcc.Key)
							continue
						}
						if actualKey.Value != expectedKey.Value {
							t.Errorf("Accessor %d: Expected Key %q, got %q", i, expectedKey.Value, actualKey.Value)
						}
					}
				}
			}
		})
	}
}
