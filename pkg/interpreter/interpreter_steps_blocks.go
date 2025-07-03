// NeuroScript Version: 0.5.2
// File version: 11
// Purpose: Corrected nil comparison logic for Position struct to check a field's zero value instead.
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

// executeIf handles the "if" step.
func (i *Interpreter) executeIf(step ast.Step, isInHandler bool, activeError *lang.RuntimeError) (result lang.Value, wasReturn bool, wasCleared bool, err error) {
	posStr := "<unknown_pos>"
	// FIX: Compare a field on the Position struct to its zero value.
	if step.Position.Line != 0 {
		posStr = step.Position.String()
	}
	i.Logger().Debug("[DEBUG-INTERP]   Executing IF", "pos", posStr)

	if step.Cond == nil {
		return nil, false, false, lang.NewRuntimeError(lang.ErrorCodeInternal, "IF step has nil Condition", nil).WithPosition(&step.Position)
	}

	condResult, evalErr := i.evaluate.Expression(step.Cond)
	if evalErr != nil {
		return nil, false, false, lang.WrapErrorWithPosition(evalErr, step.Cond.GetPos(), "evaluating IF condition")
	}

	if lang.IsTruthy(condResult) {
		i.Logger().Debug("[DEBUG-INTERP]     IF condition TRUE, executing THEN block", "pos", posStr)
		return i.executeBlock(step.Body, &step.Position, "IF_THEN", isInHandler, activeError)
	} else if step.ElseBody != nil {
		i.Logger().Debug("[DEBUG-INTERP]     IF condition FALSE, executing ELSE block", "pos", posStr)
		return i.executeBlock(step.ElseBody, &step.Position, "IF_ELSE", isInHandler, activeError)
	}

	i.Logger().Debug("[DEBUG-INTERP]     IF condition FALSE, no ELSE block", "pos", posStr)
	return &lang.NilValue{}, false, false, nil
}

// executeWhile handles the "while" step.
func (i *Interpreter) executeWhile(step ast.Step, isInHandler bool, activeError *lang.RuntimeError) (result lang.Value, wasReturn bool, wasCleared bool, err error) {
	posStr := "<unknown_pos>"
	// FIX: Compare a field on the Position struct to its zero value.
	if step.Position.Line != 0 {
		posStr = step.Position.String()
	}
	i.Logger().Debug("[DEBUG-INTERP]   Executing WHILE", "pos", posStr)

	if step.Cond == nil {
		return nil, false, false, lang.NewRuntimeError(lang.ErrorCodeInternal, "WHILE step has nil Condition", nil).WithPosition(&step.Position)
	}

	iteration := 0
	maxIterations := i.maxLoopIterations
	result = &lang.NilValue{}	// Initialize result to a valid Value

	for {
		iteration++
		if iteration > maxIterations {
			errMsg := fmt.Sprintf("WHILE loop at %s exceeded max iterations (%d)", posStr, maxIterations)
			return nil, false, wasCleared, lang.NewRuntimeError(lang.ErrorCodeInternal, errMsg, errors.New("max iterations exceeded")).WithPosition(&step.Position)
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
		blockResult, blockReturned, blockCleared, blockErr := i.executeBlock(step.Body, &step.Position, "WHILE_BODY", isInHandler, activeError)

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
	// FIX: Compare a field on the Position struct to its zero value.
	if step.Position.Line != 0 {
		posStr = step.Position.String()
	}

	loopVar := step.LoopVarName
	collectionExpr := step.Collection
	result = &lang.NilValue{}	// Initialize result to a valid Value

	i.Logger().Debug("[DEBUG-INTERP]   Executing FOR EACH", "loopVar", loopVar, "pos", posStr)

	if collectionExpr == nil {
		return nil, false, false, lang.NewRuntimeError(lang.ErrorCodeInternal, "FOR EACH step has nil Collection expression", nil).WithPosition(&step.Position)
	}
	if loopVar == "" {
		return nil, false, false, lang.NewRuntimeError(lang.ErrorCodeInternal, "FOR EACH step has empty LoopVarName", nil).WithPosition(&step.Position)
	}

	collectionVal, evalErr := i.evaluate.Expression(collectionExpr)
	if evalErr != nil {
		return nil, false, false, lang.WrapErrorWithPosition(evalErr, collectionExpr.GetPos(), fmt.Sprintf("evaluating collection for FOR EACH %s", loopVar))
	}

	iteration := 0
	maxIterations := i.maxLoopIterations

	processLoopBody := func() (bool, bool, lang.Value, error) {	// shouldBreak, shouldContinue, result, error
		res, ret, clr, bErr := i.executeBlock(step.Body, &step.Position, "FOR_BODY", isInHandler, activeError)
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
			return false, false, lang.NewRuntimeError(lang.ErrorCodeInternal, errMsg, errors.New("max iterations exceeded")).WithPosition(&step.Position)
		}
		if setErr := i.SetVariable(loopVar, item); setErr != nil {
			errMsg := fmt.Sprintf("setting loop variable '%s' in FOR EACH", loopVar)
			return false, false, lang.NewRuntimeError(lang.ErrorCodeInternal, errMsg, setErr).WithPosition(&step.Position)
		}

		shouldBreak, shouldContinue, res, err := processLoopBody()
		if res != nil {
			result = res	// Store last block result
		}
		if err != nil {
			return false, false, err
		}
		if wasReturn {
			return false, false, nil
		}
		if shouldBreak {
			return true, false, nil
		}
		if shouldContinue {
			return false, true, nil
		}
		return false, false, nil
	}

	switch c := collectionVal.(type) {
	case lang.ListValue:
		for _, item := range c.Value {
			shouldBreak, shouldContinue, err := handleIteration(item)
			if err != nil {
				return nil, false, wasCleared, err
			}
			if wasReturn {
				return result, true, wasCleared, nil
			}
			if shouldBreak {
				goto endLoop
			}
			if shouldContinue {
				continue
			}
		}
	case lang.MapValue:
		for _, item := range c.Value {	// Note: iterating over map values
			shouldBreak, shouldContinue, err := handleIteration(item)
			if err != nil {
				return nil, false, wasCleared, err
			}
			if wasReturn {
				return result, true, wasCleared, nil
			}
			if shouldBreak {
				goto endLoop
			}
			if shouldContinue {
				continue
			}
		}
	case lang.StringValue:
		for _, charRune := range c.Value {
			item := lang.StringValue{Value: string(charRune)}
			shouldBreak, shouldContinue, err := handleIteration(item)
			if err != nil {
				return nil, false, wasCleared, err
			}
			if wasReturn {
				return result, true, wasCleared, nil
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
	i.Logger().Debug("[DEBUG-INTERP]   FOR EACH loop finished normally.", "loopVar", loopVar, "pos", posStr, "iterations", iteration)
	return result, false, wasCleared, nil
}