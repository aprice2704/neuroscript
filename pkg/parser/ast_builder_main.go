// filename: pkg/parser/ast_builder_main.go
// NeuroScript Version: 0.5.2
// File version: 43
// Purpose: Updated Build method to correctly handle the refactored ast.Program.

package parser

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// ASTBuilder encapsulates the logic for building the NeuroScript AST.
type ASTBuilder struct {
	logger                       interfaces.Logger
	debugAST                     bool
	postListenerCreationTestHook func(*neuroScriptListenerImpl) // Test hook
}

// NewASTBuilder creates a new ASTBuilder instance.
func NewASTBuilder(logger interfaces.Logger) *ASTBuilder {
	if logger == nil {
		logger = logging.NewNoOpLogger()
	}
	return &ASTBuilder{
		logger:   logger,
		debugAST: true, // Forcing debug for now
	}
}

// Build takes an ANTLR parse tree and constructs the NeuroScript ast.Program AST.
func (b *ASTBuilder) Build(tree antlr.Tree) (*ast.Program, map[string]string, error) {
	if tree == nil {
		b.logger.Error("AST Builder FATAL: Cannot build AST from nil parse tree.")
		return nil, nil, fmt.Errorf("cannot build AST from nil parse tree")
	}
	b.logger.Debug("--- AST Builder: Build Start ---")

	listener := newNeuroScriptListener(b.logger, b.debugAST)

	// Execute the test hook if it's set.
	if b.postListenerCreationTestHook != nil {
		b.postListenerCreationTestHook(listener)
	}

	b.logger.Debug("AST Builder: Starting ANTLR walk...")
	walker := antlr.NewParseTreeWalker()
	walker.Walk(listener, tree)
	b.logger.Debug("AST Builder: ANTLR walk finished.")

	programAST := listener.program
	fileMetadata := listener.GetFileMetadata()

	if len(listener.errors) > 0 {
		errorMessages := make([]string, 0, len(listener.errors))
		uniqueErrors := make(map[string]bool)
		for _, err := range listener.errors {
			if err != nil {
				msg := err.Error()
				if !uniqueErrors[msg] {
					errorMessages = append(errorMessages, msg)
					uniqueErrors[msg] = true
				}
			}
		}
		if len(errorMessages) > 0 {
			combinedError := fmt.Errorf("AST build failed with %d error(s): %s", len(errorMessages), strings.Join(errorMessages, "; "))
			b.logger.Error("AST Builder: Failing build", "error", combinedError)
			return programAST, fileMetadata, combinedError
		}
	}
	// Check for a non-empty stack after the walk.
	if len(listener.ValueStack) > 0 {
		return nil, fileMetadata, fmt.Errorf("internal AST builder error: value stack size is %d at end of program", len(listener.ValueStack))
	}

	b.logger.Debug("--- AST Builder: Build process completed successfully. ---")
	return programAST, fileMetadata, nil
}
