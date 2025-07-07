// NeuroScript Version: 0.5.2
// File version: 6
// Purpose: Corrected implementation to iterate through the steps in a CommandNode's Body, fixing compiler errors.
// filename: pkg/interpreter/interpreter_commands.go
// nlines: 40
// risk_rating: MEDIUM
package interpreter

import (
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// ExecuteCommands runs all top-level commands loaded into the interpreter from a script.
// It returns the result of the very last step in the last command block.
func (i *Interpreter) ExecuteCommands() (lang.Value, error) {
	var finalResult lang.Value = &lang.NilValue{}
	var err error

	// A program can have multiple top-level 'command' blocks.
	for _, cmdNode := range i.state.commands {
		if cmdNode == nil || len(cmdNode.Body) == 0 {
			continue
		}

		// Execute the steps contained within the command block's body.
		// The `executeSteps` function will handle each step (set, call, etc.) appropriately.
		var wasReturn, wasCleared bool
		finalResult, wasReturn, wasCleared, err = i.executeSteps(cmdNode.Body, false, nil)

		if err != nil {
			// Find the position of the command block for better error reporting.
			pos := cmdNode.Pos
			if pos == nil && len(cmdNode.Body) > 0 {
				pos = cmdNode.Body[0].GetPos()
			}
			return nil, lang.WrapErrorWithPosition(err, pos, "executing top-level command block")
		}

		// Top-level commands should not have 'return' statements.
		if wasReturn {
			pos := cmdNode.Pos // Or find a more specific position if possible
			return nil, lang.NewRuntimeError(lang.ErrorCodeSyntax, "'return' statement not allowed in top-level command blocks", nil).WithPosition(pos)
		}

		// Handle clear_error if needed, though its use in commands is rare.
		if wasCleared {
			// Reset any active error state if your interpreter has one.
		}
	}

	return finalResult, nil
}
