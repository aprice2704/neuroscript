// NeuroScript Version: 0.5.2
// File version: 33
// Purpose: Corrects listener initialization and error handling to fix critical AST build failures.
// filename: pkg/core/ast_builder_main.go
// nlines: 121
// risk_rating: MEDIUM

package core

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
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
		logger = &coreNoOpLogger{}
	}
	return &ASTBuilder{
		logger:   logger,
		debugAST: true, // Forcing debug for now
	}
}

// Build takes an ANTLR parse tree and constructs the NeuroScript Program AST.
// It returns the Program, the collected file metadata, and any error.
func (b *ASTBuilder) Build(tree antlr.Tree) (*Program, map[string]string, error) {
	if tree == nil {
		b.logger.Error("AST Builder FATAL: Cannot build AST from nil parse tree.")
		return nil, nil, fmt.Errorf("cannot build AST from nil parse tree")
	}
	b.logger.Debug("--- AST Builder: Build Start ---")

	// 1. Create and fully initialize the listener and its components.
	// This is the critical fix: The program and metadata are now prepared *before* the walk.
	listener := newNeuroScriptListener(b.logger, b.debugAST)
	if listener.program == nil {
		listener.program = &Program{Procedures: make(map[string]*Procedure)}
	}
	if listener.fileMetadata == nil {
		listener.fileMetadata = make(map[string]string)
	}

	// 2. Walk the parse tree.
	b.logger.Debug("AST Builder: Starting ANTLR walk...")
	walker := antlr.NewParseTreeWalker()
	walker.Walk(listener, tree)
	b.logger.Debug("AST Builder: ANTLR walk finished.")

	// The listener now manages its own program and metadata.
	// We retrieve the final products after the walk.
	programAST := listener.program
	fileMetadata := listener.GetFileMetadata()

	// 3. Consolidate procedures from the temporary slice into the final map.
	// The listener appends procedures to a slice during the walk.
	if programAST.Procedures == nil {
		programAST.Procedures = make(map[string]*Procedure)
	}
	for _, proc := range listener.procedures {
		if proc == nil {
			listener.errors = append(listener.errors, fmt.Errorf("internal AST builder error: found nil procedure in list"))
			continue
		}
		if _, exists := programAST.Procedures[proc.Name]; exists {
			errMsg := fmt.Sprintf("duplicate procedure definition: '%s'", proc.Name)
			listener.errors = append(listener.errors, fmt.Errorf("%s", errMsg))
			continue
		}
		programAST.Procedures[proc.Name] = proc
	}
	programAST.Events = listener.events

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
		combinedError := fmt.Errorf("AST build failed with %d error(s): %s", len(errorMessages), strings.Join(errorMessages, "; "))
		b.logger.Error("AST Builder: Failing build", "error", combinedError)
		// Return the partially built program for debugging, along with the error.
		return programAST, fileMetadata, combinedError
	}

	b.logger.Debug("--- AST Builder: Build process completed successfully. ---")
	return programAST, fileMetadata, nil
}

// MapKeysListener is a utility function.
func MapKeysListener(m map[string]string) []string {
	if m == nil {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
