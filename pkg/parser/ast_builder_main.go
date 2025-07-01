// filename: pkg/core/ast_builder_main.go
// NeuroScript Version: 0.5.2
// File version: 36
// Purpose: Integrated command block consolidation into the AST build process.
// nlines: 147
// risk_rating: LOW

package parser

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/aprice2704/neuroscript/pkg/adapters"
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
		logger = adapters.NewNoOpLogger() // Use the logger from the adapters package
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

	// 1. Create and fully initialize the listener and its components.
	listener := newNeuroScriptListener(b.logger, b.debugAST)
	if listener.program == nil {
		listener.program = &ast.Program{Procedures: make(map[string]*ast.Procedure)}
	}
	if listener.fileMetadata == nil {
		listener.fileMetadata = make(map[string]string)
	}

	// 2. Walk the parse tree.
	b.logger.Debug("AST Builder: Starting ANTLR walk...")
	walker := antlr.NewParseTreeWalker()
	walker.Walk(listener, tree)
	b.logger.Debug("AST Builder: ANTLR walk finished.")

	programAST := listener.program
	fileMetadata := listener.GetFileMetadata()

	// 3. Consolidate procedures from the temporary slice into the final map.
	if programAST.Procedures == nil {
		programAST.Procedures = make(map[string]*ast.Procedure)
	}
	for _, proc := range listener.procedures {
		if proc == nil {
			listener.errors = append(listener.errors, fmt.Errorf("internal AST builder error: found nil procedure in list"))
			continue
		}
		if _, exists := programAST.Procedures[proc.Name()]; exists {
			errMsg := fmt.Sprintf("duplicate procedure definition: '%s'", proc.Name())
			listener.errors = append(listener.errors, fmt.Errorf("%s", errMsg))
			continue
		}
		programAST.Procedures[proc.Name()] = proc
	}

	// 4. Validate and consolidate event handlers.
	validEvents := make([]*ast.OnEventDecl, 0, len(listener.events))
	for _, ev := range listener.events {
		nameLit, isString := ev.EventNameExpr.(*ast.StringLiteralNode)
		if isString && nameLit.Value == "error" {
			pos := ev.Pos
			errMsg := fmt.Sprintf("misplaced 'on error' handler at line %d; 'on error' is only allowed inside a 'proc' or 'command' block", pos.Line)
			listener.errors = append(listener.errors, fmt.Errorf("%s", errMsg))
			continue // Discard the invalid handler.
		}
		validEvents = append(validEvents, ev)
	}
	programAST.Events = validEvents

	// 5. Consolidate command blocks into the final program AST.
	programAST.Commands = listener.commands

	// 6. Check for and aggregate any errors collected during the walk and validation.
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
		combinedError := fmt.Errorf("AST build failed with %d error(s): %s", len(errorMessages), strings.Join(errorMessages, "; "))
		b.logger.Error("AST Builder: Failing build", "error", combinedError)
		return programAST, fileMetadata, combinedError
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
