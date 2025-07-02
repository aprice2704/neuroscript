// filename: pkg/parser/ast_builder_main.go
// NeuroScript Version: 0.5.2
// File version: 39
// Purpose: Ensured a completely new listener is used for each build to guarantee a clean state.
// nlines: 147
// risk_rating: LOW

package parser

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// ASTBuilder encapsulates the logic for building the NeuroScript AST.
type ASTBuilder struct {
	logger   interfaces.Logger
	debugAST bool
}

// NewASTBuilder creates a new ASTBuilder instance.
func NewASTBuilder(logger interfaces.Logger) *ASTBuilder {
	if logger == nil {
		logger = logging.NewNoLogger() // Use the logger from the adapters package
	}
	return &ASTBuilder{
		logger:   logger,
		debugAST: true, // Forcing debug for now
	}
}

// Build takes an ANTLR parse tree and constructs the NeuroScript ast.Program AST.
// It returns the ast.Program, the collected file metadata, and any error.
func (b *ASTBuilder) Build(tree antlr.Tree) (*ast.Program, map[string]string, error) {
	if tree == nil {
		b.logger.Error("AST Builder FATAL: Cannot build AST from nil parse tree.")
		return nil, nil, fmt.Errorf("cannot build AST from nil parse tree")
	}
	b.logger.Debug("--- AST Builder: Build Start ---")

	// 1. Create a completely new listener for each build to ensure a clean state.
	listener := newNeuroScriptListener(b.logger, b.debugAST)

	// 2. Walk the parse tree.
	b.logger.Debug("AST Builder: Starting ANTLR walk...")
	walker := antlr.NewParseTreeWalker()
	walker.Walk(listener, tree)
	b.logger.Debug("AST Builder: ANTLR walk finished.")

	programAST := listener.program
	fileMetadata := listener.GetFileMetadata()

	// 3. Final consolidation of top-level declarations is now handled directly by the listener's exit methods.
	// The listener.procedures and listener.commands slices are no longer needed here as they
	// are added directly to the programAST during the walk.

	// 4. Check for and aggregate any errors collected during the walk.
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

	b.logger.Debug("--- AST Builder: Build process completed successfully. ---")
	return programAST, fileMetadata, nil
}

// MapKeys is a utility function.
func MapKeys(m map[string]string) []string {
	if m == nil {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
