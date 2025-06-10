// NeuroScript Version: 0.5.2
// File version: 32
// Purpose: Entry point for the AST builder, orchestrating the ANTLR listener to construct the program AST.
// filename: pkg/core/ast_builder_main.go
// nlines: 215
// risk_rating: HIGH

package core

import (
	"errors"
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// --- ASTBuilder (Exported Constructor and Build Method) ---

// ASTBuilder encapsulates the logic for building the NeuroScript AST using a listener.
type ASTBuilder struct {
	logger   logging.Logger
	debugAST bool // Option to enable detailed AST construction logging
}

// NewASTBuilder creates a new ASTBuilder instance.
func NewASTBuilder(logger logging.Logger) *ASTBuilder {
	if logger == nil {
		logger = &coreNoOpLogger{}
	}
	// Forcing debugAST true here for continued debugging.
	return &ASTBuilder{
		logger:   logger,
		debugAST: true,
	}
}

// Build takes an ANTLR parse tree and constructs the NeuroScript Program AST (*core.Program).
// It now returns the Program, the collected file metadata, and any error.
func (b *ASTBuilder) Build(tree antlr.Tree) (*Program, map[string]string, error) {
	if tree == nil {
		fmt.Printf("AST Builder ERROR: PANIC - Cannot build AST from nil parse tree.\n")
		return nil, nil, fmt.Errorf("cannot build AST from nil parse tree")
	}
	fmt.Printf("--- AST Builder: Build Start ---\n")

	// Create the listener instance.
	listener := newNeuroScriptListener(b.logger, b.debugAST) // Pass debug flag

	// Walk the parse tree with the listener.
	fmt.Printf("AST Builder DEBUG: Starting ANTLR walk...\n")
	walker := antlr.NewParseTreeWalker()
	walker.Walk(listener, tree)
	fmt.Printf("AST Builder DEBUG: ANTLR walk finished.\n")

	// Get metadata *after* the walk, before returning on error.
	fileMetadata := listener.GetFileMetadata()
	if fileMetadata == nil {
		fileMetadata = make(map[string]string)
		fmt.Printf("AST Builder WARN: Listener returned nil metadata map, initialized empty map.\n")
	}
	fmt.Printf("AST Builder DEBUG: Metadata collected from listener: %#v\n", fileMetadata)

	// Check for errors collected during the walk
	if len(listener.errors) > 0 {
		fmt.Printf("AST Builder ERROR: %d error(s) detected immediately after ANTLR walk.\n", len(listener.errors))
		errorMessages := make([]string, 0, len(listener.errors))
		for i, err := range listener.errors {
			if err != nil {
				errMsg := err.Error()
				errorMessages = append(errorMessages, errMsg)
				fmt.Printf("AST Builder DEBUG: Recorded error #%d: %s\n", i+1, errMsg)
			} else {
				errorMessages = append(errorMessages, "<nil error recorded>")
				fmt.Printf("AST Builder ERROR: Listener recorded a nil error object.\n")
			}
		}
		combinedError := fmt.Errorf("AST build failed with %d error(s) during parse walk: %s", len(errorMessages), strings.Join(errorMessages, "; "))
		fmt.Printf("AST Builder ERROR: Failing build due to errors during walk: %v\n", combinedError)
		// We still return the partially built program and metadata for inspection
		return listener.program, fileMetadata, combinedError
	}

	// Get the assembled program from the listener.
	programAST := listener.program

	if programAST == nil {
		fmt.Printf("AST Builder ERROR: PANIC - Build completed without explicit errors, but resulted in a nil program AST\n")
		return nil, fileMetadata, errors.New("AST builder internal error: program AST is unexpectedly nil after successful walk")
	}

	// Final assembly: Populate the Program's map from the listener's temporary slice.
	if programAST.Procedures == nil {
		fmt.Printf("AST Builder WARN: Program AST Procedures map was nil, initializing.\n")
		programAST.Procedures = make(map[string]*Procedure)
	}
	fmt.Printf("AST Builder DEBUG: Assembling %d procedures found by listener into the program's map.\n", len(listener.procedures))

	duplicateProcs := false
	var firstDuplicateName string

	for i, proc := range listener.procedures { // Iterate listener's temporary slice
		fmt.Printf("AST Builder DEBUG: Processing procedure #%d from listener list.\n", i+1)

		if proc != nil {
			fmt.Printf("AST Builder DEBUG: Proc #%d is named '%s' at position %s.\n", i+1, proc.Name, proc.Pos.String())
			// Check for duplicates *before* assigning.
			if existingProc, exists := programAST.Procedures[proc.Name]; exists {
				fmt.Printf("AST Builder ERROR: Found duplicate procedure definition for '%s'.\n", proc.Name)
				if !duplicateProcs { // Record the first duplicate found
					firstDuplicateName = proc.Name
				}
				duplicateProcs = true
				errMsg := fmt.Sprintf("duplicate procedure definition: '%s' found at %s conflicts with existing at %s",
					proc.Name, proc.Pos.String(), existingProc.Pos.String())
				fmt.Printf("AST Builder ERROR: Duplicate details. name=%s, new_pos=%s, existing_pos=%s\n", proc.Name, proc.Pos.String(), existingProc.Pos.String())
				listener.errors = append(listener.errors, errors.New(errMsg))
				continue // Skip adding this duplicate procedure
			}
			// Assign the non-nil proc pointer to the map only if it doesn't exist yet
			programAST.Procedures[proc.Name] = proc
			fmt.Printf("AST Builder DEBUG: Successfully added procedure '%s' to program AST.\n", proc.Name)
		} else {
			fmt.Printf("AST Builder WARN: Encountered a nil procedure pointer in the temporary list at index %d.\n", i)
			// This might be a critical error, let's record it.
			listener.errors = append(listener.errors, fmt.Errorf("internal AST builder error: found nil procedure pointer at index %d", i))
		}
	}

	// If duplicates were found OR other errors exist, aggregate errors and return.
	if len(listener.errors) > 0 {
		fmt.Printf("AST Builder ERROR: Failing build after assembly stage due to %d total errors.\n", len(listener.errors))
		uniqueErrors := make(map[string]struct{})
		finalErrorMessages := make([]string, 0, len(listener.errors))
		for _, err := range listener.errors {
			if err != nil {
				msg := err.Error()
				if _, seen := uniqueErrors[msg]; !seen {
					finalErrorMessages = append(finalErrorMessages, msg)
					uniqueErrors[msg] = struct{}{}
				}
			}
		}

		errorPrefix := "AST build failed"
		if duplicateProcs {
			errorPrefix = fmt.Sprintf("%s due to duplicate procedure(s) (first: '%s') and potentially other errors", errorPrefix, firstDuplicateName)
		}

		combinedError := fmt.Errorf("%s: %s", errorPrefix, strings.Join(finalErrorMessages, "; "))
		fmt.Printf("AST Builder ERROR: Final aggregated error: %v\n", combinedError)
		return programAST, fileMetadata, combinedError
	}

	fmt.Printf("--- AST Builder: Build process completed successfully. ---\n")
	return programAST, fileMetadata, nil // Return nil error on success
}

func MapKeysListener(m map[string]string) []string { // Renamed to avoid conflict if core.MapKeys exists
	if m == nil {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
