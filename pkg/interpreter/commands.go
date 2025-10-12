// NeuroScript Version: 0.8.0
// File version: 7
// Purpose: Un-exported ExecuteCommands as it is an internal implementation detail.
// filename: pkg/interpreter/commands.go
// nlines: 40
// risk_rating: MEDIUM
package interpreter

import (
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// executeCommands runs all top-level commands loaded into the interpreter from a script.
func (i *Interpreter) executeCommands() (lang.Value, error) {
	var finalResult lang.Value = &lang.NilValue{}
	var err error

	for _, cmdNode := range i.state.commands {
		if cmdNode == nil || len(cmdNode.Body) == 0 {
			continue
		}

		var wasReturn, wasCleared bool
		finalResult, wasReturn, wasCleared, err = i.executeSteps(cmdNode.Body, false, nil)

		if err != nil {
			pos := cmdNode.BaseNode.StartPos
			if pos == nil && len(cmdNode.Body) > 0 {
				pos = cmdNode.Body[0].GetPos()
			}
			return nil, lang.WrapErrorWithPosition(err, pos, "executing top-level command block")
		}

		if wasReturn {
			pos := cmdNode.BaseNode.StartPos
			return nil, lang.NewRuntimeError(lang.ErrorCodeSyntax, "'return' statement not allowed in top-level command blocks", nil).WithPosition(pos)
		}

		if wasCleared {
			// This state is handled by the executeSteps loop.
		}
	}

	return finalResult, nil
}
