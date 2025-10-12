// NeuroScript Version: 0.8.0
// File version: 47
// Purpose: Re-plumbed all expression evaluation to use the external 'eval' package.
// filename: pkg/interpreter/steps_blocks.go
// nlines: 200
// risk_rating: HIGH

package interpreter

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/eval"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// executeIf handles the "if" step.
func (i *Interpreter) executeIf(step ast.Step, isInHandler bool, activeError *lang.RuntimeError) (result lang.Value, wasReturn bool, wasCleared bool, err error) {
	condResult, evalErr := eval.Expression(i, step.Cond)
	if evalErr != nil {
		return nil, false, false, lang.WrapErrorWithPosition(evalErr, step.Cond.GetPos(), "evaluating IF condition")
	}

	if lang.IsTruthy(condResult) {
		return i.executeBlock(step.Body, step.GetPos(), "IF_THEN", isInHandler, activeError, 0)
	} else if step.ElseBody != nil {
		return i.executeBlock(step.ElseBody, step.GetPos(), "IF_ELSE", isInHandler, activeError, 0)
	}

	return &lang.NilValue{}, false, false, nil
}

// executeWhile handles the "while" step.
func (i *Interpreter) executeWhile(step ast.Step, isInHandler bool, activeError *lang.RuntimeError) (result lang.Value, wasReturn bool, wasCleared bool, err error) {
	if step.Cond == nil {
		return nil, false, false, lang.NewRuntimeError(lang.ErrorCodeInternal, "WHILE step has nil Condition", nil).WithPosition(step.GetPos())
	}

	result = &lang.NilValue{}

	for iteration := 0; ; iteration++ {
		if iteration >= i.maxLoopIterations {
			return nil, false, false, lang.NewRuntimeError(lang.ErrorCodeResourceExhaustion, fmt.Sprintf("exceeded max iterations (%d)", i.maxLoopIterations), lang.ErrMaxIterationsExceeded).WithPosition(step.GetPos())
		}

		condResult, evalErr := eval.Expression(i, step.Cond)
		if evalErr != nil {
			return nil, false, wasCleared, lang.WrapErrorWithPosition(evalErr, step.Cond.GetPos(), "evaluating WHILE condition")
		}

		if !lang.IsTruthy(condResult) {
			break
		}

		blockResult, blockReturned, blockCleared, blockErr := i.executeBlock(step.Body, step.GetPos(), "WHILE_BODY", isInHandler, activeError, 1)

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
		errMsg := fmt.Sprintf("FOR EACH step has nil Collection expression. This indicates a severe bug in the AST Builder. Step Details: %+v", step)
		return nil, false, false, lang.NewRuntimeError(lang.ErrorCodeInternal, errMsg, nil).WithPosition(step.GetPos())
	}
	if step.LoopVarName == "" {
		return nil, false, false, lang.NewRuntimeError(lang.ErrorCodeInternal, "FOR EACH step has empty LoopVarName", nil).WithPosition(step.GetPos())
	}

	collectionVal, evalErr := eval.Expression(i, step.Collection)
	if evalErr != nil {
		return nil, false, wasCleared, lang.WrapErrorWithPosition(evalErr, step.Collection.GetPos(), fmt.Sprintf("evaluating collection for FOR EACH %s", step.LoopVarName))
	}

	var itemsToIterate []lang.Value
	switch c := collectionVal.(type) {
	case *lang.ListValue:
		itemsToIterate = c.Value
	case lang.ListValue:
		itemsToIterate = c.Value
	case *lang.MapValue:
		itemsToIterate = make([]lang.Value, 0, len(c.Value))
		for _, v := range c.Value {
			itemsToIterate = append(itemsToIterate, v)
		}
	case lang.StringValue:
		itemsToIterate = make([]lang.Value, 0, len(c.Value))
		for _, charRune := range c.Value {
			itemsToIterate = append(itemsToIterate, lang.StringValue{Value: string(charRune)})
		}
	case *lang.NilValue:
		itemsToIterate = []lang.Value{}
	default:
		errMsg := fmt.Sprintf("cannot iterate over type %s for FOR EACH %s", lang.TypeOf(collectionVal), step.LoopVarName)
		return nil, false, wasCleared, lang.NewRuntimeError(lang.ErrorCodeType, errMsg, nil).WithPosition(step.Collection.GetPos())
	}

	result = &lang.NilValue{}

	for iteration, item := range itemsToIterate {
		if iteration >= i.maxLoopIterations {
			return nil, false, false, lang.NewRuntimeError(lang.ErrorCodeResourceExhaustion, fmt.Sprintf("exceeded max iterations (%d)", i.maxLoopIterations), lang.ErrMaxIterationsExceeded).WithPosition(step.GetPos())
		}

		if setErr := i.SetVariable(step.LoopVarName, item); setErr != nil {
			errMsg := fmt.Sprintf("setting loop variable '%s' in FOR EACH", step.LoopVarName)
			return nil, false, wasCleared, lang.NewRuntimeError(lang.ErrorCodeInternal, errMsg, setErr).WithPosition(step.GetPos())
		}

		blockResult, blockReturned, blockCleared, blockErr := i.executeBlock(step.Body, step.GetPos(), "FOR_BODY", isInHandler, activeError, 1)

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
