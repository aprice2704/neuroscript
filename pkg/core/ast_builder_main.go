// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-29 // Updated for string literal un-escaping and literal handling
package core

import (
	"errors"
	"fmt" // Ensure fmt is imported

	// Added for number parsing
	"strings"

	"github.com/antlr4-go/antlr/v4" // Corrected ANTLR import
	// Corrected path
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
		// Use the existing coreNoOpLogger from this package (defined in helpers.go)
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
		b.logger.Error("AST Builder: Cannot build AST from nil parse tree.")
		return nil, nil, fmt.Errorf("cannot build AST from nil parse tree")
	}
	b.logger.Debug("AST Builder: Starting AST build process using Listener.")

	// Create the listener instance.
	listener := newNeuroScriptListener(b.logger, b.debugAST) // Pass debug flag

	// Walk the parse tree with the listener.
	b.logger.Debug("AST Builder: Starting ANTLR walk...")
	walker := antlr.NewParseTreeWalker()
	walker.Walk(listener, tree)
	b.logger.Debug("AST Builder: ANTLR walk finished.")

	// Get metadata *after* the walk, before returning on error.
	fileMetadata := listener.GetFileMetadata()
	if fileMetadata == nil {
		fileMetadata = make(map[string]string)
		b.logger.Warn("AST Builder: Listener returned nil metadata map, initialized empty map.")
	}

	// Check for errors collected during the walk
	if len(listener.errors) > 0 {
		errorMessages := make([]string, 0, len(listener.errors))
		for _, err := range listener.errors {
			if err != nil {
				errorMessages = append(errorMessages, err.Error())
			} else {
				errorMessages = append(errorMessages, "<nil error recorded>")
				b.logger.Error("AST Builder: Listener recorded a nil error object.")
			}
		}
		combinedError := fmt.Errorf("AST build failed with %d error(s): %s", len(errorMessages), strings.Join(errorMessages, "; "))
		b.logger.Error("AST Builder: Errors detected during ANTLR walk.", "error", combinedError)
		return listener.program, fileMetadata, combinedError
	}

	// Get the assembled program from the listener.
	programAST := listener.program

	if programAST == nil {
		b.logger.Error("AST Builder: Build completed without explicit errors, but resulted in a nil program AST")
		return nil, fileMetadata, errors.New("AST builder internal error: program AST is unexpectedly nil after successful walk")
	}

	// Final assembly: Populate the Program's map from the listener's temporary slice.
	if programAST.Procedures == nil {
		b.logger.Warn("AST Builder: Program AST Procedures map was nil, initializing.")
		programAST.Procedures = make(map[string]*Procedure)
	}

	// --- Debug Logging Added Here ---
	b.logger.Debug("AST Builder: Assembling procedures found by listener.", "count", len(listener.procedures))
	// --- End Debug Logging ---

	duplicateProcs := false
	var firstDuplicateName string

	for i, proc := range listener.procedures { // Iterate listener's temporary slice

		// --- Debug Logging Added Here ---
		b.logger.Debug("AST Builder: Checking procedure from listener list.", "index", i, "proc_pointer", fmt.Sprintf("%p", proc))
		// --- End Debug Logging ---

		if proc != nil {
			// Check for duplicates *before* assigning.
			if existingProc, exists := programAST.Procedures[proc.Name]; exists {
				if !duplicateProcs { // Record the first duplicate found
					firstDuplicateName = proc.Name
				}
				duplicateProcs = true
				// Use the String() method of the Position struct
				errMsg := fmt.Sprintf("duplicate procedure definition: '%s' found at %s conflicts with existing at %s",
					proc.Name, proc.Pos.String(), existingProc.Pos.String())

				b.logger.Error("AST Builder: Duplicate procedure.", "name", proc.Name, "new_pos", proc.Pos.String(), "existing_pos", existingProc.Pos.String())
				listener.errors = append(listener.errors, errors.New(errMsg))
				continue // Skip adding this duplicate procedure
			}
			// Assign the non-nil proc pointer to the map only if it doesn't exist yet
			programAST.Procedures[proc.Name] = proc

		} else {
			b.logger.Warn("AST Builder encountered a nil procedure pointer in the temporary list.", "index", i)
			// Potentially make this fatal:
			// return nil, fileMetadata, fmt.Errorf("internal AST builder error: found nil procedure pointer at index %d", i)
		}
	}

	// If duplicates were found OR other errors exist, aggregate errors and return.
	if duplicateProcs || len(listener.errors) > 0 { // Check listener.errors again as we might have added new ones
		errorMessages := make([]string, 0, len(listener.errors))
		hasNonDuplicateError := false
		uniqueErrors := make(map[string]struct{}) // Avoid adding same error msg multiple times

		currentErrors := listener.errors
		for _, err := range currentErrors {
			if err != nil {
				msg := err.Error()
				if _, seen := uniqueErrors[msg]; !seen {
					errorMessages = append(errorMessages, msg)
					uniqueErrors[msg] = struct{}{}
					if !strings.Contains(msg, "duplicate procedure definition") {
						hasNonDuplicateError = true
					}
				}
			}
		}
		errorPrefix := "AST build failed"
		// Adjust error prefix logic based on whether there are any messages to join
		if len(errorMessages) > 0 {
			if duplicateProcs && !hasNonDuplicateError && strings.Contains(errorMessages[0], "duplicate procedure definition") {
				// If the first (and possibly only) error is the duplicate one
				errorPrefix = fmt.Sprintf("AST build failed due to duplicate procedure definition(s) (first: '%s')", firstDuplicateName)
			} else if duplicateProcs {
				errorPrefix = fmt.Sprintf("AST build failed due to duplicate procedure definition(s) (first: '%s') and other errors", firstDuplicateName)
			}
		} else if duplicateProcs { // Only duplicates, but errorMessages might be empty if they were all identical.
			errorPrefix = fmt.Sprintf("AST build failed due to duplicate procedure definition(s) (first: '%s')", firstDuplicateName)
			// Ensure there's at least one message if we indicate duplicates
			if len(errorMessages) == 0 {
				errorMessages = append(errorMessages, fmt.Sprintf("duplicate procedure: %s", firstDuplicateName))
			}
		}

		combinedError := fmt.Errorf("%s: %s", errorPrefix, strings.Join(errorMessages, "; "))
		b.logger.Error("AST Builder: Build failed.", "error", combinedError)
		return programAST, fileMetadata, combinedError
	}

	b.logger.Debug("AST Builder: Build process completed successfully.")
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

// Note: Implementations for other Enter/Exit methods (e.g., for expressions, statements,
// list_literal, map_literal, map_entry) are expected to be in this file or other ast_builder_*.go files.
// They will need to correctly use l.pushValue and l.popValue (or l.popNValues)
// to manage the AST node stack.
