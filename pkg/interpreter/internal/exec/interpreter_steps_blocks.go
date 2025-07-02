// NeuroScript Version: 0.3.1
// File version: 7
// Purpose: Corrected all function signatures to return  Value. Refactored executeFor to be type-safe, iterating directly on Value types (ListValue, MapValue, StringValue) without reflection or unwrapping.
// filename: pkg/interpreter/internal/exec/interpreter_steps_blocks.go
// nlines: 200
// risk_rating: HIGH

package exec

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// executeIf handles the "if" step.
func (i *Interpreter) executeIf(step ast.Step, isInHandler bool, activeError *lang.RuntimeError) (result lang.Value, wasReturn bool, wasCleared bool, err error) {
	posStr := "<unknown_pos>"
	if step.Pos != nil {
		posStr = step.Pos.String()
	}
	i.Logger().Debug("[DEBUG-INTERP]   Executing IF", "pos", posStr)

	if step.Cond == nil {
		return nil, false, false, lang.NewRuntimeError(lang.ErrorCodeInternal, "IF step has nil Condition", nil).WithPosition(step.Pos)
	}

	condResult, evalErr := i.evaluate.Expression(step.Cond)
	if evalErr != nil {
		return nil, false, false, lang.WrapErrorWithPosition(evalErr, step.Cond.GetPos(), "evaluating IF condition")
	}

	if lang.IsTruthy(condResult) {
		i.Logger().Debug("[DEBUG-INTERP]     IF condition TRUE, executing THEN block", "pos", posStr)
		return i.executeBlock(step.Body, step.Pos, "IF_THEN", isInHandler, activeError)
	} else if step.Else != nil {
		i.Logger().Debug("[DEBUG-INTERP]     IF condition FALSE, executing ELSE block", "pos", posStr)
		return i.executeBlock(step.Else, step.Pos, "IF_ELSE", isInHandler, activeError)
	}

	i.Logger().Debug("[DEBUG-INTERP]     IF condition FALSE, no ELSE block", "pos", posStr)
	return lang.NilValue{}, false, false, nil
}

// executeWhile handles the "while" step.
func (i *Interpreter) executeWhile(step ast.Step, isInHandler bool, activeError *lang.RuntimeError) (result lang.Value, wasReturn bool, wasCleared bool, err error) {
	posStr := "<unknown_pos>"
	if step.Pos != nil {
		posStr = step.Pos.String()
	}
	i.Logger().Debug("[DEBUG-INTERP]   Executing WHILE", "pos", posStr)

	if step.Cond == nil {
		return nil, false, false, lang.NewRuntimeError(lang.ErrorCodeInternal, "WHILE step has nil Condition", nil).WithPosition(step.Pos)
	}

	iteration := 0
	maxIterations := i.maxLoopIterations
	result = lang.NilValue{}	// Initialize result to a valid Value

	for {
		iteration++
		if iteration > maxIterations {
			errMsg := fmt.Sprintf("WHILE loop at %s exceeded max iterations (%d)", posStr, maxIterations)
			return nil, false, wasCleared, lang.NewRuntimeError(lang.ErrorCodeInternal, errMsg, errors.New("max iterations exceeded")).WithPosition(step.Pos)
		}

		condResult, evalErr := i.evaluate.Expression(step.Cond)
		if evalErr != nil {
			return nil, false, wasCleared, lang.WrapErrorWithPosition(evalErr, step.Cond.GetPos(), "evaluating WHILE condition")
		}

		if !lang.IsTruthy(condResult) {
			i.Logger().Debug("[DEBUG-INTERP]     WHILE condition FALSE, exiting loop", "pos", posStr, "iterations", iteration-1)
			break	// Exit loop
		}

		i.Logger().Debug("[DEBUG-INTERP]     WHILE condition TRUE, executing block", "pos", posStr, "iteration", iteration)
		blockResult, blockReturned, blockCleared, blockErr := i.executeBlock(step.Body, step.Pos, "WHILE_BODY", isInHandler, activeError)

		if blockCleared {
			wasCleared = true
			activeError = nil
		}

		if blockErr != nil {
			if errors.Is(blockErr, lang.ErrBreak) {
				i.Logger().Debug("[DEBUG-INTERP]     BREAK encountered in WHILE loop body", "pos", posStr)
				return result, false, wasCleared, nil
			}
			if errors.Is(blockErr, lang.ErrContinue) {
				i.Logger().Debug("[DEBUG-INTERP]     CONTINUE encountered in WHILE loop body, proceeding to next iteration", "pos", posStr)
				continue
			}
			return nil, false, wasCleared, blockErr
		}
		if blockReturned {
			return blockResult, true, wasCleared, nil
		}
		result = blockResult
	}
	return result, false, wasCleared, nil
}

// executeFor handles the "for each" step.
func (i *Interpreter) executeFor(step ast.Step, isInHandler bool, activeError *lang.RuntimeError) (result lang.Value, wasReturn bool, wasCleared bool, err error) {
	posStr := "<unknown_pos>"
	if step.Pos != nil {
		posStr = step.Pos.String()
	}

	loopVar := step.LoopVarName
	collectionExpr := step.Collection
	result = lang.NilValue{}	// Initialize result to a valid Value

	i.Logger().Debug("[DEBUG-INTERP]   Executing FOR EACH", "loopVar", loopVar, "pos", posStr)

	if collectionExpr == nil {
		return nil, false, false, lang.NewRuntimeError(lang.ErrorCodeInternal, "FOR EACH step has nil Collection expression", nil).WithPosition(step.Pos)
	}
	if loopVar == "" {
		return nil, false, false, lang.NewRuntimeError(lang.ErrorCodeInternal, "FOR EACH step has empty LoopVarName", nil).WithPosition(step.Pos)
	}

	collectionVal, evalErr := i.evaluate.Expression(collectionExpr)
	if evalErr != nil {
		return nil, false, false, lang.WrapErrorWithPosition(evalErr, collectionExpr.GetPos(), fmt.Sprintf("evaluating collection for FOR EACH %s", loopVar))
	}

	iteration := 0
	maxIterations := i.maxLoopIterations

	var loopErr error
	var loopReturned bool
	var blockResult lang.Value

	processLoopBody := func() (bool, bool, lang.Value, error) {	// shouldBreak, shouldContinue, result, error
		res, ret, clr, bErr := i.executeBlock(step.Body, step.Pos, "FOR_BODY", isInHandler, activeError)
		if clr {
			wasCleared = true
			activeError = nil
		}
		if bErr != nil {
			if errors.Is(bErr, lang.ErrBreak) {
				return true, false, res, nil
			}
			if errors.Is(bErr, lang.ErrContinue) {
				return false, true, res, nil
			}
			return false, false, nil, bErr
		}
		if ret {
			return false, false, res, nil	// Propagate return up
		}
		return false, false, res, nil
	}

	handleIteration := func(item lang.Value) (bool, bool, error) {
		iteration++
		if iteration > maxIterations {
			errMsg := fmt.Sprintf("FOR EACH loop for %s at %s exceeded max iterations (%d)", loopVar, posStr, maxIterations)
			return false, false, lang.NewRuntimeError(lang.ErrorCodeInternal, errMsg, errors.New("max iterations exceeded")).WithPosition(step.Pos)
		}
		if setErr := i.SetVariable(loopVar, item); setErr != nil {
			errMsg := fmt.Sprintf("setting loop variable '%s' in FOR EACH", loopVar)
			return false, false, lang.NewRuntimeError(lang.ErrorCodeInternal, errMsg, setErr).WithPosition(step.Pos)
		}

		shouldBreak, shouldContinue, res, err := processLoopBody()
		result = res	// Store last block result
		if err != nil {
			return false, false, err
		}
		if shouldBreak {
			return true, false, nil
		}
		if shouldContinue {
			return false, true, nil
		}
		if res != nil && loopReturned {
			return false, false, nil	// Should be handled by caller
		}
		return false, false, nil
	}

	switch c := collectionVal.(type) {
	case ListValue:
		for _, item := range c.Value {
			shouldBreak, shouldContinue, err := handleIteration(item)
			if err != nil {
				return nil, false, wasCleared, err
			}
			if shouldBreak {
				goto endLoop
			}
			if shouldContinue {
				continue
			}
		}
	case MapValue:
		for _, item := range c.Value {	// Note: iterating over map values, keys are ignored for now
			shouldBreak, shouldContinue, err := handleIteration(item)
			if err != nil {
				return nil, false, wasCleared, err
			}
			if shouldBreak {
				goto endLoop
			}
			if shouldContinue {
				continue
			}
		}
	case StringValue:
		for _, charRune := range c.Value {
			item := lang.StringValue{Value: string(charRune)}
			shouldBreak, shouldContinue, err := handleIteration(item)
			if err != nil {
				return nil, false, wasCleared, err
			}
			if shouldBreak {
				goto endLoop
			}
			if shouldContinue {
				continue
			}
		}
	default:
		errMsg := fmt.Sprintf("cannot iterate over type %s for FOR EACH %s", lang.TypeOf(collectionVal), loopVar)
		return nil, false, wasCleared, lang.NewRuntimeError(lang.ErrorCodeType, errMsg, nil).WithPosition(collectionExpr.GetPos())
	}

endLoop:
	if loopErr != nil {
		return nil, false, wasCleared, loopErr
	}
	if loopReturned {
		return blockResult, true, wasCleared, nil
	}

	i.Logger().Debug("[DEBUG-INTERP]   FOR EACH loop finished normally.", "loopVar", loopVar, "pos", posStr, "iterations", iteration)
	return result, false, wasCleared, nil
}