// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 9
// :: description: Updated command block execution to push context to stackFrames, ensuring correct stacking in loops.
// :: latestChange: Pushed contextName to stackFrames using baseStack.
// :: filename: pkg/interpreter/commands.go
// :: serialization: go

package interpreter

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// executeCommands runs all top-level commands loaded into the interpreter from a script.
func (i *Interpreter) executeCommands() (lang.Value, error) {
	var finalResult lang.Value = &lang.NilValue{}
	var err error

	// Capture the base stack before iterating commands.
	// This ensures that if we have multiple command blocks, the stack doesn't grow infinitely with siblings.
	baseStack := i.state.stackFrames

	for _, cmdNode := range i.state.commands {
		if cmdNode == nil || len(cmdNode.Body) == 0 {
			continue
		}

		// Set meaningful context for stack traces
		contextName := "Command Block"
		if cmdNode.BaseNode.StartPos != nil {
			contextName = fmt.Sprintf("Command Block (at %s)", cmdNode.BaseNode.StartPos.String())
		}
		i.state.currentProcName = contextName
		// Push context to stackFrames so called procedures inherit it in their trace.
		// We append to baseStack to treat each command block as a sibling in the call stack.
		i.state.stackFrames = append(baseStack, contextName)

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
