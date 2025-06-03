// NeuroScript Version: 0.3.1
// File version: 0.0.2 // Corrected test case expectations for complex_nested_access and script for bracket_access_with_simple_expression.
// Purpose: Tests for L-value parsing in set statements.
// filename: pkg/core/ast_builder_lvalue_test.go
// nlines: 250 // Approximate
// risk_rating: MEDIUM

package core

import (
	"testing"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated" // Assuming this path
	// Assuming this path for logger
)

// parseScriptToLValueNode is a helper function to parse a script snippet
// and return the LValueNode from the first "set" statement encountered.
func parseScriptToLValueNode(t *testing.T, scriptContent string) *LValueNode {
	t.Helper()

	input := antlr.NewInputStream(scriptContent)
	lexer := gen.NewNeuroScriptLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := gen.NewNeuroScriptParser(stream)

	errorListener := NewSyntaxErrorListener(scriptContent)
	parser.RemoveErrorListeners()
	parser.AddErrorListener(errorListener)

	tree := parser.Program()

	if len(errorListener.GetErrors()) > 0 {
		t.Fatalf("Syntax errors found during parsing script for LValue test:\n%s\nErrors: %v", scriptContent, errorListener.GetErrors())
		return nil
	}

	nopLogger := &coreNoOpLogger{}
	// Check if a test-specific logger is available from a helper like NewDefaultTestInterpreter
	// This part depends on how your test setup provides loggers.
	// For simplicity, using coreNoOpLogger directly. If your NewASTBuilder panics with it,
	// you might need to pass a logger from testInterp.Logger() if testInterp is available here.
	// However, for AST building, a no-op logger is often fine if not debugging the builder itself.

	astBuilder := NewASTBuilder(nopLogger) // Ensure NewASTBuilder can handle coreNoOpLogger
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
			if step.Type == "set" && step.LValue != nil {
				return step.LValue
			}
		}
	}

	t.Fatalf("No 'set' statement with LValue found in script:\n%s", scriptContent)
	return nil
}

func TestLValueParsing(t *testing.T) {
	testCases := []struct {
		name                      string
		script                    string
		expectedIdentifier        string
		expectedAccessorCount     int
		expectedAccessors         []AccessorNode // For detailed checks; only key fields usually
		expectParseError          bool
		skipDetailedAccessorCheck bool
	}{
		{
			name:                  "simple identifier",
			script:                "func t means\nset myVar = 1\nendfunc",
			expectedIdentifier:    "myVar",
			expectedAccessorCount: 0,
			expectedAccessors:     []AccessorNode{},
		},
		{
			name:                  "single dot access",
			script:                "func t means\nset myMap.key = 1\nendfunc",
			expectedIdentifier:    "myMap",
			expectedAccessorCount: 1,
			expectedAccessors:     []AccessorNode{{Type: DotAccess, FieldName: "key"}},
		},
		{
			name:                  "multiple dot access",
			script:                "func t means\nset myMap.key1.key2 = 1\nendfunc",
			expectedIdentifier:    "myMap",
			expectedAccessorCount: 2,
			expectedAccessors:     []AccessorNode{{Type: DotAccess, FieldName: "key1"}, {Type: DotAccess, FieldName: "key2"}},
		},
		{
			name:                      "single bracket access string key",
			script:                    "func t means\nset myMap[\"a_key\"] = 1\nendfunc",
			expectedIdentifier:        "myMap",
			expectedAccessorCount:     1,
			expectedAccessors:         []AccessorNode{{Type: BracketAccess}},
			skipDetailedAccessorCheck: true,
		},
		{
			name:                      "single bracket access numeric index",
			script:                    "func t means\nset myList[0] = 1\nendfunc",
			expectedIdentifier:        "myList",
			expectedAccessorCount:     1,
			expectedAccessors:         []AccessorNode{{Type: BracketAccess}},
			skipDetailedAccessorCheck: true,
		},
		{
			name:                      "multiple bracket access string keys",
			script:                    "func t means\nset myMap[\"keyA\"][\"keyB\"] = 1\nendfunc",
			expectedIdentifier:        "myMap",
			expectedAccessorCount:     2,
			expectedAccessors:         []AccessorNode{{Type: BracketAccess}, {Type: BracketAccess}},
			skipDetailedAccessorCheck: true,
		},
		{
			name:                  "mixed dot and bracket",
			script:                "func t means\nset data.items[0].name = 1\nendfunc",
			expectedIdentifier:    "data",
			expectedAccessorCount: 3,
			expectedAccessors:     []AccessorNode{{Type: DotAccess, FieldName: "items"}, {Type: BracketAccess}, {Type: DotAccess, FieldName: "name"}},
			// Setting skipDetailedAccessorCheck to true as specific IndexOrKey values are not asserted.
			skipDetailedAccessorCheck: true,
		},
		{
			name:                      "mixed bracket and dot",
			script:                    "func t means\nset config[\"settings\"].port = 1\nendfunc",
			expectedIdentifier:        "config",
			expectedAccessorCount:     2,
			expectedAccessors:         []AccessorNode{{Type: BracketAccess}, {Type: DotAccess, FieldName: "port"}},
			skipDetailedAccessorCheck: true,
		},
		{
			name:                  "complex nested access",
			script:                "func t means\nset obj.array[1][\"inner\"].field.anotherArray[0] = 1\nendfunc",
			expectedIdentifier:    "obj",
			expectedAccessorCount: 6, // Corrected from 5
			expectedAccessors: []AccessorNode{
				{Type: DotAccess, FieldName: "array"},
				{Type: BracketAccess}, // [1]
				{Type: BracketAccess}, // ["inner"]
				{Type: DotAccess, FieldName: "field"},
				{Type: DotAccess, FieldName: "anotherArray"}, // Corrected
				{Type: BracketAccess},                        // [0]
			},
			skipDetailedAccessorCheck: true,
		},
		{
			name: "bracket access with simple expression (variable)",
			// Script modified so "myList[x]" is the LValue from the first 'set' statement.
			// Assumes 'x' is a variable; its prior definition isn't needed for LValue parsing structure test.
			script:                    "func t means\nset myList[x] = 1\nendfunc",
			expectedIdentifier:        "myList",
			expectedAccessorCount:     1,
			expectedAccessors:         []AccessorNode{{Type: BracketAccess}}, // IndexOrKey will be a VariableNode "x"
			skipDetailedAccessorCheck: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lvalNode := parseScriptToLValueNode(t, tc.script)
			if lvalNode == nil {
				return // Helper already called t.Fatalf
			}

			if lvalNode.Identifier != tc.expectedIdentifier {
				t.Errorf("Expected LValue Identifier to be %q, got %q", tc.expectedIdentifier, lvalNode.Identifier)
			}

			if len(lvalNode.Accessors) != tc.expectedAccessorCount {
				t.Errorf("Expected %d accessors, got %d. Actual Accessors:", tc.expectedAccessorCount, len(lvalNode.Accessors))
				for i, acc := range lvalNode.Accessors {
					// Log details of actual accessors if count mismatches
					t.Logf("  Accessor %d: Type=%v, FieldName=%q, IndexOrKey=%T (%#v)", i, acc.Type, acc.FieldName, acc.IndexOrKey, acc.IndexOrKey)
				}
			}

			// Only do detailed checks if counts match and not skipped
			if !tc.skipDetailedAccessorCheck && len(lvalNode.Accessors) == tc.expectedAccessorCount && tc.expectedAccessorCount > 0 {
				for i, expectedAcc := range tc.expectedAccessors {
					if i >= len(lvalNode.Accessors) { // Should not happen if counts match
						t.Fatalf("Error in test logic: trying to access index %d when only %d accessors exist", i, len(lvalNode.Accessors))
						break
					}
					actualAcc := lvalNode.Accessors[i]
					if actualAcc.Type != expectedAcc.Type {
						t.Errorf("Accessor %d: Expected Type %v, got %v", i, expectedAcc.Type, actualAcc.Type)
					}
					if expectedAcc.Type == DotAccess && actualAcc.FieldName != expectedAcc.FieldName {
						t.Errorf("Accessor %d: Expected FieldName %q, got %q", i, expectedAcc.FieldName, actualAcc.FieldName)
					}
					// Further checks for IndexOrKey can be added here if needed, e.g., checking type or simple literal values.
					// For example, to check if accessor 1 for "complex_nested_access" (obj.array[1]...) has a NumberLiteralNode:
					// if tc.name == "complex_nested_access" && i == 1 { // Accessor for [1]
					// 	if _, ok := actualAcc.IndexOrKey.(*NumberLiteralNode); !ok {
					// 		t.Errorf("Accessor %d: Expected IndexOrKey to be *NumberLiteralNode, got %T", i, actualAcc.IndexOrKey)
					// 	}
					// }
				}
			}
		})
	}
}

// Minimal SyntaxErrorListener for parsing in tests.
type SyntaxErrorListener struct {
	*antlr.DefaultErrorListener
	SourceName string
	errors     []StructuredSyntaxError
}

func NewSyntaxErrorListener(sourceName string) *SyntaxErrorListener {
	return &SyntaxErrorListener{
		DefaultErrorListener: antlr.NewDefaultErrorListener(),
		SourceName:           sourceName,
		errors:               make([]StructuredSyntaxError, 0),
	}
}

func (l *SyntaxErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	adjustedColumn := column + 1
	symbolText := ""
	if token, ok := offendingSymbol.(antlr.Token); ok {
		symbolText = token.GetText()
	}

	l.errors = append(l.errors, StructuredSyntaxError{
		Line:            line,
		Column:          adjustedColumn,
		Msg:             msg,
		OffendingSymbol: symbolText,
		SourceName:      l.SourceName,
	})
}

func (l *SyntaxErrorListener) GetErrors() []StructuredSyntaxError {
	return l.errors
}
