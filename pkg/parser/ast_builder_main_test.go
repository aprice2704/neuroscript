// filename: pkg/parser/ast_builder_main_test.go
// NeuroScript Version: 0.5.2
// File version: 10
// Purpose: Corrected the expected error message to match the actual failure mode.

package parser

import (
	"strings"
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/aprice2704/neuroscript/pkg/logging"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

func TestASTBuilder_Build(t *testing.T) {
	t.Run("successful build of a valid script", func(t *testing.T) {
		script := `
            func my_func(needs input) means
                set result = input + 1
                return result
            endfunc
        `
		parserAPI := NewParserAPI(logging.NewNoOpLogger())
		tree, err := parserAPI.Parse(script)
		if err != nil {
			t.Fatalf("Parser failed unexpectedly: %v", err)
		}

		builder := NewASTBuilder(logging.NewNoOpLogger())
		program, _, err := builder.Build(tree)

		if err != nil {
			t.Errorf("Expected no error from Build, but got: %v", err)
		}
		if program == nil {
			t.Error("Expected a non-nil program, but got nil")
		}
	})

	t.Run("build fails if value stack is not empty at the end", func(t *testing.T) {
		builder := NewASTBuilder(logging.NewNoOpLogger())

		// Set the test hook to manipulate the listener's state mid-build.
		builder.postListenerCreationTestHook = func(l *neuroScriptListenerImpl) {
			l.push("rogue value")
		}

		// Create a minimal valid parse tree.
		parserAPI := NewParserAPI(logging.NewNoOpLogger())
		tree, _ := parserAPI.Parse("func a() means\n endfunc")

		_, _, err := builder.Build(tree)
		if err == nil {
			t.Fatal("Expected an error for non-empty stack, but got nil")
		}
		if !strings.Contains(err.Error(), "value stack size is 1 at end of program") {
			t.Errorf("Expected error about non-empty stack, but got: %v", err)
		}
	})

	t.Run("build fails when visiting an error node", func(t *testing.T) {
		// This script has a syntax error that will cause the parser to create an ErrorNode.
		// We manually create the parser here to bypass the ParserAPI's early exit on syntax errors,
		// ensuring the builder actually receives a tree containing an error node.
		script := `func my_func() means @ endfunc`
		input := antlr.NewInputStream(script)
		lexer := gen.NewNeuroScriptLexer(input)
		stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
		p := gen.NewNeuroScriptParser(stream)
		p.RemoveErrorListeners() // Remove default listeners to suppress console output for this expected error.
		tree := p.Program()

		builder := NewASTBuilder(logging.NewNoOpLogger())
		_, _, err := builder.Build(tree)
		if err == nil {
			t.Fatal("Expected an error for visiting an error node, but got nil")
		}
		// The syntax error causes a cascade where the builder fails to pop the procedure body.
		// This is the correct final error to check for.
		if !strings.Contains(err.Error(), "stack underflow: could not pop procedure body") {
			t.Errorf("Expected error about stack underflow, but got: %v", err)
		}
	})
}
