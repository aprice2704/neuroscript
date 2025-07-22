// filename: pkg/parser/ast_builder_main_test.go
// NeuroScript Version: 0.5.2
// File version: 13
// Purpose: Corrected tests to use the robust LineInfo algorithm and restored value stack check.

package parser

import (
	"strings"
	"testing"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/logging"
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
		tree, tokenStream, errs := parserAPI.parseInternal("test.ns", script)
		if len(errs) > 0 {
			t.Fatalf("Parser failed unexpectedly: %v", errs)
		}

		builder := NewASTBuilder(logging.NewNoOpLogger())
		program, _, err := builder.BuildFromParseResult(tree, tokenStream)

		if err != nil {
			t.Errorf("Expected no error from Build, but got: %v", err)
		}
		if program == nil {
			t.Error("Expected a non-nil program, but got nil")
		}
	})

	t.Run("build fails if value stack is not empty at the end", func(t *testing.T) {
		builder := NewASTBuilder(logging.NewNoOpLogger())

		builder.postListenerCreationTestHook = func(l *neuroScriptListenerImpl) {
			l.push("rogue value")
		}

		parserAPI := NewParserAPI(logging.NewNoOpLogger())
		tree, tokenStream, _ := parserAPI.parseInternal("test.ns", "func a() means\n endfunc")

		_, _, err := builder.BuildFromParseResult(tree, tokenStream)
		if err == nil {
			t.Fatal("Expected an error for non-empty stack, but got nil")
		}
		if !strings.Contains(err.Error(), "value stack size is 1 at end of program") {
			t.Errorf("Expected error about non-empty stack, but got: %v", err)
		}
	})

	t.Run("build fails when visiting an error node", func(t *testing.T) {
		script := `func my_func() means @ endfunc`
		input := antlr.NewInputStream(script)
		lexer := gen.NewNeuroScriptLexer(input)
		stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
		p := gen.NewNeuroScriptParser(stream)
		p.RemoveErrorListeners()
		tree := p.Program()

		builder := NewASTBuilder(logging.NewNoOpLogger())
		_, _, err := builder.BuildFromParseResult(tree, stream)
		if err == nil {
			t.Fatal("Expected an error for visiting an error node, but got nil")
		}
		if !strings.Contains(err.Error(), "stack underflow: could not pop procedure body") {
			t.Errorf("Expected error about stack underflow, but got: %v", err)
		}
	})
}
