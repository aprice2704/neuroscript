// NeuroScript Version: 0.7.0
// File version: 79
// Purpose: Refactored 'ask' logic into interpreter_ask.go to reduce file size.
// filename: pkg/interpreter/interpreter_exec_more.go
// nlines: 250
// risk_rating: HIGH

package interpreter

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func (i *Interpreter) executeCall(step ast.Step) (lang.Value, error) {
	if step.Call == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "call step is missing call expression", nil).WithPosition(step.GetPos())
	}
	return i.evaluate.Expression(step.Call)
}

func (i *Interpreter) executeEmit(step ast.Step) (lang.Value, error) {
	if len(step.Values) == 0 {
		return &lang.NilValue{}, nil
	}
	val, err := i.evaluate.Expression(step.Values[0])
	if err != nil {
		return nil, err
	}
	if i.customEmitFunc != nil {
		i.customEmitFunc(val)
	} else {
		i.logger.Debug("Emit statement executed", "value", val.String())
	}
	return val, nil
}

func (i *Interpreter) executeWhisper(step ast.Step) (lang.Value, error) {
	if step.WhisperStmt == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "whisper step is missing whisper statement details", nil).WithPosition(step.GetPos())
	}
	handleVal, err := i.evaluate.Expression(step.WhisperStmt.Handle)
	if err != nil {
		return nil, err
	}
	dataVal, err := i.evaluate.Expression(step.WhisperStmt.Value)
	if err != nil {
		return nil, err
	}
	if i.customWhisperFunc != nil {
		i.customWhisperFunc(handleVal, dataVal)
	} else {
		i.logger.Debug("Whisper statement executed without a custom handler", "handle", handleVal.String(), "data", dataVal.String())
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
