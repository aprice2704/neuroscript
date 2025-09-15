// NeuroScript Version: 0.7.2
// File version: 1
// Purpose: Implements the appendScript method to merge script definitions into the interpreter.
// filename: pkg/interpreter/interpreter_append.go
// nlines: 50
// risk_rating: MEDIUM

package interpreter

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// appendScript merges procedures, event handlers, and commands from a program
// AST into the interpreter's existing state. It does not clear previous definitions.
// It returns an error if a procedure being added already exists.
func (i *Interpreter) appendScript(tree *interfaces.Tree) error {
	if tree == nil || tree.Root == nil {
		i.logger.Warn("AppendScript called with a nil program AST.")
		return nil
	}

	program, ok := tree.Root.(*ast.Program)
	if !ok {
		return fmt.Errorf("interpreter.AppendScript: expected root node of type *ast.Program, but got %T", tree.Root)
	}

	// Merge procedures, checking for duplicates.
	for name, proc := range program.Procedures {
		if _, exists := i.state.knownProcedures[name]; exists {
			return lang.NewRuntimeError(lang.ErrorCodeDuplicate, fmt.Sprintf("procedure '%s' already defined", name), lang.ErrProcedureExists)
		}
		i.state.knownProcedures[name] = proc
	}

	// Append event handlers.
	for _, eventDecl := range program.Events {
		if err := i.eventManager.register(eventDecl, i); err != nil {
			return fmt.Errorf("failed to register event handler during append: %w", err)
		}
	}

	// Append top-level commands.
	if program.Commands != nil {
		i.state.commands = append(i.state.commands, program.Commands...)
	}

	return nil
}

// AppendScript merges procedures and event handlers from a new program AST
// into the interpreter's existing state. It does not clear existing definitions.
// It will return an error if a procedure in the new program already exists.
func (i *Interpreter) AppendScript(tree *interfaces.Tree) error {
	return i.appendScript(tree)
}
