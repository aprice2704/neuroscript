// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Updated AppendScript to respect AllowRedefinition flag.
// filename: pkg/interpreter/append.go
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
// AST into the interpreter's existing state.
func (i *Interpreter) appendScript(tree *interfaces.Tree) error {
	if tree == nil || tree.Root == nil {
		i.Logger().Warn("AppendScript called with a nil program AST.")
		return nil
	}

	program, ok := tree.Root.(*ast.Program)
	if !ok {
		return fmt.Errorf("interpreter.AppendScript: expected root node of type *ast.Program, but got %T", tree.Root)
	}

	for name, proc := range program.Procedures {
		if !i.AllowRedefinition {
			if _, exists := i.state.knownProcedures[name]; exists {
				return lang.NewRuntimeError(lang.ErrorCodeDuplicate, fmt.Sprintf("procedure '%s' already defined", name), lang.ErrProcedureExists)
			}
		}
		// If AllowRedefinition is true, or no collision found, overwrite.
		i.state.knownProcedures[name] = proc
	}

	for _, eventDecl := range program.Events {
		// Note: eventManager.register() handles local collision logic.
		// If needed, we might need to push the AllowRedefinition flag down to it,
		// but typically multiple handlers for the same event *type* are allowed.
		// Collision usually refers to named handlers (which Load checks).
		if err := i.eventManager.register(eventDecl, i); err != nil {
			return fmt.Errorf("failed to register event handler during append: %w", err)
		}
	}

	if program.Commands != nil {
		i.state.commands = append(i.state.commands, program.Commands...)
	}

	return nil
}

// AppendScript merges procedures and event handlers from a new program AST
// into the interpreter's existing state.
func (i *Interpreter) AppendScript(tree *interfaces.Tree) error {
	return i.appendScript(tree)
}
