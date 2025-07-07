// NeuroScript Version: 0.5.2
// File version: 37.0.0
// Purpose: Refactored the executeFor loop to simplify its logic and correctly propagate all state, mirroring executeWhile.
// filename: pkg/interpreter/interpreter_steps_blocks.go
// nlines: 200
// risk_rating: HIGH

package interpreter

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// Block Execution Contract:
// The functions in this file (executeIf, executeWhile, executeFor) are responsible for managing
// control flow constructs. They operate under a specific contract with the main execution
// loop (`recExecuteSteps` in `interpreter_exec.go`):
//
// 1. Block Execution: Each construct is given a block of steps (e.g., the body of a `while`
//    loop) to execute. It calls `i.executeBlock()`, which in turn calls `recExecuteSteps` for that block.
// 2. Standard Returns: If the block executes without issue, the construct's job is to return the
//    result of the last statement, and the execution continues normally.
// 3. Error Handling: If `recExecuteSteps` returns a standard runtime error, the construct does not
//    handle it. It immediately propagates the error up the call stack.
// 4. Control Flow Signals: `break` and `continue` are not standard errors. They are special signals
//    implemented as wrapped sentinel errors (`ErrBreak`, `ErrContinue`).
//    - The main execution loop (`recExecuteSteps`) has a special case: if it catches an error
//      that is a `break` or `continue` signal, it *immediately stops its own execution*
//      and returns the signal error up to its caller.
//    - The loop constructs in this file (`executeWhile`, `executeFor`) are the designated callers
//      that *must* catch these specific signals.
//    - Upon catching an `ErrBreak`, the loop must terminate its own execution (e.g., via goto
//      or a labeled break) and return `nil` for the error, effectively "consuming" the signal.
//    - Upon catching an `ErrContinue`, the loop must skip the rest of the current iteration,
//      start the next one, and also return `nil`, consuming the signal.
//
// This contract ensures that control signals are handled *only* by the nearest enclosing loop
// and are not accidentally caught by general-purpose `on_error` blocks.

// executeIf handles the "if" step.
func (i *Interpreter) executeIf(step ast.Step, isInHandler bool, activeError *lang.RuntimeError) (result lang.Value, wasReturn bool, wasCleared bool, err error) {
	condResult, evalErr := i.evaluate.Expression(step.Cond)
	if evalErr != nil {
		return nil, false, false, lang.WrapErrorWithPosition(evalErr, step.Cond.GetPos(), "evaluating IF condition")
	}

	if lang.IsTruthy(condResult) {
		return i.executeBlock(step.Body, &step.Position, "IF_THEN", isInHandler, activeError, 0)
	} else if step.ElseBody != nil {
		return i.executeBlock(step.ElseBody, &step.Position, "IF_ELSE", isInHandler, activeError, 0)
	}

	return &lang.NilValue{}, false, false, nil
}

// executeWhile handles the "while" step.
func (i *Interpreter) executeWhile(step ast.Step, isInHandler bool, activeError *lang.RuntimeError) (result lang.Value, wasReturn bool, wasCleared bool, err error) {
	if step.Cond == nil {
		return nil, false, false, lang.NewRuntimeError(lang.ErrorCodeInternal, "WHILE step has nil Condition", nil).WithPosition(&step.Position)
	}

	result = &lang.NilValue{}

	for iteration := 0; iteration < i.maxLoopIterations; iteration++ {
		condResult, evalErr := i.evaluate.Expression(step.Cond)
		if evalErr != nil {
			return nil, false, wasCleared, lang.WrapErrorWithPosition(evalErr, step.Cond.GetPos(), "evaluating WHILE condition")
		}

		if !lang.IsTruthy(condResult) {
			break
		}

		blockResult, blockReturned, blockCleared, blockErr := i.executeBlock(step.Body, &step.Position, "WHILE_BODY", isInHandler, activeError, 1)

		if blockCleared {
			wasCleared = true
			activeError = nil
		}

		if blockErr != nil {
			var rtErr *lang.RuntimeError
			if errors.As(blockErr, &rtErr) {
				if errors.Is(rtErr.Unwrap(), lang.ErrBreak) {
					goto endWhileLoop
				}
				if errors.Is(rtErr.Unwrap(), lang.ErrContinue) {
					continue
				}
			}
			return nil, false, wasCleared, blockErr
		}
		if blockReturned {
			return blockResult, true, wasCleared, nil
		}
		result = blockResult
	}

endWhileLoop:
	return result, false, wasCleared, nil
}

// executeFor handles the "for each" step.
func (i *Interpreter) executeFor(step ast.Step, isInHandler bool, activeError *lang.RuntimeError) (result lang.Value, wasReturn bool, wasCleared bool, err error) {
	if step.Collection == nil {
		return nil, false, false, lang.NewRuntimeError(lang.ErrorCodeInternal, "FOR EACH step has nil Collection expression", nil).WithPosition(&step.Position)
	}
	if step.LoopVarName == "" {
		return nil, false, false, lang.NewRuntimeError(lang.ErrorCodeInternal, "FOR EACH step has empty LoopVarName", nil).WithPosition(&step.Position)
	}

	collectionVal, evalErr := i.evaluate.Expression(step.Collection)
	if evalErr != nil {
		return nil, false, wasCleared, lang.WrapErrorWithPosition(evalErr, step.Collection.GetPos(), fmt.Sprintf("evaluating collection for FOR EACH %s", step.LoopVarName))
	}

	var itemsToIterate []lang.Value
	switch c := collectionVal.(type) {
	case lang.ListValue:
		itemsToIterate = c.Value
	case lang.MapValue:
		// Note: Iteration order over maps is not guaranteed.
		itemsToIterate = make([]lang.Value, 0, len(c.Value))
		for _, v := range c.Value {
			itemsToIterate = append(itemsToIterate, v)
		}
	case lang.StringValue:
		itemsToIterate = make([]lang.Value, 0, len(c.Value))
		for _, charRune := range c.Value {
			itemsToIterate = append(itemsToIterate, lang.StringValue{Value: string(charRune)})
		}
	default:
		errMsg := fmt.Sprintf("cannot iterate over type %s for FOR EACH %s", lang.TypeOf(collectionVal), step.LoopVarName)
		return nil, false, wasCleared, lang.NewRuntimeError(lang.ErrorCodeType, errMsg, nil).WithPosition(step.Collection.GetPos())
	}

	result = &lang.NilValue{}

	for _, item := range itemsToIterate {
		if setErr := i.SetVariable(step.LoopVarName, item); setErr != nil {
			errMsg := fmt.Sprintf("setting loop variable '%s' in FOR EACH", step.LoopVarName)
			return nil, false, wasCleared, lang.NewRuntimeError(lang.ErrorCodeInternal, errMsg, setErr).WithPosition(&step.Position)
		}

		blockResult, blockReturned, blockCleared, blockErr := i.executeBlock(step.Body, &step.Position, "FOR_BODY", isInHandler, activeError, 1)

		if blockCleared {
			wasCleared = true
			activeError = nil
		}

		if blockErr != nil {
			var rtErr *lang.RuntimeError
			if errors.As(blockErr, &rtErr) {
				if errors.Is(rtErr.Unwrap(), lang.ErrBreak) {
					goto endForLoop
				}
				if errors.Is(rtErr.Unwrap(), lang.ErrContinue) {
					continue
				}
			}
			return nil, false, wasCleared, blockErr
		}

		if blockReturned {
			return blockResult, true, wasCleared, nil
		}
		result = blockResult
	}

endForLoop:
	return result, false, wasCleared, nil
}
