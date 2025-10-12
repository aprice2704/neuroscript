// NeuroScript Version: 0.8.0
// File version: 80
// Purpose: Refactored to use the 'eval' package and HostContext for core execution logic.
// filename: pkg/interpreter/exec_more.go
// nlines: 80
// risk_rating: HIGH

package interpreter

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/eval"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func (i *Interpreter) executeCall(step ast.Step) (lang.Value, error) {
	if step.Call == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "call step is missing call expression", nil).WithPosition(step.GetPos())
	}
	// The entire call logic is now handled by the evaluator.
	return eval.Expression(i, step.Call)
}

func (i *Interpreter) executeWhisper(step ast.Step) (lang.Value, error) {
	if step.WhisperStmt == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "whisper step is missing whisper statement details", nil).WithPosition(step.GetPos())
	}

	handleVal, err := eval.Expression(i, step.WhisperStmt.Handle)
	if err != nil {
		return nil, err
	}
	dataVal, err := eval.Expression(i, step.WhisperStmt.Value)
	if err != nil {
		return nil, err
	}

	if i.hostContext != nil && i.hostContext.WhisperFunc != nil {
		i.hostContext.WhisperFunc(handleVal, dataVal)
	} else {
		// This default behavior should likely be removed in favor of a mandatory handler.
		i.Logger().Debug("Whisper statement executed without a custom handler", "handle", handleVal.String(), "data", dataVal.String())
	}
	return dataVal, nil
}

func (i *Interpreter) executeBlock(blockValue interface{}, parentPos *types.Position, blockType string, isInHandler bool, activeError *lang.RuntimeError, depth int) (result lang.Value, wasReturn bool, wasCleared bool, err error) {
	steps, ok := blockValue.([]ast.Step)
	if !ok {
		if blockValue == nil {
			return &lang.NilValue{}, false, false, nil
		}
		errMsg := fmt.Sprintf("internal error: invalid block format for %s - expected []Step, got %T", blockType, blockValue)
		return nil, false, false, lang.NewRuntimeError(lang.ErrorCodeInternal, errMsg, lang.ErrInternal).WithPosition(parentPos)
	}
	return i.recExecuteSteps(steps, isInHandler, activeError, depth)
}

func shouldUpdateLastResult(stepTypeLower string) bool {
	switch stepTypeLower {
	case "set", "assign", "emit", "ask", "promptuser", "call", "expression_statement", "must", "mustbe", "whisper":
		return true
	default:
		return false
	}
}

func ensureRuntimeError(err error, pos *types.Position, context string) *lang.RuntimeError {
	if rtErr, ok := err.(*lang.RuntimeError); ok {
		if rtErr.Position == nil && pos != nil {
			return rtErr.WithPosition(pos)
		}
		return rtErr
	}
	return lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("internal error during %s", context), err).WithPosition(pos)
}
